package info

import (
	"fmt"
	"math"
	"testing"
)

func TestStringer(t *testing.T) {
	tests := []struct {
		subject fmt.Stringer
		want    string
	}{
		{Qual(0), "[]"},
		{Qual(Overflow), "OV"},
		{Qual(0x02), "[2]"},
		{Qual(0x06), "[2],[3]"},
		{Qual(Blocked | NotTopical), "BL,NT"},
		{Qual(ElapsedTimeInvalid | Substituted | Invalid), "EI,SB,IV"},

		{NewSinglePtQual(On, 0), "On"},
		{NewSinglePtQual(Off, 2), "Off;[2]"},
		{NewSinglePtQual(Off, 4|8), "Off;[3],EI"},

		{NewDoublePtQual(IndeterminateOrIntermediate, 0), "IndeterminateOrIntermediate"},
		{NewDoublePtQual(DeterminatedOn, OK), "DeterminatedOn"},
		{NewDoublePtQual(DeterminatedOff, Substituted), "DeterminatedOff;SB"},
		{NewDoublePtQual(Indeterminate, Invalid|Blocked), "Indeterminate;BL,IV"},
	}

	for _, test := range tests {
		got := test.subject.String()
		if got != test.want {
			t.Errorf("%T %#v got %q, want %s",
				test.subject, test.subject, got, test.want)
		}
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
