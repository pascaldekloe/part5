package session

import (
	"bytes"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSeqNoCount(t *testing.T) {
	var golden = []struct{ first, last, want uint }{
		{0, 0, 0},
		{0, 9, 9},
		{32767, 32767, 0},
		{32766, 32767, 1},
		{32767, 0, 1},
		{32766, 1, 3},
	}
	for _, gold := range golden {
		got := seqNoCount(gold.first, gold.last)
		if got != gold.want {
			t.Errorf("got %d for #%d to #%d, want %d", got, gold.first, gold.last, gold.want)
		}
	}
}

// TestSeq ping–pongs a message back and forth.
// The I-frames should ack each other.
// On Exit ping should ack the last pong (with a U-frame).
func TestSeq(t *testing.T) {
	if testing.Short() {
		return
	}
	t.Parallel()

	connA, connB := net.Pipe()
	ping, pong, exitGroup := newTCPTestDuo(t, connA, connB, TCPConfig{})
	defer exitGroup.Wait()

	// overflow sequence numbers
	const calls = 99 + (1 << 15)

	// ping greets pong and waits for confirmation
	go func() {
		for i := 0; i < calls; i++ {
			o := NewOutbound([]byte("ping"))
			ping.Class2 <- o
			if err := <-o.Done; err != nil {
				t.Fatal("ping outbound error:", err)
			}
		}
		t.Log("pinged all; close connection")
		ping.Launch <- Exit
	}()

	// blocks until pong exit to ensure ping sends ack before close
	pongQuit := make(chan struct{})

	// pong answers ping and waits for confirmation
	go func() {
		for payload := range pong.In {
			if string(payload) != "ping" {
				t.Fatalf(`pong got %q, want "ping"`, payload)
			}
			o := NewOutbound([]byte("pong"))
			pong.Class1 <- o
			if err := <-o.Done; err != nil {
				t.Fatal("pong outbound error:", err)
			}
		}
		close(pongQuit)
	}()

	// need timeouts to report issues instead of hanging on'm
	expire := time.After(9 * time.Second)

	for i := 0; i < calls; i++ {
		select {
		case payload := <-ping.In:
			if string(payload) != "pong" {
				t.Fatalf(`ping got %q, want "pong"`, payload)
			}
		case <-expire:
			t.Fatalf("ping receive #%d timeout", i)
		}
	}

	select {
	case <-pongQuit:
		break
	case <-expire:
		t.Fatal("pong quit timeout")
	}
}

func TestBringDown(t *testing.T) {
	var logLines bytes.Buffer
	log.SetOutput(&logLines)
	Trace = true

	connA, connB := net.Pipe()
	a, b, exitGroup := newTCPTestDuo(t, connA, connB, TCPConfig{})
	defer exitGroup.Wait()

	a.Launch <- Down
	select {
	case a.Class1 <- NewOutbound([]byte("ping")):
		t.Error("outbound class Ⅰ on station A still receptive")
	default:
		break
	}

	// let STOPDT sink in
	time.Sleep(10 * time.Millisecond)

	select {
	case b.Class2 <- NewOutbound([]byte("pong")):
		t.Error("outbound class Ⅱ on station B still receptive")
	default:
		break
	}

	close(a.Launch) // exit
	time.Sleep(2 * timeoutResolution)

	o := NewOutbound([]byte("ping"))
	select {
	case a.Class2 <- o:
		select {
		case err := <-o.Done:
			if err != ErrNoConn {
				t.Errorf("outbound after exit error %v, want %v", err, ErrNoConn)
			}
		case <-time.After(10 * time.Millisecond):
			t.Error("outbound after exit timeout")
		}
	default:
		t.Error("outbound after exit blocked")
	}

	exitGroup.Wait()

	for {
		line, err := logLines.ReadString('\n')
		if err == io.EOF {
			break
		}

		switch {
		case strings.Contains(line, "U[STARTDT_ACT]"):
		case strings.Contains(line, "U[STARTDT_CON]"):
		case strings.Contains(line, "U[STOPDT_ACT]"):
		case strings.Contains(line, "U[STOPDT_CON]"):
		default:
			t.Errorf("don't want log line %q", line)
		}
	}
}

func TestUnackThreashold(t *testing.T) {
	connA, connB := net.Pipe()
	a, b, exitGroup := newTCPTestDuo(t, connA, connB, TCPConfig{
		// send acknowledge on third pending
		RecvUnackMax: 3,
	})
	defer func() {
		a.Launch <- Exit
		exitGroup.Wait()
	}()

	go func() {
		for range b.In {
			continue // discard
		}
	}()

	first := NewOutbound([]byte("arbitrary"))
	select {
	case a.Class1 <- first:
		break
	case <-time.After(10 * time.Millisecond):
		t.Fatal("first outbound blocked")
	}

	second := NewOutbound([]byte("arbitrary"))
	select {
	case a.Class2 <- second:
		break
	case <-time.After(10 * time.Millisecond):
		t.Fatal("second outbound blocked")
	}

	select {
	case err := <-first.Done:
		t.Fatal("first done before threashold limit:", err)
	case err := <-second.Done:
		t.Fatal("second done before threashold limit:", err)
	case <-time.After(10 * time.Millisecond):
		break
	}

	// should trigger acknowledge
	third := NewOutbound([]byte("arbitrary"))
	select {
	case a.Class2 <- third:
		break
	case <-time.After(10 * time.Millisecond):
		t.Fatal("third outbound blocked")
	}

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		select {
		case err := <-first.Done:
			if err != nil {
				t.Error("first outbound error:", err)
			}
		case <-time.After(10 * time.Millisecond):
			t.Error("first outbound acknowledge timeout")
		}
	}()
	go func() {
		defer wg.Done()
		select {
		case err := <-second.Done:
			if err != nil {
				t.Error("second outbound error:", err)
			}
		case <-time.After(10 * time.Millisecond):
			t.Error("second outbound acknowledge timeout")
		}
	}()
	go func() {
		defer wg.Done()
		select {
		case err := <-third.Done:
			if err != nil {
				t.Error("third outbound error:", err)
			}
		case <-time.After(10 * time.Millisecond):
			t.Error("third outbound acknowledge timeout")
		}
	}()
	wg.Wait()
}

func TestFastAckOnIdle(t *testing.T) {
	t.Parallel()

	connA, connB := net.Pipe()
	a, b, exitGroup := newTCPTestDuo(t, connA, connB, TCPConfig{})
	defer func() {
		a.Launch <- Exit
		exitGroup.Wait()
	}()

	go func() {
		for range b.In {
			continue // discard
		}
	}()

	first := NewOutbound([]byte("arbitrary"))
	select {
	case a.Class1 <- first:
		break
	case <-time.After(10 * time.Millisecond):
		t.Fatal("class Ⅰ outbound blocked")
	}

	select {
	case err := <-first.Done:
		if err != nil {
			t.Fatal("outbound error:", err)
		}
		t.Fatal("too soon")
	case <-time.After(50 * time.Millisecond):
		break
	}

	select {
	case err := <-first.Done:
		if err != nil {
			t.Fatal("outbound error:", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout")
	}
}

type temporaryProblem struct{}

func (p temporaryProblem) Error() string   { return "temporary problem test" }
func (p temporaryProblem) Timeout() bool   { panic("not used") }
func (p temporaryProblem) Temporary() bool { return true }

type badConn struct {
	net.Conn
}

func (bad badConn) Read(p []byte) (n int, err error) {
	if len(p) > 5 {
		p = p[:5]
		n, err = bad.Conn.Read(p)
		if err == nil {
			err = temporaryProblem{}
		}
	} else {
		n, err = bad.Conn.Read(p)
	}
	return
}

func (bad badConn) Write(p []byte) (n int, err error) {
	if len(p) > 5 {
		p = p[:5]
		n, err = bad.Conn.Write(p)
		if err == nil {
			err = temporaryProblem{}
		}
	} else {
		n, err = bad.Conn.Write(p)
	}
	return
}

func TestBadConn(t *testing.T) {
	t.Parallel()

	// get data transfer Up dispite temporary errors
	connA, connB := net.Pipe()
	a, _, exitGroup := newTCPTestDuo(t, badConn{connA}, badConn{connB}, TCPConfig{
		SendUnackTimeout: time.Second,
	})
	defer exitGroup.Wait()

	// expires with the retry interval and 5 byte steps
	first := NewOutbound(make([]byte, 249))
	select {
	case a.Class1 <- first:
		break
	case <-time.After(10 * time.Millisecond):
		t.Fatal("first outbound blocked")
	}

	// won't execute because of first
	second := NewOutbound([]byte("arbitrary"))
	select {
	case a.Class1 <- second:
		break
	case <-time.After(10 * time.Millisecond):
		t.Fatal("second outbound blocked")
	}

	exitGroup.Add(2)
	go func() {
		defer exitGroup.Done()
		select {
		case err := <-first.Done:
			if err != errAckExpire {
				t.Errorf("first outbound error %v, want %v", err, errAckExpire)
			}
		case <-time.After(2 * time.Second):
			t.Error("first outbound expired")
		}
	}()
	go func() {
		defer exitGroup.Done()
		select {
		case err := <-second.Done:
			if err != ErrConnLost {
				t.Errorf("second outbound error %v, want %v", err, ErrConnLost)
			}
		case <-time.After(2 * time.Second):
			t.Error("second outbound expired")
		}
	}()
}

// BenchmarkFlood tests one-sided data push.
func BenchmarkFlood(bench *testing.B) {
	connA, connB := net.Pipe()
	a, b, exitGroup := newTCPTestDuo(bench, connA, connB, TCPConfig{
		RecvUnackTimeout: time.Second,
	})
	defer func() {
		a.Launch <- Exit
		exitGroup.Wait()
	}()

	data := []byte("Hello World!")
	go func() {
		for payload := range b.In {
			if !bytes.Equal(payload, data) {
				bench.Fatalf("got %q, want %q", payload, data)
			}
			continue // discard
		}
	}()

	bench.SetBytes(int64(len(data)))
	bench.ReportAllocs()
	bench.SetParallelism(12)

	bench.ResetTimer()
	bench.RunParallel(func(p *testing.PB) {
		for p.Next() {
			o := NewOutbound(data)
			a.Class2 <- o
			if err := <-o.Done; err != nil {
				bench.Fatal(err)
			}
		}
	})
	bench.StopTimer()
}

// NewTCPTestDuo initiates a session with two stations.
func newTCPTestDuo(t testing.TB, connA, connB net.Conn, config TCPConfig) (a, b *Station, exitGroup *sync.WaitGroup) {
	exitGroup = new(sync.WaitGroup)

	// fail test on expiry and free exitGroup
	expiry := 10 * time.Second
	deadline := time.After(expiry)

	// ensure a level or fail test
	wantLevel := func(t testing.TB, stat *Station, want Level) {
		t.Helper()

		select {
		case l := <-stat.Level:
			if l != want {
				t.Errorf("reached level %s first, want %s", l, want)
				exitGroup.Wait()
				t.FailNow()
			}

		case err := <-stat.Err:
			t.Logf("error before level %s: %s", want, err)
		case <-deadline:
			t.Errorf("reach level %s timeout", want)
			exitGroup.Wait()
			t.FailNow()
		}
	}

	a = TCP(config, connA)
	b = TCP(config, connB)

	// fail test on error; reselase exitGroup on channel close
	exitGroup.Add(2)
	go func() {
		defer exitGroup.Done()

		expire := time.After(expiry)
		for {
			select {
			case err := <-a.Err:
				if err == nil {
					t.Log("station A error closed")
					return
				}
				t.Logf("station A error: %#v", err)
			case <-expire:
				t.Error("station A error close expired")
				return
			}
		}
	}()
	go func() {
		defer exitGroup.Done()

		expire := time.After(expiry)
		for {
			select {
			case err := <-b.Err:
				if err == nil {
					t.Log("station B error closed")
					return
				}
				t.Logf("station B error: %#v", err)
			case <-expire:
				t.Error("station B error close expired")
				return
			}
		}
	}()

	// must initate as Down: connected and no "data transfer"
	wantLevel(t, a, Down)
	wantLevel(t, b, Down)

	a.Launch <- Up
	wantLevel(t, b, Up)
	wantLevel(t, a, Up)

	// report level changes; release exitGroup on Exit
	exitGroup.Add(2)
	go func() {
		defer exitGroup.Done()

		expire := time.After(expiry)
		for {
			select {
			case l := <-a.Level:
				t.Log("station A level", l)
				if l == Exit {
					return
				}
			case <-expire:
				t.Error("station A exit level expired")
				return
			}
		}
	}()
	go func() {
		defer exitGroup.Done()

		expire := time.After(expiry)
		for {
			select {
			case l := <-b.Level:
				t.Log("station B level", l)
				if l == Exit {
					return
				}
			case <-expire:
				t.Error("station B exit level expired")
				return
			}
		}
	}()

	return
}
