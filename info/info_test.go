package info

import "testing"

func TestStepPos(t *testing.T) {
	_, transient, qualDesc := NewStepPos(0, true, Invalid).Split()
	if !transient {
		t.Error("got transient false, want true")
	}
	if qualDesc != Invalid {
		t.Errorf("got quality descriptor %#x, want %#x", qualDesc, Invalid)
	}

	_, transient, qualDesc = NewStepPos(0, false, Blocked).Split()
	if transient {
		t.Error("got transient true, want false")
	}
	if qualDesc != Blocked {
		t.Errorf("got quality descriptor %#x, want %#x", qualDesc, Blocked)
	}

	// test full value range
	for value := -64; value <= 63; value++ {
		p := NewStepPos(value, false, OK)
		if got, _, _ := p.Split(); got != value {
			t.Errorf("%d: got value %d from %#02x", value, got, p)
		}
	}
}
func TestStepCmd(t *testing.T) {
	c := NewStepCmd(DeterminedOn, 31, true)
	if got := c.Higher(); got != DeterminedOn {
		t.Errorf("got higher %v, want %v", got, DeterminedOn)
	}
	if got := c.Qual(); got != 31 {
		t.Errorf("got qualifier %d, want 31", got)
	}
	if got := c.Exec(); got != true {
		t.Error("got execute false, want true")
	}

	c = NewStepCmd(DeterminedOff, 0, false)
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
