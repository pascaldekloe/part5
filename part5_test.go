package part5

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/pascaldekloe/part5/info"
	"github.com/pascaldekloe/part5/session"
)

// TestMeasuredValues is a full-cycle test for the Caller–Delegate mappings.
func TestMeasuredValues(t *testing.T) {
	t.Parallel()

	const testAddr = 99

	// test definitions
	var measures = []*GoldenMeasure{
		0: {ID: info.ID{Type: info.M_SP_NA_1}, Value: info.On},
		1: {ID: info.ID{Type: info.M_DP_NA_1}, Value: info.DeterminedOn},
		2: {ID: info.ID{Type: info.M_ME_NC_1}, Value: float32(99.8)},
	}
	for i, m := range measures {
		m.ID.Addr = testAddr
		m.ID.Cause = info.Retloc | info.TestFlag
		m.MeasureAttrs.Addr = info.ObjAddr(i) + 1000
		m.MeasureAttrs.QualDesc = uint(i * 0x10)
		m.T = t
	}

	// listener configuration for each test
	delegate := &Delegate{
		Singles: map[info.ObjAddr]Single{
			measures[0].MeasureAttrs.Addr: measures[0].Single,
		},
		Doubles: map[info.ObjAddr]Double{
			measures[1].MeasureAttrs.Addr: measures[1].Double,
		},
		Floats: map[info.ObjAddr]Float{
			measures[2].MeasureAttrs.Addr: measures[2].Float,
		},
	}
	t.Log("measure setup:", delegate)

	// link controlling- with a monitored station
	controllingEnd, monitoredEnd := session.Pipe(2 * time.Second)
	caller, monitoredErrs := MustLaunch(monitoredEnd, info.Wide, nil)
	_, controllingErrs := MustLaunch(controllingEnd, info.Wide,
		map[info.CommonAddr]*Delegate{testAddr: delegate})

	// shutdown hooks
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for err := range controllingErrs {
			t.Error("error on controlling side:", err)
		}
	}()
	go func() {
		defer wg.Done()
		for err := range monitoredErrs {
			t.Error("error on monitored side:", err)
		}
	}()

	// call each address once
	for i, m := range measures {
		var err error

		switch v := m.Value.(type) {
		case info.SinglePoint:
			err = caller.Single(m.ID, v, m.MeasureAttrs)
		case info.DoublePoint:
			err = caller.Double(m.ID, v, m.MeasureAttrs)
		case float32:
			err = caller.Float(m.ID, v, m.MeasureAttrs)
		default:
			t.Fatalf("%d: unknown datatype", i)
		}

		if err != nil {
			t.Errorf("%d: %s call error: %s", i, m.ID, err)
		}
	}

	// shutdown flush
	close(monitoredEnd.Class1)
	close(monitoredEnd.Class2)
	wg.Wait()
	close(controllingEnd.Class1)
	close(controllingEnd.Class2)
	delegate.Wait()

	for i, m := range measures {
		if m.Count != 1 {
			t.Errorf("%d: %s got %d calls; want 1", i, m.ID, m.Count)
		}
	}
}

// TestCommandExecution is a full-cycle test for the Caller–Delegate mappings.
func TestCommandExecution(t *testing.T) {
	t.Parallel()

	const testAddr = 99

	// test definitions
	var executes = []*GoldenExec{
		0: {ID: info.ID{Type: info.C_SC_NA_1}, Point: info.On},
		1: {ID: info.ID{Type: info.C_DC_NA_1}, Point: info.DeterminedOn},
		2: {ID: info.ID{Type: info.C_SE_NC_1}, Point: float32(99.8)},
	}
	for i, e := range executes {
		e.ID.Addr = testAddr
		e.ID.Cause = info.Act | info.TestFlag
		e.ExecAttrs.Addr = info.ObjAddr(i) + 1000
		e.ExecAttrs.Qual = uint(i)
		e.ExecAttrs.InSelect = i%2 == 0
		e.Accept = i%3 != 0
		e.Term = i%3 == 1
		e.T = t
	}

	// listener configuration for each test
	delegate := &Delegate{
		SingleCmds: map[info.ObjAddr]SingleCmd{
			executes[0].ExecAttrs.Addr: executes[0].SingleCmd,
		},
		DoubleCmds: map[info.ObjAddr]DoubleCmd{
			executes[1].ExecAttrs.Addr: executes[1].DoubleCmd,
		},
		FloatSetpoints: map[info.ObjAddr]FloatSetpoint{
			executes[2].ExecAttrs.Addr: executes[2].FloatSetpoint,
		},
	}
	t.Log("command setup:", delegate)

	// link controlling- with a monitored station
	controllingEnd, monitoredEnd := session.Pipe(2 * time.Second)
	caller, controllingErrs := MustLaunch(controllingEnd, info.Wide, nil)
	_, monitoredErrs := MustLaunch(monitoredEnd, info.Wide,
		map[info.CommonAddr]*Delegate{testAddr: delegate})

	// shutdown hooks
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for err := range controllingErrs {
			t.Error("error on controlling side:", err)
		}
	}()
	go func() {
		defer wg.Done()
		for err := range monitoredErrs {
			t.Error("error on monitored side:", err)
		}
	}()

	var termGroup sync.WaitGroup
	// call each address once
	for i, e := range executes {
		termed := make(chan error)
		term := CmdTerm(func(err error) {
			termed <- err
			close(termed)
		})

		var err error

		switch p := e.Point.(type) {
		case info.SinglePoint:
			err = caller.SingleCmd(e.ID, p, e.ExecAttrs, term)
		case info.DoublePoint:
			err = caller.DoubleCmd(e.ID, p, e.ExecAttrs, term)
		case float32:
			err = caller.FloatSetpoint(e.ID, p, e.ExecAttrs, term)
		default:
			t.Fatalf("%d: unknown datatype", i)
		}

		if e.Accept && err != nil {
			t.Errorf("%d: %s call error: %s", i, e.ID, err)
		}
		if !e.Accept && err != ErrCmdDeny {
			t.Errorf("%d: %s call error %q, want %q", i, e.ID, err, ErrCmdDeny)
		}

		termGroup.Add(1)
		go func(testIndex int, id info.ID, wantTerm bool) {
			defer termGroup.Done()

			if wantTerm {
				select {
				case err := <-termed:
					if err != nil {
						t.Errorf("%d: %s termination error: %s", testIndex, id, err)
					}
				case <-time.After(10 * time.Millisecond):
					t.Errorf("%d: %s termination expire", testIndex, id)
				}
			} else {
				select {
				case err := <-termed:
					t.Errorf("%d: %s terminated with: %s", testIndex, id, err)
				case <-time.After(10 * time.Millisecond):
				}
			}
		}(i, e.ID, e.Term)
	}
	termGroup.Wait()

	// shutdown flush
	delegate.Wait()
	close(controllingEnd.Class1)
	close(controllingEnd.Class2)
	close(monitoredEnd.Class1)
	close(monitoredEnd.Class2)
	wg.Wait()

	for i, e := range executes {
		if e.Count != 1 {
			t.Errorf("%d: %s got %d calls; want 1", i, e.ID, e.Count)
		}
	}
}

// GoldenMeasure is a recoding listener.
type GoldenMeasure struct {
	info.ID
	MeasureAttrs
	Value interface{}

	T     *testing.T // report target
	Count int        // number of calls
}

func (gold *GoldenMeasure) check(id info.ID, m interface{}, attrs MeasureAttrs) {
	if id != gold.ID {
		gold.T.Errorf("%s called as %s", gold.ID, id)
	}
	if !reflect.DeepEqual(m, gold.Value) {
		gold.T.Errorf("%s called with value %v, want %v", gold.ID, m, gold.Value)
	}
	if !reflect.DeepEqual(attrs, gold.MeasureAttrs) {
		gold.T.Errorf("%s called with %#v, want %#v", gold.ID, attrs, gold.MeasureAttrs)
	}

	gold.Count++
}

func (gold *GoldenMeasure) Single(id info.ID, m info.SinglePoint, attrs MeasureAttrs) {
	gold.check(id, m, attrs)
}

func (gold *GoldenMeasure) Double(id info.ID, m info.DoublePoint, attrs MeasureAttrs) {
	gold.check(id, m, attrs)
}

func (gold *GoldenMeasure) Float(id info.ID, m float32, attrs MeasureAttrs) {
	gold.check(id, m, attrs)
}

// GoldenExec is a recoding listener.
type GoldenExec struct {
	info.ID
	ExecAttrs
	Point  interface{}
	Accept bool
	Term   bool

	T     *testing.T // report target
	Count int        // number of calls
}

func (gold *GoldenExec) check(id info.ID, p interface{}, attrs ExecAttrs, term ExecTerm) {
	if id != gold.ID {
		gold.T.Errorf("%s called as %s", gold.ID, id)
	}
	if !reflect.DeepEqual(p, gold.Point) {
		gold.T.Errorf("%s called with point %v, want %v", gold.ID, p, gold.Point)
	}
	if !reflect.DeepEqual(attrs, gold.ExecAttrs) {
		gold.T.Errorf("%s called with %#v, want %#v", gold.ID, attrs, gold.ExecAttrs)
	}

	gold.Count++

	if gold.Term {
		go func() {
			time.Sleep(100 * time.Microsecond)
			term()
		}()
	}
}

func (gold *GoldenExec) SingleCmd(id info.ID, p info.SinglePoint, attrs ExecAttrs, term ExecTerm) (ok bool) {
	gold.check(id, p, attrs, term)
	return gold.Accept
}

func (gold *GoldenExec) DoubleCmd(id info.ID, p info.DoublePoint, attrs ExecAttrs, term ExecTerm) (ok bool) {
	gold.check(id, p, attrs, term)
	return gold.Accept
}

func (gold *GoldenExec) FloatSetpoint(id info.ID, p float32, attrs ExecAttrs, term ExecTerm) (ok bool) {
	gold.check(id, p, attrs, term)
	return gold.Accept
}
