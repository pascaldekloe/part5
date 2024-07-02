// Package info provides the OSI presentation layer.
package info

import (
	"encoding/binary"
	"errors"
	"math"
	"unsafe"
)

// Information Object Address
type (
	Addr interface {
		// Addressing can use any of these three modes.
		Addr8 | Addr16 | Addr24

		// N gets the address as a numeric value.
		N() uint
	}

	Addr8  [1]uint8
	Addr16 [2]uint8
	Addr24 [3]uint8
)

// NewAddrN returns a numeric address mapping. The return is false
// when the numeric value does not fit the width of address type T.
func NewAddrN[T Addr](n uint) (T, bool) {
	var addr T

	// addr[0] = uint8(n)
	// Go's generics suck. Can't slice the type either.
	// https://github.com/golang/go/issues/44253

	view := unsafe.Slice(&addr[0], len(addr))
	for i := range view {
		view[i] = uint8(n)
		n <<= 8
	}

	return addr, n == 0
}

// N implements the Addr interface.
func (addr Addr8) N() uint { return uint(addr[0]) }

// Slice implements the Addr interface.
func (addr Addr8) Slice() []byte { return addr[:] }

// N implements the Addr interface.
func (addr Addr16) N() uint {
	return uint(addr[0]) | uint(addr[1])<<8
}

// N implements the Addr interface.
func (addr Addr24) N() uint {
	return uint(addr[0]) | uint(addr[1])<<8 | uint(addr[2])<<16
}

func (addr Addr16) LowOctet() uint8    { return addr[0] }
func (addr Addr16) HighOctet() uint8   { return addr[1] }
func (addr Addr24) LowOctet() uint8    { return addr[0] }
func (addr Addr24) MiddleOctet() uint8 { return addr[1] }
func (addr Addr24) HighOctet() uint8   { return addr[2] }

// CommonAddr is a station address. Zero is not used.
// The width is controlled by Params.AddrSize.
// See companion standard 101, subclause 7.2.4.
type CommonAddr uint16

// ErrAddrZero signals an uninitialized CommonAddr.
var errAddrZero = errors.New("part5: common address 0 is not used")

// GlobalAddr is the broadcast address. Use is restricted
// to C_IC_NA_1, C_CI_NA_1, C_CS_NA_1 and C_RP_NA_1.
// When in 8-bit mode 255 is mapped to this value on the fly.
const GlobalAddr CommonAddr = 65535

// ObjAddr is the information object address.
// The width is controlled by Params.ObjAddrSize.
// See companion standard 101, subclause 7.2.5.
type ObjAddr uint

// Zero means that the address is irrelevant.
const IrrelevantAddr ObjAddr = 0

// A SinglePoint information object (IEV 371-02-07) is either On or Off.
type SinglePoint uint8

// Single-Point States
const (
	Off = iota
	On
)

// SinglePointQual has a single-point information-object (IEV 371-02-07)
// with a quality descriptor. If the octet does not equal On or Off, then
// it has quality remarks.
type SinglePointQual uint8

// Value loses the quality descriptor.
func (pt SinglePointQual) Value() SinglePoint { return SinglePoint(pt & 1) }

// Qual returns the quality descriptor flags. Overflow is always zero and
// EllapesTimeInvalid does not apply.
func (pt SinglePointQual) Qual() Qual { return Qual(pt & 0xfe) }

// A DoublePoint information object (IEV 371-02-08) is either DeterminatedOn,
// DeterminatedOff, Indeterminated, or IndeterminateOrIntermediate.
// http://blog.iec61850.com/2009/04/why-do-we-need-single-point-and-double.html
type DoublePoint uint8

// Double-Point States
const (
	IndeterminateOrIntermediate = iota
	DeterminedOff
	DeterminedOn
	Indeterminate
)

// DoublePointQual has a double-point information-object (IEV 371-02-08)
// with a quality descriptor. If the octet does not equal any of the four
// states, then it has quality remarks.
type DoublePointQual uint8

// Value loses the quality descriptor.
func (pt DoublePointQual) Value() DoublePoint { return DoublePoint(pt & 3) }

// Qual returns the quality descriptor. Overflow is always zero and
// EllapesTimeInvalid does not apply.
func (pt DoublePointQual) Qual() Qual { return Qual(pt & 0xfc) }

// Qual holds flags about an information object, a.k.a. the quality descriptor.
type Qual uint8

// Quality Descriptor Flags
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
	EllapsedTimeInvalid

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

// StepPos is a measured value with a transient state indication.
// See companion standard 101, subclause 7.2.6.5.
type StepPos uint

// NewStepPos returns a new step position.
// Values out of [-64, 63] oveflow silently.
func NewStepPos(value int, transient bool) StepPos {
	pos := StepPos(value & 0x7f)
	if transient {
		pos |= 0x80
	}
	return pos
}

// Pos returns the value in [-64, 63] plus whether the equipment is transient state.
func (pos StepPos) Pos() (value int, transient bool) {
	u := uint(pos)
	if u&0x40 == 0 {
		// trim rest
		value = int(u & 0x3f)
	} else {
		// sign extend
		value = int(u) | (-1 &^ 0x3f)
	}
	transient = u&0x80 != 0
	return
}

// StepPosQual is a measured value with transient state indication,
// and with a quality descriptor.
// See companion standard 101, subclause 7.2.6.5.
type StepPosQual [2]uint8

// NewStepPosQual returns a new step position, with quality descriptor OK.
// Values out of [-64, 63] oveflow silently.
func NewStepPosQual(value int, flags Qual) StepPosQual {
	return StepPosQual{uint8(value & 0x7f), uint8(flags)}
}

// NewTransientStepPosQual sets the transient flag.
func NewTransientStepPosQual(value int, flags Qual) StepPosQual {
	return StepPosQual{uint8(value&0x7f) | 0x80, uint8(flags)}
}

// Value returns the step position value in range [-64, 63], plus whether the
// equipment is in a transient state.
func (pos StepPosQual) Value() (value int, transient bool) {
	value = int(pos[0] & 0x7f)
	signBit := value >> 6
	signExt := 0 - signBit
	value |= signExt << 6
	return value, pos[0]&0x80 != 0
}

// Qual returns the quality descriptor. EllapesTimeInvalid does not apply.
func (pos StepPosQual) Qual() Qual { return Qual(pos[1]) }

// FlagQual sets [logical OR] the quality descriptor bits from flags.
func (pos *StepPosQual) FlagQual(flags Qual) {
	pos[1] |= uint8(flags)
}

// BitsQual is a four octet bitstring with a quality descriptor.
type BitsQual [5]uint8

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

// Normal is a 16-bit normalized value.
// See companion standard 101, subclause 7.2.6.6.
type Normal int16

// Float64 returns the value in [-1, 1 − 2⁻¹⁵].
func (n Normal) Float64() float64 { return float64(n) / 32768 }

// Qualifier Of Parameter Of Measured Values
// See companion standard 101, subclause 7.2.6.24.
const (
	_          = iota // 0: not used
	Threashold        // 1: threshold value
	Smoothing         // 2: smoothing factor (filter time constant)
	LowLimit          // 3: low limit for transmission of measured values
	HighLimit         // 4: high limit for transmission of measured values

	// 5‥31: reserved for standard definitions of this companion standard (compatible range)
	// 32‥63: reserved for special use (private range)

	ChangeFlag      = 64  // marks local parameter change
	InOperationFlag = 128 // marks parameter operation
)

// Cmd is a command.
// See companion standard 101, subclause 7.2.6.26.
type Cmd uint

// Qual returns the qualifier of command.
//
//	0: no additional definition
//	1: short pulse duration (circuit-breaker), duration determined by a system parameter in the outstation
//	2: long pulse duration, duration determined by a system parameter in the outstation
//	3: persistent output
//	4‥8: reserved for standard definitions of this companion standard
//	9‥15: reserved for the selection of other predefined functions
//	16‥31: reserved for special use (private range)
func (c Cmd) Qual() uint { return uint((c >> 2) & 31) }

// Exec returns whether the command executes (or selects).
// See section 5, subclause 6.8.
func (c Cmd) Exec() bool { return c&128 == 0 }

// SetpointCmd is the qualifier of a set-point command.
// See companion standard 101, subclause 7.2.6.39.
type SetpointCmd uint

// Qual returns the qualifier of set-point command.
//
//	0: default
//	0‥63: reserved for standard definitions of this companion standard (compatible range)
//	64‥127: reserved for special use (private range)
func (c SetpointCmd) Qual() uint { return uint(c & 127) }

// Exec returns whether the command executes (or selects).
// See section 5, subclause 6.8.
func (c SetpointCmd) Exec() bool { return c&128 == 0 }
