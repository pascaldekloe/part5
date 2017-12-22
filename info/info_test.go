package info

import "testing"

func TestStepPos(t *testing.T) {
	if _, transient := NewStepPos(0, true).Pos(); !transient {
		t.Error("got transient false, want true")
	}
	if _, transient := NewStepPos(0, false).Pos(); transient {
		t.Error("got transient true, want false")
	}

	// test full value range
	for value := -64; value <= 63; value++ {
		p := NewStepPos(value, false)
		if got, _ := p.Pos(); got != value {
			t.Errorf("%d: got value %d from %#02x", value, got, p)
		}
	}
}
