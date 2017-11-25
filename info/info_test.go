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

func TestRegulCmd(t *testing.T) {
	c := NewRegulCmd(DeterminedOn, 31, true)
	if got := c.Higher(); got != DeterminedOn {
		t.Errorf("got higher %v, want %v", got, DeterminedOn)
	}
	if got := c.Qual(); got != 31 {
		t.Errorf("got qualifier %d, want 31", got)
	}
	if got := c.Exec(); got != true {
		t.Error("got execute false, want true")
	}

	c = NewRegulCmd(DeterminedOff, 0, false)
	if got := c.Higher(); got != DeterminedOff {
		t.Errorf("got higher %v, want %v", got, DeterminedOff)
	}
	if got := c.Qual(); got != 0 {
		t.Errorf("got qualifier %d, want 0", got)
	}
	if got := c.Exec(); got != false {
		t.Error("got execute true, want false")
	}
}
