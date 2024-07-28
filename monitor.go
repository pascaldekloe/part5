package part5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/pascaldekloe/part5/info"
)

// The structure of the monitoring interface matches table 16 from companion
// standard 101 closely.
type (
	// A Monitor consumes information in monitor direction as listed in table 8
	// from companion standard 101.
	Monitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
		SinglePtMonitor[Orig, Com, Obj]
		SinglePtChangeMonitor[Orig, Com, Obj]
		DoublePtMonitor[Orig, Com, Obj]
		ProtEquipMonitor[Orig, Com, Obj]
		StepMonitor[Orig, Com, Obj]
		BitsMonitor[Orig, Com, Obj]
		NormMonitor[Orig, Com, Obj]
		ScaledMonitor[Orig, Com, Obj]
		FloatMonitor[Orig, Com, Obj]
	}

	// SinglePtMonitor consumes single points.
	SinglePtMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
		// SinglePt gets called for type identifier 1: M_SP_NA_1.
		SinglePt(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtQual)
		// SinglePtAtMinute gets called for type identifier 2: M_SP_TA_1.
		SinglePtAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtQual, info.CP24Time2a)
		// SinglePtAtMoment gets called for type identifier 30: M_SP_TB_1.
		SinglePtAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtQual, info.CP56Time2a)
	}

	// SinglePtChangeMonitor consumes single points with change detection.
	SinglePtChangeMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
		// SinglePtChangePack gets called for type identifier 20: M_PS_NA_1.
		SinglePtChangePack(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtChangePack, info.Qual)
	}

	// DoublePtMonitor consumes double points.
	DoublePtMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
		// DoublePt gets called for type identifier 3: M_DP_NA_1.
		DoublePt(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual)
		// DoublePtAtMinute gets called for type identifier 4: M_DP_TA_1.
		DoublePtAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual, info.CP24Time2a)
		// DoublePtAtMoment gets called for type identifier 31: M_DP_TB_1.
		DoublePtAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual, info.CP56Time2a)
	}

	// ProtEquipMonitor consumes double points about protection equipment.
	// The uint16 with the ellapsed time should contain milliseconds in
	// range 0..60000. See info.EllapsedTimeInvalid in the DoublePtQual
	// before using such value.
	ProtEquipMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
		// ProtEquipAtMinute gets called for type identifier 17: M_EP_TA_1.
		ProtEquipAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual, uint16, info.CP24Time2a)
		// ProtEquipAtMoment gets called for type identifier 38: M_EP_TD_1.
		ProtEquipAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual, uint16, info.CP56Time2a)
	}

	// StepMonitor consumes step positions.
	StepMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
		// Step gets called for type identifier 5: M_ST_NA_1.
		Step(info.DataUnit[Orig, Com, Obj], Obj, info.StepQual)
		// StepAtMinute gets called for type identifier 6: M_ST_TA_1.
		StepAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.StepQual, info.CP24Time2a)
		// StepAtMoment gets called for type identifier 32: M_ST_TB_1.
		StepAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.StepQual, info.CP56Time2a)
	}

	// BitsMonitor consumes bitstrings.
	BitsMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
		// Bits gets called for type identifier 7: M_BO_NA_1.
		Bits(info.DataUnit[Orig, Com, Obj], Obj, info.BitsQual)
		// BitsAtMinute gets called for type identifier 8: M_BO_TA_1.
		BitsAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.BitsQual, info.CP24Time2a)
		// BitsAtMoment gets called for type identifier 33: M_BO_TB_1.
		BitsAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.BitsQual, info.CP56Time2a)
	}

	// NormMonitor consumes normalized values.
	NormMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
		// NormUnqual gets called for type identifier 21: M_ME_ND_1.
		NormUnqual(info.DataUnit[Orig, Com, Obj], Obj, info.Norm)
		// Norm gets called for type identifier 9: M_ME_NA_1.
		Norm(info.DataUnit[Orig, Com, Obj], Obj, info.NormQual)
		// NormAtMinute gets called for type identifier 10: M_ME_TA_1.
		NormAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.NormQual, info.CP24Time2a)
		// NormAtMoment gets called for type identifier 34: M_ME_TD_1.
		NormAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.NormQual, info.CP56Time2a)
	}

	// ScaledMonitor consumes scaled values.
	ScaledMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
		// Scaled gets called for type identifier 11: M_ME_NB_1.
		Scaled(info.DataUnit[Orig, Com, Obj], Obj, int16, info.Qual)
		// ScaledAtMinute gets called for type identifier 12: M_ME_TB_1.
		ScaledAtMinute(info.DataUnit[Orig, Com, Obj], Obj, int16, info.Qual, info.CP24Time2a)
		// ScaledAtMoment gets called for type identifier 35: M_ME_TE_1.
		ScaledAtMoment(info.DataUnit[Orig, Com, Obj], Obj, int16, info.Qual, info.CP56Time2a)
	}

	// FloatMonitor consumes floating points.
	FloatMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
		// Float gets called for type identifier 13: M_ME_NC_1.
		Float(info.DataUnit[Orig, Com, Obj], Obj, float32, info.Qual)
		// FloatAtMinute gets called for type identifier 14: M_ME_TC_1.
		FloatAtMinute(info.DataUnit[Orig, Com, Obj], Obj, float32, info.Qual, info.CP24Time2a)
		// FloatAtMoment gets called for type identifier 36: M_ME_TF_1.
		FloatAtMoment(info.DataUnit[Orig, Com, Obj], Obj, float32, info.Qual, info.CP56Time2a)
	}
)

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

func (l logger[Orig, Com, Obj]) SinglePtAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.SinglePtQual, tag info.CP24Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(), tag)
}

func (l logger[Orig, Com, Obj]) SinglePtAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.SinglePtQual, tag info.CP56Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(), tag)
}

func (l logger[Orig, Com, Obj]) SinglePtChangePack(u info.DataUnit[Orig, Com, Obj], addr Obj, pts info.SinglePtChangePack, q info.Qual) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %016b~%016b %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, pts>>16, pts&0xffff, q)
}

func (l logger[Orig, Com, Obj]) DoublePt(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.DoublePtQual) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual())
}

func (l logger[Orig, Com, Obj]) DoublePtAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.DoublePtQual, tag info.CP24Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(), tag)
}

func (l logger[Orig, Com, Obj]) DoublePtAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.DoublePtQual, tag info.CP56Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(), tag)
}

func (l logger[Orig, Com, Obj]) ProtEquipAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.DoublePtQual, ms uint16, tag info.CP24Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %dms %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(), ms, tag)
}

func (l logger[Orig, Com, Obj]) ProtEquipAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.DoublePtQual, ms uint16, tag info.CP56Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %dms %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(), ms, tag)
}

func (l logger[Orig, Com, Obj]) Step(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.StepQual) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual())
}

func (l logger[Orig, Com, Obj]) StepAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.StepQual, tag info.CP24Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(), tag)
}

func (l logger[Orig, Com, Obj]) StepAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.StepQual, tag info.CP56Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Value(), p.Qual(), tag)
}

func (l logger[Orig, Com, Obj]) Bits(u info.DataUnit[Orig, Com, Obj], addr Obj, b info.BitsQual) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %#x %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, b.Array(), b.Qual())
}

func (l logger[Orig, Com, Obj]) BitsAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, b info.BitsQual, tag info.CP24Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %#x %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, b.Array(), b.Qual(), tag)
}

func (l logger[Orig, Com, Obj]) BitsAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, b info.BitsQual, tag info.CP56Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %#x %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, b.Array(), b.Qual(), tag)
}

func (l logger[Orig, Com, Obj]) NormUnqual(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.Norm) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %f\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, n.Float64())
}

func (l logger[Orig, Com, Obj]) Norm(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.NormQual) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %f %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, n.Link().Float64(), n.Qual())
}

func (l logger[Orig, Com, Obj]) NormAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.NormQual, tag info.CP24Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %f %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, n.Link().Float64(), n.Qual(), tag)
}

func (l logger[Orig, Com, Obj]) NormAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.NormQual, tag info.CP56Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %f %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, n.Link().Float64(), n.Qual(), tag)
}

func (l logger[Orig, Com, Obj]) Scaled(u info.DataUnit[Orig, Com, Obj], addr Obj, v int16, q info.Qual) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %d %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, v, q)
}

func (l logger[Orig, Com, Obj]) ScaledAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, v int16, q info.Qual, tag info.CP24Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %d %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, v, q, tag)
}

func (l logger[Orig, Com, Obj]) ScaledAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, v int16, q info.Qual, tag info.CP56Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %d %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, v, q, tag)
}

func (l logger[Orig, Com, Obj]) Float(u info.DataUnit[Orig, Com, Obj], addr Obj, f float32, q info.Qual) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %g %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, f, q)
}

func (l logger[Orig, Com, Obj]) FloatAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, f float32, q info.Qual, tag info.CP24Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %g %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, f, q, tag)
}

func (l logger[Orig, Com, Obj]) FloatAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, f float32, q info.Qual, tag info.CP56Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %g %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, f, q, tag)
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
			for i := 0; i+len(addr)+1 <= len(u.Info); i += len(addr) + 1 {
				mon.SinglePt(u,
					Obj(u.Info[i:i+len(addr)]),
					info.SinglePtQual(u.Info[i+len(addr)]),
				)
			}
		}

	case info.M_SP_TA_1: // single-point with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_SP_TA_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+4) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+4 <= len(u.Info); i += len(addr) + 4 {
			mon.SinglePtAtMinute(u,
				Obj(u.Info[i:i+len(addr)]),
				info.SinglePtQual(u.Info[i+len(addr)]),
				info.CP24Time2a(u.Info[i+len(addr)+1:i+len(addr)+4]),
			)
		}

	case info.M_SP_TB_1: // single-point with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_SP_TB_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+8 <= len(u.Info); i += len(addr) + 8 {
			mon.SinglePtAtMoment(u,
				Obj(u.Info[i:i+len(addr)]),
				info.SinglePtQual(u.Info[i+len(addr)]),
				info.CP56Time2a(u.Info[i+len(addr)+1:i+len(addr)+8]),
			)
		}

	case info.M_PS_NA_1: // single-points with status change
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 5)
			if err != nil {
				return err
			}
			for i := len(addr); i+5 <= len(u.Info); i += 5 {
				mon.SinglePtChangePack(u, addr,
					info.SinglePtChangePack(
						binary.BigEndian.Uint32(
							u.Info[i:i+4],
						),
					),
					info.Qual(u.Info[i+4]),
				)
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+5) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+5 <= len(u.Info); i += len(addr) + 5 {
				mon.SinglePtChangePack(u,
					Obj(u.Info[i:i+len(addr)]),
					info.SinglePtChangePack(
						binary.BigEndian.Uint32(
							u.Info[i+len(addr):i+len(addr)+4],
						),
					),
					info.Qual(u.Info[i+len(addr)+4]),
				)
			}
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
			for i := 0; i+len(addr)+1 <= len(u.Info); i += len(addr) + 1 {
				mon.DoublePt(u,
					Obj(u.Info[i:i+len(addr)]),
					info.DoublePtQual(u.Info[i+len(addr)]),
				)
			}
		}

	case info.M_DP_TA_1: // double-point with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_DP_TA_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+4) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+4 <= len(u.Info); i += len(addr) + 4 {
			mon.DoublePtAtMinute(u,
				Obj(u.Info[i:i+len(addr)]),
				info.DoublePtQual(u.Info[i+len(addr)]),
				info.CP24Time2a(u.Info[i+len(addr)+1:i+len(addr)+4]),
			)
		}

	case info.M_DP_TB_1: // double-point with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_DP_TB_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+8 <= len(u.Info); i += len(addr) + 8 {
			mon.DoublePtAtMoment(u,
				Obj(u.Info[i:i+len(addr)]),
				info.DoublePtQual(u.Info[i+len(addr)]),
				info.CP56Time2a(u.Info[i+len(addr)+1:i+len(addr)+8]),
			)
		}

	case info.M_EP_TA_1: // protection equipment with 3-octet time tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_EP_TA_1 not allowed")
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+6) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+6 <= len(u.Info); i += len(addr) + 6 {
				mon.ProtEquipAtMinute(u,
					Obj(u.Info[i:i+len(addr)]),
					info.DoublePtQual(u.Info[i+len(addr)]),
					binary.BigEndian.Uint16(u.Info[i+len(addr)+1:i+len(addr)+3]),
					info.CP24Time2a(u.Info[i+len(addr)+3:i+len(addr)+6]),
				)
			}
		}

	case info.M_EP_TD_1: // protection equipment with 7-octet time tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_EP_TD_1 not allowed")
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+10) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+10 <= len(u.Info); i += len(addr) + 10 {
				mon.ProtEquipAtMoment(u,
					Obj(u.Info[i:i+len(addr)]),
					info.DoublePtQual(u.Info[i+len(addr)]),
					binary.BigEndian.Uint16(u.Info[i+len(addr)+1:i+len(addr)+3]),
					info.CP56Time2a(u.Info[i+len(addr)+3:i+len(addr)+10]),
				)
			}
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
			for i := 0; i+len(addr)+2 <= len(u.Info); i += len(addr) + 2 {
				mon.Step(u,
					Obj(u.Info[i:i+len(addr)]),
					info.StepQual(u.Info[i+len(addr):i+len(addr)+2]),
				)
			}
		}

	case info.M_ST_TA_1: // step position with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ST_TA_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+5) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+5 <= len(u.Info); i += len(addr) + 5 {
			mon.StepAtMinute(u,
				Obj(u.Info[i:i+len(addr)]),
				info.StepQual(u.Info[i+len(addr):i+len(addr)+2]),
				info.CP24Time2a(u.Info[i+len(addr)+2:i+len(addr)+5]),
			)
		}

	case info.M_ST_TB_1: // step position with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ST_TB_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+9) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+9 <= len(u.Info); i += len(addr) + 9 {
			mon.StepAtMoment(u,
				Obj(u.Info[i:i+len(addr)]),
				info.StepQual(u.Info[i+len(addr):i+len(addr)+2]),
				info.CP56Time2a(u.Info[i+len(addr)+2:i+len(addr)+9]),
			)
		}

	case info.M_BO_NA_1: // bit string
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
			for i := 0; i+len(addr)+5 <= len(u.Info); i += len(addr) + 5 {
				mon.Bits(u,
					Obj(u.Info[i:i+len(addr)]),
					info.BitsQual(u.Info[i+len(addr):i+len(addr)+5]),
				)
			}
		}

	case info.M_BO_TA_1: // bit string with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_BO_TA_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+8 <= len(u.Info); i += len(addr) + 8 {
			mon.BitsAtMinute(u,
				Obj(u.Info[i:i+len(addr)]),
				info.BitsQual(u.Info[i+len(addr):i+len(addr)+5]),
				info.CP24Time2a(u.Info[i+len(addr)+5:i+len(addr)+8]),
			)
		}

	case info.M_BO_TB_1: // bit string with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_BO_TB_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+12) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+12 <= len(u.Info); i += len(addr) + 12 {
			mon.BitsAtMoment(u,
				Obj(u.Info[i:i+len(addr)]),
				info.BitsQual(u.Info[i+len(addr):i+len(addr)+5]),
				info.CP56Time2a(u.Info[i+len(addr)+5:i+len(addr)+12]),
			)
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
			for i := 0; i+len(addr)+2 <= len(u.Info); i += len(addr) + 2 {
				mon.NormUnqual(u,
					Obj(u.Info[i:i+len(addr)]),
					info.Norm(u.Info[i+len(addr):i+len(addr)+2]),
				)
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
			for i := 0; i+len(addr)+3 <= len(u.Info); i += len(addr) + 3 {
				mon.Norm(u,
					Obj(u.Info[i:i+len(addr)]),
					info.NormQual(u.Info[i+len(addr):i+len(addr)+3]),
				)
			}
		}

	case info.M_ME_TA_1: // normalized value with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ME_TA_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+6) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+6 <= len(u.Info); i += len(addr) + 6 {
			mon.NormAtMinute(u,
				Obj(u.Info[i:i+len(addr)]),
				info.NormQual(u.Info[i+len(addr):i+len(addr)+3]),
				info.CP24Time2a(u.Info[i+len(addr)+3:i+len(addr)+6]),
			)
		}

	case info.M_ME_TD_1: // normalized value with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ME_TD_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+10) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+10 <= len(u.Info); i += len(addr) + 10 {
			mon.NormAtMoment(u,
				Obj(u.Info[i:i+len(addr)]),
				info.NormQual(u.Info[i+len(addr):i+len(addr)+3]),
				info.CP56Time2a(u.Info[i+len(addr)+3:i+len(addr)+10]),
			)
		}

	case info.M_ME_NB_1: // scaled value
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 3)
			if err != nil {
				return err
			}
			for i := len(addr); i+3 <= len(u.Info); i += 3 {
				mon.Scaled(u, addr,
					int16(
						binary.LittleEndian.Uint16(
							u.Info[i:i+2],
						),
					),
					info.Qual(u.Info[i+2]),
				)
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+3) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+3 <= len(u.Info); i += len(addr) + 3 {
				mon.Scaled(u,
					Obj(u.Info[i:i+len(addr)]),
					int16(
						binary.LittleEndian.Uint16(
							u.Info[i+len(addr):i+len(addr)+2],
						),
					),
					info.Qual(u.Info[i+len(addr)+2]),
				)
			}
		}

	case info.M_ME_TB_1: // scaled value with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ME_TB_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+6) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+6 <= len(u.Info); i += len(addr) + 6 {
			mon.ScaledAtMinute(u,
				Obj(u.Info[i:i+len(addr)]),
				int16(
					binary.LittleEndian.Uint16(
						u.Info[i+len(addr):i+len(addr)+2],
					),
				),
				info.Qual(u.Info[i+len(addr)+2]),
				info.CP24Time2a(u.Info[i+len(addr)+3:i+len(addr)+6]),
			)
		}

	case info.M_ME_TE_1: // scaled value with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ME_TE_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+10) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+10 <= len(u.Info); i += len(addr) + 10 {
			mon.ScaledAtMoment(u,
				Obj(u.Info[i:i+len(addr)]),
				int16(
					binary.LittleEndian.Uint16(
						u.Info[i+len(addr):i+len(addr)+2],
					),
				),
				info.Qual(u.Info[i+len(addr)+2]),
				info.CP56Time2a(u.Info[i+len(addr)+3:i+len(addr)+10]),
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
					math.Float32frombits(
						binary.LittleEndian.Uint32(
							u.Info[i:i+4],
						),
					),
					info.Qual(u.Info[i+4]),
				)
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+5) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+5 <= len(u.Info); i += len(addr) + 5 {
				mon.Float(u,
					Obj(u.Info[i:i+len(addr)]),
					math.Float32frombits(
						binary.LittleEndian.Uint32(
							u.Info[i+len(addr):i+len(addr)+4],
						),
					),
					info.Qual(u.Info[i+len(addr)+4]),
				)
			}
		}

	case info.M_ME_TC_1: // floating point with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ME_TC_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+8 <= len(u.Info); i += len(addr) + 8 {
			mon.FloatAtMinute(u,
				Obj(u.Info[i:i+len(addr)]),
				math.Float32frombits(
					binary.LittleEndian.Uint32(
						u.Info[i+len(addr):i+len(addr)+4],
					),
				),
				info.Qual(u.Info[i+len(addr)+4]),
				info.CP24Time2a(u.Info[i+len(addr)+5:i+len(addr)+8]),
			)
		}

	case info.M_ME_TF_1: // floating point with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ME_TF_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+12) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+12 <= len(u.Info); i += len(addr) + 12 {
			mon.FloatAtMoment(u,
				Obj(u.Info[i:i+len(addr)]),
				math.Float32frombits(
					binary.LittleEndian.Uint32(
						u.Info[i+len(addr):i+len(addr)+4],
					),
				),
				info.Qual(u.Info[i+len(addr)+4]),
				info.CP56Time2a(u.Info[i+len(addr)+5:i+len(addr)+12]),
			)
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
