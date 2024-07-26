package part5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/pascaldekloe/part5/info"
)

// A Monitor accepts information in monitor direction as listed in table 8 from
// companion standard 101.
type Monitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	SinglePtMonitor[Orig, Com, Obj]
	DoublePtMonitor[Orig, Com, Obj]
	StepMonitor[Orig, Com, Obj]
	BitsMonitor[Orig, Com, Obj]
	NormMonitor[Orig, Com, Obj]
	ScaledMonitor[Orig, Com, Obj]
	FloatMonitor[Orig, Com, Obj]
}

// SinglePtMonitor accepts single points.
type SinglePtMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// SinglePt gets called for type identifier 1: M_SP_NA_1.
	SinglePt(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtQual)
	// SinglePtAtMinute gets called for type identifier 2: M_SP_TA_1.
	SinglePtAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtQual, info.CP24Time2a)
	// SinglePtAtMoment gets called for type identifier 30: M_SP_TB_1.
	SinglePtAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtQual, info.CP56Time2a)
}

// DoublePtMonitor accepts double points.
type DoublePtMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// DoublePt gets called for type identifier 3: M_DP_NA_1.
	DoublePt(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual)
	// DoublePtAtMinute gets called for type identifier 4: M_DP_TA_1.
	DoublePtAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual, info.CP24Time2a)
	// DoublePtAtMoment gets called for type identifier 31: M_DP_TB_1.
	DoublePtAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual, info.CP56Time2a)
}

// StepMonitor accepts step positions.
type StepMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// Step gets called for type identifier 5: M_ST_NA_1.
	Step(info.DataUnit[Orig, Com, Obj], Obj, info.StepQual)
	// StepAtMinute gets called for type identifier 6: M_ST_TA_1.
	StepAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.StepQual, info.CP24Time2a)
	// StepAtMoment gets called for type identifier 32: M_ST_TB_1.
	StepAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.StepQual, info.CP56Time2a)
}

// BitsMonitor accepts bitstrings.
type BitsMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// Bits gets called for type identifier 7: M_BO_NA_1.
	Bits(info.DataUnit[Orig, Com, Obj], Obj, info.BitsQual)
	// BitsAtMinute gets called for type identifier 8: M_BO_TA_1.
	BitsAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.BitsQual, info.CP24Time2a)
	// BitsAtMoment gets called for type identifier 33: M_BO_TB_1.
	BitsAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.BitsQual, info.CP56Time2a)
}

// NormMonitor accepts normalized values.
type NormMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// NormUnqual gets called for type identifier 21: M_ME_ND_1.
	NormUnqual(info.DataUnit[Orig, Com, Obj], Obj, info.Norm)
	// Norm gets called for type identifier 9: M_ME_NA_1.
	Norm(info.DataUnit[Orig, Com, Obj], Obj, info.NormQual)
	// NormAtMinute gets called for type identifier 10: M_ME_TA_1.
	NormAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.NormQual, info.CP24Time2a)
	// NormAtMoment gets called for type identifier 34: M_ME_TD_1.
	NormAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.NormQual, info.CP56Time2a)
}

// ScaledMonitor accepts scaled values.
type ScaledMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// Scaled gets called for type identifier 11: M_ME_NB_1.
	Scaled(info.DataUnit[Orig, Com, Obj], Obj, int16, info.Qual)
	// ScaledAtMinute gets called for type identifier 12: M_ME_TB_1.
	ScaledAtMinute(info.DataUnit[Orig, Com, Obj], Obj, int16, info.Qual, info.CP24Time2a)
	// ScaledAtMoment gets called for type identifier 35: M_ME_TE_1.
	ScaledAtMoment(info.DataUnit[Orig, Com, Obj], Obj, int16, info.Qual, info.CP56Time2a)
}

// FloatMonitor accepts floating points.
type FloatMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// Float gets called for type identifier 13: M_ME_NC_1.
	Float(info.DataUnit[Orig, Com, Obj], Obj, float32, info.Qual)
	// FloatAtMinute gets called for type identifier 14: M_ME_TC_1.
	FloatAtMinute(info.DataUnit[Orig, Com, Obj], Obj, float32, info.Qual, info.CP24Time2a)
	// FloatAtMoment gets called for type identifier 36: M_ME_TF_1.
	FloatAtMoment(info.DataUnit[Orig, Com, Obj], Obj, float32, info.Qual, info.CP56Time2a)
}

type logger[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	W io.Writer
}

// NewLogger returns a Monitor which writes on each invocation as a text line in
// a human readable formon.
func NewLogger[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](_ info.Params[Orig, Com, Obj], w io.Writer) Monitor[Orig, Com, Obj] {
	return logger[Orig, Com, Obj]{w}
}

func (l logger[Orig, Com, Obj]) SinglePt(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.SinglePtQual) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual())
}

func (l logger[Orig, Com, Obj]) SinglePtAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.SinglePtQual, t info.CP24Time2a) {
	min, secInMilli := t.MinuteAndMillis()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s :%02d:%02d.%03d\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(),
		min, secInMilli/1000, secInMilli%1000)
}

func (l logger[Orig, Com, Obj]) SinglePtAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.SinglePtQual, t info.CP56Time2a) {
	y, m, d := t.Calendar()
	H, M, secInMilli := t.ClockAndMillis()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(),
		y, m, d, H, M, secInMilli/1000, secInMilli%1000)
}

func (l logger[Orig, Com, Obj]) DoublePt(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.DoublePtQual) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual())
}

func (l logger[Orig, Com, Obj]) DoublePtAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.DoublePtQual, t info.CP24Time2a) {
	min, secInMilli := t.MinuteAndMillis()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s :%02d:%02d.%03d\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(),
		min, secInMilli/1000, secInMilli%1000)
}

func (l logger[Orig, Com, Obj]) DoublePtAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.DoublePtQual, t info.CP56Time2a) {
	y, m, d := t.Calendar()
	H, M, secInMilli := t.ClockAndMillis()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(),
		y, m, d, H, M, secInMilli/1000, secInMilli%1000)
}

func (l logger[Orig, Com, Obj]) Step(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.StepQual) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual())
}

func (l logger[Orig, Com, Obj]) StepAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.StepQual, t info.CP24Time2a) {
	min, secInMilli := t.MinuteAndMillis()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s :%02d:%02d.%03d\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(),
		min, secInMilli/1000, secInMilli%1000)
}

func (l logger[Orig, Com, Obj]) StepAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.StepQual, t info.CP56Time2a) {
	y, m, d := t.Calendar()
	H, M, secInMilli := t.ClockAndMillis()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(),
		y, m, d, H, M, secInMilli/1000, secInMilli%1000)
}

func (l logger[Orig, Com, Obj]) Bits(u info.DataUnit[Orig, Com, Obj], addr Obj, b info.BitsQual) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %#x %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, b.Array(), b.Qual())
}

func (l logger[Orig, Com, Obj]) BitsAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, b info.BitsQual, t info.CP24Time2a) {
	min, secInMilli := t.MinuteAndMillis()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %#x %s :%02d:%02d.%03d\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, b.Array(), b.Qual(),
		min, secInMilli/1000, secInMilli%1000)
}

func (l logger[Orig, Com, Obj]) BitsAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, b info.BitsQual, t info.CP56Time2a) {
	y, m, d := t.Calendar()
	H, M, secInMilli := t.ClockAndMillis()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %#x %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, b.Array(), b.Qual(),
		y, m, d, H, M, secInMilli/1000, secInMilli%1000)
}

func (l logger[Orig, Com, Obj]) NormUnqual(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.Norm) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %f\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, n.Float64())
}

func (l logger[Orig, Com, Obj]) Norm(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.NormQual) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %f %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, n.Link().Float64(), n.Qual())
}

func (l logger[Orig, Com, Obj]) NormAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.NormQual, t info.CP24Time2a) {
	min, secInMilli := t.MinuteAndMillis()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %f %s :%02d:%02d.%03d\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, n.Link().Float64(), n.Qual(),
		min, secInMilli/1000, secInMilli%1000)
}

func (l logger[Orig, Com, Obj]) NormAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.NormQual, t info.CP56Time2a) {
	y, m, d := t.Calendar()
	H, M, secInMilli := t.ClockAndMillis()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %f %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, n.Link().Float64(), n.Qual(),
		y, m, d, H, M, secInMilli/1000, secInMilli%1000)
}

func (l logger[Orig, Com, Obj]) Scaled(u info.DataUnit[Orig, Com, Obj], addr Obj, v int16, q info.Qual) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %d %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, v, q)
}

func (l logger[Orig, Com, Obj]) ScaledAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, v int16, q info.Qual, t info.CP24Time2a) {
	min, secInMilli := t.MinuteAndMillis()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %d %s :%02d:%02d.%03d\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, v, q,
		min, secInMilli/1000, secInMilli%1000)
}

func (l logger[Orig, Com, Obj]) ScaledAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, v int16, q info.Qual, t info.CP56Time2a) {
	y, m, d := t.Calendar()
	H, M, secInMilli := t.ClockAndMillis()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %d %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, v, q,
		y, m, d, H, M, secInMilli/1000, secInMilli%1000)
}

func (l logger[Orig, Com, Obj]) Float(u info.DataUnit[Orig, Com, Obj], addr Obj, f float32, q info.Qual) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %g %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, f, q)
}

func (l logger[Orig, Com, Obj]) FloatAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, f float32, q info.Qual, t info.CP24Time2a) {
	min, secInMilli := t.MinuteAndMillis()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %g %s :%02d:%02d.%03d\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, f, q,
		min, secInMilli/1000, secInMilli%1000)
}

func (l logger[Orig, Com, Obj]) FloatAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, f float32, q info.Qual, t info.CP56Time2a) {
	y, m, d := t.Calendar()
	H, M, secInMilli := t.ClockAndMillis()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %g %s %02d-%02d-%02dT%02d:%02d:%02d.%03d\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, f, q,
		y, m, d, H, M, secInMilli/1000, secInMilli%1000)
}

// ErrNotMonitor rejects an info.DataUnit based on its type identifier.
var ErrNotMonitor = errors.New("part5: ASDU type identifier not in monitor information range 1..44")

var errInfoSize = errors.New("part5: size of ASDU payload doesn't match the variable structure qualifier")

// ApplyDataUnit propagates u to the corresponding mon method. Note that one
// DataUnit can have multiple information elements, in which case the interface
// is invoked once for each element in order of appearance. DataUnits with zero
// information elements do not cause any listener invocation. Errors other than
// ErrNotMonitor reject malformed content in the DataUnit.
func ApplyDataUnit[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](mon Monitor[Orig, Com, Obj], u info.DataUnit[Orig, Com, Obj]) error {
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
