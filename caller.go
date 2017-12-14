package part5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"sync"
	"time"

	"github.com/pascaldekloe/part5/info"
	"github.com/pascaldekloe/part5/session"
)

// ErrCmdRace signals a concurrent command execution.
var ErrCmdRace = errors.New("part5: pending command execution on same address")

// ErrCmdDeny signals a remote acceptance rejection.
var ErrCmdDeny = errors.New("part5: command execution denied")

// Validation Rejection
var (
	errType     = errors.New("part5: type identifier doesn't match call or time tag")
	errSingle   = errors.New("part5: unknown single-point state")
	errDouble   = errors.New("part5: unknown double-point state")
	errQualDesc = errors.New("part5: illegal quality descriptor flags")
	errCmdCause = errors.New("part5: cause of transmission for command not act(ivate)")
	errCmdQual  = errors.New("part5: command qualifier exceeds [0, 31]")
)

// Command Acknowledge Problems
var (
	errCmdExpire  = errors.New("part5: launch confirmation of command execution timeout; status unknown")
	errTermExpire = errors.New("part5: termination confirmation of command execution timeout; status unknown")
)

// Caller submits messages. All of its methods are thread-safe.
type Caller struct {
	params     *info.Params
	transport  *session.Transport
	interrupts sync.RWMutex
	confirm    sync.Map
}

// Send submits a message and awaits receival confirmation.
func (c *Caller) Send(u *info.ASDU) error {
	data, err := u.MarshalBinary()
	if err != nil {
		return err
	}

	o := session.NewOutbound(data)
	c.interrupts.RLock()
	switch u.Cause {
	case info.Spont, info.Req, info.Retloc, info.Retrem:
		c.transport.Class1 <- o
	default:
		c.transport.Class2 <- o
	}
	c.interrupts.RUnlock()
	return <-o.Done
}

// Single sends a type M_SP_NA_1, M_SP_TA_1 or M_SP_TB_1.
func (c *Caller) Single(id info.ID, v info.SinglePoint, attrs MeasureAttrs) error {
	u := info.NewASDU(c.params, id)
	if err := u.AppendAddr(attrs.Addr); err != nil {
		return err
	}

	if v&^1 != 0 {
		return errSingle
	}
	if attrs.QualDesc&^0xf0 != 0 {
		return errQualDesc
	}
	u.Info = append(u.Info, byte(uint(v)|attrs.QualDesc))

	switch id.Type {
	case info.M_SP_NA_1:
		if attrs.Time != nil {
			return errType
		}

	case info.M_SP_TA_1:
		panic("TODO: append 24-bit timestamp")

	case info.M_SP_TB_1:
		panic("TODO: append 56-bit timestamp")

	default:
		return errType
	}

	return c.Send(u)
}

// Double sends a type M_DP_NA_1, M_DP_TA_1 or M_DP_TB_1.
func (c *Caller) Double(id info.ID, v info.DoublePoint, attrs MeasureAttrs) error {
	u := info.NewASDU(c.params, id)
	if err := u.AppendAddr(attrs.Addr); err != nil {
		return err
	}

	if v&^3 != 0 {
		return errDouble
	}
	if attrs.QualDesc&^0xf0 != 0 {
		return errQualDesc
	}
	u.Info = append(u.Info, byte(uint(v)|attrs.QualDesc))

	switch id.Type {
	case info.M_DP_NA_1:
		if attrs.Time != nil {
			return errType
		}

	case info.M_DP_TA_1:
		panic("TODO: append 24-bit timestamp")

	case info.M_DP_TB_1:
		panic("TODO: append 56-bit timestamp")

	default:
		return errType
	}

	return c.Send(u)
}

// Step sends a type M_ST_NA_1, M_ST_TA_1 or M_ST_TB_1.
func (c *Caller) Step(id info.ID, v info.StepPos, attrs MeasureAttrs) {
	panic("TODO: not implemented")
}

// Bits sends a type M_BO_NA_1, M_BO_TA_1 or M_BO_TB_1.
func (c *Caller) Bits(id info.ID, v uint32, attrs MeasureAttrs) {
	panic("TODO: not implemented")
}

// Normal sends a type M_ME_NA_1, M_ME_TA_1, M_ME_TD_1 or M_ME_ND_1.
// The quality descriptor must default to info.OK for type M_ME_ND_1.
func (c *Caller) Normal(id info.ID, v info.Normal, attrs MeasureAttrs) {
	panic("TODO: not implemented")
}

// Scaled sends a type M_ME_NB_1, M_ME_TB_1 or M_ME_TE_1.
func (c *Caller) Scaled(id info.ID, v int16, attrs MeasureAttrs) {
	panic("TODO: not implemented")
}

// Float sends a type M_ME_NC_1, M_ME_TC_1 or M_ME_TF_1.
func (c *Caller) Float(id info.ID, v float32, attrs MeasureAttrs) error {
	u := info.NewASDU(c.params, id)
	if err := u.AppendAddr(attrs.Addr); err != nil {
		return err
	}

	if attrs.QualDesc&^0xf1 != 0 {
		return errQualDesc
	}
	i := len(u.Info)
	u.Info = append(u.Info, 0, 0, 0, 0, byte(attrs.QualDesc))
	binary.LittleEndian.PutUint32(u.Info[i:], math.Float32bits(v))

	switch id.Type {
	case info.M_ME_NC_1:
		if attrs.Time != nil {
			return errType
		}

	case info.M_ME_TC_1:
		panic("TODO: append 24-bit timestamp")

	case info.M_ME_TF_1:
		panic("TODO: append 56-bit timestamp")

	default:
		return errType
	}

	return c.Send(u)
}

// SingleCmd sends a type C_SC_NA_1 or C_SC_TA_1.
func (c *Caller) SingleCmd(id info.ID, p info.SinglePoint, attrs ExecAttrs, term CmdTerm) error {
	if id.Cause&^info.TestFlag != info.Act {
		return errCmdCause
	}

	u := info.NewASDU(c.params, id)
	if err := u.AppendAddr(attrs.Addr); err != nil {
		return err
	}

	if p&^1 != 0 {
		return errSingle
	}
	if attrs.Qual > 31 {
		return errCmdQual
	}
	u.Info = append(u.Info, byte(attrs.Qual<<2|uint(p)))

	switch id.Type {
	case info.C_SC_NA_1:
		if attrs.Time != nil {
			return errType
		}

	case info.C_SC_TA_1:
		panic("TODO: append 56-bit timestamp")

	default:
		return errType
	}

	selIndex := -1
	if attrs.InSelect {
		selIndex = u.ObjAddrSize
	}
	return c.launchCmd(u, selIndex, term)
}

// DoubleCmd sends a type C_DC_NA_1 or C_DC_TA_1.
func (c *Caller) DoubleCmd(id info.ID, p info.DoublePoint, attrs ExecAttrs, term CmdTerm) error {
	panic("TODO: not implemented")
}

// RegulCmd sends a type C_RC_NA_1 or C_RC_TA_1.
func (c *Caller) RegulCmd(id info.ID, up info.DoublePoint, attrs ExecAttrs, term CmdTerm) error {
	panic("TODO: not implemented")
}

// NormalSetpoint sends a type C_SE_NA_1 or C_SE_TA_1.
func (c *Caller) NormalSetpoint(id info.ID, p info.Normal, attrs ExecAttrs, term CmdTerm) error {
	panic("TODO: not implemented")
}

// ScaledSetpoint sends a type C_SE_NB_1 or C_SE_TB_1.
func (c *Caller) ScaledSetpoint(id info.ID, p int16, attrs ExecAttrs, term CmdTerm) error {
	panic("TODO: not implemented")
}

// FloatSetpoint sends a type C_SE_NC_1 or C_SE_TC_1.
func (c *Caller) FloatSetpoint(id info.ID, p float32, attrs ExecAttrs, term CmdTerm) error {
	panic("TODO: not implemented")
}

// BitsCmd sends a type C_BO_NA_1 or C_BO_TA_1.
func (c *Caller) BitsCmd(id info.ID, bits uint32, addr info.ObjAddr, tag *time.Time, term CmdTerm) error {
	panic("TODO: not implemented")
}

func (c *Caller) launchCmd(exe *info.ASDU, selIndex int, term CmdTerm) error {
	if selIndex < 0 {
		c.interrupts.RLock()
		defer c.interrupts.RUnlock()
	} else {
		c.interrupts.Lock()
		defer c.interrupts.Unlock()

		sel := exe.Reply(exe.Cause)   // clone
		sel.Info[sel.AddrSize] |= 128 // select flag

		if err := c.sendCmd(sel, nil); err != nil {
			return err
		}
	}

	return c.sendCmd(exe, term)
}

func (c *Caller) sendCmd(req *info.ASDU, term CmdTerm) (err error) {
	data, err := req.MarshalBinary()
	if err != nil {
		return err
	}

	// register a response channel & defer cleanup
	ch := make(chan *info.ASDU, 1)
	key := cmdKey(req)
	if _, loaded := c.confirm.LoadOrStore(key, ch); loaded {
		return ErrCmdRace
	}
	defer func() {
		if err != nil || term == nil {
			c.confirm.Delete(key)
			return
		}
		// keep listening for activation termination

		go func() {
			defer c.confirm.Delete(key)

			// BUG(pascaldekloe): Hardcoded command termination timeout of 60s.
			expire := time.NewTimer(time.Minute)
			defer expire.Stop()

			select {
			case resp := <-ch:
				want := info.Actterm | req.Cause&info.TestFlag
				if resp.Cause == want {
					term(nil)
				} else {
					term(fmt.Errorf("part5: received cause of transmission %s, want %s", resp.Cause, want))
				}
			case <-expire.C:
				term(errTermExpire)
			}
		}()
	}()

	out := session.NewOutbound(data)
	c.transport.Class2 <- out
	if err := <-out.Done; err != nil {
		return err
	}

	// BUG(pascaldekloe): Hardcoded command activation timeout of 10s.
	expire := time.NewTimer(10 * time.Second)
	defer expire.Stop()

	select {
	case <-expire.C:
		return errCmdExpire

	case resp := <-ch:
		confirm := info.Actcon | req.Cause&info.TestFlag
		switch resp.Cause {
		case confirm:
			return nil
		case confirm | info.NegFlag:
			return ErrCmdDeny
		}
		return fmt.Errorf("part5: received cause of transmission %s, want %s", resp.Cause, confirm)
	}
}

func cmdKey(u *info.ASDU) uint64 {
	key := uint64(u.GetObjAddrAt(0)<<16) | uint64(u.Orig<<8) | uint64(u.Type)

	hash := fnv.New32a()
	hash.Write(u.Info)
	return key | uint64(hash.Sum32())<<32
}
