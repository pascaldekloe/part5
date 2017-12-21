package info

import (
	"strings"
	"testing"
)

var goldenASDUs = []struct {
	u *ASDU
	s string
}{
	{
		&ASDU{
			Params: Wide,
			ID:     ID{1001, 7, M_SP_NA_1, Percyc},
			Info:   []byte{1, 2, 3, 4},
		},
		"7@1001 M_SP_NA_1 <percyc> 197121:0x04",
	}, {
		&ASDU{
			Params: Narrow,
			ID:     ID{42, 0, M_DP_NA_1, Back},
			Info:   []byte{1, 2, 3, 4},
		},
		"@42 M_DP_NA_1 <back> 1:0x02 3:0x04",
	}, {
		&ASDU{
			Params: Narrow,
			ID:     ID{250, 0, M_ST_NA_1, Spont},
			Info:   []byte{1, 2, 3, 4, 5},
		},
		"@250 M_ST_NA_1 <spont> 1:0x0203 4:0x05 <EOF>",
	}, {
		&ASDU{
			Params:  Narrow,
			ID:      ID{12, 0, M_ME_NC_1, Init},
			InfoSeq: true,
			Info:    []byte{99, 0, 1, 2, 3, 4, 5},
		},
		"@12 M_ME_NC_1 <init> 99:0x0001020304 100:0x05 <EOF>",
	},
}

func TestASDUStrings(t *testing.T) {
	for _, gold := range goldenASDUs {
		if got := gold.u.String(); got != gold.s {
			t.Errorf("got %q, want %q", got, gold.s)
		}
	}
}

func TestASDUEncoding(t *testing.T) {
	for _, gold := range goldenASDUs {
		if strings.Contains(gold.s, " <EOF>") {
			continue
		}

		bytes, err := gold.u.MarshalBinary()
		if err != nil {
			t.Error(gold.s, "marshal error:", err)
			continue
		}

		u := NewASDU(gold.u.Params, ID{})
		if err = u.UnmarshalBinary(bytes); err != nil {
			t.Error(gold.s, "unmarshal error:", err)
			continue
		}

		if got := u.String(); got != gold.s {
			t.Errorf("got %q, want %q", got, gold.s)
		}
	}
}
