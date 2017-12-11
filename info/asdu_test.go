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
	want := NewASDU(Narrow, ID{
		Addr:  3,
		Type:  M_ME_NC_1,
		Cause: Spont | NegFlag | TestFlag,
	})
	want.Info, err = hex.DecodeString("100000c6428111286b6ece00")
	if err != nil {
		t.Fatal(err)
	}

	got := &ASDU{Params: Narrow}
	if err := got.UnmarshalBinary(serial); err != nil {
		t.Fatal("unmarshal:", err)
	}
	if g, w := got.String(), want.String(); g != w {
		t.Errorf("got %s, want %s", g, w)
	}
}

var goldenASDUStrings = []struct {
	feed *ASDU
	want string
}{
	{
		&ASDU{
			Params: Wide,
			ID:     ID{1001, 7, M_SP_NA_1, Percyc},
			Info:   []byte{1, 2, 3, 4},
		},
		"7@1001 M_SP_NA_1 <percyc> 197121:0x04",
	},
	{
		&ASDU{
			Params: Narrow,
			ID:     ID{42, 0, M_DP_NA_1, Back},
			Info:   []byte{1, 2, 3, 4},
		},
		"@42 M_DP_NA_1 <back> 1:0x02 3:0x04",
	},
	{
		&ASDU{
			Params: Narrow,
			ID:     ID{255, 0, M_ST_NA_1, Spont},
			Info:   []byte{1, 2, 3, 4, 5},
		},
		"@255 M_ST_NA_1 <spont> 1:0x0203 4:0x05 <EOF>",
	},
	{
		&ASDU{
			Params:  Narrow,
			ID:      ID{12, 0, M_ME_NC_1, Init},
			InfoSeq: true,
			Info:    []byte{99, 0, 1, 2, 3, 4, 5},
		},
		"@12 M_ME_NC_1 <init> 99:0x0001020304 100:0x05 <EOF>",
	},
}

func TestGoldenASDUStrings(t *testing.T) {
	for _, gold := range goldenASDUStrings {
		if got := gold.feed.String(); got != gold.want {
			t.Errorf("got %q, want %q", got, gold.want)
		}
	}
}
