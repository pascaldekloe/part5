package info

import (
	"encoding/hex"
	"testing"
)

func TestParseASDU(t *testing.T) {
	serial, err := hex.DecodeString("0d02c303100000c6428111286b6ece00")
	if err != nil {
		t.Fatal(err)
	}
	want := NewASDU(Narrow, 3, M_ME_NC_1, Spont|NegFlag|TestFlag)
	want.Info, err = hex.DecodeString("100000c6428111286b6ece00")
	if err != nil {
		t.Fatal(err)
	}

	got := NewASDU(Narrow, 0, 0, 0)
	if err := got.UnmarshalBinary(serial); err != nil {
		t.Fatal("unmarshal:", err)
	}
	if g, w := got.String(), want.String(); g != w {
		t.Errorf("got %s, want %s", g, w)
	}
}
