package part5

import (
	"fmt"
	"io"

	"github.com/pascaldekloe/part5/info"
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
