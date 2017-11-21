package info

import (
	"encoding/hex"
	"reflect"
	"testing"
)

func TestParseASDU(t *testing.T) {
	serial, err := hex.DecodeString("0d02c303100000c6428111286b6ece00")
	if err != nil {
		t.Fatal(err)
	}

	got := Narrow.NewASDU(0, 0, 0)
	if err := got.UnmarshalBinary(serial); err != nil {
		t.Fatal("unmarshal:", err)
	}

	want := Narrow.NewASDU(M_ME_NC_1, Spont|NegFlag|TestFlag, 3)
	want.Info = serial[4:]

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %s, want %s", got, want)
	}
}
