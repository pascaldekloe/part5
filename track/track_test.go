package track

import (
	"reflect"
	"sort"
	"testing"

	"github.com/pascaldekloe/part5/info"
)

func TestInro(t *testing.T) {
	want := []string{
		"@99 C_IC_NA_1 <actcon,test> 0:0x1a",
		"@99 M_ME_NC_1 <inro6,test> 42:0x0000803f00",
		"@99 M_ME_NC_1 <inro6,test> 43:0x0000004040",
		"@99 M_ME_NC_1 <inro6,test> 44:0x0000404000",
		"@99 C_IC_NA_1 <actterm,test> 0:0x1a",
	}

	u1 := info.NewASDU(info.Narrow, info.ID{
		Addr:  99,
		Type:  info.M_ME_NC_1,
		Cause: info.Percyc | info.TestFlag,
	})
	u1.AddFloat(42, info.MeasuredFloat{1, info.OK})
	u1.AddFloat(44, info.MeasuredFloat{4, info.OK})

	u2 := info.NewASDU(info.Narrow, info.ID{
		Addr:  99,
		Type:  info.M_ME_NC_1,
		Cause: info.Back | info.TestFlag,
	})
	u2.SetFloats(43, info.MeasuredFloat{2, info.NotTopical}, info.MeasuredFloat{3, info.OK})

	var h Head
	h.Add(u1)
	h.Add(u2)

	req := info.NewASDU(info.Narrow, info.ID{
		Addr:  99,
		Type:  info.C_IC_NA_1,
		Cause: info.Act | info.TestFlag,
	})
	req.SetInro(26)
	out := make(chan *info.ASDU, 100)
	h.Inro(req, out)
	close(out)

	got := make([]string, 0, len(out))
	for u := range out {
		got = append(got, u.String())
	}

	if len(got) != 5 {
		t.Fatalf("want 5 ASDUs, got %q", got)
	}
	sort.Strings(got[1:4])
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got:\n  %q\nwant:\n  %q", got, want)
	}
}
