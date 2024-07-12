package info

import (
	"math"
	"testing"
)

func TestQualString(t *testing.T) {
	if s := Qual(0).String(); s != "OK" {
		t.Errorf("0 got %q, want OK", s)
	}
	if s := Qual(Overflow).String(); s != "OV" {
		t.Errorf("Overflow got %q, want OV", s)
	}
	if s := Qual(0x02).String(); s != "RES2" {
		t.Errorf("0x02 got %q, want RES2", s)
	}
	if s := Qual(0x06).String(); s != "RES2,RES3" {
		t.Errorf("0x06 got %q, want RES2,RES3", s)
	}
	if s := Qual(Blocked | NotTopical).String(); s != "BL,NT" {
		t.Errorf("Blocked|NotTopical got %q, want BL,NT", s)
	}
	if s := Qual(ElapsedTimeInvalid | Substituted | Invalid).String(); s != "EI,SB,IV" {
		t.Errorf("ElapsedTimeInvalid|Substituted|Invalid got %q, want EI,SB,IV", s)
	}
}

// TestStep tests the full value range.
func TestStep(t *testing.T) {
	for value := -64; value <= 63; value++ {
		if v, tr := NewStepQual(value, NotTopical).Value().Pos(); v != value || tr {
			t.Errorf("got position and tranient (%d, %t), want (%d, false)", v, tr, value)
		}
		if v, tr := NewTransientStepQual(value, OK).Value().Pos(); v != value || !tr {
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
