package info

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/bits"
	"time"
)

var (
	// Narrow is the smalles configuration.
	Narrow = &Params{CauseSize: 1, AddrSize: 1, ObjAddrSize: 1, TimeZone: time.UTC}
	// Wide is the largest configuration.
	Wide = &Params{CauseSize: 2, AddrSize: 2, ObjAddrSize: 3, TimeZone: time.UTC}
)

// Params defines network-specific fixed system parameters.
// See companion standard 101, subclause 7.1.
type Params struct {
	// Number of octets for an ASDU common address.
	// The standard requires "a" in [1, 2].
	AddrSize int

	// Number of octets for an ASDU cause of transmission.
	// The standard requires "b" in [1, 2].
	// Value 2 includes/activates the originator address.
	CauseSize int

	// Number of octets for an ASDU information object address.
	// The standard requires "c" in [1, 3].
	ObjAddrSize int

	// TimeZone controls the time tag interpretation.
	// The standard fails to mention this one.
	TimeZone *time.Location
}

// ErrParam signals an illegal value in Params.
var errParam = errors.New("part5: fixed system parameter out of range")

// Valid returns the validation result.
func (p Params) Valid() error {
	switch {
	default:
		return nil
	case p.CauseSize < 1 || p.CauseSize > 2:
		return errParam
	case p.AddrSize < 1 || p.AddrSize > 2:
		return errParam
	case p.ObjAddrSize < 1 || p.ObjAddrSize > 3:
		return errParam
	case p.TimeZone == nil:
		return errParam
	}
}

var (
	errOrigFit    = errors.New("part5: originator address not allowed with cause size 1 system parameter")
	errAddrFit    = errors.New("part5: common address exceeds size system parameter")
	errObjAddrFit = errors.New("part5: information object address exceeds size system parameter")
	errObjFit     = errors.New("part5: information object index not in [1, 127]")
)

// ValidAddr returns the validation result of a station address.
func (p Params) ValidAddr(addr CommonAddr) error {
	if addr == 0 {
		return errAddrZero
	}
	if bits.Len(uint(addr)) > p.AddrSize*8 {
		return errAddrFit
	}
	return nil
}

// ObjAddr decodes an information object address from buf.
// The function panics when the byte array is too small
// or when the address size parameter is out of bounds.
func (p *Params) ObjAddr(buf []byte) ObjAddr {
	addr := ObjAddr(buf[0])

	switch p.ObjAddrSize {
	default:
		panic(errParam)
	case 1:
		break
	case 2:
		addr |= ObjAddr(buf[1]) << 8
	case 3:
		addr |= ObjAddr(buf[1])<<8 | ObjAddr(buf[2])<<16
	}

	return addr
}

// A VarStructQual, or variable structure qualifier, defines the ASDU payload
// with a Count() and a Sequence flag.
type VarStructQual uint8

// Variable Structure Qualifier Flags
const (
	// The SQ flag signals that the information applies to a consecutive
	// sequence of addresses, in which each information object address is
	// one higher than its previous element. The address in the ASDU gives
	// the offset (for the first information object).
	Sequence = 0x80
)

// Count returns the number of information objects, which can be zero.
func (q VarStructQual) Count() int { return int(q & 0x7f) }

// ID identifies the application data.
type ID struct {
	Type   TypeID
	Struct VarStructQual
	Cause  Cause

	// Originator Address [1, 255] or 0 for the default.
	// The applicability is controlled by Params.CauseSize.
	Orig uint8

	Addr CommonAddr // station address
}

// String returns a compact label.
func (id ID) String() string {
	if id.Orig == 0 {
		return fmt.Sprintf("@%d %s <%s>", id.Addr, id.Type, id.Cause)
	}
	return fmt.Sprintf("%d@%d %s <%s>", id.Orig, id.Addr, id.Type, id.Cause)
}

// ASDU (Application Service Data Unit) is an application message.
type ASDU struct {
	*Params
	ID
	bootstrap [17]byte // prevents Info malloc
	Info      []byte   // information object serial
}

// NewASDU returns a new ASDU with the provided parameters.
func NewASDU(p *Params, id ID) *ASDU {
	u := ASDU{Params: p, ID: id}
	u.Info = u.bootstrap[:0]
	return &u
}

var errInroGroup = errors.New("part5: interrogation group number exceeds 16")

// MustNewInro returns a new interrogation command [C_IC_NA_1].
// Use group 1 to 16, or 0 for the default.
func MustNewInro(p *Params, addr CommonAddr, orig uint8, group uint) *ASDU {
	if group > 16 {
		panic(errInroGroup)
	}

	u := ASDU{
		Params: p,
		ID: ID{
			Type:   C_IC_NA_1,
			Struct: 0x01,
			Cause:  Act,
			Addr:   addr,
			Orig:   orig,
		},
	}

	u.Info = u.bootstrap[:p.ObjAddrSize+1]
	u.Info[p.ObjAddrSize] = byte(group + uint(Inrogen))
	return &u
}

// Respond returns a new "responding" ASDU which addresses "initiating" u.
func (u *ASDU) Respond(t TypeID, c Cause) *ASDU {
	return NewASDU(u.Params, ID{
		Addr:  u.Addr,
		Orig:  u.Orig,
		Type:  t,
		Cause: c | u.Cause&TestFlag,
	})
}

// Reply returns a new "responding" ASDU which addresses "initiating" u
// with a copy of Info.
func (u *ASDU) Reply(c Cause) *ASDU {
	r := NewASDU(u.Params, u.ID)
	r.Cause = c | u.Cause&TestFlag
	r.Struct = u.Struct & Sequence
	r.Info = append(r.Info, u.Info...)
	return r
}

// String returns a full description.
func (u *ASDU) String() string {
	dataSize := ObjSize[u.Type]
	if dataSize == 0 {
		if u.Struct&Sequence == 0 {
			return fmt.Sprintf("%s: %#x", u.ID, u.Info)
		}
		return fmt.Sprintf("%s seq: %#x", u.ID, u.Info)
	}

	if len(u.Info) < u.ObjAddrSize {
		if u.Struct&Sequence == 0 {
			return fmt.Sprintf("%s: %#x <EOF>", u.ID, u.Info)
		}
		return fmt.Sprintf("%s seq: %#x <EOF>", u.ID, u.Info)
	}
	addr := u.ObjAddr(u.Info)

	buf := bytes.NewBufferString(u.ID.String())

	for i := u.ObjAddrSize; ; {
		if i+dataSize > len(u.Info) {
			fmt.Fprintf(buf, " %d:%#x <EOF>", addr, u.Info[i:])
			break
		}
		fmt.Fprintf(buf, " %d:%#x", addr, u.Info[i:i+dataSize])
		i += dataSize

		if i == len(u.Info) {
			break
		}

		if u.Struct&Sequence != 0 {
			addr++
		} else {
			if i+u.ObjAddrSize > len(u.Info) {
				fmt.Fprintf(buf, " %#x <EOF>", u.Info[i:])
				break
			}
			addr = u.ObjAddr(u.Info[i : i+u.ObjAddrSize])
			i += u.ObjAddrSize
		}
	}

	return buf.String()
}

// MarshalBinary honors the encoding.BinaryMarshaler interface.
func (u *ASDU) MarshalBinary() (data []byte, err error) {
	switch {
	case u.Cause == 0:
		return nil, errCauseZero
	case u.Addr == 0:
		return nil, errAddrZero
	}

	data = make([]byte, 2+u.CauseSize+u.AddrSize+len(u.Info))
	data[0] = byte(u.Type)
	data[1] = byte(u.Struct)
	data[2] = byte(u.Cause)
	i := 3 // write index

	switch u.CauseSize {
	case 1:
		if u.Orig != 0 {
			return nil, errOrigFit
		}
	case 2:
		data[i] = u.Orig
		i++
	default:
		return nil, errParam
	}

	switch u.AddrSize {
	case 1:
		if u.Addr == GlobalAddr {
			data[i] = 255
		} else if u.Addr >= 255 {
			return nil, errAddrFit
		} else {
			data[i] = byte(u.Addr)
		}
		i++
	case 2:
		data[i] = byte(u.Addr)
		i++
		data[i] = byte(u.Addr >> 8)
		i++
	default:
		return nil, errParam
	}

	copy(data[i:], u.Info)

	return data, nil
}

// UnmarshalBinary honors the encoding.BinaryUnmarshaler interface.
// Params must be set in advance. All other fields are initialized.
func (u *ASDU) UnmarshalBinary(data []byte) error {
	if len(data) < 3 {
		if len(data) == 0 {
			return io.EOF
		}
		return io.ErrUnexpectedEOF
	}

	u.Type = TypeID(data[0])
	u.Struct = VarStructQual(data[1])
	u.Cause = Cause(data[2])
	i := 3 // read index

	switch u.CauseSize {
	case 1:
		u.Orig = 0
	case 2:
		if len(data) < 4 {
			return io.ErrUnexpectedEOF
		}
		u.Orig = data[3]
		i = 4
	default:
		return errParam
	}

	switch u.AddrSize {
	case 1:
		if len(data) < i {
			return io.ErrUnexpectedEOF
		}
		u.Addr = CommonAddr(data[i])
		i++
		if u.Addr == 255 {
			// map 8-bit variant to 16-bit equivalent
			u.Addr = GlobalAddr
		}
	case 2:
		if len(data) < i+1 {
			return io.ErrUnexpectedEOF
		}
		u.Addr = CommonAddr(data[i]) | CommonAddr(data[i+1])<<8
		i += 2
	default:
		return errParam
	}

	u.Info = append(u.bootstrap[:0], data[i:]...)
	return nil
}

// AddObjAddr appends an information object address to Info.
func (u *ASDU) AddObjAddr(addr ObjAddr) error {
	switch u.ObjAddrSize {
	default:
		return errParam

	case 1:
		if addr > 255 {
			return errObjAddrFit
		}
		u.Info = append(u.Info, byte(addr))

	case 2:
		if addr > 65535 {
			return errObjAddrFit
		}
		u.Info = append(u.Info, byte(addr), byte(addr>>8))

	case 3:
		if addr > 16777215 {
			return errObjAddrFit
		}
		u.Info = append(u.Info, byte(addr), byte(addr>>8), byte(addr>>16))
	}

	return nil
}
