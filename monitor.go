package part5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/pascaldekloe/part5/info"
)

// Monitor processes information in monitor direction,
// as listed in table 8 from companion standard 101.
type Monitor[COT info.Cause, Common info.ComAddr, Object info.Addr] struct {
	// SinglePt gets called for type identifier 1: M_SP_NA_1.
	SinglePt func(info.DataUnit[COT, Common, Object], Object, info.SinglePtQual)
	// SinglePtWithTime gets called for type identifier 2: M_SP_TA_1.
	SinglePtWithTime func(info.DataUnit[COT, Common, Object], Object, info.SinglePtQual, CP24Time2a)
	// SinglePtWithDate gets called for type identifier 30: M_SP_TB_1.
	SinglePtWithDate func(info.DataUnit[COT, Common, Object], Object, info.SinglePtQual, CP56Time2a)

	// DoublePt gets called for type identifier 3: M_DP_NA_1.
	DoublePt func(info.DataUnit[COT, Common, Object], Object, info.DoublePtQual)
	// DoublePtWithTime gets called for type identifier 4: M_DP_TA_1.
	DoublePtWithTime func(info.DataUnit[COT, Common, Object], Object, info.DoublePtQual, CP24Time2a)
	// DoublePtWithDate gets called for type identifier 31: M_DP_TB_1.
	DoublePtWithDate func(info.DataUnit[COT, Common, Object], Object, info.DoublePtQual, CP56Time2a)

	// Step gets called for type identifier 5: M_ST_NA_1.
	Step func(info.DataUnit[COT, Common, Object], Object, info.StepQual)
	// StepWithTime gets called for type identifier 6: M_ST_TA_1.
	StepWithTime func(info.DataUnit[COT, Common, Object], Object, info.StepQual, CP24Time2a)
	// StepWithDate gets called for type identifier 32: M_ST_TB_1.
	StepWithDate func(info.DataUnit[COT, Common, Object], Object, info.StepQual, CP56Time2a)

	// Bits gets called for type identifier 7: M_BO_NA_1.
	Bits func(info.DataUnit[COT, Common, Object], Object, info.BitsQual)
	// BitsWithTime gets called for type identifier 8: M_BO_TA_1.
	BitsWithTime func(info.DataUnit[COT, Common, Object], Object, info.BitsQual, CP24Time2a)
	// BitsWithDate gets called for type identifier 33: M_BO_TB_1.
	BitsWithDate func(info.DataUnit[COT, Common, Object], Object, info.BitsQual, CP56Time2a)

	NormUnqual func(info.DataUnit[COT, Common, Object], Object, info.Norm)
	// Norm gets called for type identifier 9: M_ME_NA_1.
	Norm func(info.DataUnit[COT, Common, Object], Object, info.NormQual)
	// NormWithTime gets called for type identifier 10: M_ME_TA_1.
	NormWithTime func(info.DataUnit[COT, Common, Object], Object, info.NormQual, CP24Time2a)
	// NormWithDate gets called for type identifier 34: M_ME_TD_1.
	NormWithDate func(info.DataUnit[COT, Common, Object], Object, info.NormQual, CP56Time2a)

	// Scaled gets called for type identifier 11: M_ME_NB_1.
	Scaled func(info.DataUnit[COT, Common, Object], Object, int16, info.Qual)
	// ScaledWithTime gets called for type identifier 12: M_ME_TB_1.
	ScaledWithTime func(info.DataUnit[COT, Common, Object], Object, int16, info.Qual, CP24Time2a)
	// ScaledWithDate gets called for type identifier 35: M_ME_TE_1.
	ScaledWithDate func(info.DataUnit[COT, Common, Object], Object, int16, info.Qual, CP56Time2a)

	// Float gets called for type identifier 13: M_ME_NC_1.
	Float func(info.DataUnit[COT, Common, Object], Object, float32, info.Qual)
	// FloatWithTime gets called for type identifier 14: M_ME_TC_1.
	FloatWithTime func(info.DataUnit[COT, Common, Object], Object, float32, info.Qual, CP24Time2a)
	// FloatWithDate gets called for type identifier 36: M_ME_TF_1.
	FloatWithDate func(info.DataUnit[COT, Common, Object], Object, float32, info.Qual, CP56Time2a)
}

// NewMonitor instantiatiates all listener functions with a no-op placeholder.
func NewMonitor[COT info.Cause, Common info.ComAddr, Object info.Addr]() Monitor[COT, Common, Object] {
	return Monitor[COT, Common, Object]{
		SinglePt:         func(info.DataUnit[COT, Common, Object], Object, info.SinglePtQual) {},
		SinglePtWithTime: func(info.DataUnit[COT, Common, Object], Object, info.SinglePtQual, CP24Time2a) {},
		SinglePtWithDate: func(info.DataUnit[COT, Common, Object], Object, info.SinglePtQual, CP56Time2a) {},

		DoublePt:         func(info.DataUnit[COT, Common, Object], Object, info.DoublePtQual) {},
		DoublePtWithTime: func(info.DataUnit[COT, Common, Object], Object, info.DoublePtQual, CP24Time2a) {},
		DoublePtWithDate: func(info.DataUnit[COT, Common, Object], Object, info.DoublePtQual, CP56Time2a) {},

		Step:         func(info.DataUnit[COT, Common, Object], Object, info.StepQual) {},
		StepWithTime: func(info.DataUnit[COT, Common, Object], Object, info.StepQual, CP24Time2a) {},
		StepWithDate: func(info.DataUnit[COT, Common, Object], Object, info.StepQual, CP56Time2a) {},

		Bits:         func(info.DataUnit[COT, Common, Object], Object, info.BitsQual) {},
		BitsWithTime: func(info.DataUnit[COT, Common, Object], Object, info.BitsQual, CP24Time2a) {},
		BitsWithDate: func(info.DataUnit[COT, Common, Object], Object, info.BitsQual, CP56Time2a) {},

		NormUnqual:   func(info.DataUnit[COT, Common, Object], Object, info.Norm) {},
		Norm:         func(info.DataUnit[COT, Common, Object], Object, info.NormQual) {},
		NormWithTime: func(info.DataUnit[COT, Common, Object], Object, info.NormQual, CP24Time2a) {},
		NormWithDate: func(info.DataUnit[COT, Common, Object], Object, info.NormQual, CP56Time2a) {},

		Scaled:         func(info.DataUnit[COT, Common, Object], Object, int16, info.Qual) {},
		ScaledWithTime: func(info.DataUnit[COT, Common, Object], Object, int16, info.Qual, CP24Time2a) {},
		ScaledWithDate: func(info.DataUnit[COT, Common, Object], Object, int16, info.Qual, CP56Time2a) {},

		Float:         func(info.DataUnit[COT, Common, Object], Object, float32, info.Qual) {},
		FloatWithTime: func(info.DataUnit[COT, Common, Object], Object, float32, info.Qual, CP24Time2a) {},
		FloatWithDate: func(info.DataUnit[COT, Common, Object], Object, float32, info.Qual, CP56Time2a) {},
	}
}

// NewLog writes each measurement event as a text line in a human readable formon.
func NewLog[COT info.Cause, Common info.ComAddr, Object info.Addr](w io.Writer) Monitor[COT, Common, Object] {
	return Monitor[COT, Common, Object]{
		SinglePt: func(u info.DataUnit[COT, Common, Object], addr Object, p info.SinglePtQual) {
			fmt.Fprintf(w, "%s %s %#x/%#x %s %s\n",
				u.Type, u.Cause, u.Addr, addr, p.Value(), p.Qual())
		},
		SinglePtWithTime: func(u info.DataUnit[COT, Common, Object], addr Object, p info.SinglePtQual, t CP24Time2a) {
			min, secInMilli := t.MinuteAndMillis()
			fmt.Fprintf(w, "%s %s %#x/%#x %s %s :%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Addr, addr, p.Value(), p.Qual(),
				min, secInMilli/1000, secInMilli%1000)
		},
		SinglePtWithDate: func(u info.DataUnit[COT, Common, Object], addr Object, p info.SinglePtQual, t CP56Time2a) {
			y, m, d := t.Calendar()
			H, M, secInMilli := t.ClockAndMillis()
			fmt.Fprintf(w, "%s %s %#x/%#x %s %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Addr, addr, p.Value(), p.Qual(),
				y, m, d, H, M, secInMilli/1000, secInMilli%1000)
		},

		DoublePt: func(u info.DataUnit[COT, Common, Object], addr Object, p info.DoublePtQual) {
			fmt.Fprintf(w, "%s %s %#x/%#x %s %s\n",
				u.Type, u.Cause, u.Addr, addr, p.Value(), p.Qual())
		},
		DoublePtWithTime: func(u info.DataUnit[COT, Common, Object], addr Object, p info.DoublePtQual, t CP24Time2a) {
			min, secInMilli := t.MinuteAndMillis()
			fmt.Fprintf(w, "%s %s %#x/%#x %s %s :%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Addr, addr, p.Value(), p.Qual(),
				min, secInMilli/1000, secInMilli%1000)
		},
		DoublePtWithDate: func(u info.DataUnit[COT, Common, Object], addr Object, p info.DoublePtQual, t CP56Time2a) {
			y, m, d := t.Calendar()
			H, M, secInMilli := t.ClockAndMillis()
			fmt.Fprintf(w, "%s %s %#x/%#x %s %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Addr, addr, p.Value(), p.Qual(),
				y, m, d, H, M, secInMilli/1000, secInMilli%1000)
		},

		Step: func(u info.DataUnit[COT, Common, Object], addr Object, p info.StepQual) {
			fmt.Fprintf(w, "%s %s %#x/%#x %s %s\n",
				u.Type, u.Cause, u.Addr, addr, p.Value(), p.Qual())
		},
		StepWithTime: func(u info.DataUnit[COT, Common, Object], addr Object, p info.StepQual, t CP24Time2a) {
			min, secInMilli := t.MinuteAndMillis()
			fmt.Fprintf(w, "%s %s %#x/%#x %s %s :%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Addr, addr, p.Value(), p.Qual(),
				min, secInMilli/1000, secInMilli%1000)
		},
		StepWithDate: func(u info.DataUnit[COT, Common, Object], addr Object, p info.StepQual, t CP56Time2a) {
			y, m, d := t.Calendar()
			H, M, secInMilli := t.ClockAndMillis()
			fmt.Fprintf(w, "%s %s %#x/%#x %s %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Addr, addr, p.Value(), p.Qual(),
				y, m, d, H, M, secInMilli/1000, secInMilli%1000)
		},

		Bits: func(u info.DataUnit[COT, Common, Object], addr Object, b info.BitsQual) {
			fmt.Fprintf(w, "%s %s %#x/%#x %#x %s\n",
				u.Type, u.Cause, u.Addr, addr, b.Array(), b.Qual())
		},
		BitsWithTime: func(u info.DataUnit[COT, Common, Object], addr Object, b info.BitsQual, t CP24Time2a) {
			min, secInMilli := t.MinuteAndMillis()
			fmt.Fprintf(w, "%s %s %#x/%#x %#x %s :%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Addr, addr, b.Array(), b.Qual(),
				min, secInMilli/1000, secInMilli%1000)
		},
		BitsWithDate: func(u info.DataUnit[COT, Common, Object], addr Object, b info.BitsQual, t CP56Time2a) {
			y, m, d := t.Calendar()
			H, M, secInMilli := t.ClockAndMillis()
			fmt.Fprintf(w, "%s %s %#x/%#x %#x %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Addr, addr, b.Array(), b.Qual(),
				y, m, d, H, M, secInMilli/1000, secInMilli%1000)
		},

		NormUnqual: func(u info.DataUnit[COT, Common, Object], addr Object, n info.Norm) {
			fmt.Fprintf(w, "%s %s %#x/%#x %f\n",
				u.Type, u.Cause, u.Addr, addr, n.Float64())
		},
		Norm: func(u info.DataUnit[COT, Common, Object], addr Object, n info.NormQual) {
			fmt.Fprintf(w, "%s %s %#x/%#x %f %s\n",
				u.Type, u.Cause, u.Addr, addr, n.Link().Float64(), n.Qual())
		},
		NormWithTime: func(u info.DataUnit[COT, Common, Object], addr Object, n info.NormQual, t CP24Time2a) {
			min, secInMilli := t.MinuteAndMillis()
			fmt.Fprintf(w, "%s %s %#x/%#x %f %s :%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Addr, addr, n.Link().Float64(), n.Qual(),
				min, secInMilli/1000, secInMilli%1000)
		},
		NormWithDate: func(u info.DataUnit[COT, Common, Object], addr Object, n info.NormQual, t CP56Time2a) {
			y, m, d := t.Calendar()
			H, M, secInMilli := t.ClockAndMillis()
			fmt.Fprintf(w, "%s %s %#x/%#x %f %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Addr, addr, n.Link().Float64(), n.Qual(),
				y, m, d, H, M, secInMilli/1000, secInMilli%1000)
		},

		Scaled: func(u info.DataUnit[COT, Common, Object], addr Object, v int16, q info.Qual) {
			fmt.Fprintf(w, "%s %s %#x/%#x %d %s\n",
				u.Type, u.Cause, u.Addr, addr, v, q)
		},
		ScaledWithTime: func(u info.DataUnit[COT, Common, Object], addr Object, v int16, q info.Qual, t CP24Time2a) {
			min, secInMilli := t.MinuteAndMillis()
			fmt.Fprintf(w, "%s %s %#x/%#x %d %s :%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Addr, addr, v, q,
				min, secInMilli/1000, secInMilli%1000)
		},
		ScaledWithDate: func(u info.DataUnit[COT, Common, Object], addr Object, v int16, q info.Qual, t CP56Time2a) {
			y, m, d := t.Calendar()
			H, M, secInMilli := t.ClockAndMillis()
			fmt.Fprintf(w, "%s %s %#x/%#x %d %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Addr, addr, v, q,
				y, m, d, H, M, secInMilli/1000, secInMilli%1000)
		},

		Float: func(u info.DataUnit[COT, Common, Object], addr Object, f float32, q info.Qual) {
			fmt.Fprintf(w, "%s %s %#x/%#x %g %s\n",
				u.Type, u.Cause, u.Addr, addr, f, q)
		},
		FloatWithTime: func(u info.DataUnit[COT, Common, Object], addr Object, f float32, q info.Qual, t CP24Time2a) {
			min, secInMilli := t.MinuteAndMillis()
			fmt.Fprintf(w, "%s %s %#x/%#x %g %s :%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Addr, addr, f, q,
				min, secInMilli/1000, secInMilli%1000)
		},
		FloatWithDate: func(u info.DataUnit[COT, Common, Object], addr Object, f float32, q info.Qual, t CP56Time2a) {
			y, m, d := t.Calendar()
			H, M, secInMilli := t.ClockAndMillis()
			fmt.Fprintf(w, "%s %s %#x/%#x %g %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
				u.Type, u.Cause, u.Addr, addr, f, q,
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
func (mon *Monitor[COT, Common, Object]) OnDataUnit(u info.DataUnit[COT, Common, Object]) error {
	// monitor type identifiers (M_*) are in range 1..44
	if u.Type-1 > 43 {
		return ErrNotMonitor
	}

	// NOTE: Go can't get the array length from a generic as a constant yet.
	var addr Object

	switch u.Type {
	case info.M_SP_NA_1: // single-point
		if u.Var&info.Sequence != 0 {
			if len(u.Info) != len(addr)+u.Var.Count() {
				return errInfoSize
			}
			addr := Object(u.Info[:len(addr)])
			for _, b := range u.Info[len(addr):] {
				mon.SinglePt(u, addr, info.SinglePtQual(b))
				addr, _ = u.Params.AddrOf(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Var.Count()*(len(addr)+1) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+1 {
				mon.SinglePt(u, Object(u.Info[:len(addr)]),
					info.SinglePtQual(u.Info[len(addr)]),
				)
				u.Info = u.Info[len(addr)+1:]
			}
		}

	case info.M_SP_TA_1: // single-point with 3 octet time-tag
		if u.Var&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_SP_TA_1 not allowed")
		}
		if len(u.Info) != u.Var.Count()*(len(addr)+4) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+4 {
			mon.SinglePtWithTime(u, Object(u.Info[:len(addr)]),
				info.SinglePtQual(u.Info[len(addr)]),
				CP24Time2a(u.Info[len(addr)+1:len(addr)+4]),
			)
			u.Info = u.Info[len(addr)+4:]
		}

	case info.M_SP_TB_1: // single-point with 7 octet time-tag
		if u.Var&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_SP_TB_1 not allowed")
		}
		if len(u.Info) != u.Var.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for len(u.Info) >= (len(addr) + 8) {
			mon.SinglePtWithDate(u, Object(u.Info[:len(addr)]),
				info.SinglePtQual(u.Info[len(addr)]),
				CP56Time2a(u.Info[len(addr)+1:len(addr)+8]),
			)
			u.Info = u.Info[len(addr)+8:]
		}

	case info.M_DP_NA_1: // double-point
		if u.Var&info.Sequence != 0 {
			if len(u.Info) != len(addr)+u.Var.Count() {
				return errInfoSize
			}
			addr := Object(u.Info[:len(addr)])
			for _, b := range u.Info[len(addr):] {
				mon.DoublePt(u, addr, info.DoublePtQual(b))
				addr, _ = u.Params.AddrOf(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Var.Count()*(len(addr)+1) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+1 {
				mon.DoublePt(u, Object(u.Info[:len(addr)]),
					info.DoublePtQual(u.Info[len(addr)]),
				)
				u.Info = u.Info[len(addr)+1:]
			}
		}

	case info.M_DP_TA_1: // double-point with 3 octet time-tag
		if u.Var&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_DP_TA_1 not allowed")
		}
		if len(u.Info) != u.Var.Count()*(len(addr)+4) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+4 {
			mon.DoublePtWithTime(u, Object(u.Info[:len(addr)]),
				info.DoublePtQual(u.Info[len(addr)]),
				CP24Time2a(u.Info[len(addr)+1:len(addr)+4]),
			)
			u.Info = u.Info[len(addr)+4:]
		}

	case info.M_DP_TB_1: // double-point with 7 octet time-tag
		if u.Var&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_DP_TB_1 not allowed")
		}
		if len(u.Info) != u.Var.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+8 {
			mon.DoublePtWithDate(u, Object(u.Info[:len(addr)]),
				info.DoublePtQual(u.Info[len(addr)]),
				CP56Time2a(u.Info[len(addr)+1:len(addr)+8]),
			)
			u.Info = u.Info[len(addr)+8:]
		}

	case info.M_ST_NA_1: // step position
		if u.Var&info.Sequence != 0 {
			if len(u.Info) != len(addr)+u.Var.Count()*2 {
				return errInfoSize
			}
			addr := Object(u.Info[:len(addr)])
			for i := len(addr); i+2 <= len(u.Info); i += 2 {
				mon.Step(u, addr, info.StepQual(u.Info[i:i+2]))
				addr, _ = u.Params.AddrOf(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Var.Count()*(len(addr)+2) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+2 {
				mon.Step(u, Object(u.Info[:len(addr)]),
					info.StepQual(u.Info[len(addr):len(addr)+2]),
				)
				u.Info = u.Info[len(addr)+2:]
			}
		}

	case info.M_ST_TA_1: // step position with 3 octet time-tag
		if u.Var&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ST_TA_1 not allowed")
		}
		if len(u.Info) != u.Var.Count()*(len(addr)+5) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+5 {
			mon.StepWithTime(u, Object(u.Info[:len(addr)]),
				info.StepQual(u.Info[len(addr):len(addr)+2]),
				CP24Time2a(u.Info[len(addr)+2:len(addr)+5]),
			)
			u.Info = u.Info[len(addr)+5:]
		}

	case info.M_ST_TB_1: // step position with 7 octet time-tag
		if u.Var&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ST_TB_1 not allowed")
		}
		if len(u.Info) != u.Var.Count()*(len(addr)+9) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+9 {
			mon.StepWithDate(u, Object(u.Info[:len(addr)]),
				info.StepQual(u.Info[len(addr):len(addr)+2]),
				CP56Time2a(u.Info[len(addr)+2:len(addr)+9]),
			)
			u.Info = u.Info[len(addr)+9:]
		}

	case info.M_BO_NA_1: // bitstring
		if u.Var&info.Sequence != 0 {
			if len(u.Info) != len(addr)+u.Var.Count()*5 {
				return errInfoSize
			}
			addr := Object(u.Info[:len(addr)])
			for i := len(addr); i+5 <= len(u.Info); i += 5 {
				mon.Bits(u, addr, info.BitsQual(u.Info[i:i+5]))
				addr, _ = u.Params.AddrOf(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Var.Count()*(len(addr)+5) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+5 {
				mon.Bits(u, Object(u.Info),
					info.BitsQual(u.Info[len(addr):len(addr)+5]),
				)
				u.Info = u.Info[len(addr)+5:]
			}
		}

	case info.M_BO_TA_1: // bit string with 3 octet time-tag
		if u.Var&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_BO_TA_1 not allowed")
		}
		if len(u.Info) != u.Var.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+8 {
			mon.BitsWithTime(u, Object(u.Info[:len(addr)]),
				info.BitsQual(u.Info[len(addr):len(addr)+5]),
				CP24Time2a(u.Info[len(addr)+5:len(addr)+8]),
			)
			u.Info = u.Info[len(addr)+8:]
		}

	case info.M_BO_TB_1: // bit string with 7 octet time-tag
		if u.Var&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_BO_TB_1 not allowed")
		}
		if len(u.Info) != u.Var.Count()*(len(addr)+12) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+12 {
			mon.BitsWithDate(u, Object(u.Info[:len(addr)]),
				info.BitsQual(u.Info[len(addr):len(addr)+5]),
				CP56Time2a(u.Info[len(addr)+5:len(addr)+12]),
			)
			u.Info = u.Info[len(addr)+12:]
		}

	case info.M_ME_ND_1: // normalized value without quality descriptor
		if u.Var&info.Sequence != 0 {
			if len(u.Info) != len(addr)+u.Var.Count()*2 {
				return errInfoSize
			}
			addr := Object(u.Info[:len(addr)])
			for i := len(addr); i+1 < len(u.Info); i += 2 {
				mon.NormUnqual(u, addr, info.Norm(u.Info[i:i+2]))
				addr, _ = u.Params.AddrOf(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Var.Count()*(len(addr)+2) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+2 {
				mon.NormUnqual(u, Object(u.Info[:len(addr)]),
					info.Norm(u.Info[len(addr):len(addr)+2]),
				)
				u.Info = u.Info[len(addr)+2:]
			}
		}

	case info.M_ME_NA_1: // normalized value
		if u.Var&info.Sequence != 0 {
			if len(u.Info) != len(addr)+u.Var.Count()*3 {
				return errInfoSize
			}
			addr := Object(u.Info[:len(addr)])
			for i := len(addr); i+3 <= len(u.Info); i += 3 {
				mon.Norm(u, addr, info.NormQual(u.Info[i:i+3]))
				addr, _ = u.Params.AddrOf(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Var.Count()*(len(addr)+3) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+3 {
				mon.Norm(u, Object(u.Info[:len(addr)]),
					info.NormQual(u.Info[len(addr):len(addr)+3]),
				)
				u.Info = u.Info[len(addr)+3:]
			}
		}

	case info.M_ME_TA_1: // normalized value with 3 octet time-tag
		if u.Var&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ME_TA_1 not allowed")
		}
		if len(u.Info) != u.Var.Count()*(len(addr)+6) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+6 {
			mon.NormWithTime(u, Object(u.Info[:len(addr)]),
				info.NormQual(u.Info[len(addr):len(addr)+3]),
				CP24Time2a(u.Info[len(addr)+3:len(addr)+6]),
			)
			u.Info = u.Info[len(addr)+6:]
		}

	case info.M_ME_TD_1: // normalized value with 7 octet time-tag
		if u.Var&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ME_TD_1 not allowed")
		}
		if len(u.Info) != u.Var.Count()*(len(addr)+10) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+10 {
			mon.NormWithDate(u, Object(u.Info[:len(addr)]),
				info.NormQual(u.Info[len(addr):len(addr)+3]),
				CP56Time2a(u.Info[len(addr)+3:len(addr)+10]),
			)
			u.Info = u.Info[len(addr)+10:]
		}

	case info.M_ME_NB_1: // scaled value
		if u.Var&info.Sequence != 0 {
			if len(u.Info) != len(addr)+u.Var.Count()*3 {
				return errInfoSize
			}
			addr := Object(u.Info[:len(addr)])
			for i := len(addr); i+3 <= len(u.Info); i += 3 {
				mon.Scaled(u, addr,
					int16(binary.LittleEndian.Uint16(u.Info[i:i+2])),
					info.Qual(u.Info[i+2]),
				)
				addr, _ = u.Params.AddrOf(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Var.Count()*(len(addr)+3) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+3 {
				mon.Scaled(u, Object(u.Info[:len(addr)]),
					int16(binary.LittleEndian.Uint16(u.Info[len(addr):len(addr)+2])),
					info.Qual(u.Info[len(addr)+2]),
				)
				u.Info = u.Info[len(addr)+3:]
			}
		}

	case info.M_ME_TB_1: // scaled value with 3 octet time-tag
		if u.Var&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ME_TB_1 not allowed")
		}
		if len(u.Info) != u.Var.Count()*(len(addr)+6) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+6 {
			mon.ScaledWithTime(u, Object(u.Info[:len(addr)]),
				int16(binary.LittleEndian.Uint16(u.Info[len(addr):len(addr)+2])),
				info.Qual(u.Info[len(addr)+2]),
				CP24Time2a(u.Info[len(addr)+3:len(addr)+6]),
			)
			u.Info = u.Info[len(addr)+6:]
		}

	case info.M_ME_TE_1: // scaled value with 7 octet time-tag
		if u.Var&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ME_TE_1 not allowed")
		}
		if len(u.Info) != u.Var.Count()*(len(addr)+10) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+10 {
			mon.ScaledWithDate(u, Object(u.Info[:len(addr)]),
				int16(binary.LittleEndian.Uint16(u.Info[len(addr):len(addr)+2])),
				info.Qual(u.Info[len(addr)+2]),
				CP56Time2a(u.Info[len(addr)+3:len(addr)+10]),
			)
			u.Info = u.Info[len(addr)+10:]
		}

	case info.M_ME_NC_1: // floating-point
		if u.Var&info.Sequence != 0 {
			if len(u.Info) != len(addr)+u.Var.Count()*5 {
				return errInfoSize
			}
			addr := Object(u.Info[:len(addr)])
			for i := len(addr); i+5 <= len(u.Info); i += 5 {
				mon.Float(u, addr,
					math.Float32frombits(binary.LittleEndian.Uint32(u.Info[i:i+4])),
					info.Qual(u.Info[i+4]),
				)
				addr, _ = u.Params.AddrOf(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Var.Count()*(len(addr)+5) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+5 {
				mon.Float(u, Object(u.Info[:len(addr)]),
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
		if u.Var&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ME_TC_1 not allowed")
		}
		if len(u.Info) != u.Var.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+8 {
			mon.FloatWithTime(u, Object(u.Info[:len(addr)]),
				math.Float32frombits(
					binary.LittleEndian.Uint32(
						u.Info[len(addr):len(addr)+4],
					),
				),
				info.Qual(u.Info[len(addr)+4]),
				CP24Time2a(u.Info[len(addr)+5:len(addr)+8]),
			)
			u.Info = u.Info[len(addr)+8:]
		}

	case info.M_ME_TF_1: // floating point with 7 octet time-tag
		if u.Var&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ME_TF_1 not allowed")
		}
		if len(u.Info) != u.Var.Count()*(len(addr)+12) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+12 {
			mon.FloatWithDate(u, Object(u.Info[:len(addr)]),
				math.Float32frombits(
					binary.LittleEndian.Uint32(
						u.Info[len(addr):len(addr)+4],
					),
				),
				info.Qual(u.Info[len(addr)+4]),
				CP56Time2a(u.Info[len(addr)+5:len(addr)+12]),
			)
			u.Info = u.Info[len(addr)+12:]
		}

	}
	return nil
}
