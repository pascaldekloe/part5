package info

import (
	"fmt"
	"testing"
)

// Wide tests multi-byte addressing.
var Wide Params[OrigAddr8, ComAddr16, ObjAddr16]

var goldenDataUnits = []struct {
	unit DataUnit[OrigAddr8, ComAddr16, ObjAddr16]
	desc string
}{
	{
		DataUnit[OrigAddr8, ComAddr16, ObjAddr16]{
			Type:  M_SP_NA_1,
			Enc:   1,
			Cause: Cyclic,
			Orig:  Wide.MustOrigAddrN(7),
			Addr:  Wide.MustComAddrN(1001),
			Info:  []byte{1, 2, 3},
		},
		"M_SP_NA_1 cyclic 7 1001: 0x03@513 .",
	}, {
		DataUnit[OrigAddr8, ComAddr16, ObjAddr16]{
			Type:  M_DP_NA_1,
			Enc:   2,
			Cause: Back,
			Addr:  Wide.MustComAddrN(42),
			Info:  []byte{1, 2, 3, 4, 5, 6},
		},
		"M_DP_NA_1 back 0 42: 0x03@513 0x06@1284 .",
	},

	// empty not explicitly prohibited by spec
	{
		DataUnit[OrigAddr8, ComAddr16, ObjAddr16]{
			Type:  M_SP_NA_1,
			Enc:   0,
			Cause: Cyclic,
			Addr:  Wide.MustComAddrN(404),
			Info:  []byte{},
		},
		"M_SP_NA_1 cyclic 0 404: .",
	}, {
		DataUnit[OrigAddr8, ComAddr16, ObjAddr16]{
			Type:  M_DP_NA_1,
			Enc:   0 | 128,
			Cause: Back,
			Addr:  Wide.MustComAddrN(302),
			Info:  []byte{1, 2},
		},
		"M_DP_NA_1 back 0 302: SQ@513 .",
	},

	// address sequence with overflow is undefined behavour
	{
		DataUnit[OrigAddr8, ComAddr16, ObjAddr16]{
			Type:  M_DP_NA_1,
			Enc:   2 | 128,
			Cause: Back,
			Addr:  Wide.MustComAddrN(666),
			Info:  []byte{0xff, 0xff, 3, 4},
		},
		"M_DP_NA_1 back 0 666: SQ@65535 0x03 0x04 @^ !",
	},

	// incomplete or additional data
	{
		DataUnit[OrigAddr8, ComAddr16, ObjAddr16]{
			Type:  M_ST_NA_1,
			Enc:   2,
			Cause: Spont,
			Orig:  Wide.MustOrigAddrN(21),
			Addr:  Wide.MustComAddrN(250),
			Info:  []byte{1, 2, 3, 4, 5},
		},
		"M_ST_NA_1 spont 21 250: 0x0102030405 ~2 ?",
	}, {
		DataUnit[OrigAddr8, ComAddr16, ObjAddr16]{
			Type:  M_ST_NA_1,
			Enc:   1 | 128,
			Cause: Spont,
			Orig:  Wide.MustOrigAddrN(22),
			Addr:  Wide.MustComAddrN(251),
			Info:  []byte{1},
		},
		"M_ST_NA_1 spont 22 251: SQ @ 0x01<EOF> ~1 !",
	}, {
		DataUnit[OrigAddr8, ComAddr16, ObjAddr16]{
			Type:  M_ME_NC_1,
			Enc:   2 | 128,
			Cause: Init,
			Addr:  Wide.MustComAddrN(12),
			Info:  []byte{99, 0, 1, 2, 3, 4, 5},
		},
		"M_ME_NC_1 init 0 12: SQ@99 0x0102030405 ~2 ?",
	},
}

// Verify labels for both valid and invalid encodings.
func TestDataUnitDesc(t *testing.T) {
	for _, gold := range goldenDataUnits {
		got := fmt.Sprintf("%s", gold.unit)
		if got != gold.desc {
			t.Errorf("got %q, want %q", got, gold.desc)
		}
	}
}

// Encode (with Append) and decode (with Adopt) should be lossless.
func TestDataUnitEncoding(t *testing.T) {
	for _, gold := range goldenDataUnits {
		asdu := gold.unit.Append(nil)
		t.Logf("%s got encoded as %#x", gold.desc, asdu)

		got := Wide.NewDataUnit()
		err := got.Adopt(asdu)
		switch {
		case err != nil:
			t.Errorf("encoding of %s got parse error: %s",
				gold.desc, err)

		case got.Type != gold.unit.Type,
			got.Enc != gold.unit.Enc,
			got.Cause != gold.unit.Cause,
			got.Orig != gold.unit.Orig,
			got.Addr != gold.unit.Addr,
			string(got.Info) != string(gold.unit.Info):
			t.Errorf("%#s became %#s after codec cycle",
				gold.unit, got)
		}
	}
}
