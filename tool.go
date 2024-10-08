package part5

import (
	"fmt"
	"io"

	"github.com/pascaldekloe/part5/info"
)

// MonitorDelegate passes the Monitor interface to any its sub-interfaces. All
// fields are optional. Nil causes silent discards.
type MonitorDelegate[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	SinglePtMonitor[Orig, Com, Obj]
	SinglePtChangeMonitor[Orig, Com, Obj]
	DoublePtMonitor[Orig, Com, Obj]
	StepMonitor[Orig, Com, Obj]
	BitsMonitor[Orig, Com, Obj]
	NormMonitor[Orig, Com, Obj]
	ScaledMonitor[Orig, Com, Obj]
	FloatMonitor[Orig, Com, Obj]
	TotalsMonitor[Orig, Com, Obj]
	ProtectMonitor[Orig, Com, Obj]
	ProtectStartMonitor[Orig, Com, Obj]
	ProtectOutMonitor[Orig, Com, Obj]
	InitEndMonitor[Orig, Com, Obj]
}

// NewMonitorDelegate returns a new delegate with each sub-interface nil.
func NewMonitorDelegate[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](_ info.System[Orig, Com, Obj]) *MonitorDelegate[Orig, Com, Obj] {
	return NewMonitorDelegateDefault[Orig, Com, Obj](nil)
}

// NewMonitorDelegateDefault returns a new delegate with each sub-interface set
// to a def(ault) value. Note that def may be nil.
func NewMonitorDelegateDefault[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](def Monitor[Orig, Com, Obj]) *MonitorDelegate[Orig, Com, Obj] {
	return &MonitorDelegate[Orig, Com, Obj]{
		SinglePtMonitor:       def,
		SinglePtChangeMonitor: def,
		DoublePtMonitor:       def,
		StepMonitor:           def,
		BitsMonitor:           def,
		NormMonitor:           def,
		ScaledMonitor:         def,
		FloatMonitor:          def,
		TotalsMonitor:         def,
		ProtectMonitor:        def,
		ProtectStartMonitor:   def,
		ProtectOutMonitor:     def,
		InitEndMonitor:        def,
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) SinglePt(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.SinglePtQual) {
	if del.SinglePtMonitor != nil {
		del.SinglePtMonitor.SinglePt(u, addr, p)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) SinglePtAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.SinglePtQual, tag info.CP24Time2a) {
	if del.SinglePtMonitor != nil {
		del.SinglePtMonitor.SinglePtAtMinute(u, addr, p, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) SinglePtAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.SinglePtQual, tag info.CP56Time2a) {
	if del.SinglePtMonitor != nil {
		del.SinglePtMonitor.SinglePtAtMoment(u, addr, p, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) SinglePtChangePack(u info.DataUnit[Orig, Com, Obj], addr Obj, pack info.SinglePtChangePack, q info.Qual) {
	if del.SinglePtChangeMonitor != nil {
		del.SinglePtChangeMonitor.SinglePtChangePack(u, addr, pack, q)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) DoublePt(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.DoublePtQual) {
	if del.DoublePtMonitor != nil {
		del.DoublePtMonitor.DoublePt(u, addr, p)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) DoublePtAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.DoublePtQual, tag info.CP24Time2a) {
	if del.DoublePtMonitor != nil {
		del.DoublePtMonitor.DoublePtAtMinute(u, addr, p, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) DoublePtAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.DoublePtQual, tag info.CP56Time2a) {
	if del.DoublePtMonitor != nil {
		del.DoublePtMonitor.DoublePtAtMoment(u, addr, p, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) Step(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.StepQual) {
	if del.StepMonitor != nil {
		del.StepMonitor.Step(u, addr, p)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) StepAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.StepQual, tag info.CP24Time2a) {
	if del.StepMonitor != nil {
		del.StepMonitor.StepAtMinute(u, addr, p, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) StepAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.StepQual, tag info.CP56Time2a) {
	if del.StepMonitor != nil {
		del.StepMonitor.StepAtMoment(u, addr, p, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) Bits(u info.DataUnit[Orig, Com, Obj], addr Obj, b info.BitsQual) {
	if del.BitsMonitor != nil {
		del.BitsMonitor.Bits(u, addr, b)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) BitsAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, b info.BitsQual, tag info.CP24Time2a) {
	if del.BitsMonitor != nil {
		del.BitsMonitor.BitsAtMinute(u, addr, b, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) BitsAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, b info.BitsQual, tag info.CP56Time2a) {
	if del.BitsMonitor != nil {
		del.BitsMonitor.BitsAtMoment(u, addr, b, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) NormUnqual(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.Norm) {
	if del.NormMonitor != nil {
		del.NormMonitor.NormUnqual(u, addr, n)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) Norm(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.NormQual) {
	if del.NormMonitor != nil {
		del.NormMonitor.Norm(u, addr, n)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) NormAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.NormQual, tag info.CP24Time2a) {
	if del.NormMonitor != nil {
		del.NormMonitor.NormAtMinute(u, addr, n, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) NormAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.NormQual, tag info.CP56Time2a) {
	if del.NormMonitor != nil {
		del.NormMonitor.NormAtMoment(u, addr, n, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) Scaled(u info.DataUnit[Orig, Com, Obj], addr Obj, v int16, q info.Qual) {
	if del.ScaledMonitor != nil {
		del.ScaledMonitor.Scaled(u, addr, v, q)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) ScaledAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, v int16, q info.Qual, tag info.CP24Time2a) {
	if del.ScaledMonitor != nil {
		del.ScaledMonitor.ScaledAtMinute(u, addr, v, q, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) ScaledAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, v int16, q info.Qual, tag info.CP56Time2a) {
	if del.ScaledMonitor != nil {
		del.ScaledMonitor.ScaledAtMoment(u, addr, v, q, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) Float(u info.DataUnit[Orig, Com, Obj], addr Obj, f float32, q info.Qual) {
	if del.FloatMonitor != nil {
		del.FloatMonitor.Float(u, addr, f, q)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) FloatAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, f float32, q info.Qual, tag info.CP24Time2a) {
	if del.FloatMonitor != nil {
		del.FloatMonitor.FloatAtMinute(u, addr, f, q, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) FloatAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, f float32, q info.Qual, tag info.CP56Time2a) {
	if del.FloatMonitor != nil {
		del.FloatMonitor.FloatAtMoment(u, addr, f, q, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) Totals(u info.DataUnit[Orig, Com, Obj], addr Obj, c info.Counter) {
	if del.TotalsMonitor != nil {
		del.TotalsMonitor.Totals(u, addr, c)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) TotalsAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, c info.Counter, tag info.CP24Time2a) {
	if del.TotalsMonitor != nil {
		del.TotalsMonitor.TotalsAtMinute(u, addr, c, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) TotalsAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, c info.Counter, tag info.CP56Time2a) {
	if del.TotalsMonitor != nil {
		del.TotalsMonitor.TotalsAtMoment(u, addr, c, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) ProtectAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, e info.ProtectEvent, tag info.CP24Time2a) {
	if del.ProtectMonitor != nil {
		del.ProtectMonitor.ProtectAtMinute(u, addr, e, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) ProtectAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, e info.ProtectEvent, tag info.CP56Time2a) {
	if del.ProtectMonitor != nil {
		del.ProtectMonitor.ProtectAtMoment(u, addr, e, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) ProtectStartAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, e info.ProtectStartEvent, tag info.CP24Time2a) {
	if del.ProtectStartMonitor != nil {
		del.ProtectStartMonitor.ProtectStartAtMinute(u, addr, e, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) ProtectStartAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, e info.ProtectStartEvent, tag info.CP56Time2a) {
	if del.ProtectStartMonitor != nil {
		del.ProtectStartMonitor.ProtectStartAtMoment(u, addr, e, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) ProtectOutAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, e info.ProtectOutEvent, tag info.CP24Time2a) {
	if del.ProtectOutMonitor != nil {
		del.ProtectOutMonitor.ProtectOutAtMinute(u, addr, e, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) ProtectOutAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, e info.ProtectOutEvent, tag info.CP56Time2a) {
	if del.ProtectOutMonitor != nil {
		del.ProtectOutMonitor.ProtectOutAtMoment(u, addr, e, tag)
	}
}

func (del *MonitorDelegate[Orig, Com, Obj]) InitEnd(u info.DataUnit[Orig, Com, Obj], c info.InitCause) {
	if del.InitEndMonitor != nil {
		del.InitEndMonitor.InitEnd(u, c)
	}
}

type logger[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	W io.Writer
}

// NewLogger returns a Monitor which writes on each invocation as a text line in
// a human readable formon.
func NewLogger[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](_ info.System[Orig, Com, Obj], w io.Writer) Monitor[Orig, Com, Obj] {
	return logger[Orig, Com, Obj]{w}
}

func (l logger[Orig, Com, Obj]) SinglePt(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.SinglePtQual) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Pt(), p.Qual())
}

func (l logger[Orig, Com, Obj]) SinglePtAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.SinglePtQual, tag info.CP24Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Pt(), p.Qual(), tag)
}

func (l logger[Orig, Com, Obj]) SinglePtAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.SinglePtQual, tag info.CP56Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Pt(), p.Qual(), tag)
}

func (l logger[Orig, Com, Obj]) SinglePtChangePack(u info.DataUnit[Orig, Com, Obj], addr Obj, pack info.SinglePtChangePack, q info.Qual) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %016b~%016b %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, pack>>16, pack&0xffff, q)
}

func (l logger[Orig, Com, Obj]) DoublePt(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.DoublePtQual) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Pt(), p.Qual())
}

func (l logger[Orig, Com, Obj]) DoublePtAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.DoublePtQual, tag info.CP24Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Pt(), p.Qual(), tag)
}

func (l logger[Orig, Com, Obj]) DoublePtAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.DoublePtQual, tag info.CP56Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Pt(), p.Qual(), tag)
}

func (l logger[Orig, Com, Obj]) Step(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.StepQual) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Step(), p.Qual())
}

func (l logger[Orig, Com, Obj]) StepAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.StepQual, tag info.CP24Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Step(), p.Qual(), tag)
}

func (l logger[Orig, Com, Obj]) StepAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.StepQual, tag info.CP56Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, p.Step(), p.Qual(), tag)
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
		u.Type, u.Cause, u.Orig, u.Addr, addr, n.Ref().Float64(), n.Qual())
}

func (l logger[Orig, Com, Obj]) NormAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.NormQual, tag info.CP24Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %f %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, n.Ref().Float64(), n.Qual(), tag)
}

func (l logger[Orig, Com, Obj]) NormAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.NormQual, tag info.CP56Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %f %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, n.Ref().Float64(), n.Qual(), tag)
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

func (l logger[Orig, Com, Obj]) Totals(u info.DataUnit[Orig, Com, Obj], addr Obj, c info.Counter) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, c)
}

func (l logger[Orig, Com, Obj]) TotalsAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, c info.Counter, tag info.CP24Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, c, tag)
}

func (l logger[Orig, Com, Obj]) TotalsAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, c info.Counter, tag info.CP56Time2a) {
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, c, tag)
}

func (l logger[Orig, Com, Obj]) ProtectAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, e info.ProtectEvent, tag info.CP24Time2a) {
	duration, _ := e.Elapsed()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %dms %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, e.State().Pt(), e.Qual(), duration.Millis(), tag)
}

func (l logger[Orig, Com, Obj]) ProtectAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, e info.ProtectEvent, tag info.CP56Time2a) {
	duration, _ := e.Elapsed()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %dms %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, e.State().Pt(), e.Qual(), duration.Millis(), tag)
}

func (l logger[Orig, Com, Obj]) ProtectStartAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, e info.ProtectStartEvent, tag info.CP24Time2a) {
	duration, _ := e.Relay()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %dms %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, e.Flags(), e.Qual(), duration.Millis(), tag)
}

func (l logger[Orig, Com, Obj]) ProtectStartAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, e info.ProtectStartEvent, tag info.CP56Time2a) {
	duration, _ := e.Relay()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %dms %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, e.Flags(), e.Qual(), duration.Millis(), tag)
}

func (l logger[Orig, Com, Obj]) ProtectOutAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, e info.ProtectOutEvent, tag info.CP24Time2a) {
	duration, _ := e.Relay()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %dms %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, e.Flags(), e.Qual(), duration.Millis(), tag)
}

func (l logger[Orig, Com, Obj]) ProtectOutAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, e info.ProtectOutEvent, tag info.CP56Time2a) {
	duration, _ := e.Relay()
	fmt.Fprintf(l.W, "%s %s %x %#x/%#x %s %s %dms %s\n",
		u.Type, u.Cause, u.Orig, u.Addr, addr, e.Flags(), e.Qual(), duration.Millis(), tag)
}

func (l logger[Orig, Com, Obj]) InitEnd(u info.DataUnit[Orig, Com, Obj], c info.InitCause) {
	fmt.Fprintf(l.W, "%s %s %x %#x %d\n",
		u.Type, u.Cause, u.Orig, u.Addr, c)
}
