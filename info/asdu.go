package info

import (
	"errors"
	"fmt"
	"io"
	"time"
)

var (
	// Narrow is the smalles configuration.
	Narrow = &Params{CauseSize: 1, CommonAddrSize: 1, ObjAddrSize: 1, TimeZone: time.UTC}
	// Wide is the largest configuration.
	Wide = &Params{CauseSize: 2, CommonAddrSize: 2, ObjAddrSize: 3, TimeZone: time.UTC}
)

// Params defines network-specific fixed system parameters.
// See companion standard 101, subclause 7.1.
type Params struct {
	// Number of octets for an ASDU common address.
	// The standard specifies "a" in the range of (1, 2).
	CommonAddrSize int

	// Number of octets for an ASDU cause of transmission.
	// The standard specifies "b" in the range of (1, 2).
	// Value 2 includes/activates the originator address.
	CauseSize int

	// Number of octets for an ASDU information object address.
	// The standard specifies "c" in the range of (1, 3).
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
	case p.CommonAddrSize < 1 || p.CommonAddrSize > 2:
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
	errObjFit     = errors.New("part5: information object index exceeds range (0, 127)")
)

// ID identifies the application data.
type ID struct {
	Addr  CommonAddr // station address
	Type  TypeID     // information content
	Cause Cause      // submission category
}

// ASDU (Application Service Data Unit) is an application message.
type ASDU struct {
	*Params
	ID

	// Originator Address (1, 255) or 0 for the default.
	// The precense is controlled by Params.CauseSize.
	Orig uint8

	InfoSeq bool // marks Info as a sequence

	bootstrap [17]byte // common case
	Info      []byte   // information object serial
}

// NewASDU returns a new ASDU with the provided parameters.
func NewASDU(p *Params, addr CommonAddr, t TypeID, c Cause) *ASDU {
	u := ASDU{
		Params: p,
		ID:     ID{Addr: addr, Type: t, Cause: c},
	}
	u.Info = u.bootstrap[:0]
	return &u
}

// Reply returns a new ASDU which addresses u.
func (u *ASDU) Reply(t TypeID, c Cause) *ASDU {
	r := NewASDU(u.Params, u.Addr, t, c)
	r.Orig = u.Orig
	return r
}

// String returns a compact description.
func (u *ASDU) String() string {
	seqParam := ""
	if u.InfoSeq {
		seqParam = ",seq"
	}
	return fmt.Sprintf("%s[%d]%s %s: %#x", u.Type, u.Addr, seqParam, u.Cause, u.Info)
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

	data = make([]byte, 2+u.CauseSize+u.CommonAddrSize+len(u.Info))
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

	switch u.CommonAddrSize {
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
	n := 2 + u.CauseSize + u.CommonAddrSize
	if n > len(data) {
		return io.EOF
	}

	u.Type = TypeID(data[0])

	// fixed element size
	objSize := ObjSize[u.Type]
	if objSize == 0 {
		return ErrType
	}

	// read the variable structure qualifier
	if vsq := data[1]; vsq > 127 {
		u.InfoSeq = true
		objCount := int(vsq & 127)
		u.Info = make([]byte, u.ObjAddrSize+(objCount*objSize))
	} else {
		u.InfoSeq = false
		objCount := int(vsq)
		u.Info = make([]byte, objCount*(u.ObjAddrSize+objSize))
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

	switch u.CommonAddrSize {
	default:
		return errParam
	case 1:
		addr := CommonAddr(data[n-1])
		if addr == 255 {
			// map 8-bit variant to 16-bit equivalent
			addr = GlobalAddr
		}
		u.Addr = addr

	case 2:
		u.Addr = CommonAddr(data[n-2]) | CommonAddr(data[n-1])<<8
	}

	if copy(u.Info, data[n:]) < len(u.Info) {
		return io.EOF
	}
	return nil
}

// PutObjAddrAt encodes addr into Info at index i.
// The function panics when the byte array is too small.
func (u *ASDU) PutObjAddrAt(i int, addr ObjAddr) error {
	u.Info[i] = byte(addr)

	switch u.ObjAddrSize {
	default:
		return errParam

	case 1:
		if addr > 255 {
			return errObjAddrFit
		}

	case 2:
		if addr > 65535 {
			return errObjAddrFit
		}
		u.Info[i+1] = byte(addr >> 8)

	case 3:
		if addr > 16777215 {
			return errObjAddrFit
		}
		u.Info[i+1] = byte(addr >> 8)
		u.Info[i+2] = byte(addr >> 16)
	}

	return nil
}

// GetObjAddrAt decodes from Info at index i.
// The function panics when the byte array is too small
// or when the address size parameter is out of bounds.
func (u *ASDU) GetObjAddrAt(i int) ObjAddr {
	addr := ObjAddr(u.Info[i])

	switch u.ObjAddrSize {
	default:
		panic(errParam)
	case 1:
		break
	case 2:
		addr |= ObjAddr(u.Info[i+1]) << 8
	case 3:
		addr |= ObjAddr(u.Info[i+1])<<8 | ObjAddr(u.Info[i+2])<<16
	}

	return addr
}
