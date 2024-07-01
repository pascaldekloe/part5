package info

import (
	"math"
	"testing"
)

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

// TestNorm tests the full value range.
func TestNorm(t *testing.T) {
	b := Norm{0x00, 0x80}
	f := b.Float64()
	if f != -1 {
		t.Errorf("%#x: got %f, want -1", b, f)
	}

	for i := int(math.MinInt16 + 1); i <= int(math.MaxInt16); i++ {
		b.SetInt16(int16(i))
		got := b.Float64()
		if got <= f || got >= 1 {
			t.Errorf("%#x: got %f, want in range (%f, 1)", b, got, f)
		}
		f = got
	}
}
