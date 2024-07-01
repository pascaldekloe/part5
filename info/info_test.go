package info

import "testing"

// TestStepPos tests the full value range.
func TestStepPos(t *testing.T) {
	for value := -64; value <= 63; value++ {
		if v, tr := NewStepPosQual(value).Pos(); v != value || tr {
			t.Errorf("got position and tranient (%d, %t), want (%d, false)", v, tr, value)
		}
		if v, tr := NewTransientStepPosQual(value).Pos(); v != value || !tr {
			t.Errorf("got position and tranient (%d, %t), want (%d, true)", v, tr, value)
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
