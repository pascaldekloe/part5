// Package info provides the OSI presentation layer.
//
// The originator addresss, the common address and and the information-object
// address (OrigAddr, ComAddr and ObjAddr) each can be formatted with the fmt
// package. The numeric notation in hexadecimal "%x" and its upper-casing "%X"
// do not omit any leading zeroes. Decimals can be fixed in width too with
// leading space as "% d". The “alternated format” for addresses is an octet
// listing in big-endian order. Decimal octets "%#d" get sepatared by the dot
// character ".", while hexadecimals "%#x" and "%#X" get separated by the colon
// character ":".
package info

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Params define the system-specific parameters. Specifically, the generics set
// the address width, which can be zero [none] for the originator.
type Params[Orig OrigAddr, Com ComAddr, Obj ObjAddr] struct{}

// ErrComAddrZero denies the zero value as an address. Use is explicitly
// prohibited in subsection 7.2.4.
var errComAddrZero = errors.New("part5: common address <0> is not used")

// The “common address” addresses stations. Zero is not used.
// All information objects/addresses reside in a common address.
// See companion standard 101, subclause 7.2.4.
type (
	// A ComAddr can be instantiated with ComAddrN of Params.
	ComAddr interface {
		// The address width is fixed per system.
		ComAddr8 | ComAddr16

		// N gets the address as a numeric value.
		N() uint

		// The global address is a broadcast address,
		// directed to all stations of a specific system.
		// Use is restricted to type C_IC_NA_1, C_CI_NA_1,
		// C_CS_NA_1 and C_RP_NA_1.
		Global() bool
	}

	ComAddr8  [1]uint8
	ComAddr16 [2]uint8
)

// ComAddrN returns either a numeric match, or false when n overflows the
// address width of Com.
func (p Params[Orig, Com, Obj]) ComAddrN(n uint) (Com, bool) {
	var addr Com
	for i := 0; i < len(addr); i++ {
		addr[i] = uint8(n)
		n >>= 8
	}
	return addr, n == 0
}

// MustComAddrN is like ComAddrN, yet it panics on overflow.
func (p Params[Orig, Com, Obj]) MustComAddrN(n uint) Com {
	addr, ok := p.ComAddrN(n)
	if !ok {
		panic("overflow of common address")
	}
	return addr
}

// N implements the ComAddr interface.
func (addr ComAddr8) N() uint {
	return uint(addr[0])
}

// N implements the ComAddr interface.
func (addr ComAddr16) N() uint {
	return uint(addr[0]) | uint(addr[1])<<8
}

// Global implements the ComAddr interface.
func (addr ComAddr8) Global() bool {
	return addr[0] == 0xff
}

// Global implements the ComAddr interface.
func (addr ComAddr16) Global() bool {
	return addr[0] == 0xff && addr[1] == 0xff
}

// LowOctet provides an alternative to numeric addressing.
func (addr ComAddr16) LowOctet() uint8 { return addr[0] }

// HighOctet provides an alternative to numeric addressing.
func (addr ComAddr16) HighOctet() uint8 { return addr[1] }

// “The information object address is used as a destination address in
// control direction and a source address in the monitor direction.”
// — companion standard 101, subclause 7.2.5.
type (
	// An ObjAddr can be instantiated with ObjAddrN of Params.
	ObjAddr interface {
		// The address width is fixed per system.
		ObjAddr8 | ObjAddr16 | ObjAddr24

		// N gets the address as a numeric value.
		// Zero marks the address as irrelevant.
		N() uint
	}

	ObjAddr8  [1]uint8
	ObjAddr16 [2]uint8
	ObjAddr24 [3]uint8
)

// ObjAddrN returns either a numeric match, or false when n overflows the
// address width of Obj.
func (p Params[Orig, Com, Obj]) ObjAddrN(n uint) (Obj, bool) {
	var addr Obj
	for i := 0; i < len(addr); i++ {
		addr[i] = uint8(n)
		n >>= 8
	}
	return addr, n == 0
}

// MustObjAddrN is like ObjAddrN, yet it panics on overflow.
func (p Params[Orig, Com, Obj]) MustObjAddrN(n uint) Obj {
	addr, ok := p.ObjAddrN(n)
	if !ok {
		panic("overflow of information-object")
	}
	return addr
}

// N implements the Addr interface.
func (addr ObjAddr8) N() uint {
	return uint(addr[0])
}

// N implements the Addr interface.
func (addr ObjAddr16) N() uint {
	return uint(addr[0]) | uint(addr[1])<<8
}

// N implements the Addr interface.
func (addr ObjAddr24) N() uint {
	return uint(addr[0]) | uint(addr[1])<<8 | uint(addr[2])<<16
}

// LowOctet provides an alternative to numeric addressing.
func (addr ObjAddr16) LowOctet() uint8 { return addr[0] }

// HighOctet provides an alternative to numeric addressing.
func (addr ObjAddr16) HighOctet() uint8 { return addr[1] }

// LowOctet provides an alternative to numeric addressing.
func (addr ObjAddr24) LowOctet() uint8 { return addr[0] }

// MiddleOctet provides an alternative to numeric addressing.
func (addr ObjAddr24) MiddleOctet() uint8 { return addr[1] }

// HighOctet provides an alternative to numeric addressing.
func (addr ObjAddr24) HighOctet() uint8 { return addr[2] }

// SinglePt has single-point information (IEV 371-02-07), which should be either
// Off or On.
type SinglePt uint8

// Single-Point States
const (
	Off = iota
	On
)

// String returns the label, which is either "Off", "On" or "<illegal>".
func (pt SinglePt) String() string {
	switch pt {
	case Off:
		return "Off"
	case On:
		return "On"
	default:
		return "<illegal>"
	}
}

// SinglePtQual has single-point information (IEV 371-02-07) with a quality
// descriptor. If the octet does not equal Off or On, then it has quality
// remarks.
type SinglePtQual uint8

// Value loses the quality descriptor. The return is either Off or On.
func (pt SinglePtQual) Value() SinglePt { return SinglePt(pt & 1) }

// Qual returns the quality descriptor flags. Overflow is always zero and
// ElapsedTimeInvalid does not apply.
func (pt SinglePtQual) Qual() Qual { return Qual(pt & 0xfe) }

// DoublePt is double-point information (IEV 371-02-08), which should be either
// IndeterminateOrIntermediate, DeterminatedOff, DeterminatedOn or Indeterminate.
// http://blog.iec61850.com/2009/04/why-do-we-need-single-point-and-double.html
type DoublePt uint8

// Double-Point States
const (
	IndeterminateOrIntermediate = iota
	DeterminatedOff
	DeterminatedOn
	Indeterminate
)

// String returns the label which is either "IndeterminateOrIntermediate",
// "DeterminedOff", "DeterminedOn", "Indeterminate" or "<illegal>".
func (pt DoublePt) String() string {
	switch pt {
	case IndeterminateOrIntermediate:
		return "IndeterminateOrIntermediate"
	case DeterminatedOff:
		return "DeterminatedOff"
	case DeterminatedOn:
		return "DeterminatedOn"
	case Indeterminate:
		return "Indeterminate"
	default:
		return "<illegal>"
	}
}

// DoublePtQual has a double-point information-object (IEV 371-02-08)
// with a quality descriptor. If the octet does not equal any of the four
// states, then it has quality remarks.
type DoublePtQual uint8

// Value loses the quality descriptor. The return is either
// IndeterminateOrIntermediate, DeterminatedOff, DeterminatedOn or
// Indeterminate.
func (pt DoublePtQual) Value() DoublePt { return DoublePt(pt & 3) }

// Qual returns the quality descriptor. Overflow is always zero and
// ElapsedTimeInvalid does not apply.
func (pt DoublePtQual) Qual() Qual { return Qual(pt & 0xfc) }

// Regul is a regulating-step command (IEV 371-03-13) state, which should be
// either Lower or Higher. State/value 0 and 3 are not permitted.
type Regul uint8

// Regulating-Step States
const (
	_ = iota // not permitted
	Lower
	Higher
	_ // not permitted
)

// String returns the label, which is either "Lower", "Higher" or "<illegal>".
func (r Regul) String() string {
	switch r {
	case Lower:
		return "Lower"
	case Higher:
		return "Higher"
	default:
		return "<illegal>"
	}
}

// Qual holds flags about an information object, a.k.a. the quality descriptor.
type Qual uint8

// Quality-Descriptor Flags
const (
	// The OV flag signals that the value of the information object is
	// beyond a predefined range of value (mainly applicable to analog
	// values).
	Overflow Qual = 1 << iota

	_ // reserve
	_ // reserve

	// The EI flag signals that the acquisition function recognized abnormal
	// conditions. The elapsed time of the information object is not defined
	// under this condition.
	ElapsedTimeInvalid

	// The BL flag signals that the value of the information object is
	// blocked for transmission. The value remains in the state that was
	// acquired before it was blocked. Blocking and deblocking may be
	// initiated for example by a local lock or a local automatic cause.
	Blocked

	// The SB flag signals that the value of the information object was
	// provided by the input of an operator (dispatcher) instead of an
	// automatic source.
	Substituted

	// The NT flag signals that the most recent update was unsuccessful,
	// i.e., not updated during a specified time interval, or if it is
	// unavailable.
	NotTopical

	// The IV flag signals that the acquisition function recognized abnormal
	// conditions of the information source, i.e, missing or non-operating
	// updating devices. The value of the information object is not defined
	// under this condition.
	Invalid

	// codes from standard
	OV = Overflow
	EI = ElapsedTimeInvalid
	BL = Blocked
	SB = Substituted
	NT = NotTopical
	IV = Invalid

	OK Qual = 0 // no remarks
)

// String returns the codes from the standard, comma separated, with "[2]" and
// "[3]" for the two reserved bits, and "[]" for none.
func (flags Qual) String() string {
	switch flags {
	case OK: // no flags
		return "[]"
	case OV:
		return "OV"
	case EI:
		return "EI"
	case BL:
		return "BL"
	case SB:
		return "SB"
	case NT:
		return "NT"
	case IV:
		return "IV"
	}
	// got multiple flags or reserved flags

	var buf strings.Builder
	if flags&OV != 0 {
		buf.WriteString(",OV")
	}
	if flags&2 != 0 {
		buf.WriteString(",[2]")
	}
	if flags&4 != 0 {
		buf.WriteString(",[3]")
	}
	if flags&EI != 0 {
		buf.WriteString(",EI")
	}
	if flags&BL != 0 {
		buf.WriteString(",BL")
	}
	if flags&SB != 0 {
		buf.WriteString(",SB")
	}
	if flags&NT != 0 {
		buf.WriteString(",NT")
	}
	if flags&IV != 0 {
		buf.WriteString(",IV")
	}
	return buf.String()[1:]
}

// Step contains a step position, which is a numeric value with transient state
// indication. See companion standard 101, subclause 7.2.6.5.
type Step uint8

// NewStep returns a new step position for equipment not in transient state.
// Values out of [-64, 63] overflow silently.
func NewStep(value int) Step { return Step(value & 0x7f) }

// NewTransientStep returns a new step position for equipment in transient state.
// Values out of [-64, 63] overflow silently.
func NewTransientStep(value int) Step { return Step(value&0x7f) | 0x80 }

// Pos returns the value in [-64, 63] plus whether the equipment is transient state.
func (p Step) Pos() (value int, transient bool) {
	value = int(p & 0x7f)
	signBit := value >> 6
	signExt := 0 - signBit
	value |= signExt << 6
	return value, p&0x80 != 0
}

// String gets the decimal value, including a "#T" suffix when transient.
func (p Step) String() string {
	v, t := p.Pos()
	if !t {
		return strconv.Itoa(v)
	}
	return fmt.Sprintf("%d#T", v)
}

// SetpQual is a step position with a quality descriptor.
type StepQual [2]uint8

// NewStepQual returns a new step position for equipment not in transient state.
// Values out of [-64, 63] oveflow silently.
func NewStepQual(value int, flags Qual) StepQual {
	return StepQual{uint8(value & 0x7f), uint8(flags)}
}

// NewTransientStepQual returns a new step position for equipment in transient
// state. Values out of [-64, 63] oveflow silently.
func NewTransientStepQual(value int, flags Qual) StepQual {
	return StepQual{uint8(value | 0x80), uint8(flags)}
}

// Value loses the quality descriptor.
func (p StepQual) Value() Step { return Step(p[0]) }

// Qual returns the quality descriptor. ElapsedTimeInvalid does not apply.
func (p StepQual) Qual() Qual { return Qual(p[1]) }

// FlagQual sets [logical OR] the quality descriptor bits from flags.
func (p *StepQual) FlagQual(flags Qual) {
	p[1] |= uint8(flags)
}

// BitsQual is a four octet bitstring with a quality descriptor.
type BitsQual [5]uint8

// Array gets a copy of the bitstring.
func (b *BitsQual) Array() [4]uint8 { return ([4]byte)(b[:4]) }

// Slice gets a reference to the data.
func (b *BitsQual) Slice() []uint8 { return b[:4] }

// BigEndian returns the bitstring as an integer.
func (b BitsQual) BigEndian() uint32 {
	return binary.BigEndian.Uint32(b[:4])
}

// SetBigEndian updates the bitstring with an integer.
func (b *BitsQual) SetBigEndian(v uint32) {
	binary.BigEndian.PutUint32(b[:4], v)
}

// Qual returns the quality descriptor. ElapsedTimeInvalid does not apply.
func (b BitsQual) Qual() Qual { return Qual(b[4]) }

// FlagQual sets [logical OR] the quality descriptor bits from flags.
func (b *BitsQual) FlagQual(flags Qual) {
	b[4] |= uint8(flags)
}

// Norm is a 16-bit normalized value.
type Norm [2]uint8

// Int16 returns the numeric value as is [unscaled].
func (norm Norm) Int16() int16 {
	return int16(binary.LittleEndian.Uint16(norm[:2]))
}

// SetInt16 updates the normalized value with the numeric value as is
// [unscaled].
func (norm *Norm) SetInt16(v int16) {
	binary.LittleEndian.PutUint16(norm[:2], uint16(v))
}

// Float64 returns the value in range [-1, 1 − 2⁻¹⁵].
func (norm Norm) Float64() float64 {
	return float64(norm.Int16()) / 32768
}

// SetFloat64 updates the value in range [-1, 1 − 2⁻¹⁵].
func (norm *Norm) SetFloat64(v float64) {
	scaled := math.Round(v * 32768)
	switch {
	case scaled <= math.MinInt16:
		norm.SetInt16(math.MinInt16)
	case scaled >= math.MaxInt16:
		norm.SetInt16(math.MaxInt16)
	default:
		norm.SetInt16(int16(scaled))
	}
}

// Norm is a 16-bit normalized value with a quality desciptor.
type NormQual [3]uint8

// Ref returns a reference to the normalized-value withouth the quality
// descriptor.
func (norm *NormQual) Ref() *Norm { return (*Norm)(norm[:2]) }

// Qual returns the quality descriptor. ElapsedTimeInvalid does not apply.
func (norm NormQual) Qual() Qual { return Qual(norm[2]) }

// FlagQual sets [logical OR] the quality descriptor bits from flags.
func (norm *NormQual) FlagQual(flags Qual) {
	norm[2] |= uint8(flags)
}

// ProtEquipStart has flags for start events from protection equipment when it
// detects faults.
type ProtEquipStart uint8

// Start Events Of Protection Equipment
// See companion standard 101, subclause 7.2.6.11.
const (
	GenStartFlag   ProtEquipStart = 1 << iota // GS: general start of operation
	L1StartFlag                               // SL1: start of operation phase L1
	L2StartFlag                               // SL2: start of operation phase L2
	L3StartFlag                               // SL3: start of operation phase L3
	EarthStartFlag                            // SIE: start of operation IE (earth current)
	RevStartFlag                              // SRD: start of operation in reverse direction

	// codes from standard
	GS  = GenStartFlag
	SL1 = L1StartFlag
	SL2 = L2StartFlag
	SL3 = L3StartFlag
	SIE = EarthStartFlag
	SRD = RevStartFlag
)

// String returns the codes from the standard, comma separated, with "[7]" and
// "[8]" for the two reserved bits, and "<>" for none.
func (flags ProtEquipStart) String() string {
	switch flags {
	case 0:
		return "[]"
	case GS:
		return "GS"
	case SL1:
		return "SL1"
	case SL2:
		return "SL2"
	case SL3:
		return "SL3"
	case SIE:
		return "SIE"
	case SRD:
		return "SRD"
	}
	// got multiple flags or reserved flags

	var buf strings.Builder
	if flags&GS != 0 {
		buf.WriteString(",GS")
	}
	if flags&SL1 != 0 {
		buf.WriteString(",SL1")
	}
	if flags&SL2 != 0 {
		buf.WriteString(",SL2")
	}
	if flags&SL3 != 0 {
		buf.WriteString(",SL3")
	}
	if flags&SIE != 0 {
		buf.WriteString(",SIE")
	}
	if flags&SRD != 0 {
		buf.WriteString(",SRD")
	}
	if flags&64 != 0 {
		buf.WriteString(",[7]")
	}
	if flags&128 != 0 {
		buf.WriteString(",[8]")
	}
	return buf.String()[1:]
}

// ProtEquipOut has flags for commands from protection equipment to output the
// circuit information.
type ProtEquipOut uint8

// Output Circuit Information Of Protection Equipment
// See companion standard 101, subclause 7.2.6.12.
const (
	GenOutFlag ProtEquipOut = 1 << iota // GC: general command to output circuit
	L1OutFlag                           // CL1: command to output circuit phase L1
	L2OutFlag                           // CL2: command to output circuit phase L2
	L3OutFlag                           // CL3: command to output circuit phase L3

	// codes from standard
	GC  = GenOutFlag
	CL1 = L1OutFlag
	CL2 = L2OutFlag
	CL3 = L3OutFlag
)

// String returns the codes from the standard, comma separated, with "[5]"
// "[6]", "[7]" and "[8]" for the four reserved bits, and "[]" for none.
func (flags ProtEquipOut) String() string {
	switch flags {
	case 0:
		return "[]"
	case GC:
		return "GC"
	case CL1:
		return "CL1"
	case CL2:
		return "CL2"
	case CL3:
		return "CL3"
	}
	// got multiple flags or reserved flags

	var buf strings.Builder
	if flags&GC != 0 {
		buf.WriteString(",GC")
	}
	if flags&CL1 != 0 {
		buf.WriteString(",CL1")
	}
	if flags&CL2 != 0 {
		buf.WriteString(",CL2")
	}
	if flags&CL3 != 0 {
		buf.WriteString(",CL3")
	}
	if flags&16 != 0 {
		buf.WriteString(",[5]")
	}
	if flags&32 != 0 {
		buf.WriteString(",[6]")
	}
	if flags&64 != 0 {
		buf.WriteString(",[7]")
	}
	if flags&128 != 0 {
		buf.WriteString(",[8]")
	}
	return buf.String()[1:]
}

// ParamQual is the qualifier of parameter of measured values conform companion
// standard 101, subclause 7.2.6.24.
type ParamQual uint8

// Qualifier Of Parameter Of Measured Values
// See companion standard 101, subclause 7.2.6.24.
const (
	_          ParamQual = iota // 0: not used
	Threashold                  // 1: threshold value
	Smoothing                   // 2: smoothing factor (filter time constant)
	LowLimit                    // 3: low limit for transmission of measured values
	HighLimit                   // 4: high limit for transmission of measured values

	// 5..31: reserved for standard definitions of this companion standard (compatible range)
	// 32..63: reserved for special use (private range)

	ChangeFlag  ParamQual = 64  // marks local parameter change
	NotInOpFlag ParamQual = 128 // marks parameter operation

	// codes from standard
	LPC = ChangeFlag
	POP = NotInOpFlag
)

// CmdQual is the qualifier of command conform companion standard 101, subclause
// 7.2.6.26. The zero value executes without any additional definition, i.e.,
// .Select() == false and .Additional() == 0.
type CmdQual uint8

// Additional returns the additional definition in the qualifier of command,
// range 0..31.
//
//	0: no additional definition
//	1: short pulse duration (circuit-breaker), duration determined by a system parameter in the outstation
//	2: long pulse duration, duration determined by a system parameter in the outstation
//	3: persistent output
//	4..8: reserved for standard definitions of the IEC companion standard
//	9..15: reserved for the selection of other predefined functions
//	16..31: reserved for special use (private range)
func (q CmdQual) Additional() uint { return uint((q >> 2) & 31) }

// SetAdditional replaces the additional definition in the qualifier of command.
// Any bits from q values over 31 get discarded silently.
func (q *CmdQual) SetAdditional(a uint) {
	// discard current, if any
	*q &^= 31 << 2
	// trim range, and merge
	*q |= CmdQual((a & 31) << 2)
}

// Select gets the S/E flag, which causes the command to select instead of
// execute. See “Command transmission” in section 5, subclause 6.8.
func (q CmdQual) Select() bool { return q&128 != 0 }

// FlagSelect sets the S/E flag, which causes the command to select instead of
// execute. See “Command transmission” in section 5, subclause 6.8.
func (q *CmdQual) FlagSelect() { *q |= 128 }

// SetPtQual is the qualifier of set-point command conform companion standard
// 101, subclause 7.2.6.39. The zero value exectutes the “default”, i.e.,
// .Select() == false and .N() == 0.
type SetPtQual uint8

// N returns the QL value, range 0..127.
//
//	0: default
//	1..63: reserved for standard definitions of the IEC companion standard (compatible range)
//	64..127: reserved for special use (private range)
func (q SetPtQual) N() uint { return uint(q & 127) }

// SetN resplaces the QL value, range 0..127.
// Any bits from n values over 127 get discarded silently.
func (q *SetPtQual) SetN(n uint) { *q = *q&128 | SetPtQual(n&127) }

// Select gets the S/E flag, which causes the command to select instead of
// execute. See “Command transmission” in section 5, subclause 6.8.
func (q SetPtQual) Select() bool { return q&128 != 0 }

// FlagSelect sets the S/E flag, which causes the command to select instead of
// execute. See “Command transmission” in section 5, subclause 6.8.
func (q *SetPtQual) FlagSelect() { *q |= 128 }

// SinglePtChangePack holds 16 single-point values, each including a flag to
// signal changes since the last reporting of this information. Note how changes
// like On-Off-On or Off-On-Off could go undetected without such tracking on the
// equipment.
//
// The compound information element (CP32) is read in big-endian order, conform
// “Status and status change detection” of companion standard 101, subclause
// 7.2.6.40.
type SinglePtChangePack uint32

// Element returns the single point together with its change-dected flag by
// sequence number. The function panics when n is not in range 1..16.
func (pack SinglePtChangePack) Element(n int) (status SinglePt, change bool) {
	if n < 1 || n > 16 {
		panic("number out of bounds")
	}
	// values are defined in big-endian bit-order
	status = SinglePt(pack>>(32-n)) & 1
	change = (pack>>(16-n))&1 != 0
	return
}
