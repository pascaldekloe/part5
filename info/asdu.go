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

// ID identifies the application data.
type ID struct {
	Addr CommonAddr // station address

	// Originator Address [1, 255] or 0 for the default.
	// The applicability is controlled by Params.CauseSize.
	Orig uint8

	Type  TypeID // information content
	Cause Cause  // submission category
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
	InfoSeq   bool     // marks Info as a sequence
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
		ID:     ID{addr, orig, C_IC_NA_1, Act},
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
	r.InfoSeq = u.InfoSeq
	r.Info = append(r.Info, u.Info...)
	return r
}

// String returns a full description.
func (u *ASDU) String() string {
	dataSize := ObjSize[u.Type]
	if dataSize == 0 {
		if !u.InfoSeq {
			return fmt.Sprintf("%s: %#x", u.ID, u.Info)
		}
		return fmt.Sprintf("%s seq: %#x", u.ID, u.Info)
	}

	end := len(u.Info)
	addrSize := u.ObjAddrSize
	if end < addrSize {
		if !u.InfoSeq {
			return fmt.Sprintf("%s: %#x <EOF>", u.ID, u.Info)
		}
		return fmt.Sprintf("%s seq: %#x <EOF>", u.ID, u.Info)
	}
	addr := u.ObjAddr(u.Info)

	buf := bytes.NewBufferString(u.ID.String())

	for i := addrSize; ; {
		start := i
		i += dataSize
		if i > end {
			fmt.Fprintf(buf, " %d:%#x <EOF>", addr, u.Info[start:])
			break
		}
		fmt.Fprintf(buf, " %d:%#x", addr, u.Info[start:i])
		if i == end {
			break
		}

		if u.InfoSeq {
			addr++
		} else {
			start = i
			i += addrSize
			if i > end {
				fmt.Fprintf(buf, " %#x <EOF>", u.Info[start:i])
				break
			}
			addr = u.ObjAddr(u.Info[start:])
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

	// calculate the size declaration byte
	// named "variable structure qualifier"
	var vsq byte
	{
		// fixed element size
		objSize := ObjSize[u.Type]
		if objSize == 0 {
			return nil, ErrType
		}

		// See companion standard 101, subclause 7.2.2.
		if u.InfoSeq {
			objCount := (len(u.Info) - u.ObjAddrSize) / objSize
			if objCount >= 128 {
				return nil, errObjFit
			}
			vsq = byte(objCount) | 128
		} else {
			objCount := len(u.Info) / (u.ObjAddrSize + objSize)
			if objCount >= 128 {
				return nil, errObjFit
			}
			vsq = byte(objCount)
		}
	}

	data = make([]byte, 2+u.CauseSize+u.AddrSize+len(u.Info))
	data[0] = byte(u.Type)
	data[1] = vsq
	data[2] = byte(u.Cause)

	i := 3
	switch u.CauseSize {
	default:
		return nil, errParam
	case 1:
		if u.Orig != 0 {
			return nil, errOrigFit
		}
	case 2:
		data[i] = u.Orig
		i++
	}

	switch u.AddrSize {
	default:
		return nil, errParam
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
	}

	copy(data[i:], u.Info)

	return data, nil
}

// UnmarshalBinary honors the encoding.BinaryUnmarshaler interface.
// Params must be set in advance. All other fields are initialized.
func (u *ASDU) UnmarshalBinary(data []byte) error {
	// data unit identifier size check
	lenDUI := 2 + u.CauseSize + u.AddrSize
	if lenDUI > len(data) {
		return io.EOF
	}
	u.Info = append(u.bootstrap[:0], data[lenDUI:]...)

	u.Type = TypeID(data[0])

	// fixed element size
	objSize := ObjSize[u.Type]
	if objSize == 0 {
		return ErrType
	}

	var size int
	// read the variable structure qualifier
	if vsq := data[1]; vsq > 127 {
		u.InfoSeq = true
		objCount := int(vsq & 127)
		size = u.ObjAddrSize + (objCount * objSize)
	} else {
		u.InfoSeq = false
		objCount := int(vsq)
		size = objCount * (u.ObjAddrSize + objSize)
	}

	switch {
	case size == 0:
		return errObjFit
	case size > len(u.Info):
		return io.EOF
	case size < len(u.Info):
		// not explicitly prohibited
		u.Info = u.Info[:size]
	}

	u.Cause = Cause(data[2])

	switch u.CauseSize {
	default:
		return errParam
	case 1:
		u.Orig = 0
	case 2:
		u.Orig = data[3]
	}

	switch u.AddrSize {
	default:
		return errParam
	case 1:
		addr := CommonAddr(data[lenDUI-1])
		if addr == 255 {
			// map 8-bit variant to 16-bit equivalent
			addr = GlobalAddr
		}
		u.Addr = addr
	case 2:
		u.Addr = CommonAddr(data[lenDUI-2]) | CommonAddr(data[lenDUI-1])<<8
	}

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
