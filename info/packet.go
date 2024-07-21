package info

import (
	"encoding/binary"
	"io"
	"math"
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

// The originator address defaults to zero.
// Value 1..255 addresses a specific part of the system.
type (
	// An OrigAddr can be instantiated with OrigAddrN from Params.
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
func (p Params[Orig, Com, Obj]) OrigAddrN(n uint) (Orig, bool) {
	var addr Orig
	for i := 0; i < len(addr); i++ {
		addr[i] = uint8(n)
		n >>= 8
	}
	return addr, n == 0
}

// MustOrigAddrN is like OrigAddrN, yet it panics on overflow.
func (p Params[Orig, Com, Obj]) MustOrigAddrN(n uint) Orig {
	addr, ok := p.OrigAddrN(n)
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
	// configuration inherited from generics
	Params[Orig, Com, Obj]

	Type  TypeID // payload class
	Enc          // payload structure
	Cause        // cause of transmission
	Orig  Orig   // originator address
	Addr  Com    // station address

	// The previous causes 2 to 4 bytes of padding.
	bootstrap [10]byte

	Info []byte // encoded payload
}

func (u *DataUnit[Orig, Com, Obj]) setAddr(addr Obj) {
	u.Info = u.bootstrap[:0]
	u.addAddr(addr)
}

func (u *DataUnit[Orig, Com, Obj]) addAddr(addr Obj) {
	for i := 0; i < len(addr); i++ {
		u.Info = append(u.Info, addr[i])
	}
}

// NewDataUnit returns a new ASDU.
func (p Params[Orig, Com, Obj]) NewDataUnit() DataUnit[Orig, Com, Obj] {
	return DataUnit[Orig, Com, Obj]{}
}

// SingleCmd sets Type, Enc,  and Info with single command C_SC_NA_1 Act
// conform companion standard 101, subsection 7.3.2.1.
func (p Params[Orig, Com, Obj]) SingleCmd(addr Obj, c Cmd) DataUnit[Orig, Com, Obj] {
	u := p.cmd(addr, c)
	u.Type = C_SC_NA_1
	return u
}

// DoubleCmd sets Type, Enc,  and Info with double command C_DC_NA_1 Act
// conform companion standard 101, subsection 7.3.2.2.
func (p Params[Orig, Com, Obj]) DoubleCmd(addr Obj, c Cmd) DataUnit[Orig, Com, Obj] {
	u := p.cmd(addr, c)
	u.Type = C_DC_NA_1
	return u
}

// RegulCmd sets Type, Enc,  and Info with regulating-step command
// C_RC_NA_1 Act conform companion standard 101, subsection 7.3.2.3.
func (p Params[Orig, Com, Obj]) RegulCmd(addr Obj, c Cmd) DataUnit[Orig, Com, Obj] {
	u := p.cmd(addr, c)
	u.Type = C_RC_NA_1
	return u
}

func (p Params[Orig, Com, Obj]) cmd(addr Obj, c Cmd) DataUnit[Orig, Com, Obj] {
	u := p.NewDataUnit()
	u.Enc = 1     // fixed
	u.Cause = Act // could Deact
	u.setAddr(addr)
	u.Info = append(u.Info, byte(c))
	return u
}

// NormSetPt sets Type, Enc,  and Info with a set-point command C_SE_NA_1 Act
// conform companion standard 101, subsection 7.3.2.4.
func (p Params[Orig, Com, Obj]) NormSetPt(addr Obj, value Norm, q SetPtQual) DataUnit[Orig, Com, Obj] {
	u := p.NewDataUnit()
	u.Type = C_SE_NA_1
	u.Enc = 1     // fixed
	u.Cause = Act // could Deact
	u.setAddr(addr)
	u.Info = append(u.Info, value[0], value[1])
	u.Info = append(u.Info, byte(q))
	return u
}

// ScaledSetPt sets Type, Enc,  and Info with set-point command C_SE_NB_1 Act
// conform companion standard 101, subsection 7.3.2.5.
func (p Params[Orig, Com, Obj]) ScaledSetPt(addr Obj, value int16, q SetPtQual) DataUnit[Orig, Com, Obj] {
	u := p.NewDataUnit()
	u.Type = C_SE_NB_1
	u.Enc = 1     // fixed
	u.Cause = Act // could Deact
	u.setAddr(addr)
	u.Info = append(u.Info, byte(value), byte(value>>8))
	u.Info = append(u.Info, byte(q))
	return u
}

// FloatSetPt sets Type, Enc,  and Info with set-point command C_SE_NC_1 Act
// conform companion standard 101, subsection 7.3.2.6.
func (p Params[Orig, Com, Obj]) FloatSetPt(addr Obj, value float32, q SetPtQual) DataUnit[Orig, Com, Obj] {
	u := p.NewDataUnit()
	u.Type = C_SE_NC_1
	u.Enc = 1     // fixed
	u.Cause = Act // could Deact
	u.setAddr(addr)
	u.Info = binary.LittleEndian.AppendUint32(u.Info,
		math.Float32bits(value))
	u.Info = append(u.Info, byte(q))
	return u
}

// Inro sets Type, Enc,  and Info with interrogation-command C_IC_NA_1 Act
// conform companion standard 101, subsection 7.3.4.1.
func (p Params[Orig, Com, Obj]) Inro() DataUnit[Orig, Com, Obj] {
	return p.InroGroup(0)
}

// InroGroup sets Type, Enc,  and Info with interrogation-command C_IC_NA_1
// Act conform companion standard 101, subsection 7.3.4.1. Group can be disabled
// with 0 for (global) station interrogation. Otherwise, use a group identifier
// in range [1..16].
func (p Params[Orig, Com, Obj]) InroGroup(group uint) DataUnit[Orig, Com, Obj] {
	u := p.NewDataUnit()
	u.Type = C_IC_NA_1
	u.Enc = 1     // fixed
	u.Cause = Act // could Deact
	var addr Obj  // fixed to zero
	u.setAddr(addr)

	// qualifier of interrogation is described in
	// companion standard 101, section 7.2.6.22
	u.Info = append(u.Info, byte(group+20))
	return u
}

// TestCmd sets Type, Enc,  and Info with test-command C_TS_NA_1 Act
// conform companion standard 101, subsection 7.3.4.5.
func (p Params[Orig, Com, Obj]) TestCmd() DataUnit[Orig, Com, Obj] {
	u := p.NewDataUnit()
	u.Type = C_TS_NA_1
	u.Enc = 1     // fixed
	u.Cause = Act // fixed
	var addr Obj  // fixed to zero
	u.setAddr(addr)
	// fixed test bit pattern defined in companion
	// standard 101, subsection 7.2.6.14
	u.Info = append(u.Info, 0b1010_1010, 0b0101_0101)
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
	u.Info = asdu[3+len(u.Orig)+len(u.Addr):]
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
