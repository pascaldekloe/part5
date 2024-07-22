package part5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/pascaldekloe/part5/info"
)

// Single Point
type (
	// SinglePt gets called for type identifier 1: M_SP_NA_1.
	SinglePt[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtQual)
	// SinglePtAtMinute gets called for type identifier 2: M_SP_TA_1.
	SinglePtAtMinute[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtQual, info.CP24Time2a)
	// SinglePtAtMoment gets called for type identifier 30: M_SP_TB_1.
	SinglePtAtMoment[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtQual, info.CP56Time2a)
)

// Double Point
type (
	// DoublePt gets called for type identifier 3: M_DP_NA_1.
	DoublePt[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual)
	// DoublePtAtMinute gets called for type identifier 4: M_DP_TA_1.
	DoublePtAtMinute[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual, info.CP24Time2a)
	// DoublePtAtMoment gets called for type identifier 31: M_DP_TB_1.
	DoublePtAtMoment[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual, info.CP56Time2a)
)

// Step Position
type (
	// Step gets called for type identifier 5: M_ST_NA_1.
	Step[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, info.StepQual)
	// StepAtMinute gets called for type identifier 6: M_ST_TA_1.
	StepAtMinute[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, info.StepQual, info.CP24Time2a)
	// StepAtMoment gets called for type identifier 32: M_ST_TB_1.
	StepAtMoment[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, info.StepQual, info.CP56Time2a)
)

// Bitstring
type (
	// Bits gets called for type identifier 7: M_BO_NA_1.
	Bits[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, info.BitsQual)
	// BitsAtMinute gets called for type identifier 8: M_BO_TA_1.
	BitsAtMinute[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, info.BitsQual, info.CP24Time2a)
	// BitsAtMoment gets called for type identifier 33: M_BO_TB_1.
	BitsAtMoment[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, info.BitsQual, info.CP56Time2a)
)

// Normalized Value
type (
	// NormUnqual gets called for type identifier 21: M_ME_ND_1.
	NormUnqual[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, info.Norm)
	// Norm gets called for type identifier 9: M_ME_NA_1.
	Norm[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, info.NormQual)
	// NormAtMinute gets called for type identifier 10: M_ME_TA_1.
	NormAtMinute[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, info.NormQual, info.CP24Time2a)
	// NormAtMoment gets called for type identifier 34: M_ME_TD_1.
	NormAtMoment[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, info.NormQual, info.CP56Time2a)
)

// Scaled Value
type (
	// Scaled gets called for type identifier 11: M_ME_NB_1.
	Scaled[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, int16, info.Qual)
	// ScaledAtMinute gets called for type identifier 12: M_ME_TB_1.
	ScaledAtMinute[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, int16, info.Qual, info.CP24Time2a)
	// ScaledAtMoment gets called for type identifier 35: M_ME_TE_1.
	ScaledAtMoment[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, int16, info.Qual, info.CP56Time2a)
)

// Floating Point
type (
	// Float gets called for type identifier 13: M_ME_NC_1.
	Float[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, float32, info.Qual)
	// FloatAtMinute gets called for type identifier 14: M_ME_TC_1.
	FloatAtMinute[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, float32, info.Qual, info.CP24Time2a)
	// FloatAtMoment gets called for type identifier 36: M_ME_TF_1.
	FloatAtMoment[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] func(info.DataUnit[Orig, Com, Obj], Obj, float32, info.Qual, info.CP56Time2a)
)

// Monitor processes information in monitor direction,
// as listed in table 8 from companion standard 101.
type Monitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	info.Params[Orig, Com, Obj]

	SinglePt[Orig, Com, Obj]
	SinglePtAtMinute[Orig, Com, Obj]
	SinglePtAtMoment[Orig, Com, Obj]

	DoublePt[Orig, Com, Obj]
	DoublePtAtMinute[Orig, Com, Obj]
	DoublePtAtMoment[Orig, Com, Obj]

	Step[Orig, Com, Obj]
	StepAtMinute[Orig, Com, Obj]
	StepAtMoment[Orig, Com, Obj]

	Bits[Orig, Com, Obj]
	BitsAtMinute[Orig, Com, Obj]
	BitsAtMoment[Orig, Com, Obj]

	NormUnqual[Orig, Com, Obj]
	Norm[Orig, Com, Obj]
	NormAtMinute[Orig, Com, Obj]
	NormAtMoment[Orig, Com, Obj]

	Scaled[Orig, Com, Obj]
	ScaledAtMinute[Orig, Com, Obj]
	ScaledAtMoment[Orig, Com, Obj]

	Float[Orig, Com, Obj]
	FloatAtMinute[Orig, Com, Obj]
	FloatAtMoment[Orig, Com, Obj]
}

// NewMonitor instantiatiates all listener functions with a no-op placeholder.
func NewMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](p info.Params[Orig, Com, Obj]) Monitor[Orig, Com, Obj] {
	return Monitor[Orig, Com, Obj]{
		Params: p, // nop

		SinglePt:         func(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtQual) {},
		SinglePtAtMinute: func(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtQual, info.CP24Time2a) {},
		SinglePtAtMoment: func(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtQual, info.CP56Time2a) {},

		DoublePt:         func(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual) {},
		DoublePtAtMinute: func(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual, info.CP24Time2a) {},
		DoublePtAtMoment: func(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual, info.CP56Time2a) {},

		Step:         func(info.DataUnit[Orig, Com, Obj], Obj, info.StepQual) {},
		StepAtMinute: func(info.DataUnit[Orig, Com, Obj], Obj, info.StepQual, info.CP24Time2a) {},
		StepAtMoment: func(info.DataUnit[Orig, Com, Obj], Obj, info.StepQual, info.CP56Time2a) {},

		Bits:         func(info.DataUnit[Orig, Com, Obj], Obj, info.BitsQual) {},
		BitsAtMinute: func(info.DataUnit[Orig, Com, Obj], Obj, info.BitsQual, info.CP24Time2a) {},
		BitsAtMoment: func(info.DataUnit[Orig, Com, Obj], Obj, info.BitsQual, info.CP56Time2a) {},

		NormUnqual:   func(info.DataUnit[Orig, Com, Obj], Obj, info.Norm) {},
		Norm:         func(info.DataUnit[Orig, Com, Obj], Obj, info.NormQual) {},
		NormAtMinute: func(info.DataUnit[Orig, Com, Obj], Obj, info.NormQual, info.CP24Time2a) {},
		NormAtMoment: func(info.DataUnit[Orig, Com, Obj], Obj, info.NormQual, info.CP56Time2a) {},

		Scaled:         func(info.DataUnit[Orig, Com, Obj], Obj, int16, info.Qual) {},
		ScaledAtMinute: func(info.DataUnit[Orig, Com, Obj], Obj, int16, info.Qual, info.CP24Time2a) {},
		ScaledAtMoment: func(info.DataUnit[Orig, Com, Obj], Obj, int16, info.Qual, info.CP56Time2a) {},

		Float:         func(info.DataUnit[Orig, Com, Obj], Obj, float32, info.Qual) {},
		FloatAtMinute: func(info.DataUnit[Orig, Com, Obj], Obj, float32, info.Qual, info.CP24Time2a) {},
		FloatAtMoment: func(info.DataUnit[Orig, Com, Obj], Obj, float32, info.Qual, info.CP56Time2a) {},
	}
}

// NewLog writes each measurement event as a text line in a human readable formon.
func NewLog[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](p info.Params[Orig, Com, Obj], w io.Writer) Monitor[Orig, Com, Obj] {
	return Monitor[Orig, Com, Obj]{
		Params: p, // nop

		SinglePt: func(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.SinglePtQual) {
			fmt.Fprintf(w, "%s %s %x %#x/%#x %s %s\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual())
		},
		SinglePtAtMinute: func(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.SinglePtQual, t info.CP24Time2a) {
			min, secInMilli := t.MinuteAndMillis()
			fmt.Fprintf(w, "%s %s %x %#x/%#x %s %s :%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(),
				min, secInMilli/1000, secInMilli%1000)
		},
		SinglePtAtMoment: func(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.SinglePtQual, t info.CP56Time2a) {
			y, m, d := t.Calendar()
			H, M, secInMilli := t.ClockAndMillis()
			fmt.Fprintf(w, "%s %s %x %#x/%#x %s %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(),
				y, m, d, H, M, secInMilli/1000, secInMilli%1000)
		},

		DoublePt: func(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.DoublePtQual) {
			fmt.Fprintf(w, "%s %s %x %#x/%#x %s %s\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual())
		},
		DoublePtAtMinute: func(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.DoublePtQual, t info.CP24Time2a) {
			min, secInMilli := t.MinuteAndMillis()
			fmt.Fprintf(w, "%s %s %x %#x/%#x %s %s :%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(),
				min, secInMilli/1000, secInMilli%1000)
		},
		DoublePtAtMoment: func(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.DoublePtQual, t info.CP56Time2a) {
			y, m, d := t.Calendar()
			H, M, secInMilli := t.ClockAndMillis()
			fmt.Fprintf(w, "%s %s %x %#x/%#x %s %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(),
				y, m, d, H, M, secInMilli/1000, secInMilli%1000)
		},

		Step: func(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.StepQual) {
			fmt.Fprintf(w, "%s %s %x %#x/%#x %s %s\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual())
		},
		StepAtMinute: func(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.StepQual, t info.CP24Time2a) {
			min, secInMilli := t.MinuteAndMillis()
			fmt.Fprintf(w, "%s %s %x %#x/%#x %s %s :%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(),
				min, secInMilli/1000, secInMilli%1000)
		},
		StepAtMoment: func(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.StepQual, t info.CP56Time2a) {
			y, m, d := t.Calendar()
			H, M, secInMilli := t.ClockAndMillis()
			fmt.Fprintf(w, "%s %s %x %#x/%#x %s %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(),
				y, m, d, H, M, secInMilli/1000, secInMilli%1000)
		},

		Bits: func(u info.DataUnit[Orig, Com, Obj], addr Obj, b info.BitsQual) {
			fmt.Fprintf(w, "%s %s %x %#x/%#x %#x %s\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, b.Array(), b.Qual())
		},
		BitsAtMinute: func(u info.DataUnit[Orig, Com, Obj], addr Obj, b info.BitsQual, t info.CP24Time2a) {
			min, secInMilli := t.MinuteAndMillis()
			fmt.Fprintf(w, "%s %s %x %#x/%#x %#x %s :%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, b.Array(), b.Qual(),
				min, secInMilli/1000, secInMilli%1000)
		},
		BitsAtMoment: func(u info.DataUnit[Orig, Com, Obj], addr Obj, b info.BitsQual, t info.CP56Time2a) {
			y, m, d := t.Calendar()
			H, M, secInMilli := t.ClockAndMillis()
			fmt.Fprintf(w, "%s %s %x %#x/%#x %#x %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, b.Array(), b.Qual(),
				y, m, d, H, M, secInMilli/1000, secInMilli%1000)
		},

		NormUnqual: func(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.Norm) {
			fmt.Fprintf(w, "%s %s %x %#x/%#x %f\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, n.Float64())
		},
		Norm: func(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.NormQual) {
			fmt.Fprintf(w, "%s %s %x %#x/%#x %f %s\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, n.Link().Float64(), n.Qual())
		},
		NormAtMinute: func(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.NormQual, t info.CP24Time2a) {
			min, secInMilli := t.MinuteAndMillis()
			fmt.Fprintf(w, "%s %s %x %#x/%#x %f %s :%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, n.Link().Float64(), n.Qual(),
				min, secInMilli/1000, secInMilli%1000)
		},
		NormAtMoment: func(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.NormQual, t info.CP56Time2a) {
			y, m, d := t.Calendar()
			H, M, secInMilli := t.ClockAndMillis()
			fmt.Fprintf(w, "%s %s %x %#x/%#x %f %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, n.Link().Float64(), n.Qual(),
				y, m, d, H, M, secInMilli/1000, secInMilli%1000)
		},

		Scaled: func(u info.DataUnit[Orig, Com, Obj], addr Obj, v int16, q info.Qual) {
			fmt.Fprintf(w, "%s %s %x %#x/%#x %d %s\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, v, q)
		},
		ScaledAtMinute: func(u info.DataUnit[Orig, Com, Obj], addr Obj, v int16, q info.Qual, t info.CP24Time2a) {
			min, secInMilli := t.MinuteAndMillis()
			fmt.Fprintf(w, "%s %s %x %#x/%#x %d %s :%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, v, q,
				min, secInMilli/1000, secInMilli%1000)
		},
		ScaledAtMoment: func(u info.DataUnit[Orig, Com, Obj], addr Obj, v int16, q info.Qual, t info.CP56Time2a) {
			y, m, d := t.Calendar()
			H, M, secInMilli := t.ClockAndMillis()
			fmt.Fprintf(w, "%s %s %x %#x/%#x %d %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, v, q,
				y, m, d, H, M, secInMilli/1000, secInMilli%1000)
		},

		Float: func(u info.DataUnit[Orig, Com, Obj], addr Obj, f float32, q info.Qual) {
			fmt.Fprintf(w, "%s %s %x %#x/%#x %g %s\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, f, q)
		},
		FloatAtMinute: func(u info.DataUnit[Orig, Com, Obj], addr Obj, f float32, q info.Qual, t info.CP24Time2a) {
			min, secInMilli := t.MinuteAndMillis()
			fmt.Fprintf(w, "%s %s %x %#x/%#x %g %s :%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, f, q,
				min, secInMilli/1000, secInMilli%1000)
		},
		FloatAtMoment: func(u info.DataUnit[Orig, Com, Obj], addr Obj, f float32, q info.Qual, t info.CP56Time2a) {
			y, m, d := t.Calendar()
			H, M, secInMilli := t.ClockAndMillis()
			fmt.Fprintf(w, "%s %s %x %#x/%#x %g %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Orig, u.Addr, addr, f, q,
				y, m, d, H, M, secInMilli/1000, secInMilli%1000)
		},
	}
}

// ErrNotMonitor rejects an info.DataUnit based on its type identifier.
var ErrNotMonitor = errors.New("part5: ASDU type identifier not in monitor information range 1..44")

var errInfoSize = errors.New("part5: size of ASDU payload doesn't match the variable structure qualifier")

// OnDataUnit applies the payload to the respective listener. Note that one
// DataUnit can have multiple information elements, in which case the listener
// is invoked once for each element in order of appearance. DataUnits with zero
// information elements do not cause any listener invocation. Errors other than
// ErrNotMonitor reject malformed content in the DataUnit.
func (mon *Monitor[Orig, Com, Obj]) OnDataUnit(u info.DataUnit[Orig, Com, Obj]) error {
	// monitor type identifiers (M_*) are in range 1..44
	if u.Type-1 > 43 {
		return ErrNotMonitor
	}

	// NOTE: Go can't get the array length from a generic as a constant yet.
	var addr Obj

	switch u.Type {
	case info.M_SP_NA_1: // single-point
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 1)
			if err != nil {
				return err
			}
			for _, b := range u.Info[len(addr):] {
				mon.SinglePt(u, addr, info.SinglePtQual(b))
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+1) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+1 {
				mon.SinglePt(u, Obj(u.Info[:len(addr)]),
					info.SinglePtQual(u.Info[len(addr)]),
				)
				u.Info = u.Info[len(addr)+1:]
			}
		}

	case info.M_SP_TA_1: // single-point with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_SP_TA_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+4) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+4 {
			mon.SinglePtAtMinute(u, Obj(u.Info[:len(addr)]),
				info.SinglePtQual(u.Info[len(addr)]),
				info.CP24Time2a(u.Info[len(addr)+1:len(addr)+4]),
			)
			u.Info = u.Info[len(addr)+4:]
		}

	case info.M_SP_TB_1: // single-point with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_SP_TB_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for len(u.Info) >= (len(addr) + 8) {
			mon.SinglePtAtMoment(u, Obj(u.Info[:len(addr)]),
				info.SinglePtQual(u.Info[len(addr)]),
				info.CP56Time2a(u.Info[len(addr)+1:len(addr)+8]),
			)
			u.Info = u.Info[len(addr)+8:]
		}

	case info.M_DP_NA_1: // double-point
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 1)
			if err != nil {
				return err
			}
			for _, b := range u.Info[len(addr):] {
				mon.DoublePt(u, addr, info.DoublePtQual(b))
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+1) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+1 {
				mon.DoublePt(u, Obj(u.Info[:len(addr)]),
					info.DoublePtQual(u.Info[len(addr)]),
				)
				u.Info = u.Info[len(addr)+1:]
			}
		}

	case info.M_DP_TA_1: // double-point with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_DP_TA_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+4) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+4 {
			mon.DoublePtAtMinute(u, Obj(u.Info[:len(addr)]),
				info.DoublePtQual(u.Info[len(addr)]),
				info.CP24Time2a(u.Info[len(addr)+1:len(addr)+4]),
			)
			u.Info = u.Info[len(addr)+4:]
		}

	case info.M_DP_TB_1: // double-point with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_DP_TB_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+8 {
			mon.DoublePtAtMoment(u, Obj(u.Info[:len(addr)]),
				info.DoublePtQual(u.Info[len(addr)]),
				info.CP56Time2a(u.Info[len(addr)+1:len(addr)+8]),
			)
			u.Info = u.Info[len(addr)+8:]
		}

	case info.M_ST_NA_1: // step position
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 2)
			if err != nil {
				return err
			}
			for i := len(addr); i+2 <= len(u.Info); i += 2 {
				mon.Step(u, addr, info.StepQual(u.Info[i:i+2]))
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+2) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+2 {
				mon.Step(u, Obj(u.Info[:len(addr)]),
					info.StepQual(u.Info[len(addr):len(addr)+2]),
				)
				u.Info = u.Info[len(addr)+2:]
			}
		}

	case info.M_ST_TA_1: // step position with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ST_TA_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+5) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+5 {
			mon.StepAtMinute(u, Obj(u.Info[:len(addr)]),
				info.StepQual(u.Info[len(addr):len(addr)+2]),
				info.CP24Time2a(u.Info[len(addr)+2:len(addr)+5]),
			)
			u.Info = u.Info[len(addr)+5:]
		}

	case info.M_ST_TB_1: // step position with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ST_TB_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+9) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+9 {
			mon.StepAtMoment(u, Obj(u.Info[:len(addr)]),
				info.StepQual(u.Info[len(addr):len(addr)+2]),
				info.CP56Time2a(u.Info[len(addr)+2:len(addr)+9]),
			)
			u.Info = u.Info[len(addr)+9:]
		}

	case info.M_BO_NA_1: // bitstring
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 5)
			if err != nil {
				return err
			}
			for i := len(addr); i+5 <= len(u.Info); i += 5 {
				mon.Bits(u, addr, info.BitsQual(u.Info[i:i+5]))
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+5) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+5 {
				mon.Bits(u, Obj(u.Info),
					info.BitsQual(u.Info[len(addr):len(addr)+5]),
				)
				u.Info = u.Info[len(addr)+5:]
			}
		}

	case info.M_BO_TA_1: // bit string with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_BO_TA_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+8 {
			mon.BitsAtMinute(u, Obj(u.Info[:len(addr)]),
				info.BitsQual(u.Info[len(addr):len(addr)+5]),
				info.CP24Time2a(u.Info[len(addr)+5:len(addr)+8]),
			)
			u.Info = u.Info[len(addr)+8:]
		}

	case info.M_BO_TB_1: // bit string with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_BO_TB_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+12) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+12 {
			mon.BitsAtMoment(u, Obj(u.Info[:len(addr)]),
				info.BitsQual(u.Info[len(addr):len(addr)+5]),
				info.CP56Time2a(u.Info[len(addr)+5:len(addr)+12]),
			)
			u.Info = u.Info[len(addr)+12:]
		}

	case info.M_ME_ND_1: // normalized value without quality descriptor
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 2)
			if err != nil {
				return err
			}
			for i := len(addr); i+1 < len(u.Info); i += 2 {
				mon.NormUnqual(u, addr, info.Norm(u.Info[i:i+2]))
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+2) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+2 {
				mon.NormUnqual(u, Obj(u.Info[:len(addr)]),
					info.Norm(u.Info[len(addr):len(addr)+2]),
				)
				u.Info = u.Info[len(addr)+2:]
			}
		}

	case info.M_ME_NA_1: // normalized value
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 3)
			if err != nil {
				return err
			}
			for i := len(addr); i+3 <= len(u.Info); i += 3 {
				mon.Norm(u, addr, info.NormQual(u.Info[i:i+3]))
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+3) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+3 {
				mon.Norm(u, Obj(u.Info[:len(addr)]),
					info.NormQual(u.Info[len(addr):len(addr)+3]),
				)
				u.Info = u.Info[len(addr)+3:]
			}
		}

	case info.M_ME_TA_1: // normalized value with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ME_TA_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+6) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+6 {
			mon.NormAtMinute(u, Obj(u.Info[:len(addr)]),
				info.NormQual(u.Info[len(addr):len(addr)+3]),
				info.CP24Time2a(u.Info[len(addr)+3:len(addr)+6]),
			)
			u.Info = u.Info[len(addr)+6:]
		}

	case info.M_ME_TD_1: // normalized value with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ME_TD_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+10) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+10 {
			mon.NormAtMoment(u, Obj(u.Info[:len(addr)]),
				info.NormQual(u.Info[len(addr):len(addr)+3]),
				info.CP56Time2a(u.Info[len(addr)+3:len(addr)+10]),
			)
			u.Info = u.Info[len(addr)+10:]
		}

	case info.M_ME_NB_1: // scaled value
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 3)
			if err != nil {
				return err
			}
			for i := len(addr); i+3 <= len(u.Info); i += 3 {
				mon.Scaled(u, addr,
					int16(binary.LittleEndian.Uint16(u.Info[i:i+2])),
					info.Qual(u.Info[i+2]),
				)
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+3) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+3 {
				mon.Scaled(u, Obj(u.Info[:len(addr)]),
					int16(binary.LittleEndian.Uint16(u.Info[len(addr):len(addr)+2])),
					info.Qual(u.Info[len(addr)+2]),
				)
				u.Info = u.Info[len(addr)+3:]
			}
		}

	case info.M_ME_TB_1: // scaled value with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ME_TB_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+6) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+6 {
			mon.ScaledAtMinute(u, Obj(u.Info[:len(addr)]),
				int16(binary.LittleEndian.Uint16(u.Info[len(addr):len(addr)+2])),
				info.Qual(u.Info[len(addr)+2]),
				info.CP24Time2a(u.Info[len(addr)+3:len(addr)+6]),
			)
			u.Info = u.Info[len(addr)+6:]
		}

	case info.M_ME_TE_1: // scaled value with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ME_TE_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+10) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+10 {
			mon.ScaledAtMoment(u, Obj(u.Info[:len(addr)]),
				int16(binary.LittleEndian.Uint16(u.Info[len(addr):len(addr)+2])),
				info.Qual(u.Info[len(addr)+2]),
				info.CP56Time2a(u.Info[len(addr)+3:len(addr)+10]),
			)
			u.Info = u.Info[len(addr)+10:]
		}

	case info.M_ME_NC_1: // floating-point
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 5)
			if err != nil {
				return err
			}
			for i := len(addr); i+5 <= len(u.Info); i += 5 {
				mon.Float(u, addr,
					math.Float32frombits(binary.LittleEndian.Uint32(u.Info[i:i+4])),
					info.Qual(u.Info[i+4]),
				)
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+5) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+5 {
				mon.Float(u, Obj(u.Info[:len(addr)]),
					math.Float32frombits(
						binary.LittleEndian.Uint32(
							u.Info[len(addr):len(addr)+4],
						),
					),
					info.Qual(u.Info[len(addr)+4]),
				)
				u.Info = u.Info[len(addr)+5:]
			}
		}

	case info.M_ME_TC_1: // floating point with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ME_TC_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+8 {
			mon.FloatAtMinute(u, Obj(u.Info[:len(addr)]),
				math.Float32frombits(
					binary.LittleEndian.Uint32(
						u.Info[len(addr):len(addr)+4],
					),
				),
				info.Qual(u.Info[len(addr)+4]),
				info.CP24Time2a(u.Info[len(addr)+5:len(addr)+8]),
			)
			u.Info = u.Info[len(addr)+8:]
		}

	case info.M_ME_TF_1: // floating point with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ME_TF_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+12) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+12 {
			mon.FloatAtMoment(u, Obj(u.Info[:len(addr)]),
				math.Float32frombits(
					binary.LittleEndian.Uint32(
						u.Info[len(addr):len(addr)+4],
					),
				),
				info.Qual(u.Info[len(addr)+4]),
				info.CP56Time2a(u.Info[len(addr)+5:len(addr)+12]),
			)
			u.Info = u.Info[len(addr)+12:]
		}

	}
	return nil
}


func addrSeqStart[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](u *info.DataUnit[Orig, Com, Obj], encSize int) (addr Obj, err error) {
	if len(u.Info) != len(addr)+u.Enc.Count()*encSize {
		return addr, errInfoSize
	}
	addr = Obj(u.Info[:len(addr)])

	// overflow check
	lastN := addr.N() + uint(u.Enc.Count()) - 1
	if _, ok := u.Params.ObjAddrN(lastN); !ok {
		return addr, info.ErrAddrSeq
	}
	return addr, nil
}
