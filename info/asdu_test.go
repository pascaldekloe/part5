package info

import "testing"

func mustAddr16(n uint) ComAddr16 {
	var p Params[Cause16, ComAddr16, Addr16]
	addr, ok := p.NewComAddr(n)
	if !ok {
		panic(n)
	}
	return addr
}

var goldenASDUs = []struct {
	unit DataUnit[Cause16, ComAddr16, Addr16]
	desc string
}{
	{
		DataUnit[Cause16, ComAddr16, Addr16]{
			Type:  M_SP_NA_1,
			Var:   1,
			Cause: Cause16{Percyc, 7},
			Addr:  mustAddr16(1001),
			Info:  []byte{1, 2, 3},
		},
		"M_SP_NA_1 @1001 percyc #7: 513:0x03 .",
	}, {
		DataUnit[Cause16, ComAddr16, Addr16]{
			Type:  M_DP_NA_1,
			Var:   2,
			Cause: Cause16{Back, 0},
			Addr:  mustAddr16(42),
			Info:  []byte{1, 2, 3, 4, 5, 6},
		},
		"M_DP_NA_1 @42 back #0: 513:0x03 1284:0x06 .",
	}, {
		DataUnit[Cause16, ComAddr16, Addr16]{
			Type:  M_ST_NA_1,
			Var:   2,
			Cause: Cause16{Spont, 21},
			Addr:  mustAddr16(250),
			Info:  []byte{1, 2, 3, 4, 5},
		},
		"M_ST_NA_1 @250 spont #21: 513:0x0304 0x05<EOF> !",
	}, {
		DataUnit[Cause16, ComAddr16, Addr16]{
			Type:  M_ST_NA_1,
			Var:   1,
			Cause: Cause16{Spont, 22},
			Addr:  mustAddr16(251),
			Info:  []byte{1, 2, 3, 4, 5, 6, 7},
		},
		"M_ST_NA_1 @251 spont #22: 513:0x0304 0x050607<EOF> 𝚫 +1 !",
	}, {
		DataUnit[Cause16, ComAddr16, Addr16]{
			Type:  M_ME_NC_1,
			Var:   2 | Sequence,
			Cause: Cause16{Init, 60},
			Addr:  mustAddr16(12),
			Info:  []byte{99, 0, 1, 2, 3, 4, 5, 6},
		},
		"M_ME_NC_1 @12 init #60: 99:0x0102030405 100:0x06<EOF> !",
	}, {
		DataUnit[Cause16, ComAddr16, Addr16]{
			Type:  M_ME_NC_1,
			Var:   2 | Sequence,
			Cause: Cause16{Init, 61},
			Addr:  mustAddr16(12),
			Info:  []byte{99, 0, 1, 2, 3, 4, 5},
		},
		"M_ME_NC_1 @12 init #61: 99:0x0102030405 𝚫 −1 !",
	},
}

// Verify labels for valid and invalid encodings.
func TestDataUnitString(t *testing.T) {
	for _, gold := range goldenASDUs {
		if got := gold.unit.String(); got != gold.desc {
			t.Errorf("got %q, want %q", got, gold.desc)
		}
	}
}

// Encode (with Append) and decode (with Adopt) should be lossless.
func TestASDUEncoding(t *testing.T) {
	for _, gold := range goldenASDUs {
		asdu := gold.unit.Append(nil)
		t.Logf("%s got encoded as %#x", gold.desc, asdu)

		var got DataUnit[Cause16, ComAddr16, Addr16]
		err := got.Adopt(asdu)
		switch {
		case err != nil:
			t.Errorf("encoding of %s got parse error: %s",
				gold.desc, err)

		case got.Type != gold.unit.Type,
			got.Var != gold.unit.Var,
			got.Cause != gold.unit.Cause,
			got.Addr != gold.unit.Addr,
			string(got.Info) != string(gold.unit.Info):
			t.Errorf("%+v became %+v after codec cycle",
				gold.unit, got)
		}
	}
}
