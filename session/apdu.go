package session

import (
	"errors"
	"fmt"
	"io"
)

// Start is the APDU magic number.
const start byte = 0x68

var (
	errStart  = errors.New("part5: APDU start mismatch")
	errLength = errors.New("part5: APDU length out of range")
)

// APDU (Application Protocol Data Unit) is a transport layer datagram.
// The first 6 octets contain the APCI (Application Protocol Control
// Information) header. The following ASDU (Application Service Data Unit)
// body may take up to 249 octets.
type apdu [255]byte

// String returns a compact description.
func (u *apdu) String() string {
	switch u.Format() {
	case uFrame:
		return fmt.Sprintf("U[%s]", u.Function())
	case sFrame:
		return fmt.Sprintf("S[recv=%04X]", u.RecvSeqNo())
	case iFrame:
		return fmt.Sprintf("I[recv=%04X, send=%04X] %#x", u.RecvSeqNo(), u.SendSeqNo(), u[6:u[1]+2])
	default:
		panic("unknown format")
	}
}

// Format defines the APDU type.
type format rune

const (
	// A U-frame is an ASDU in the U format, for the unnumbered control functions.
	uFrame format = 'U'
	// An S-frame is an ASDU in the S format, for numbered supervisory functions.
	sFrame format = 'S'
	// An I-frame is an ASDU in the I format, for numbered information transfers.
	iFrame format = 'I'
)

// Format returns the APDU type.
func (u *apdu) Format() format {
	switch u[2] & 3 {
	case 1:
		return sFrame
	case 3:
		return uFrame
	default:
		return iFrame
	}
}

// Function is a unnumbered control function.
// See companion standard 104, figure 8.
type function uint

const (
	// bringUp is a data transfer start request.
	// See companion standard 104, subclause 5.3.
	bringUp function = 4 << iota

	// bringUpOK is an bringUp confirmation response.
	bringUpOK

	// bringDown is a data transfer stop request.
	// See companion standard 104, subclause 5.3.
	bringDown

	// bringDownOK is a bringDown confirmation response.
	bringDownOK

	// keepAlive is a keep-alive request.
	// See companion standard 104, subclause 5.2.
	keepAlive

	// keepAliveOK is a keepAlive response.
	keepAliveOK
)

// String returns the IEC identification token.
func (f function) String() string {
	switch f {
	case bringUp:
		return "STARTDT_ACT"
	case bringUpOK:
		return "STARTDT_CON"
	case bringDown:
		return "STOPDT_ACT"
	case bringDownOK:
		return "STOPDT_CON"
	case keepAlive:
		return "TESTFR_ACT"
	case keepAliveOK:
		return "TESTFR_CON"
	default:
		return fmt.Sprintf("<illegal %#x>", uint(f))
	}
}

// Function is valid only for the U format.
func (u *apdu) Function() function {
	return function(u[2] & 0xfc)
}

// SendSeqNo returns the 15-bit send sequence number for
// protection against loss and duplication of messages.
// See companion standard 104, subclause 5.1.
// Only valid with the I format.
func (u *apdu) SendSeqNo() uint {
	var x = uint(u[2]) | uint(u[3])<<8
	return x >> 1
}

// RecvSeqNo returns the 15-bit receive sequence number for
// protection against loss and duplication of messages.
// See companion standard 104, subclause 5.1.
// Only valid with the I and S format.
func (u *apdu) RecvSeqNo() uint {
	var x = uint(u[4]) | uint(u[5])<<8
	return x >> 1
}

// Payload returns a copy of the ASDU serial.
func (u *apdu) Payload() []byte {
	p := make([]byte, u[1]-4)
	copy(p, u[6:])
	return p
}

// InitFunc sets the U format.
func (u *apdu) InitFunc(f function) {
	u[0] = start
	u[1] = 4
	u[2] = byte(f | 3)
	u[3] = 0
	u[4] = 0
	u[5] = 0
}

// InitAck sets the S format.
func (u *apdu) InitAck(seqNo uint) {
	u[0] = start
	u[1] = 4
	u[2] = 1
	u[3] = 0
	u[4] = uint8(seqNo << 1)
	u[5] = uint8(seqNo >> 7)
}

// InitASDU sets the I format.
func (u *apdu) InitASDU(asdu []byte, sendSeqNo, recvSeqNo uint) error {
	if len(asdu) > 249 {
		return errLength
	}

	u[0] = start
	u[1] = uint8(len(asdu) + 4)
	u[2] = uint8(sendSeqNo << 1)
	u[3] = uint8(sendSeqNo >> 7)
	u[4] = uint8(recvSeqNo << 1)
	u[5] = uint8(recvSeqNo >> 7)
	copy(u[6:], asdu)

	return nil
}

// Marshal encodes the datagram and returns the number of octets written.
// Skip is the number of octets already written. Its use is inteded to
// recover from partial writes.
func (u *apdu) Marshal(w io.Writer, skip int) (int, error) {
	end := int(u[1]) + 2
	if end < 6 || end > len(u) {
		return 0, errLength
	}

	if skip >= end {
		return 0, nil
	}
	n, err := w.Write(u[skip:end])
	return n, err
}

// Unmarshal decodes the following and returns the number of octets read.
// Skip is the number of octets already received. Its use is inteded to
// recover from partial reads.
func (u *apdu) Unmarshal(r io.Reader, skip int) (int, error) {
	if skip < 2 {
		n, err := io.ReadFull(r, u[skip:2])
		skip += n
		if err != nil {
			if skip != 0 && u[0] != start {
				return skip, errStart
			}
			return skip, err
		}
	}
	// got first two bytes

	if u[0] != start {
		return skip, errStart
	}
	length := int(u[1])
	if length < 4 || length > len(u)-2 {
		return skip, errLength
	}

	// get payload
	if skip >= length+2 {
		return skip, nil
	}
	n, err := io.ReadFull(r, u[skip:length+2])
	skip += n
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return skip, err
}
