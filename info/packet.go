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

// Cause Of Transmission Flags
const (
	// The P/N flag negates confirmation when set.
	// In case of irrelevance, the bit should be zero.
	negFlag = 0x40

	// The T flag classifies the unit for testing when set.
	testFlag = 0x80
)

// Cause of transmission codes a justification for the emission. The 2-octet
// version can include an originator address.
// See companion standard 101, subclause 7.2.3 for a full description.
type (
	// A COT can be instantiated with COTOf or COTOrigOf from Params.
	COT interface {
		// The width is fixed per system.
		COT8 | COT16

		// Cause returns the 6-bit code.
		Cause() Cause

		// The P/N flag negates confirmation.
		IsNeg() bool

		// The T flag classifies the unit for testing.
		IsTest() bool

		// The number of the origanator address defaults to zero.
		// Value 1..255 addresses a specific part of the system.
		OrigAddr() uint8
	}

	COT8  [1]uint8 // without originator address
	COT16 [2]uint8 // with originator address
)

func (p Params[T, Com, Obj]) COTOf(c Cause) T {
	var addr T
	addr[0] = uint8(c)
	return addr
}

// COTOrigOf sets the cause of transmission with an origin address.
// The return is false when orig is non-zero with COT8.
func (p Params[T, Com, Obj]) COTOrigOf(c Cause, orig uint) (T, bool) {
	var addr T
	addr[0] = uint8(c)
	if len(addr) < 2 {
		return addr, orig == 0
	}
	// workaround limited generics
	if i := 1; i < len(addr) {
		addr[i] = uint8(orig)
	}
	return addr, true
}

// MustCOTOrigOf is like COTOrigOf, yet it panics on failure.
func (p Params[T, Com, Obj]) MustCOTOrigOf(c Cause, orig uint) T {
	addr, ok := p.COTOrigOf(c, orig)
	if !ok {
		panic("originator address not applicable")
	}
	return addr
}

// Cause implements the COT interface.
func (c COT8) Cause() Cause { return Cause(c[0] & 0x3f) }

// Cause implements the COT interface.
func (c COT16) Cause() Cause { return Cause(c[0] & 0x3f) }

// IsNeg implements the COT interface.
func (c COT8) IsNeg() bool { return c[0]&negFlag != 0 }

// IsNeg implements the COT interface.
func (c COT16) IsNeg() bool { return c[0]&negFlag != 0 }

// IsTest implements the COT interface.
func (c COT8) IsTest() bool { return c[0]&testFlag != 0 }

// IsTest implements the COT interface.
func (c COT16) IsTest() bool { return c[0]&testFlag != 0 }

// OrigAddr implements the COT interface. The return is always zero.
func (c COT8) OrigAddr() uint8 { return 0 }

// OrigAddr implements the COT interface.
func (c COT16) OrigAddr() uint8 { return c[1] }

// A DataUnit has a transmission packet called ASDU, short for Application
// Service Data Unit. Information objects are encoded conform the TypeID and
// and Enc values. Encoding is likely to contain one or more Object addresses.
type DataUnit[T COT, Com ComAddr, Obj ObjAddr] struct {
	// configuration inherited from generics
	Params[T, Com, Obj]

	Type TypeID // payload class
	Enc  Enc    // payload structure
	COT  T      // cause of transmission
	Addr Com    // station address

	// The previous causes 2 to 4 bytes of padding.
	bootstrap [10]byte

	Info []byte // encoded payload
}

func (u *DataUnit[T, Com, Obj]) setAddr(addr Obj) {
	u.Info = u.bootstrap[:0]
	u.addAddr(addr)
}

func (u *DataUnit[T, Com, Obj]) addAddr(addr Obj) {
	for i := 0; i < len(addr); i++ {
		u.Info = append(u.Info, addr[i])
	}
}

// NewDataUnit returns a new ASDU.
func (p Params[T, Com, Obj]) NewDataUnit() DataUnit[T, Com, Obj] {
	return DataUnit[T, Com, Obj]{}
}

// SingleCmd sets Type, Enc, COT and Info with single command C_SC_NA_1 Act
// conform companion standard 101, subsection 7.3.2.1.
func (p Params[T, Com, Obj]) SingleCmd(addr Obj, c Cmd) DataUnit[T, Com, Obj] {
	u := p.cmd(addr, c)
	u.Type = C_SC_NA_1
	return u
}

// DoubleCmd sets Type, Enc, COT and Info with double command C_DC_NA_1 Act
// conform companion standard 101, subsection 7.3.2.2.
func (p Params[T, Com, Obj]) DoubleCmd(addr Obj, c Cmd) DataUnit[T, Com, Obj] {
	u := p.cmd(addr, c)
	u.Type = C_DC_NA_1
	return u
}

// RegulCmd sets Type, Enc, COT and Info with regulating-step command
// C_RC_NA_1 Act conform companion standard 101, subsection 7.3.2.3.
func (p Params[T, Com, Obj]) RegulCmd(addr Obj, c Cmd) DataUnit[T, Com, Obj] {
	u := p.cmd(addr, c)
	u.Type = C_RC_NA_1
	return u
}

func (p Params[T, Com, Obj]) cmd(addr Obj, c Cmd) DataUnit[T, Com, Obj] {
	u := p.NewDataUnit()
	u.Enc = 1            // fixed
	u.COT = p.COTOf(Act) // could Deact
	u.setAddr(addr)
	u.Info = append(u.Info, byte(c))
	return u
}

// NormSetPt sets Type, Enc, COT and Info with a set-point command C_SE_NA_1 Act
// conform companion standard 101, subsection 7.3.2.4.
func (p Params[T, Com, Obj]) NormSetPt(addr Obj, value Norm, q SetPtQual) DataUnit[T, Com, Obj] {
	u := p.NewDataUnit()
	u.Type = C_SE_NA_1
	u.Enc = 1            // fixed
	u.COT = p.COTOf(Act) // could Deact
	u.setAddr(addr)
	u.Info = append(u.Info, value[0], value[1])
	u.Info = append(u.Info, byte(q))
	return u
}

// ScaledSetPt sets Type, Enc, COT and Info with set-point command C_SE_NB_1 Act
// conform companion standard 101, subsection 7.3.2.5.
func (p Params[T, Com, Obj]) ScaledSetPt(addr Obj, value int16, q SetPtQual) DataUnit[T, Com, Obj] {
	u := p.NewDataUnit()
	u.Type = C_SE_NB_1
	u.Enc = 1            // fixed
	u.COT = p.COTOf(Act) // could Deact
	u.setAddr(addr)
	u.Info = append(u.Info, byte(value), byte(value>>8))
	u.Info = append(u.Info, byte(q))
	return u
}

// FloatSetPt sets Type, Enc, COT and Info with set-point command C_SE_NC_1 Act
// conform companion standard 101, subsection 7.3.2.6.
func (p Params[T, Com, Obj]) FloatSetPt(addr Obj, value float32, q SetPtQual) DataUnit[T, Com, Obj] {
	u := p.NewDataUnit()
	u.Type = C_SE_NC_1
	u.Enc = 1            // fixed
	u.COT = p.COTOf(Act) // could Deact
	u.setAddr(addr)
	u.Info = binary.LittleEndian.AppendUint32(u.Info,
		math.Float32bits(value))
	u.Info = append(u.Info, byte(q))
	return u
}

// Inro sets Type, Enc, COT and Info with interrogation-command C_IC_NA_1 Act
// conform companion standard 101, subsection 7.3.4.1.
func (p Params[T, Com, Obj]) Inro() DataUnit[T, Com, Obj] {
	return p.InroGroup(0)
}

// InroGroup sets Type, Enc, COT and Info with interrogation-command C_IC_NA_1
// Act conform companion standard 101, subsection 7.3.4.1. Group can be disabled
// with 0 for (global) station interrogation. Otherwise, use a group identifier
// in range [1..16].
func (p Params[T, Com, Obj]) InroGroup(group uint) DataUnit[T, Com, Obj] {
	u := p.NewDataUnit()
	u.Type = C_IC_NA_1
	u.Enc = 1            // fixed
	u.COT = p.COTOf(Act) // could Deact
	var addr Obj         // fixed to zero
	u.setAddr(addr)

	// qualifier of interrogation is described in
	// companion standard 101, section 7.2.6.22
	u.Info = append(u.Info, byte(group+20))
	return u
}

// TestCmd sets Type, Enc, COT and Info with test-command C_TS_NA_1 Act
// conform companion standard 101, subsection 7.3.4.5.
func (p Params[T, Com, Obj]) TestCmd() DataUnit[T, Com, Obj] {
	u := p.NewDataUnit()
	u.Type = C_TS_NA_1
	u.Enc = 1            // fixed
	u.COT = p.COTOf(Act) // fixed
	var addr Obj         // fixed to zero
	u.setAddr(addr)
	// fixed test bit pattern defined in companion
	// standard 101, subsection 7.2.6.14
	u.Info = append(u.Info, 0b1010_1010, 0b0101_0101)
	return u
}

// Adopt reads the Data Unit Identifier from the ASDU into the fields.
// The remainder of the bytes is sliced as Info without any validation.
func (u *DataUnit[T, Com, Obj]) Adopt(asdu []byte) error {
	if len(asdu) < 2+len(u.COT)+len(u.Addr) {
		if len(asdu) == 0 {
			return io.EOF
		}
		return io.ErrUnexpectedEOF
	}

	// copy header
	u.Type = TypeID(asdu[0])
	u.Enc = Enc(asdu[1])
	for i := 0; i < len(u.COT); i++ {
		u.COT[i] = asdu[i+2]
	}
	for i := 0; i < len(u.Addr); i++ {
		u.Addr[i] = asdu[i+2+len(u.COT)]
	}

	// reject values whom are "not used"
	switch {
	case u.Type == 0:
		return errTypeZero
	case u.COT.Cause() == 0:
		return errCauseZero
	case u.Addr.N() == 0:
		return errComAddrZero
	}

	// slice payload
	u.Info = asdu[2+len(u.COT)+len(u.Addr):]
	return nil
}

// Append the ASDU encoding to buf and return the extended buffer.
func (u DataUnit[T, Com, Obj]) Append(buf []byte) []byte {
	buf = append(buf, byte(u.Type), byte(u.Enc))
	for i := 0; i < len(u.COT); i++ {
		buf = append(buf, u.COT[i])
	}
	for i := 0; i < len(u.Addr); i++ {
		buf = append(buf, u.Addr[i])
	}
	buf = append(buf, u.Info...)
	return buf
}
