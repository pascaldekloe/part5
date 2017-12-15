package session

import (
	"errors"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

// TimeoutResolution is seconds according to companion standard 104,
// subclause 6.9, caption "Definition of time outs". However, thenths
// of a second make this system much more responsive i.c.w. S-frames.
const timeoutResolution = 100 * time.Millisecond

// RetryTicker is for temporary connectivity error recovery.
// Low intervals may cause too many errors.
// High intervals may cause unnecessary hold up.
var retryTicker = time.NewTicker(200 * time.Millisecond)

var (
	errSeqNo           = errors.New("part5: fatal incomming sequence number disruption")
	errAckNo           = errors.New("part5: fatal incomming acknowledge either earlier than previous or later than send")
	errAckExpire       = errors.New("part5: fatal transmission timeout t₁")
	errBringUpExpire   = errors.New("part5: fatal STARTDT acknowledge timeout t₁")
	errBringDownExpire = errors.New("part5: fatal STOPDT acknowledge timeout t₁")
	errKeepAliveExpire = errors.New("part5: fatal TESTFR acknowledge timeout t₁")
	errIllegalFunc     = errors.New("part5: illegal function ignored")
)

type tcp struct {
	conf TCPConf
	conn net.Conn

	// Transport counterparts
	in     chan<- []byte
	class1 <-chan *Outbound
	class2 <-chan *Outbound
	err    chan<- error
	// Station counterparts
	level  chan<- Level
	launch <-chan Level

	recv chan apdu // for recvLoop
	send chan apdu // for sendLoop
	// closed when send is no longer read
	sendQuit chan struct{}

	// see subclause 5.1 — Protection against loss and duplication of messages
	seqNoOut uint // sequence number of next outbound I-frame
	seqNoIn  uint // sequence number of next inbound I-frame
	ackNoOut uint // outbound sequence number yet to be confirmed
	ackNoIn  uint // inbound sequence number yet to be confirmed

	// maps send I-frames to their respective sequence number
	pending [1 << 15]struct {
		send time.Time
		done chan<- error // outbound callback
	}

	idleSince time.Time
}

// TCP returns a session with status Down.
func TCP(conf *TCPConf, conn net.Conn) *Station {
	conf.check()

	class1Chan := make(chan *Outbound)
	class2Chan := make(chan *Outbound)
	inChan := make(chan []byte)
	errChan := make(chan error, 8)
	launchChan := make(chan Level)
	levelChan := make(chan Level)

	t := tcp{
		conf:   *conf,
		conn:   conn,
		level:  levelChan,
		launch: launchChan,

		in:     inChan,
		class1: class1Chan,
		class2: class2Chan,
		err:    errChan,

		recv:     make(chan apdu, conf.RecvUnackMax),
		send:     make(chan apdu, conf.SendUnackMax), // may not block!
		sendQuit: make(chan struct{}),

		idleSince: time.Now(),
	}

	go t.recvLoop()
	go t.sendLoop()
	go t.run()

	// exit strategy:
	// (1) recvLoop closes tcp.recv feed for run on return
	// (2) run closes tcp.send feed for sendLoop on return
	// (3) sendLoop closes tcp.conn for (1) on return
	// because run (2) may quit for various reasons it flushes
	// tcp.recv to prevent blocked routines
	// run also controls the levelChan and closes errChan and inChan

	return &Station{
		Transport: Transport{inChan, class1Chan, class2Chan, errChan},
		Addr:      conn.RemoteAddr(),
		Level:     levelChan,
		Launch:    launchChan,
	}
}

// RecvLoop feeds t.recv.
func (t *tcp) recvLoop() {
	defer close(t.recv)

	var datagram apdu // reusable instance
	for {
		byteCount, err := datagram.Unmarshal(t.conn, 0)
		for err != nil {
			// See: https://github.com/golang/go/issues/4373
			if err != io.EOF && err != io.ErrClosedPipe || strings.Contains(err.Error(), "use of closed network connection") {
				t.err <- err
			}

			if e, ok := err.(net.Error); !ok || !e.Temporary() {
				return
			}
			// temporary error may be recoverable

			<-retryTicker.C

			byteCount, err = datagram.Unmarshal(t.conn, byteCount)
		}
		if Trace {
			log.Printf("%s@%s: received %s", t.conn.RemoteAddr(), t.conn.LocalAddr(), datagram.String())
		}

		t.recv <- datagram // copy
	}
}

// SendLoop drains t.send.
func (t *tcp) sendLoop() {
	defer close(t.sendQuit)

	for datagram := range t.send {
		byteCount, err := datagram.Marshal(t.conn, 0)
		for err != nil {
			// See: https://github.com/golang/go/issues/4373
			if err != io.EOF && err != io.ErrClosedPipe || strings.Contains(err.Error(), "use of closed network connection") {
				t.err <- err
			}

			if e, ok := err.(net.Error); !ok || !e.Temporary() {
				return
			}
			// temporary error may be recoverable

			<-retryTicker.C

			byteCount, err = datagram.Marshal(t.conn, byteCount)
		}
		if Trace {
			log.Printf("%s@%s: send %s", t.conn.RemoteAddr(), t.conn.LocalAddr(), datagram.String())
		}
	}
}

// Run is the big fat state machine.
func (t *tcp) run() {
	// connected and no "data transfer" yet
	level := Down
	t.level <- level

	checkTicker := time.NewTicker(timeoutResolution)

	defer func() {
		checkTicker.Stop()

		// try gracefull shutdown
		if level > Down {
			select {
			case t.send <- newFunc(bringDown):
				break
			default:
				break // best effort
			}
			t.level <- Down
		}

		if t.ackNoIn != t.seqNoIn {
			select {
			case t.send <- newAck(t.seqNoIn):
				t.ackNoIn = t.seqNoIn
			default:
				break // best effort
			}
		}

		close(t.send)
		time.Sleep(timeoutResolution)
		t.conn.Close()

		// await receive loop
		for datagram := range t.recv {
			if f := datagram.Format(); f == iFrame || f == sFrame {
				t.updateAckNoOut(datagram.RecvSeqNo())
			}
			// discard
		}

		// report to API
		close(t.level) // sends Exit [0] level
		for i := t.ackNoOut; i != t.seqNoOut; i++ {
			t.pending[i].done <- ErrConnLost
		}
		close(t.in)
		close(t.err)
		go func() {
			for o := range t.class1 {
				o.err <- ErrNoConn
			}
		}()
		go func() {
			for o := range t.class2 {
				o.err <- ErrNoConn
			}
		}()
	}()

	// transmission timestamps for timeout calculation
	var (
		willNotTimeout = time.Now().Add(time.Hour * 24 * 365 * 100)
		unackRecvd     = willNotTimeout
		bringUpSend    = willNotTimeout
		bringDownSend  = willNotTimeout
		keepAliveSend  = willNotTimeout
	)

	for {
		// nil channel blocks (for SendUnackMax and Down case)
		var class1, class2 <-chan *Outbound

		if level >= Up && seqNoCount(t.ackNoOut, t.seqNoOut) <= t.conf.SendUnackMax {
			// may send; unblock
			class1, class2 = t.class1, t.class2

			// won't block because cap(t.send) = SendUnackMax
			select {
			case o, ok := <-class1:
				if !ok {
					return
				}
				// favour class Ⅰ
				t.submit(o)
				continue

			default:
				// try class Ⅱ
				select {
				case o, ok := <-class2:
					if !ok {
						return
					}
					t.submit(o)
					continue

				default:
					break // nothing available right now
				}
			}
		}

		select {
		case o, ok := <-class1:
			if !ok {
				return
			}
			t.submit(o)

		case o, ok := <-class2:
			if !ok {
				return
			}
			t.submit(o)

		case now := <-checkTicker.C:
			// check all timeouts

			timeout := t.conf.SendUnackTimeout
			if now.Sub(bringUpSend) >= timeout {
				t.err <- errBringUpExpire
				return
			}
			if now.Sub(bringDownSend) >= timeout {
				t.err <- errBringDownExpire
				return
			}
			if now.Sub(keepAliveSend) >= timeout {
				t.err <- errKeepAliveExpire
				return
			}

			// check oldest unacknowledged outbound
			if t.ackNoOut != t.seqNoOut && now.Sub(t.pending[t.ackNoOut].send) >= timeout {
				t.pending[t.ackNoOut].done <- errAckExpire
				t.ackNoOut++
				return
			}

			// check oldest unacknowledged inbound
			if t.ackNoIn != t.seqNoIn && (now.Sub(unackRecvd) >= t.conf.RecvUnackTimeout || now.Sub(t.idleSince) >= timeoutResolution) {
				t.send <- newAck(t.seqNoIn)
				t.ackNoIn = t.seqNoIn
				t.idleSince = time.Now()
			}

			if now.Sub(t.idleSince) >= t.conf.IdleTimeout {
				t.send <- newFunc(keepAlive)
				keepAliveSend = time.Now()
				t.idleSince = keepAliveSend
			}

		case l := <-t.launch:
			var f function
			switch l {
			case Exit:
				return
			case Down:
				level = Down
				f = bringDown
			default: // Up or higher
				f = bringUp
			}
			t.send <- newFunc(f)
			t.idleSince = time.Now()

		case datagram, ok := <-t.recv:
			if !ok {
				return
			}

			t.idleSince = time.Now()

			switch datagram.Format() {
			case sFrame:
				if !t.updateAckNoOut(datagram.RecvSeqNo()) {
					return
				}

			case iFrame:
				if level < Up {
					break // discard
				}

				if !t.updateAckNoOut(datagram.RecvSeqNo()) {
					return
				}

				if datagram.SendSeqNo() != t.seqNoIn {
					t.err <- errSeqNo
					return
				}

				t.in <- datagram.Payload()

				if t.ackNoIn == t.seqNoIn {
					// first unacked
					unackRecvd = time.Now()
				}
				t.seqNoIn = (t.seqNoIn + 1) & 32767

				if seqNoCount(t.ackNoIn, t.seqNoIn) >= t.conf.RecvUnackMax {
					t.send <- newAck(t.seqNoIn)
					t.ackNoIn = t.seqNoIn
					t.idleSince = time.Now()
				}

			case uFrame:
				switch datagram.Function() {
				case bringUp:
					level = Up
					t.level <- level
					t.send <- newFunc(bringUpOK)
					t.idleSince = time.Now()

				case bringUpOK:
					level = Up
					t.level <- level
					bringUpSend = willNotTimeout

				case bringDown:
					level = Down
					t.level <- level
					t.send <- newFunc(bringDownOK)
					t.idleSince = time.Now()

				case bringDownOK:
					level = Down
					t.level <- level
					bringDownSend = willNotTimeout

				case keepAlive:
					t.send <- newFunc(keepAliveOK)
					t.idleSince = time.Now()

				case keepAliveOK:
					keepAliveSend = willNotTimeout

				default:
					t.err <- errIllegalFunc
				}
			}
		}
	}
}

func (t *tcp) submit(o *Outbound) {
	seqNo := t.seqNoOut

	datagram, err := packASDU(o.Payload, seqNo, t.seqNoIn)
	if err != nil {
		o.err <- err // buffered channel
		return
	}
	t.ackNoIn = t.seqNoIn
	t.seqNoOut = (seqNo + 1) & 32767

	p := &t.pending[seqNo&32767]
	p.done = o.err
	p.send = time.Now()

	t.send <- datagram
	t.idleSince = time.Now()
}

func (t *tcp) updateAckNoOut(n uint) (ok bool) {
	last := t.ackNoOut
	if n == last {
		return true
	}
	// new acks

	// validate
	if latest := t.seqNoOut; seqNoCount(last, latest) < seqNoCount(n, latest) {
		// don't allow unack nor ack ahead
		t.err <- errAckNo
		return false
	}

	// confirm reception
	for last != n {
		close(t.pending[last].done)
		t.pending[last].done = nil
		last = (last + 1) & 32767
	}

	t.ackNoOut = n
	return true
}

func seqNoCount(nextAckNo, nextSeqNo uint) uint {
	if nextAckNo > nextSeqNo {
		nextSeqNo += 32768
	}
	return nextSeqNo - nextAckNo
}
