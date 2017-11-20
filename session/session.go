// Package session provides the OSI transport and session layers.
package session

import (
	"errors"
	"fmt"
	"net"
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
type Level int

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
// In is closed first on Exit followed by Err. Class1 and Class2
// MUST be closed by the user after, and only after Exit.
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

	// Err captures protocol failure.
	// Err must be read or operation may block and behave
	// in an unexpected way.
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
