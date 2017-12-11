package part5

import (
	"reflect"
	"testing"

	"github.com/pascaldekloe/part5/info"
)

func TestDelegateMeasures(t *testing.T) {
	gold := []*GoldenMeasure{
		NewGoldenMeasure(t, info.M_SP_NA_1, info.On),
		NewGoldenMeasure(t, info.M_DP_TA_1, info.DeterminedOff),
		NewGoldenMeasure(t, info.M_ME_TF_1, float32(99.8)),
	}

	var d = &Delegate{
		Singles: map[info.ObjAddr]Single{gold[0].MeasureAttrs.Addr: gold[0].Single},
		Doubles: map[info.ObjAddr]Double{gold[1].MeasureAttrs.Addr: gold[1].Double},
		Floats:  map[info.ObjAddr]Float{gold[2].MeasureAttrs.Addr: gold[2].Float},
	}
	t.Log("setup:", d.String())

	for _, g := range gold {
		d.single(gold[0].ID, info.On, g.MeasureAttrs)
		d.double(gold[1].ID, info.DeterminedOff, g.MeasureAttrs)
		d.float(gold[2].ID, 99.8, g.MeasureAttrs)
	}

	for _, g := range gold {
		if g.Count != 1 {
			t.Errorf("%s listener called %d times", g.Type, g.Count)
		}
	}
}

func TestDelegateDirectExecute(t *testing.T) {
	gold := NewGoldenExec(t, info.C_SE_NC_1, float32(42))
	gold.ID.Addr = 99
	gold.Cause = info.Act
	gold.ExecAttrs = ExecAttrs{
		Addr:     7,
		InSelect: false,
		Qual:     9,
	}

	exec := &info.ASDU{Params: info.Narrow, ID: gold.ID}
	exec.Info = []byte{7, 0x00, 0x00, 0x28, 0x42, 0x09}

	d := &Delegate{
		FloatSetpoints: map[info.ObjAddr]FloatSetpoint{
			gold.ExecAttrs.Addr: gold.FloatSetpoint,
		},
	}
	t.Log("setup:", d.String())

	out := make(chan *info.ASDU, 4)
	if err := d.floatSetpoint(exec, nil, out); err != nil {
		t.Error("delegate error:", err)
	}

	if gold.Count != 1 {
		t.Errorf("listener called %d times", gold.Count)
	}

	want := []string{
		"@99 C_SE_NC_1 <actcon> 7:0x0000284209",
	}
	if got := getASDUStrings(out); !reflect.DeepEqual(got, want) {
		t.Errorf("response is %q, want %q", got, want)
	}
}

func TestDelegateSelectProcedure(t *testing.T) {
	gold := NewGoldenExec(t, info.C_DC_NA_1, info.DeterminedOn)
	gold.ID.Addr = 99
	gold.Cause = info.Act
	gold.ExecAttrs = ExecAttrs{
		Addr:     7,
		InSelect: true,
		Qual:     0,
	}
	gold.Accept = false

	sel := &info.ASDU{Params: info.Narrow, ID: gold.ID}
	sel.Info = []byte{7, 0x82}
	exec := &info.ASDU{Params: info.Narrow, ID: gold.ID}
	exec.Info = []byte{7, 0x02}

	d := &Delegate{
		DoubleCmds: map[info.ObjAddr]DoubleCmd{
			gold.ExecAttrs.Addr: gold.DoubleCmd,
		},
	}
	t.Log("setup:", d.String())

	in := make(chan *info.ASDU, 1)
	in <- exec
	out := make(chan *info.ASDU, 4)
	if err := d.doubleCmd(sel, in, out); err != nil {
		t.Error("delegate error:", err)
	}

	if gold.Count != 1 {
		t.Errorf("listener called %d times", gold.Count)
	}

	want := []string{
		"@99 C_DC_NA_1 <actcon> 7:0x82",
		"@99 C_DC_NA_1 <actcon,neg> 7:0x02",
	}
	if got := getASDUStrings(out); !reflect.DeepEqual(got, want) {
		t.Errorf("response is %q, want %q", got, want)
	}
}

// GetASDUStrings returns the string representation
// of each buffered entry and closes the channel.
func getASDUStrings(out chan *info.ASDU) []string {
	var a []string
	close(out)
	for u := range out {
		a = append(a, u.String())
	}
	return a
}
