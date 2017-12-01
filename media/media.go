// Package media provides the OSI media layers from section 1 & 2.
package media

import (
	"errors"
	"io"
	"strconv"
)

// Fixed serial configuration parameters defined in section 1.
// See also companion standard 101, annex 1.1 — UART definition.
const (
	DataSize = 8   // bits in charater
	StopBits = 1   // character termination
	Parity   = 'E' // even check bit
)

// FrameBits is the number of bits per character needed at the physical layer.
const FrameBits = 1 + DataSize + 1 + StopBits

// FT is a format class for Packet encoding.
type FT interface {
	Encode(io.Writer, Packet) error
	Decode(io.Reader) (Packet, error)
}

// Packet is an network layer PDU (protocol data unit).
type Packet struct {
	Addr Addr   // address of secondary station
	Ctrl Ctrl   // protocol-control information
	Data []byte // optional user data
}

// Addr is a secondary station address.
// The actual presence and size is system-dependent. When not in use (with size 0)
// then GlobalAddr is the mandatory value in all cases.
//
// The reference may be to a single station, a group of stations or all other
// stations. Group addresses should be defined by agreements between users and
// vendors.
type Addr uint16

// Signals Addr is out of bounds.
var ErrAddrFit = errors.New("part5: station address doesn't fit system-specific octet size")

// Network honors the net.Addr interface.
func (a Addr) Network() string { return "serial" }

// String honors the net.Addr interface.
func (a Addr) String() string {
	if a == GlobalAddr {
		return "<global>"
	}
	return strconv.Itoa(int(a))
}

// GlobalAddr is the broadcast address.
// Use this value on 1 octet addressing systems too [not 255].
// This value is applicable in 3 cases, namely
// (1) for single (control) character frames
// (2) for systems with addressing disabled
// (3) for unbalanced systems in combination with
const GlobalAddr Addr = 0xffff

// Ctrl is the control field.
type Ctrl byte

// FromPrimaryFlag marks the frame direction.
const FromPrimaryFlag Ctrl = 1 << 6

// FuncMask gets the function code from the control field.
const FuncMask Ctrl = 1<<4 - 1

// Protocol-control from primary (initiating) stations.
// See section 2, clause 5.1.2 & 6.1.2.
const (
	// Function codes from section 2, table 1 & 3.
	Reset     Ctrl = iota // 0: reset of remote link
	ResetProc             // 1: reset of user process
	Test                  // 2: test function for link (for balanced systems)
	Give                  // 3: user data: send/confirm expected
	Toss                  // 4: user data: send/no reply expected
	_                     // 5: reserved
	_                     // 6: reserved for special use by agreement
	_                     // 7: reserved for special use by agreement
	AvailReq              // 8: poll access demand (for unbalanced systems)
	StatusReq             // 9: request status of link
	Class1Req             // 10: request submission of user data (for unbalanced systems)
	Class2Req             // 11: request submission of user data (for unbalanced systems)
	_                     // 12: reserved
	_                     // 13: reserved
	_                     // 14: reserved for special use by agreement
	_                     // 15: reserved for special use by agreement

	// Flipped on every frame when FrameCountValidFlag to delete losses and duplications.
	FrameCountFlag Ctrl = 1 << 5
	// Marks the applicability of FrameCountFlag for Test, Give, Class1Req and Class2Req.
	FrameCountValidFlag Ctrl = 1 << 4
)

// Protocol-control from secondary (responding) stations.
// See section 2, clause 5.1.2 & 6.1.2.
// For permissible combinations see companion standard 101, table 3 & 4.
const (
	// Function codes from section 2, table 2 & 4.
	Ack    Ctrl = iota // 0: positive acknowledgement
	Nack               // 1: message not accepted, link busy
	_                  // 2: reserved
	_                  // 3: reserved
	_                  // 4: reserved
	_                  // 5: reserved
	_                  // 6: reserved for special use by agreement
	_                  // 7: reserved for special use by agreement
	Data               // 8: submission of user data (for unbalanced systems)
	NoData             // 9: no pending user data (for unbalanced systems)
	_                  // 10: reserved
	OK                 // 11: status of link or access demand
	_                  // 12: reserved
	_                  // 13: reserved for special use by agreement
	Down               // 14: link service not functioning
	NoImpl             // 15: link service not implemented

	// Data flow control: further messages may cause buffer overflow.
	OverflowFlag Ctrl = 1 << 4

	// With unbalanced transmission, secondary stations [slaves] may report
	// the availability of class 1 data with this flag.
	XSDemandFlag Ctrl = 1 << 5
)
