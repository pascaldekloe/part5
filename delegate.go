package part5

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/pascaldekloe/part5/info"
)

// Delegate is a listener configuration for one common address.
type Delegate struct {
	// Inbound messages are reported to each entry in All first. The
	// functions are called in order of appearance. An ASDU is dropped
	// immediately when the return is false. Thus All can be used as a
	// filter, a preprocessor or as an extension point.
	All []func(*info.ASDU) (pass bool)

	// Called in order of appearance for type C_TS_NA_1.
	// See section 5, subclause 6.11.
	Tests []func() (ok bool)

	// listeners for measured values
	Singles map[info.ObjAddr]Single
	Doubles map[info.ObjAddr]Double
	Steps   map[info.ObjAddr]Step
	Bits    map[info.ObjAddr]Bits
	Normals map[info.ObjAddr]Normal
	Scaleds map[info.ObjAddr]Scaled
	Floats  map[info.ObjAddr]Float

	// listeners for commands
	SingleCmds      map[info.ObjAddr]SingleCmd
	DoubleCmds      map[info.ObjAddr]DoubleCmd
	RegulCmds       map[info.ObjAddr]RegulCmd
	NormalSetpoints map[info.ObjAddr]NormalSetpoint
	ScaledSetpoints map[info.ObjAddr]ScaledSetpoint
	FloatSetpoints  map[info.ObjAddr]FloatSetpoint
	BitsCmds        map[info.ObjAddr]BitsCmd
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

var errSelectTimeout = errors.New("part5: command execute after select timeout")
var errSelectInterrupt = errors.New("part5: command select–execute procedure interrupted")

// SelectProc passes a command select procedure when applicable and
// returns the execute message or nil for abort.
// The position of the qualifier of command is given with cmdIndex.
func selectProc(req *info.ASDU, cmdIndex int, in <-chan *info.ASDU, out chan<- *info.ASDU) (exec *info.ASDU, err error) {
	switch req.Cause {
	case info.Act:
		break

	case info.Deact:
		// nothing in progress
		out <- req.Reply(req.Type, info.Deactcon|info.NegFlag)
		return nil, nil

	default:
		out <- req.Reply(req.Type, info.UnkCause)
		return nil, nil
	}

	if req.Info[cmdIndex]&128 == 0 {
		// direct command; no select
		return req, nil
	}

	// acknowledge select
	out <- req.Reply(req.Type, info.Actcon)

	// "not interruptable and controlled by a time out"
	timeout := time.NewTimer(10 * time.Second)
	select {
	case <-timeout.C:
		return nil, errSelectTimeout

	case exec = <-in:
		if !timeout.Stop() {
			<-timeout.C
		}
		if exec == nil {
			return nil, io.EOF
		}
	}

	if exec.Type != req.Type ||
		!bytes.Equal(exec.Info[:cmdIndex], req.Info[:cmdIndex]) ||
		exec.Info[cmdIndex] != req.Info[cmdIndex]&127 ||
		!bytes.Equal(exec.Info[cmdIndex+1:], req.Info[cmdIndex+1:]) {
		return nil, errSelectInterrupt
	}

	switch exec.Cause {
	case info.Act:
		return exec, nil // success
	case info.Deact:
		// break off
		out <- req.Reply(req.Type, info.Deactcon)
		return nil, nil
	default:
		out <- req.Reply(req.Type, info.UnkCause)
		return nil, nil
	}
}

// Called in isolation per originating address.
func (d *Delegate) singleCmd(req *info.ASDU, in <-chan *info.ASDU, out chan<- *info.ASDU) error {
	addr := req.GetObjAddrAt(0)
	f, ok := d.SingleCmds[addr]
	if !ok {
		f, ok = d.SingleCmds[info.IrrelevantAddr]
	}
	if !ok {
		out <- req.Reply(req.Type, info.UnkInfo)
		return nil
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

	req, err := selectProc(req, req.ObjAddrSize, in, out)
	if err != nil {
		return err
	}
	if req == nil {
		return nil
	}

	terminate := func() {
		out <- req.Reply(req.Type, info.Actterm)
	}

	if f(req.ID, cmd.Point(), attrs, terminate) {
		out <- req.Reply(req.Type, info.Actcon)
	} else {
		out <- req.Reply(req.Type, info.Actcon|info.NegFlag)
	}
	return nil
}

// Called in isolation per originating address.
func (d *Delegate) doubleCmd(req *info.ASDU, in <-chan *info.ASDU, out chan<- *info.ASDU) error {
	addr := req.GetObjAddrAt(0)
	f, ok := d.DoubleCmds[addr]
	if !ok {
		f, ok = d.DoubleCmds[info.IrrelevantAddr]
	}
	if !ok {
		out <- req.Reply(req.Type, info.UnkInfo)
		return nil
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

	req, err := selectProc(req, req.ObjAddrSize, in, out)
	if err != nil {
		return err
	}
	if req == nil {
		return nil
	}

	terminate := func() {
		out <- req.Reply(req.Type, info.Actterm)
	}

	if f(req.ID, cmd.Point(), attrs, terminate) {
		out <- req.Reply(req.Type, info.Actcon)
	} else {
		out <- req.Reply(req.Type, info.Actcon|info.NegFlag)
	}
	return nil
}

// Called in isolation per originating address.
func (d *Delegate) floatSetpoint(req *info.ASDU, in <-chan *info.ASDU, out chan<- *info.ASDU) error {
	addr := req.GetObjAddrAt(0)
	f, ok := d.FloatSetpoints[addr]
	if !ok {
		f, ok = d.FloatSetpoints[info.IrrelevantAddr]
	}
	if !ok {
		out <- req.Reply(req.Type, info.UnkInfo)
		return nil
	}

	req, err := selectProc(req, req.ObjAddrSize, in, out)
	if err != nil {
		return err
	}
	if req == nil {
		return nil
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

	terminate := func() {
		out <- req.Reply(req.Type, info.Actterm)
	}

	m := math.Float32frombits(binary.LittleEndian.Uint32(req.Info[req.ObjAddrSize:]))
	if f(req.ID, m, attrs, terminate) {
		out <- req.Reply(req.Type, info.Actcon)
	} else {
		out <- req.Reply(req.Type, info.Actcon|info.NegFlag)
	}
	return nil
}
