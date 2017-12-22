package part5

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/pascaldekloe/part5/info"
)

// Delegate is a listener configuration for one common address.
type Delegate struct {
	sync.WaitGroup

	// All inbound messages are reported to these entries first. The
	// functions are called in order of appearance. An ASDU is dropped
	// immediately when the return is false. Thus All can be used as a
	// filter, a preprocessor or as an extension point.
	All []func(*info.ASDU) (pass bool)

	// Tests are called in order of appearance for type C_TS_NA_1.
	// See section 5, subclause 6.11.
	Tests []func() (ok bool)

	// Listeners For Measured Values [primary]
	// Use info.IrrelevantAddr to catch'm all.
	Singles map[info.ObjAddr]Single
	Doubles map[info.ObjAddr]Double
	Steps   map[info.ObjAddr]Step
	Bits    map[info.ObjAddr]Bits
	Normals map[info.ObjAddr]Normal
	Scaleds map[info.ObjAddr]Scaled
	Floats  map[info.ObjAddr]Float

	// Listeners For Commands [secondary]
	// Use info.IrrelevantAddr to catch'm all.
	SingleCmds      map[info.ObjAddr]SingleCmd
	DoubleCmds      map[info.ObjAddr]DoubleCmd
	RegulCmds       map[info.ObjAddr]RegulCmd
	NormalSetpoints map[info.ObjAddr]NormalSetpoint
	ScaledSetpoints map[info.ObjAddr]ScaledSetpoint
	FloatSetpoints  map[info.ObjAddr]FloatSetpoint
	BitsCmds        map[info.ObjAddr]BitsCmd
}

// Valid returns the validation result.
func (d *Delegate) Valid() error {
	registered := make(map[info.ObjAddr]struct{})
	for addr := range d.Singles {
		if _, ok := registered[addr]; ok {
			return fmt.Errorf("part5: multiple listeners on information object address %d", addr)
		}
		registered[addr] = struct{}{}
	}
	for addr := range d.Doubles {
		if _, ok := registered[addr]; ok {
			return fmt.Errorf("part5: multiple listeners on information object address %d", addr)
		}
		registered[addr] = struct{}{}
	}
	for addr := range d.Steps {
		if _, ok := registered[addr]; ok {
			return fmt.Errorf("part5: multiple listeners on information object address %d", addr)
		}
		registered[addr] = struct{}{}
	}
	for addr := range d.Bits {
		if _, ok := registered[addr]; ok {
			return fmt.Errorf("part5: multiple listeners on information object address %d", addr)
		}
		registered[addr] = struct{}{}
	}
	for addr := range d.Normals {
		if _, ok := registered[addr]; ok {
			return fmt.Errorf("part5: multiple listeners on information object address %d", addr)
		}
		registered[addr] = struct{}{}
	}
	for addr := range d.Scaleds {
		if _, ok := registered[addr]; ok {
			return fmt.Errorf("part5: multiple listeners on information object address %d", addr)
		}
		registered[addr] = struct{}{}
	}
	for addr := range d.Floats {
		if _, ok := registered[addr]; ok {
			return fmt.Errorf("part5: multiple listeners on information object address %d", addr)
		}
		registered[addr] = struct{}{}
	}

	return nil
}

// String returns a compact description of the setup.
func (d *Delegate) String() string {
	buf := new(bytes.Buffer)
	buf.WriteByte('[')

	for addr := range d.Singles {
		fmt.Fprintf(buf, " single@%d", addr)
	}
	for addr := range d.Doubles {
		fmt.Fprintf(buf, " double@%d", addr)
	}
	for addr := range d.Steps {
		fmt.Fprintf(buf, " step@%d", addr)
	}
	for addr := range d.Bits {
		fmt.Fprintf(buf, " bits@%d", addr)
	}
	for addr := range d.Normals {
		fmt.Fprintf(buf, " normal@%d", addr)
	}
	for addr := range d.Scaleds {
		fmt.Fprintf(buf, " scaled@%d", addr)
	}
	for addr := range d.Floats {
		fmt.Fprintf(buf, " float@%d", addr)
	}

	for addr := range d.SingleCmds {
		fmt.Fprintf(buf, " single_cmd@%d", addr)
	}
	for addr := range d.DoubleCmds {
		fmt.Fprintf(buf, " double_cmd@%d", addr)
	}
	for addr := range d.RegulCmds {
		fmt.Fprintf(buf, " regul_cmd@%d", addr)
	}
	for addr := range d.NormalSetpoints {
		fmt.Fprintf(buf, " normal_set-point@%d", addr)
	}
	for addr := range d.ScaledSetpoints {
		fmt.Fprintf(buf, " scaled_set-point@%d", addr)
	}
	for addr := range d.FloatSetpoints {
		fmt.Fprintf(buf, " float_set-point@%d", addr)
	}
	for addr := range d.BitsCmds {
		fmt.Fprintf(buf, " bits_cmd@%d", addr)
	}

	buf.WriteString(" ]")
	return buf.String()
}

func (d *Delegate) single(id info.ID, m info.SinglePoint, attrs MeasureAttrs) {
	f, ok := d.Singles[attrs.Addr]
	if !ok {
		f, ok = d.Singles[info.IrrelevantAddr]
	}
	if ok {
		f(id, m, attrs)
	}
}

func (d *Delegate) double(id info.ID, m info.DoublePoint, attrs MeasureAttrs) {
	f, ok := d.Doubles[attrs.Addr]
	if !ok {
		f, ok = d.Doubles[info.IrrelevantAddr]
	}
	if ok {
		f(id, m, attrs)
	}
}

func (d *Delegate) step(id info.ID, m info.StepPos, attrs MeasureAttrs) {
	f, ok := d.Steps[attrs.Addr]
	if !ok {
		f, ok = d.Steps[info.IrrelevantAddr]
	}
	if ok {
		f(id, m, attrs)
	}
}

func (d *Delegate) bits(id info.ID, m uint32, attrs MeasureAttrs) {
	f, ok := d.Bits[attrs.Addr]
	if !ok {
		f, ok = d.Bits[info.IrrelevantAddr]
	}
	if ok {
		f(id, m, attrs)
	}
}

func (d *Delegate) normal(id info.ID, m info.Normal, attrs MeasureAttrs) {
	f, ok := d.Normals[attrs.Addr]
	if !ok {
		f, ok = d.Normals[info.IrrelevantAddr]
	}
	if ok {
		f(id, m, attrs)
	}
}

func (d *Delegate) scaled(id info.ID, m int16, attrs MeasureAttrs) {
	f, ok := d.Scaleds[attrs.Addr]
	if !ok {
		f, ok = d.Scaleds[info.IrrelevantAddr]
	}
	if ok {
		f(id, m, attrs)
	}
}

func (d *Delegate) float(id info.ID, m float32, attrs MeasureAttrs) {
	f, ok := d.Floats[attrs.Addr]
	if !ok {
		f, ok = d.Floats[info.IrrelevantAddr]
	}
	if ok {
		f(id, m, attrs)
	}
}

var errSelectExpire = errors.New("part5: command execute after select timeout")
var errSelectLost = errors.New("part5: command select abort due connection loss")

// SelectProc passes a command select procedure when applicable and
// returns the execute message or nil for abort.
// The position of the qualifier of command is given with cmdIndex.
func selectProc(req *info.ASDU, cmdIndex int, c *Caller, errCh chan<- error) (exec *info.ASDU) {
	switch req.Cause &^ info.TestFlag {
	case info.Act:
		break

	case info.Deact:
		// nothing in progress
		sendNB(req.Reply(info.Deactcon|info.NegFlag), c, errCh)
		return nil

	default:
		sendNB(req.Reply(info.UnkCause), c, errCh)
		return nil
	}

	if req.Info[cmdIndex]&128 == 0 {
		// direct command; no select
		return req
	}
	// "not interruptable and controlled by a time out"

	// acknowledge select
	if err := c.Send(req.Reply(info.Actcon)); err != nil {
		errCh <- err
		return nil
	}

	// BUG(pascaldekloe): Hardcoded select timeout of 10s.
	expire := time.NewTimer(10 * time.Second)
	defer expire.Stop()
	select {
	case <-expire.C:
		errCh <- errSelectExpire
		return nil

	case bytes, ok := <-c.transport.In:
		if !ok {
			errCh <- errSelectLost
			return nil
		}

		exec = &info.ASDU{Params: c.params}
		if err := exec.UnmarshalBinary(bytes); err != nil {
			errCh <- err
			return nil
		}
	}

	if exec.Type != req.Type ||
		!bytes.Equal(exec.Info[:cmdIndex], req.Info[:cmdIndex]) ||
		exec.Info[cmdIndex] != req.Info[cmdIndex]&127 ||
		!bytes.Equal(exec.Info[cmdIndex+1:], req.Info[cmdIndex+1:]) {
		errCh <- fmt.Errorf("part5: %s interrupted by %s", req, exec)
		return nil
	}

	switch exec.Cause ^ req.Cause&info.TestFlag {
	case info.Act:
		return exec // success

	case info.Deact:
		// break off
		sendNB(req.Reply(info.Deactcon), c, errCh)
		return nil

	default:
		sendNB(req.Reply(info.UnkCause), c, errCh)
		return nil
	}
}

func (d *Delegate) singleCmd(req *info.ASDU, c *Caller, errCh chan<- error) {
	addr := req.GetObjAddrAt(0)
	f, ok := d.SingleCmds[addr]
	if !ok {
		f, ok = d.SingleCmds[info.IrrelevantAddr]
	}
	if !ok {
		sendNB(req.Reply(info.UnkInfo), c, errCh)
		return
	}

	cmd := info.SingleCmd{Cmd: info.Cmd(req.Info[req.ObjAddrSize])}
	attrs := ExecAttrs{
		Addr:     addr,
		Qual:     cmd.Qual(),
		InSelect: !cmd.Exec(),
	}
	if req.Type == info.C_SC_TA_1 {
		// TODO(pascaldekloe): time tag
	}

	req = selectProc(req, req.ObjAddrSize, c, errCh)
	if req == nil {
		return
	}

	d.Add(1)
	go func() {
		defer d.Done()

		terminate := func() {
			sendNB(req.Reply(info.Actterm), c, errCh)
		}
		if f(req.ID, cmd.Point(), attrs, terminate) {
			sendNB(req.Reply(info.Actcon), c, errCh)
		} else {
			sendNB(req.Reply(info.Actcon|info.NegFlag), c, errCh)
		}
	}()
}

func (d *Delegate) doubleCmd(req *info.ASDU, c *Caller, errCh chan<- error) {
	addr := req.GetObjAddrAt(0)
	f, ok := d.DoubleCmds[addr]
	if !ok {
		f, ok = d.DoubleCmds[info.IrrelevantAddr]
	}
	if !ok {
		sendNB(req.Reply(info.UnkInfo), c, errCh)
		return
	}

	cmd := info.DoubleCmd{Cmd: info.Cmd(req.Info[req.ObjAddrSize])}
	attrs := ExecAttrs{
		Addr:     addr,
		Qual:     cmd.Qual(),
		InSelect: !cmd.Exec(),
	}
	if req.Type == info.C_DC_TA_1 {
		// TODO(pascaldekloe): time tag
	}

	req = selectProc(req, req.ObjAddrSize, c, errCh)
	if req == nil {
		return
	}

	d.Add(1)
	go func() {
		defer d.Done()

		terminate := func() {
			sendNB(req.Reply(info.Actterm), c, errCh)
		}
		if f(req.ID, cmd.Point(), attrs, terminate) {
			sendNB(req.Reply(info.Actcon), c, errCh)
		} else {
			sendNB(req.Reply(info.Actcon|info.NegFlag), c, errCh)
		}
	}()
}

func (d *Delegate) floatSetpoint(req *info.ASDU, c *Caller, errCh chan<- error) {
	addr := req.GetObjAddrAt(0)
	f, ok := d.FloatSetpoints[addr]
	if !ok {
		f, ok = d.FloatSetpoints[info.IrrelevantAddr]
	}
	if !ok {
		sendNB(req.Reply(info.UnkInfo), c, errCh)
		return
	}

	cmd := info.SetpointCmd(req.Info[req.ObjAddrSize+4])
	attrs := ExecAttrs{
		Addr:     addr,
		Qual:     cmd.Qual(),
		InSelect: !cmd.Exec(),
	}
	if req.Type == info.C_SE_TC_1 {
		// TODO(pascaldekloe): time tag
	}

	req = selectProc(req, req.ObjAddrSize+4, c, errCh)
	if req == nil {
		return
	}

	d.Add(1)
	go func() {
		defer d.Done()

		terminate := func() {
			sendNB(req.Reply(info.Actterm), c, errCh)
		}
		p := math.Float32frombits(binary.LittleEndian.Uint32(req.Info[req.ObjAddrSize:]))
		if f(req.ID, p, attrs, terminate) {
			sendNB(req.Reply(info.Actcon), c, errCh)
		} else {
			sendNB(req.Reply(info.Actcon|info.NegFlag), c, errCh)
		}
	}()
}

// SendNB does a non-blocking outbound submission.
func sendNB(u *info.ASDU, c *Caller, errCh chan<- error) {
	go func() {
		if err := c.Send(u); err != nil {
			errCh <- err
		}
	}()
}
