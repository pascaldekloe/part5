package info

import (
	"encoding/binary"
	"io"
	"math"
)

// Variable Structure Qualifier Flags
const (
	// The SQ flag signals that the information applies to a consecutive
	// sequence of addresses, in which each information-object address is
	// one higher than the previous element.
	Sequence = 0x80
)

// Var is the variable structure qualifier, which defines the payload
// layout with an object Count and a Sequence flag.
type Var uint8

// Count returns the number of information objects, which can be zero.
func (v Var) Count() int { return int(v & 0x7f) }

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

func (p Params[COT, Common, Object]) COTOf(c Cause) COT {
	var addr COT
	addr[0] = uint8(c)
	return addr
}

// COTOrigOf sets the cause of transmission with an origin address.
// The return is false when orig is non-zero with COT8.
func (p Params[COT, Common, Object]) COTOrigOf(c Cause, orig uint) (COT, bool) {
	var addr COT
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
func (p Params[COT, Common, Object]) MustCOTOrigOf(c Cause, orig uint) COT {
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
// and Var values. Encoding is likely to contain one or more Object addresses.
type DataUnit[C COT, Common ComAddr, Object Addr] struct {
	// configuration inherited from generics
	Params[C, Common, Object]

	Type TypeID // payload class
	Var  Var    // payload layout
	COT  C      // cause of transmission
	Addr Common // station address

	// The previous causes 2 to 4 bytes of padding.
	bootstrap [10]byte

	Info []byte // encoded payload
}

func (u *DataUnit[COT, Common, Object]) setAddr(addr Object) {
	u.Info = u.bootstrap[:0]
	u.addAddr(addr)
}

func (u *DataUnit[COT, Common, Object]) addAddr(addr Object) {
	for i := 0; i < len(addr); i++ {
		u.Info = append(u.Info, addr[i])
	}
}

// NewDataUnit returns a new ASDU.
func (p Params[COT, Common, Object]) NewDataUnit() DataUnit[COT, Common, Object] {
	return DataUnit[COT, Common, Object]{}
}

// SingleCmd sets Type, Var, COT and Info with single command C_SC_NA_1 Act
// conform companion standard 101, subsection 7.3.2.1.
func (p Params[COT, Common, Object]) SingleCmd(addr Object, c Cmd) DataUnit[COT, Common, Object] {
	u := p.cmd(addr, c)
	u.Type = C_SC_NA_1
	return u
}

// DoubleCmd sets Type, Var, COT and Info with double command C_DC_NA_1 Act
// conform companion standard 101, subsection 7.3.2.2.
func (p Params[COT, Common, Object]) DoubleCmd(addr Object, c Cmd) DataUnit[COT, Common, Object] {
	u := p.cmd(addr, c)
	u.Type = C_DC_NA_1
	return u
}

// RegulCmd sets Type, Var, COT and Info with regulating-step command
// C_RC_NA_1 Act conform companion standard 101, subsection 7.3.2.3.
func (p Params[COT, Common, Object]) RegulCmd(addr Object, c Cmd) DataUnit[COT, Common, Object] {
	u := p.cmd(addr, c)
	u.Type = C_RC_NA_1
	return u
}

func (p Params[COT, Common, Object]) cmd(addr Object, c Cmd) DataUnit[COT, Common, Object] {
	u := p.NewDataUnit()
	u.Var = 1            // fixed
	u.COT = p.COTOf(Act) // could Deact
	u.setAddr(addr)
	u.Info = append(u.Info, byte(c))
	return u
}

// NormSetPt sets Type, Var, COT and Info with a set-point command C_SE_NA_1 Act
// conform companion standard 101, subsection 7.3.2.4.
func (p Params[COT, Common, Object]) NormSetPt(addr Object, value Norm, q SetPtQual) DataUnit[COT, Common, Object] {
	u := p.NewDataUnit()
	u.Type = C_SE_NA_1
	u.Var = 1            // fixed
	u.COT = p.COTOf(Act) // could Deact
	u.setAddr(addr)
	u.Info = append(u.Info, value[0], value[1])
	u.Info = append(u.Info, byte(q))
	return u
}

// ScaledSetPt sets Type, Var, COT and Info with set-point command C_SE_NB_1 Act
// conform companion standard 101, subsection 7.3.2.5.
func (p Params[COT, Common, Object]) ScaledSetPt(addr Object, value int16, q SetPtQual) DataUnit[COT, Common, Object] {
	u := p.NewDataUnit()
	u.Type = C_SE_NB_1
	u.Var = 1            // fixed
	u.COT = p.COTOf(Act) // could Deact
	u.setAddr(addr)
	u.Info = append(u.Info, byte(value), byte(value>>8))
	u.Info = append(u.Info, byte(q))
	return u
}

// FloatSetPt sets Type, Var, COT and Info with set-point command C_SE_NC_1 Act
// conform companion standard 101, subsection 7.3.2.6.
func (p Params[COT, Common, Object]) FloatSetPt(addr Object, value float32, q SetPtQual) DataUnit[COT, Common, Object] {
	u := p.NewDataUnit()
	u.Type = C_SE_NC_1
	u.Var = 1            // fixed
	u.COT = p.COTOf(Act) // could Deact
	u.setAddr(addr)
	u.Info = binary.LittleEndian.AppendUint32(u.Info,
		math.Float32bits(value))
	u.Info = append(u.Info, byte(q))
	return u
}

// Inro sets Type, Var, COT and Info with interrogation-command C_IC_NA_1 Act
// conform companion standard 101, subsection 7.3.4.1.
func (p Params[COT, Common, Object]) Inro() DataUnit[COT, Common, Object] {
	return p.InroGroup(0)
}

// InroGroup sets Type, Var, COT and Info with interrogation-command C_IC_NA_1
// Act conform companion standard 101, subsection 7.3.4.1. Group can be disabled
// with 0 for (global) station interrogation. Otherwise, use a group identifier
// in range [1..16].
func (p Params[COT, Common, Object]) InroGroup(group uint) DataUnit[COT, Common, Object] {
	u := p.NewDataUnit()
	u.Type = C_IC_NA_1
	u.Var = 1            // fixed
	u.COT = p.COTOf(Act) // could Deact
	var addr Object      // fixed to zero
	u.setAddr(addr)

	// qualifier of interrogation is described in
	// companion standard 101, section 7.2.6.22
	u.Info = append(u.Info, byte(group+20))
	return u
}

// TestCmd sets Type, Var, COT and Info with test-command C_TS_NA_1 Act
// conform companion standard 101, subsection 7.3.4.5.
func (p Params[COT, Common, Object]) TestCmd() DataUnit[COT, Common, Object] {
	u := p.NewDataUnit()
	u.Type = C_TS_NA_1
	u.Var = 1            // fixed
	u.COT = p.COTOf(Act) // fixed
	var addr Object      // fixed to zero
	u.setAddr(addr)
	// fixed test bit pattern defined in companion
	// standard 101, subsection 7.2.6.14
	u.Info = append(u.Info, 0b1010_1010, 0b0101_0101)
	return u
}

// Adopt reads the Data Unit Identifier from the ASDU into the fields.
// The remainder of the bytes is sliced as Info without any validation.
func (u *DataUnit[COT, Common, Object]) Adopt(asdu []byte) error {
	if len(asdu) < 2+len(u.COT)+len(u.Addr) {
		if len(asdu) == 0 {
			return io.EOF
		}
		return io.ErrUnexpectedEOF
	}

	// copy header
	u.Type = TypeID(asdu[0])
	u.Var = Var(asdu[1])
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

// Append the ASDU encoding to buf.
func (u DataUnit[COT, Common, Object]) Append(buf []byte) []byte {
	buf = append(buf, byte(u.Type), byte(u.Var))
	for i := 0; i < len(u.COT); i++ {
		buf = append(buf, u.COT[i])
	}
	for i := 0; i < len(u.Addr); i++ {
		buf = append(buf, u.Addr[i])
	}
	buf = append(buf, u.Info...)
	return buf
}
