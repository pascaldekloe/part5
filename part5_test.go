package part5

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/pascaldekloe/part5/info"
)

// GoldenMeasure is a recoding listener.
type GoldenMeasure struct {
	info.ID
	MeasureAttrs
	Val interface{}

	T     *testing.T // report target
	Count int        // number of calls
}

// NewGoldenExec returns a new recording listener with the given prameters set,
// Count at 0 and the rest randomized.
func NewGoldenMeasure(t *testing.T, id info.TypeID, val interface{}) *GoldenMeasure {
	g := &GoldenMeasure{
		ID: info.ID{
			Addr:  info.CommonAddr(rand.Intn(1 << 23)),
			Type:  id,
			Cause: info.Cause(rand.Intn(1 << 7)),
		},
		MeasureAttrs: MeasureAttrs{
			Addr:     info.ObjAddr(rand.Intn(1 << 23)),
			QualDesc: uint(rand.Intn(1<<3) << 4),
		},
		Val: val,
		T:   t,
	}

	if id.String()[5] == 'T' && rand.Intn(2) == 0 {
		now := time.Now()
		g.Time = &now
	}

	return g
}

func (gold *GoldenMeasure) check(id info.ID, m interface{}, attrs MeasureAttrs) {
	if id != gold.ID {
		gold.T.Errorf("called with %#v, want %#v", id, gold.ID)
	}
	if !reflect.DeepEqual(m, gold.Val) {
		gold.T.Errorf("%s[%d] called with value %v, want %v", gold.Type, gold.ID.Addr, m, gold.Val)
	}
	if !reflect.DeepEqual(attrs, gold.MeasureAttrs) {
		gold.T.Errorf("%s[%d] called with %#v, want %#v", gold.Type, gold.ID.Addr, attrs, gold.MeasureAttrs)
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
	Val    interface{}
	Accept bool

	T     *testing.T // report target
	Count int        // number of calls
}

// NewGoldenExec returns a new recording listener with the given prameters set,
// Count at 0 and the rest randomized.
func NewGoldenExec(t *testing.T, id info.TypeID, val interface{}) *GoldenExec {
	g := &GoldenExec{
		ID: info.ID{
			Addr:  info.CommonAddr(rand.Intn(1 << 23)),
			Type:  id,
			Cause: info.Cause(rand.Intn(1 << 7)),
		},
		ExecAttrs: ExecAttrs{
			Addr:     info.ObjAddr(rand.Intn(1 << 23)),
			Qual:     uint(rand.Intn(64)),
			InSelect: rand.Intn(2) == 0,
		},
		Accept: rand.Intn(2) == 0,
		Val:    val,
		T:      t,
	}

	if id.String()[5] == 'T' && rand.Intn(2) == 0 {
		now := time.Now()
		g.Time = &now
	}

	return g
}

func (gold *GoldenExec) check(id info.ID, m interface{}, attrs ExecAttrs) {
	if id != gold.ID {
		gold.T.Errorf("called with %#v, want %#v", id, gold.ID)
	}
	if !reflect.DeepEqual(m, gold.Val) {
		gold.T.Errorf("%s[%d] called with value %v, want %v", gold.Type, gold.ID.Addr, m, gold.Val)
	}
	if !reflect.DeepEqual(attrs, gold.ExecAttrs) {
		gold.T.Errorf("%s[%d] called with %#v, want %#v", gold.Type, gold.ID.Addr, attrs, gold.ExecAttrs)
	}

	gold.Count++
}

func (gold *GoldenExec) SingleCmd(id info.ID, m info.SinglePoint, attrs ExecAttrs, term ExecTerm) (ok bool) {
	gold.check(id, m, attrs)
	return gold.Accept
}

func (gold *GoldenExec) DoubleCmd(id info.ID, m info.DoublePoint, attrs ExecAttrs, term ExecTerm) (ok bool) {
	gold.check(id, m, attrs)
	return gold.Accept
}

func (gold *GoldenExec) FloatSetpoint(id info.ID, m float32, attrs ExecAttrs, term ExecTerm) (ok bool) {
	gold.check(id, m, attrs)
	return gold.Accept
}
