package part5

import (
	"encoding/binary"
	"math"

	"github.com/pascaldekloe/part5/info"
)

// Exchange supplies communication with another station.
type Exchange[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	info.Params[Orig, Com, Obj]

	Orig Orig // originator address, if any
	Com  Com  // common adress
}

// NewDataUnit returns a new ASDU without payload; .Info is empty.
func (x Exchange[Orig, Com, Obj]) NewDataUnit(t info.TypeID, e info.Enc, c info.Cause) info.DataUnit[Orig, Com, Obj] {
	u := x.Params.NewDataUnit()
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

// SingleCmd returns single command C_SC_NA_1 act
// conform chapter 7.3.2.1 from companion standard 101.
func (cmd Command[Orig, Com, Obj]) SingleCmd(addr Obj, pt info.SinglePt, q info.CmdQual) info.DataUnit[Orig, Com, Obj] {
	u := cmd.act(info.C_SC_NA_1, addr)
	u.Info = append(u.Info, byte(pt&1)|byte(q&^1))
	return u
}

// DoubleCmd returns double command C_DC_NA_1 act
// conform chapter 7.3.2.2 from companion standard 101.
func (cmd Command[Orig, Com, Obj]) DoubleCmd(addr Obj, pt info.DoublePt, q info.CmdQual) info.DataUnit[Orig, Com, Obj] {
	u := cmd.act(info.C_DC_NA_1, addr)
	u.Info = append(u.Info, byte(pt&3)|byte(q&^3))
	return u
}

// RegulCmd returns regulating-step command C_RC_NA_1 act
// conform chapter 7.3.2.3 from companion standard 101.
// Regul must be either Lower or Higher. The other two states
// (0 and 3) are not permitted.
func (cmd Command[Orig, Com, Obj]) RegulCmd(addr Obj, r info.Regul, q info.CmdQual) info.DataUnit[Orig, Com, Obj] {
	u := cmd.act(info.C_RC_NA_1, addr)
	u.Info = append(u.Info, byte(r&3)|byte(q&^3))
	return u
}

// NormSetPt returns set-point command C_SE_NA_1 act
// conform chapter 7.3.2.4 from companion standard 101.
func (cmd Command[Orig, Com, Obj]) NormSetPt(addr Obj, value info.Norm, q info.SetPtQual) info.DataUnit[Orig, Com, Obj] {
	u := cmd.act(info.C_SE_NA_1, addr)
	u.Info = append(u.Info, value[0], value[1])
	u.Info = append(u.Info, byte(q))
	return u
}

// ScaledSetPt returns set-point command C_SE_NB_1 act
// conform chapter 7.3.2.5 from companion standard 101.
func (cmd Command[Orig, Com, Obj]) ScaledSetPt(addr Obj, value int16, q info.SetPtQual) info.DataUnit[Orig, Com, Obj] {
	u := cmd.act(info.C_SE_NB_1, addr)
	u.Info = append(u.Info, byte(value), byte(value>>8))
	u.Info = append(u.Info, byte(q))
	return u
}

// FloatSetPt returns set-point command C_SE_NC_1 act
// conform chapter 7.3.2.6 from companion standard 101.
func (cmd Command[Orig, Com, Obj]) FloatSetPt(addr Obj, value float32, q info.SetPtQual) info.DataUnit[Orig, Com, Obj] {
	u := cmd.act(info.C_SE_NC_1, addr)
	u.Info = binary.LittleEndian.AppendUint32(u.Info,
		math.Float32bits(value))
	u.Info = append(u.Info, byte(q))
	return u
}

// Inro returns interrogation command C_IC_NA_1 act
// conform chapter 7.3.4.1 from companion standard 101.
func (cmd Command[Orig, Com, Obj]) Inro() info.DataUnit[Orig, Com, Obj] {
	return cmd.InroGroup(0)
}

// InroGroup returns interrogation command C_IC_NA_1 act
// conform chapter 7.3.4.1 from companion standard 101.
// Group can be disabled with 0 for (global) station interrogation.
// Otherwise, use a group identifier in range [1..16].
func (cmd Command[Orig, Com, Obj]) InroGroup(group uint) info.DataUnit[Orig, Com, Obj] {
	var addr Obj // fixed to zero
	u := cmd.act(info.C_IC_NA_1, addr)
	// qualifier of interrogation is described in
	// companion standard 101, section 7.2.6.22
	u.Info = append(u.Info, byte(group+20))
	return u
}

// TestCmd returns test command C_TS_NA_1 act
// conform chapter 7.3.4.5 from companion standard 101.
func (cmd Command[Orig, Com, Obj]) TestCmd() info.DataUnit[Orig, Com, Obj] {
	var addr Obj // fixed to zero
	u := cmd.act(info.C_TS_NA_1, addr)
	// fixed test bit pattern defined in chapter 7.2.6.1
	// from companion standard 1014
	u.Info = append(u.Info, 0b1010_1010, 0b0101_0101)
	return u
}
