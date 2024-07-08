package info

import (
	"bytes"
	"fmt"
	"io"
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
	Info  []byte // encoded payload
}

// NewDataUnit returns an ASDU directed to a common address.
func (p Params[COT, Common, Object]) NewDataUnit(addr Common) DataUnit[COT, Common, Object] {
	return DataUnit[COT, Common, Object]{Addr: addr}
}

// InroAct sets an interrogation [C_IC_NA_1] activation request.
// Use either range [1..16], or 0 for station interrogation (global).
//
// The Common address and the originator address (if any) both
// remain unchanged. TestFlag can be set afterwards when applicable.
func (u *DataUnit[COT, Common, Object]) InroActGroup(group uint) {
	u.inroGroup(group)
	u.Cause[0] = Act
}

// InroDeact sets an interrogation [C_IC_NA_1] deactivation request.
// Use either range [1..16], or 0 for station interrogation (global).
//
// The (common) address and the originator address (if any) both
// remain unchanged. TestFlag can be set afterwards when applicable.
func (u *DataUnit[COT, Common, Object]) InroDeactGroup(group uint) {
	u.inroGroup(group)
	u.Cause[0] = Deact
}

func (u *DataUnit[COT, Common, Object]) inroGroup(group uint) {
	u.Type = C_IC_NA_1
	u.Var = 0x01

	// information object address is fixed to zero
	var addr Object
	if cap(u.Info) < len(addr)+1 {
		u.Info = make([]byte, len(addr), len(addr)+1)
	} else {
		u.Info = append(u.Info[:len(addr)])
		for i := range u.Info {
			u.Info[i] = 0
		}
	}

	// qualifier of interrogation is described in
	// companion standard 101, section 7.2.6.22
	u.Info = append(u.Info, byte(group+20))
}

// Respond returns a new "responding" ASDU which addresses "initiating" ASDU u.
// The response includes the TestFlag from u.
func (u DataUnit[COT, Common, Object]) Respond(t TypeID, cause uint8) DataUnit[COT, Common, Object] {
	// u is a copy allready—no pointer receiver
	u.Type = t
	u.Cause[0] &= TestFlag
	u.Cause[0] |= cause
	u.Info = nil
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
