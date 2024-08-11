package part5

import (
	"encoding/binary"
	"math"

	"github.com/pascaldekloe/part5/info"
)

// Exchange supplies communication with another station.
type Exchange[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	info.System[Orig, Com, Obj] // parameters

	// The originator address is optional, depending on the System.
	Orig Orig
	// The common address must be non zero.
	Com Com
}

// NewDataUnit returns a new ASDU without payload; .Info is empty.
func (x Exchange[Orig, Com, Obj]) NewDataUnit(t info.TypeID, e info.Enc, c info.Cause) info.DataUnit[Orig, Com, Obj] {
	u := x.System.NewDataUnit()
	u.Type = t
	u.Enc = e
	u.Cause = c
	u.Orig = x.Orig
	u.Addr = x.Com
	return u
}

// Command has the controlling perspective of an Exchange.
type Command[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	Exchange[Orig, Com, Obj]
}

// Command returns the controlling perspective of an Exchange.
func (x Exchange[Orig, Com, Obj]) Command() Command[Orig, Com, Obj] {
	return Command[Orig, Com, Obj]{x}
}

// Activation commands address one information object.
func (cmd Command[Orig, Com, Obj]) act(t info.TypeID, addr Obj) info.DataUnit[Orig, Com, Obj] {
	u := cmd.Exchange.NewDataUnit(t, 1, info.Act)
	for i := 0; i < len(addr); i++ {
		u.Info = append(u.Info, addr[i])
	}
	return u
}

// SingleCmd returns single command: C_SC_NA_1 act(ivation),
// conform chapter 7.3.2.1 of companion standard 101.
func (cmd Command[Orig, Com, Obj]) SingleCmd(addr Obj, pt info.SinglePt, q info.CmdQual) info.DataUnit[Orig, Com, Obj] {
	u := cmd.act(info.C_SC_NA_1, addr)
	u.Info = append(u.Info, byte(pt&1)|byte(q&^1))
	return u
}

// DoubleCmd returns double command: C_DC_NA_1 act(ivation),
// conform chapter 7.3.2.2 of companion standard 101.
func (cmd Command[Orig, Com, Obj]) DoubleCmd(addr Obj, pt info.DoublePt, q info.CmdQual) info.DataUnit[Orig, Com, Obj] {
	u := cmd.act(info.C_DC_NA_1, addr)
	u.Info = append(u.Info, byte(pt&3)|byte(q&^3))
	return u
}

// RegulCmd returns regulating-step command: C_RC_NA_1 act(ivation),
// conform chapter 7.3.2.3 of companion standard 101.
// Regul must be either Lower or Higher. The other two states
// (0 and 3) are not permitted.
func (cmd Command[Orig, Com, Obj]) RegulCmd(addr Obj, r info.Regul, q info.CmdQual) info.DataUnit[Orig, Com, Obj] {
	u := cmd.act(info.C_RC_NA_1, addr)
	u.Info = append(u.Info, byte(r&3)|byte(q&^3))
	return u
}

// NormSetPt returns set-point command: C_SE_NA_1 act(ivation),
// conform chapter 7.3.2.4 of companion standard 101.
func (cmd Command[Orig, Com, Obj]) NormSetPt(addr Obj, value info.Norm, q info.SetPtQual) info.DataUnit[Orig, Com, Obj] {
	u := cmd.act(info.C_SE_NA_1, addr)
	u.Info = append(u.Info, value[0], value[1])
	u.Info = append(u.Info, byte(q))
	return u
}

// ScaledSetPt returns set-point command: C_SE_NB_1 act(ivation),
// conform chapter 7.3.2.5 of companion standard 101.
func (cmd Command[Orig, Com, Obj]) ScaledSetPt(addr Obj, value int16, q info.SetPtQual) info.DataUnit[Orig, Com, Obj] {
	u := cmd.act(info.C_SE_NB_1, addr)
	u.Info = append(u.Info, byte(value), byte(value>>8))
	u.Info = append(u.Info, byte(q))
	return u
}

// FloatSetPt returns set-point command: C_SE_NC_1 act(ivation),
// conform chapter 7.3.2.6 of companion standard 101.
func (cmd Command[Orig, Com, Obj]) FloatSetPt(addr Obj, value float32, q info.SetPtQual) info.DataUnit[Orig, Com, Obj] {
	u := cmd.act(info.C_SE_NC_1, addr)
	u.Info = binary.LittleEndian.AppendUint32(u.Info,
		math.Float32bits(value))
	u.Info = append(u.Info, byte(q))
	return u
}

// Inro returns interrogation command: C_IC_NA_1 act(ivation),
// conform chapter 7.3.4.1 of companion standard 101.
func (cmd Command[Orig, Com, Obj]) Inro() info.DataUnit[Orig, Com, Obj] {
	return cmd.InroGroup(0)
}

// InroGroup returns interrogation command: C_IC_NA_1 act(ivation),
// conform chapter 7.3.4.1 of companion standard 101.
// Group can be disabled with 0 for (global) station interrogation.
// Otherwise, use a group identifier in range [1..16].
func (cmd Command[Orig, Com, Obj]) InroGroup(group uint) info.DataUnit[Orig, Com, Obj] {
	var addr Obj // fixed to zero
	u := cmd.act(info.C_IC_NA_1, addr)
	// The qualifier of interrogation codes are listed
	// at chapter 7.2.6.22 of companion standard 101.
	u.Info = append(u.Info, byte(group+20))
	return u
}

// TestCmd returns test command: C_TS_NA_1 act(ivation),
// conform chapter 7.3.4.5 of companion standard 101.
func (cmd Command[Orig, Com, Obj]) TestCmd() info.DataUnit[Orig, Com, Obj] {
	var addr Obj // fixed to zero
	u := cmd.act(info.C_TS_NA_1, addr)
	// fixed bit-pattern from chapter 7.2.6.14 of companion standard 101
	u.Info = append(u.Info, 0b1010_1010, 0b0101_0101)
	return u
}
