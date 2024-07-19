// Package info provides the OSI presentation layer.
//
// Common addresses and information-object addresses (ComAddr and Addr) can be
// formatted with the standard fmt package. The numeric notation in hexadecimal
// "%x" and its upper-casing "%X" do not omit any leading zeroes. Decimals can
// fixed in width too with leading space "% d". The “alternated format” is an
// octet listing in big-endian order. Decimal octets "%#d" get sepatared by the
// dot character ".", while hexadecimals "%#x" and "%#X" get separated by the
// colon character ":".
package info

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Params define the system-specific parameters. Specifically, the Cause of
// transmission (COT) sets the presence or absense of an originator address.
// The address generics set the numeric width for stations and information
// objects respectively.
type Params[COT Cause, Common ComAddr, Object Addr] struct{}

// ErrComAddrZero denies zero as an address.
var errComAddrZero = errors.New("part5: common address 0 is not used")

// The “common address” addresses stations. Zero is not used.
// All information objects/addresses reside in a common address.
// See companion standard 101, subclause 7.2.4.
type (
	// A ComAddr can be instantiated with ComAddrOf of Params.
	ComAddr interface {
		// The address width is fixed per system.
		ComAddr8 | ComAddr16

		// N gets the address as a numeric value.
		N() uint

		// The global address is a broadcast address,
		// directed to all stations of a specific system.
		// Use is restricted to type C_IC_NA_1, C_CI_NA_1,
		// C_CS_NA_1 and C_RP_NA_1.
		IsGlobal() bool
	}

	ComAddr8  [1]uint8
	ComAddr16 [2]uint8
)

// ComAddrOf returns either a numeric match, or false when n exceeds the address
// width of Common.
func (p Params[COT, Common, Object]) ComAddrOf(n uint) (Common, bool) {
	var addr Common
	for i := 0; i < len(addr); i++ {
		addr[i] = uint8(n)
		n >>= 8
	}
	return addr, n == 0
}

// MustComAddrOf is like CommAddrOf, yet it panics on overflow.
func (p Params[COT, Common, Object]) MustComAddrOf(n uint) Common {
	addr, ok := p.ComAddrOf(n)
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

// IsGlobal implements the ComAddr interface.
func (addr ComAddr8) IsGlobal() bool {
	return addr[0] == 0xff
}

// IsGlobal implements the ComAddr interface.
func (addr ComAddr16) IsGlobal() bool {
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
	// An Addr can be instantiated with AddrOf of Params.
	Addr interface {
		// The address width is fixed per system.
		Addr8 | Addr16 | Addr24

		// N gets the address as a numeric value.
		// Zero marks the address as irrelevant.
		N() uint
	}

	Addr8  [1]uint8
	Addr16 [2]uint8
	Addr24 [3]uint8
)

// AddrOf returns either a numeric match, or false when n exceeds the address
// width of Object.
func (p Params[COT, Common, Object]) AddrOf(n uint) (Object, bool) {
	var addr Object
	for i := 0; i < len(addr); i++ {
		addr[i] = uint8(n)
		n >>= 8
	}
	return addr, n == 0
}

// MustAddrOf is like AddrOf, yet it panics on overflow.
func (p Params[COT, Common, Object]) MustAddrOf(n uint) Object {
	addr, ok := p.AddrOf(n)
	if !ok {
		panic("overflow of information-object")
	}
	return addr
}

// N implements the Addr interface.
func (addr Addr8) N() uint {
	return uint(addr[0])
}

// N implements the Addr interface.
func (addr Addr16) N() uint {
	return uint(addr[0]) | uint(addr[1])<<8
}

// N implements the Addr interface.
func (addr Addr24) N() uint {
	return uint(addr[0]) | uint(addr[1])<<8 | uint(addr[2])<<16
}

// LowOctet provides an alternative to numeric addressing.
func (addr Addr16) LowOctet() uint8 { return addr[0] }

// HighOctet provides an alternative to numeric addressing.
func (addr Addr16) HighOctet() uint8 { return addr[1] }

// LowOctet provides an alternative to numeric addressing.
func (addr Addr24) LowOctet() uint8 { return addr[0] }

// MiddleOctet provides an alternative to numeric addressing.
func (addr Addr24) MiddleOctet() uint8 { return addr[1] }

// HighOctet provides an alternative to numeric addressing.
func (addr Addr24) HighOctet() uint8 { return addr[2] }

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
// EllapedTimeInvalid does not apply.
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
// EllapesTimeInvalid does not apply.
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
	Overflow = 1 << iota

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

	OK = 0 // no remarks
)

// String returns the flag tokens separated by comma, with "OK" for no flags.
func (q Qual) String() string {
	switch q {
	case OK: // no flags
		return "OK"
	case Overflow:
		return "OV"
	case ElapsedTimeInvalid:
		return "EI"
	case Blocked:
		return "BL"
	case Substituted:
		return "SB"
	case NotTopical:
		return "NT"
	case Invalid:
		return "IV"
	}
	// got multiple flags

	var buf strings.Builder
	if q&Overflow != 0 {
		buf.WriteString(",OV")
	}
	if q&2 != 0 {
		buf.WriteString(",RES2")
	}
	if q&4 != 0 {
		buf.WriteString(",RES3")
	}
	if q&ElapsedTimeInvalid != 0 {
		buf.WriteString(",EI")
	}
	if q&Blocked != 0 {
		buf.WriteString(",BL")
	}
	if q&Substituted != 0 {
		buf.WriteString(",SB")
	}
	if q&NotTopical != 0 {
		buf.WriteString(",NT")
	}
	if q&Invalid != 0 {
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

// Qual returns the quality descriptor. EllapesTimeInvalid does not apply.
func (p StepQual) Qual() Qual { return Qual(p[1]) }

// FlagQual sets [logical OR] the quality descriptor bits from flags.
func (p *StepQual) FlagQual(flags Qual) {
	p[1] |= uint8(flags)
}

// BitsQual is a four octet bitstring with a quality descriptor.
type BitsQual [5]uint8

// Array gets a copy of the bitstring.
func (bits *BitsQual) Array() [4]uint8 { return ([4]byte)(bits[:4]) }

// Slice gets a reference to the data.
func (bits *BitsQual) Slice() []uint8 { return bits[:4] }

// BigEndian returns the bitstring as an integer.
func (bits BitsQual) BigEndian() uint32 {
	return binary.BigEndian.Uint32(bits[:4])
}

// SetBigEndian updates the bitstring with an integer.
func (bits *BitsQual) SetBigEndian(v uint32) {
	binary.BigEndian.PutUint32(bits[:4], v)
}

// Qual returns the quality descriptor. EllapesTimeInvalid does not apply.
func (bits BitsQual) Qual() Qual { return Qual(bits[4]) }

// FlagQual sets [logical OR] the quality descriptor bits from flags.
func (bits *BitsQual) FlagQual(flags Qual) {
	bits[4] |= uint8(flags)
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

// Link returns a reference to the normalized-value withouth the quality
// descriptor.
func (norm *NormQual) Link() *Norm { return (*Norm)(norm[:2]) }

// Qual returns the quality descriptor. EllapesTimeInvalid does not apply.
func (norm NormQual) Qual() Qual { return Qual(norm[2]) }

// FlagQual sets [logical OR] the quality descriptor bits from flags.
func (norm *NormQual) FlagQual(flags Qual) {
	norm[2] |= uint8(flags)
}

// Qualifier Of Parameter Of Measured Values
// See companion standard 101, subclause 7.2.6.24.
const (
	_          = iota // 0: not used
	Threashold        // 1: threshold value
	Smoothing         // 2: smoothing factor (filter time constant)
	LowLimit          // 3: low limit for transmission of measured values
	HighLimit         // 4: high limit for transmission of measured values

	// 5..31: reserved for standard definitions of this companion standard (compatible range)
	// 32..63: reserved for special use (private range)

	ChangeFlag      = 64  // marks local parameter change
	InOperationFlag = 128 // marks parameter operation
)

// Cmd is a command.
// See companion standard 101, subclause 7.2.6.26.
type Cmd uint8

// Select flags the command to select (instead of execute).
// See “Command transmission” in section 5, subclause 6.8.
const Select = 128

func NewSingleCmd(p SinglePt) Cmd { return Cmd(p & 1) }
func NewDoubleCmd(p DoublePt) Cmd { return Cmd(p & 3) }
func NewRegulCmd(r Regul) Cmd     { return Cmd(r & 3) }

// SinglePt returns the value assuming a single-point command.
// The return is either Off or On.
func (c Cmd) SinglePt() SinglePt { return SinglePt(c & 1) }

// DoublePt returns the value assuming a double-point command.
// The return is either IndeterminateOrIntermediate, DeterminatedOff,
// DeterminatedOn or Indeterminate.
func (c Cmd) DoublePt() DoublePt { return DoublePt(c & 3) }

// Regul returns the value assuming a regulating-step command. The return should
// be either Lower or Higher. The other two states (0 and 3) are not permitted.
func (c Cmd) Regul() Regul { return Regul(c & 3) }

// Def returns the additional definition in the qualifier of command, if any.
//
//	0: no additional definition
//	1: short pulse duration (circuit-breaker), duration determined by a system parameter in the outstation
//	2: long pulse duration, duration determined by a system parameter in the outstation
//	3: persistent output
//	4..8: reserved for standard definitions of the IEC companion standard
//	9..15: reserved for the selection of other predefined functions
//	16..31: reserved for special use (private range)
func (c Cmd) Def() uint { return uint((c >> 2) & 31) }

// SetDef replaces the additional definition in the qualifier of command.
// Any bits from q values over 31 get discarded silently.
func (c *Cmd) SetDef(q uint) {
	// discard current additional definition, if any
	*c &^= 31 << 2
	// trim additional-definition range, and merge
	*c |= Cmd((q & 31) << 2)
}

// SetPtQual is the qualifier of a set-point command, which includes the Select
// flag. See companion standard 101, subclause 7.2.6.39.
type SetPtQual uint8

// N returns the qualifier code.
//
//	0: default
//	1..63: reserved for standard definitions of the IEC companion standard (compatible range)
//	64..127: reserved for special use (private range)
func (c SetPtQual) N() uint { return uint(c & 127) }
