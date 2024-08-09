package info

import (
	"errors"
	"io"
)

// Enc is the variable-structure qualifier (VQL), which defines the payload
// encoding in a DataUnit with an object Count and an address sequence flag.
type Enc uint8

// Count returns the number of information objects, which can be zero.
func (e Enc) Count() int { return int(e & 0x7f) }

// AddrSeq returns the SQ flag, which signals that the address of each encoded
// object is one higher than its previous. Specifying only the first information
// object address is more efficient than the address–object encoding without the
// SQ flag.
func (e Enc) AddrSeq() bool { return e&0x80 != 0 }

// ErrAddrSeq rejects a packet encoding that would otherwise cause undefined
// behaviour with the ObjAddr width.
var ErrAddrSeq = errors.New("part5: address sequence [VQL SQ] overflows the addres space for information objects")

type (
	// OrigAddr can be instantiated with OrigAddrN from System.
	// The originator address defaults to zero.
	// Value 1..255 addresses a specific part of the system.
	OrigAddr interface {
		// The presence is fixed per system.
		OrigAddr0 | OrigAddr8

		// N gets the address as a numeric value.
		N() uint
	}

	OrigAddr0 [0]uint8 // without originator address
	OrigAddr8 [1]uint8 // with originator address
)

// OrigAddrN returns either a numeric match, or false when n overflows the
// address width of Orig. Note that any non-zero value for OrigAddr0 gets false.
func (_ System[Orig, Com, Obj]) OrigAddrN(n uint) (Orig, bool) {
	var addr Orig
	for i := 0; i < len(addr); i++ {
		addr[i] = uint8(n)
		n >>= 8
	}
	return addr, n == 0
}

// MustOrigAddrN is like OrigAddrN, yet it panics on overflow.
func (_ System[Orig, Com, Obj]) MustOrigAddrN(n uint) Orig {
	addr, ok := System[Orig, Com, Obj]{}.OrigAddrN(n)
	if !ok {
		panic("overflow of originator address")
	}
	return addr
}

// N implements the ComAddr interface.
func (addr OrigAddr0) N() uint { return 0 }

// N implements the ComAddr interface.
func (addr OrigAddr8) N() uint { return uint(addr[0]) }

// A DataUnit has a transmission packet called ASDU, short for Application
// Service Data Unit. Information objects are encoded conform the TypeID and
// and Enc values. Encoding is likely to contain one or more Object addresses.
type DataUnit[Orig OrigAddr, Com ComAddr, Obj ObjAddr] struct {
	// configuration inheritance with generics
	System[Orig, Com, Obj]

	Type  TypeID // payload class
	Enc          // payload structure
	Cause        // cause of transmission
	Orig  Orig   // originator address
	Addr  Com    // station address

	// The previous causes 2 to 4 bytes of padding.
	bootstrap [10]byte

	Info []byte // encoded payload
}

// NewDataUnit returns a new ASDU.
func (_ System[Orig, Com, Obj]) NewDataUnit() DataUnit[Orig, Com, Obj] {
	var u DataUnit[Orig, Com, Obj]
	u.Info = u.bootstrap[:0]
	return u
}

// Adopt reads the Data Unit Identifier from the ASDU into the fields.
// The remainder of the bytes is sliced as Info without any validation.
func (u *DataUnit[Orig, Com, Obj]) Adopt(asdu []byte) error {
	if len(asdu) < 3+len(u.Orig)+len(u.Addr) {
		if len(asdu) == 0 {
			return io.EOF
		}
		return io.ErrUnexpectedEOF
	}

	// copy header
	u.Type = TypeID(asdu[0])
	u.Enc = Enc(asdu[1])
	u.Cause = Cause(asdu[2])
	for i := 0; i < len(u.Orig); i++ {
		u.Orig[i] = asdu[i+3]
	}
	for i := 0; i < len(u.Addr); i++ {
		u.Addr[i] = asdu[i+3+len(u.Orig)]
	}

	// reject values whom are "not used"
	switch {
	case u.Type == 0:
		return errTypeZero
	case u.Cause&63 == 0:
		return errCauseZero
	case u.Addr.N() == 0:
		return errComAddrZero
	}

	// slice payload
	u.Info = asdu[3+len(u.Orig)+len(u.Addr) : len(asdu) : len(asdu)]
	return nil
}

// Append the ASDU encoding to buf and return the extended buffer.
func (u DataUnit[Orig, Com, Obj]) Append(buf []byte) []byte {
	buf = append(buf, byte(u.Type), byte(u.Enc), byte(u.Cause))
	for i := 0; i < len(u.Orig); i++ {
		buf = append(buf, u.Orig[i])
	}
	for i := 0; i < len(u.Addr); i++ {
		buf = append(buf, u.Addr[i])
	}
	buf = append(buf, u.Info...)
	return buf
}

// Mirrors compares all fields for equality with the exception of Cause. For
// Cause, only the TestFlag is compared for equality. Command responses should
// mirror their respective requests.
func (u DataUnit[Orig, Com, Obj]) Mirrors(o DataUnit[Orig, Com, Obj]) bool {
	return u.Type == o.Type &&
		u.Enc == o.Enc &&
		u.Cause&TestFlag == o.Cause&TestFlag &&
		u.Orig == o.Orig &&
		u.Addr == o.Addr &&
		string(u.Info) == string(o.Info)
}
