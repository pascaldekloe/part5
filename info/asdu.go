package info

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// Var is the variable structure qualifier, which defines the payload
// layout with an object Count and a Sequence flag.
type Var uint8

// Variable Structure Qualifier Flags
const (
	// The SQ flag signals that the information applies to a consecutive
	// sequence of addresses, in which each information-object address is
	// one higher than the previous element.
	Sequence = 0x80
)

// Count returns the number of information objects, which can be zero.
func (v Var) Count() int { return int(v & 0x7f) }

// A DataUnit has a transmission packet called ASDU, short for Application
// Service Data Unit. Information objects are encoded conform the TypeID and
// and Var values. Encoding is likely to contain one or more Object addresses.
type DataUnit[COT Cause, Common ComAddr, Object Addr] struct {
	// configuration inherited from generics
	Params[COT, Common, Object]

	Type  TypeID // payload class
	Var   Var    // payload layout
	Cause COT    // cause of transmission
	Addr  Common // station address

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

// SingleCmd sets Type, Var, Cause and Info with single command C_SC_NA_1 Act
// conform companion standard 101, subsection 7.3.2.1.
func (p Params[COT, Common, Object]) SingleCmd(addr Object, c Cmd) DataUnit[COT, Common, Object] {
	u := p.cmd(addr, c)
	u.Type = C_SC_NA_1
	return u
}

// DoubleCmd sets Type, Var, Cause and Info with double command C_DC_NA_1 Oct
// conform companion standard 101, subsection 7.3.2.2.
func (p Params[COT, Common, Object]) DoubleCmd(addr Object, c Cmd) DataUnit[COT, Common, Object] {
	u := p.cmd(addr, c)
	u.Type = C_DC_NA_1
	return u
}

// RegulCmd sets Type, Var, Cause and Info with regulating-step command
// C_RC_NA_1 Act conform companion standard 101, subsection 7.3.2.3.
func (p Params[COT, Common, Object]) RegulCmd(addr Object, c Cmd) DataUnit[COT, Common, Object] {
	u := p.cmd(addr, c)
	u.Type = C_RC_NA_1
	return u
}

func (p Params[COT, Common, Object]) cmd(addr Object, c Cmd) DataUnit[COT, Common, Object] {
	u := p.NewDataUnit()
	u.Var = 1        // fixed
	u.Cause[0] = Act // could Deact
	u.setAddr(addr)
	u.Info = append(u.Info, byte(c))
	return u
}

// NormSetPt sets Type, Var, Cause and Info with a set-point command C_SE_NA_1
// Act conform companion standard 101, subsection 7.3.2.4.
func (p Params[COT, Common, Object]) NormSetPt(addr Object, value Norm, q SetPtQual) DataUnit[COT, Common, Object] {
	u := p.NewDataUnit()
	u.Type = C_SE_NA_1
	u.Var = 1        // fixed
	u.Cause[0] = Act // could Deact
	u.setAddr(addr)
	u.Info = append(u.Info, value[0], value[1])
	u.Info = append(u.Info, byte(q))
	return u
}

// ScaledSetPt sets Type, Var, Cause and Info with set-point command C_SE_NB_1
// Act conform companion standard 101, subsection 7.3.2.5.
func (p Params[COT, Common, Object]) ScaledSetPt(addr Object, value int16, q SetPtQual) DataUnit[COT, Common, Object] {
	u := p.NewDataUnit()
	u.Type = C_SE_NB_1
	u.Var = 1        // fixed
	u.Cause[0] = Act // could Deact
	u.setAddr(addr)
	u.Info = append(u.Info, byte(value), byte(value>>8))
	u.Info = append(u.Info, byte(q))
	return u
}

// FloatSetPt sets Type, Var, Cause and Info with set-point command C_SE_NC_1
// Act conform companion standard 101, subsection 7.3.2.6.
func (p Params[COT, Common, Object]) FloatSetPt(addr Object, value float32, q SetPtQual) DataUnit[COT, Common, Object] {
	u := p.NewDataUnit()
	u.Type = C_SE_NC_1
	u.Var = 1        // fixed
	u.Cause[0] = Act // could Deact
	u.setAddr(addr)
	u.Info = binary.LittleEndian.AppendUint32(u.Info,
		math.Float32bits(value))
	u.Info = append(u.Info, byte(q))
	return u
}

// Inro sets Type, Var, Cause and Info with interrogation-command C_IC_NA_1 Act
// conform companion standard 101, subsection 7.3.4.1.
func (p Params[COT, Common, Object]) Inro() DataUnit[COT, Common, Object] {
	return p.InroGroup(0)
}

// InroGroup sets Type, Var, Cause and Info with interrogation-command C_IC_NA_1
// Act conform companion standard 101, subsection 7.3.4.1. Group can be disabled
// with 0 for (global) station interrogation. Otherwise, use a group identifier
// in range [1..16].
func (p Params[COT, Common, Object]) InroGroup(group uint) DataUnit[COT, Common, Object] {
	u := p.NewDataUnit()
	u.Type = C_IC_NA_1
	u.Var = 1        // fixed
	u.Cause[0] = Act // could Deact
	var addr Object  // fixed to zero
	u.setAddr(addr)

	// qualifier of interrogation is described in
	// companion standard 101, section 7.2.6.22
	u.Info = append(u.Info, byte(group+20))
	return u
}

// TestCmd sets Type, Var, Cause and Info with test-command C_TS_NA_1 Act
// conform companion standard 101, subsection 7.3.4.5.
func (p Params[COT, Common, Object]) TestCmd() DataUnit[COT, Common, Object] {
	u := p.NewDataUnit()
	u.Type = C_TS_NA_1
	u.Var = 1        // fixed
	u.Cause[0] = Act // fixed
	var addr Object  // fixed to zero
	u.setAddr(addr)
	// fixed test bit pattern defined in companion
	// standard 101, subsection 7.2.6.14
	u.Info = append(u.Info, 0b1010_1010, 0b0101_0101)
	return u
}

// Adopt reads the Data Unit Identifier from the ASDU into the fields.
// The remainder of the bytes is sliced as Info without any validation.
func (u *DataUnit[COT, Common, Object]) Adopt(asdu []byte) error {
	if len(asdu) < 2+len(u.Cause)+len(u.Addr) {
		if len(asdu) == 0 {
			return io.EOF
		}
		return io.ErrUnexpectedEOF
	}

	// copy header
	u.Type = TypeID(asdu[0])
	u.Var = Var(asdu[1])
	for i := 0; i < len(u.Cause); i++ {
		u.Cause[i] = asdu[i+2]
	}
	for i := 0; i < len(u.Addr); i++ {
		u.Addr[i] = asdu[i+2+len(u.Cause)]
	}

	// reject values whom are "not used"
	switch {
	case u.Type == 0:
		return errTypeZero
	case u.Cause.Value()&^(NegFlag|TestFlag) == 0:
		return errCauseZero
	case u.Addr.N() == 0:
		return errComAddrZero
	}

	// slice payload
	u.Info = asdu[2+len(u.Cause)+len(u.Cause):]
	return nil
}

// Append the ASDU encoding to buf.
func (u DataUnit[COT, Common, Object]) Append(buf []byte) []byte {
	buf = append(buf, byte(u.Type), byte(u.Var))
	for i := 0; i < len(u.Cause); i++ {
		buf = append(buf, u.Cause[i])
	}
	for i := 0; i < len(u.Addr); i++ {
		buf = append(buf, u.Addr[i])
	}
	buf = append(buf, u.Info...)
	return buf
}

// String returns a full description.
func (u DataUnit[COT, Common, Object]) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s @%d %s:",
		u.Type, u.Addr.N(), u.Cause)

	dataSize := ObjSize[u.Type]
	switch {
	case dataSize == 0:
		// structure unknown
		fmt.Fprintf(&buf, " %#x ?", u.Info)

	case u.Var&Sequence == 0:
		// objects paired with an address each
		var i int // read index
		for obj_n := 0; ; obj_n++ {
			var addr Object
			if i+len(addr)+dataSize > len(u.Info) {
				if i < len(u.Info) {
					fmt.Fprintf(&buf, " %#x<EOF>", u.Info[i:])
					obj_n++
				}

				diff := obj_n - u.Var.Count()
				switch {
				case diff < 0:
					fmt.Fprintf(&buf, " 𝚫 −%d !", -diff)
				case diff > 0:
					fmt.Fprintf(&buf, " 𝚫 +%d !", diff)
				case i < len(u.Info):
					buf.WriteString(" !")
				default:
					buf.WriteString(" .")
				}

				break
			}

			for j := 0; j < len(addr); j++ {
				addr[j] = u.Info[i+j]
			}
			i += len(addr)
			fmt.Fprintf(&buf, " %d:%#x",
				addr.N(), u.Info[i:i+dataSize])
			i += dataSize
		}

	default:
		// offset address in Sequence
		var addr Object
		if len(addr) > len(u.Info) {
			fmt.Fprintf(&buf, " %#x<EOF> !", u.Info)
			break
		}
		for i := 0; i < len(addr); i++ {
			addr[i] = u.Info[i]
		}
		i := len(addr) // read index

		for obj_n := 0; ; obj_n++ {
			if i+dataSize > len(u.Info) {
				if i < len(u.Info) {
					fmt.Fprintf(&buf, " %d:%#x<EOF>",
						addr.N(), u.Info[i:])
					obj_n++
				}

				diff := obj_n - u.Var.Count()
				switch {
				case diff < 0:
					fmt.Fprintf(&buf, " 𝚫 −%d !", -diff)
				case diff > 0:
					fmt.Fprintf(&buf, " 𝚫 +%d !", diff)
				case i < len(u.Info):
					buf.WriteString(" !")
				default:
					buf.WriteString(" .")
				}

				break
			}

			fmt.Fprintf(&buf, " %d:%#x",
				addr.N(), u.Info[i:i+dataSize])
			i += dataSize
			addr, _ = u.NewAddr(addr.N() + 1)
		}
	}

	return buf.String()
}
