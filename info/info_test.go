package info

import "testing"

// TestStepPos tests the full value range.
func TestStepPos(t *testing.T) {
	if _, transient := NewStepPos(0, true).Pos(); !transient {
		t.Error("got transient false, want true")
	}
	if _, transient := NewStepPos(0, false).Pos(); transient {
		t.Error("got transient true, want false")
	}

	for _, transient := range []bool{false, true} {
		for value := -64; value <= 63; value++ {
			gotValue, gotTransient := NewStepPos(value, transient).Pos()
			if gotValue != value || gotTransient != transient {
				t.Errorf("StepPos(%d, %t) became (%d, %t)", value, transient, gotValue, gotTransient)
			}
		}
	}
}

// TestNormal tests the full value range.
func TestNormal(t *testing.T) {
	v := Normal(-1 << 15)
	last := v.Float64()
	if last != -1 {
		t.Errorf("%#04x: got %f, want -1", uint16(v), last)
	}

	for v != 1<<15-1 {
		v++
		got := v.Float64()
		if got <= last || got >= 1 {
			t.Errorf("%#04x: got %f (%#04x was %f)", uint16(v), got, uint16(v-1), last)
		}
		last = got
	}
}
