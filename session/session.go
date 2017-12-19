// Package session provides the OSI transport and session layers.
package session

import (
	"errors"
	"fmt"
	"net"
	"time"
)

var (
	// ErrConnLost signals unknown reception status.
	ErrConnLost = errors.New("part5: connection lost")

	// ErrNoConn signals unable to perform.
	ErrNoConn = errors.New("part5: no connection")
)

// Trace activates wire logging.
var Trace = false

// Level is the availability status.
type Level uint

const (
	Exit Level = iota // no connection
	Down              // not operational
	Up                // operational
)

// String returns a name.
func (l Level) String() string {
	switch l {
	case Exit:
		return "exit"
	case Down:
		return "down"
	case Up:
		return "up"
	default:
		return fmt.Sprintf("level%+d", l)
	}
}

// Station is an RTU (remote terminal unit) or SCADA (Supervisory
// Control And Data Acquisition) system.
type Station struct {
	Transport

	Addr net.Addr // remote address

	// Level propagates status changes and must be read or
	// operation blocks and may behave in an unexpected way.
	Level <-chan Level

	// Launch targets a level.
	// Any failure to do so terminates the connection.
	Launch chan<- Level
}

// Transport layer as datagram channels.
// Channel In and Err MUST be read continuously or operation may block and
// behave in an unexpected way. On Exit, In is closed first follewed by Err.
// Both Class1 and Class2 MUST be closed by the user and doing so ensures an
// Exit.
type Transport struct {
	// In captures inbound datagrams in order of appearance.
	In <-chan []byte

	// Class1 blocks until the payload is accepted
	// and sealed with timeout protection mechanisms.
	// Class Ⅰ data transmission is typically used for
	// events or for messages with high priority.
	Class1 chan<- *Outbound

	// Class2 blocks until the payload is accepted
	// and sealed with timeout protection mechanisms.
	// Class Ⅱ data transmission is typically used for
	// cyclic transmission or for low priority messages.
	Class2 chan<- *Outbound

	// Err captures all protocol failures which are not
	// directly related to an Outbound submission.
	Err <-chan error
}

// Outbound is a single-use data submission handle.
type Outbound struct {
	Payload []byte // datagram

	// In case of failure exactly one error is send.
	// It is safe to wait for the channel to close once
	// the submission (either class Ⅰ or Ⅱ) is accepted.
	Done <-chan error

	err chan<- error // hidden Done counterpart
}

// NewOutbound returns a new Outbound ready to use [once].
func NewOutbound(payload []byte) *Outbound {
	// not allowed to block
	ch := make(chan error, 1)
	return &Outbound{payload, ch, ch}
}

var errPipeTimeout = errors.New("part5: pipe exchange timeout")

// Pipe creates a synchronous, in-memory, full duplex session.
// Reads on one end are matched with writes on the other, copying
// data directly between the two; there is no internal buffering.
// The timeout sets the upper limit for an Outbound block.
func Pipe(timeout time.Duration) (*Transport, *Transport) {
	aQuit := make(chan struct{})
	aFeed := make(chan []byte)
	aClass1 := make(chan *Outbound)
	aClass2 := make(chan *Outbound)

	bQuit := make(chan struct{})
	bFeed := make(chan []byte)
	bClass1 := make(chan *Outbound)
	bClass2 := make(chan *Outbound)

	go func() {
		defer close(aQuit)
		feedPipe(bFeed, timeout, aClass1, aClass2, bQuit)
	}()
	go func() {
		defer close(bQuit)
		feedPipe(aFeed, timeout, bClass1, bClass2, aQuit)
	}()

	aIn := make(chan []byte)
	aErr := make(chan error)

	bIn := make(chan []byte)
	bErr := make(chan error)

	go func() {
		for {
			select {
			case <-aQuit:
				close(aIn)
				close(aErr)
				return
			case bytes := <-aFeed:
				aIn <- bytes
			}
		}
	}()

	go func() {
		for {
			select {
			case <-bQuit:
				close(bIn)
				close(bErr)
				return
			case bytes := <-bFeed:
				bIn <- bytes
			}
		}
	}()

	return &Transport{In: aIn, Class1: aClass1, Class2: aClass2, Err: aErr},
		&Transport{In: bIn, Class1: bClass1, Class2: bClass2, Err: bErr}
}

func feedPipe(feed chan []byte, timeout time.Duration, class1, class2 chan *Outbound, remoteQuit chan struct{}) {
	defer func() {
		go func() {
			for out := range class1 {
				out.err <- ErrNoConn
				close(out.err)
			}
			for out := range class2 {
				out.err <- ErrNoConn
				close(out.err)
			}
		}()
	}()

	// init reusable instance
	expire := time.NewTimer(time.Minute)
	expire.Stop()

	for {
		var out *Outbound
		var ok bool
		select {
		case out, ok = <-class1:
			break
		default:
			select {
			case <-remoteQuit:
				return
			case out, ok = <-class1:
				break
			case out, ok = <-class2:
				break
			}
		}
		if !ok { // local quit
			return
		}

		data := make([]byte, len(out.Payload))
		copy(data, out.Payload)

		if expire.Reset(timeout) {
			panic("pending expiry timer")
		}

		select {
		case feed <- data:
			if !expire.Stop() {
				<-expire.C
			}
			close(out.err)

		case <-remoteQuit:
			if !expire.Stop() {
				<-expire.C
			}
			out.err <- ErrConnLost
			close(out.err)
			return

		case <-expire.C:
			out.err <- errPipeTimeout
			close(out.err)
		}
	}
}
