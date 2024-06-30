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
		"@99 M_ME_NC_1 <inro6,test> 42:0x0101010100",
		"@99 M_ME_NC_1 <inro6,test> 43:0x0303030340",
		"@99 M_ME_NC_1 <inro6,test> 44:0x0404040400",
		"@99 C_IC_NA_1 <actterm,test> 0:0x1a",
	}

	u1 := info.NewASDU(info.Narrow, info.ID{
		Type:   info.M_ME_NC_1,
		Struct: 2,
		Cause:  info.Percyc | info.TestFlag,
		Addr:   99,
	})
	u1.Info = []byte{42, 1, 1, 1, 1, info.OK, 44, 2, 2, 2, 2, info.OK}

	u2 := info.NewASDU(info.Narrow, info.ID{
		Type:   info.M_ME_NC_1,
		Struct: 2 | info.Sequence,
		Cause:  info.Back | info.TestFlag,
		Addr:   99,
	})
	u2.Info = []byte{43, 3, 3, 3, 3, info.NotTopical, 4, 4, 4, 4, info.OK}

	var h Head
	h.Add(u1)
	h.Add(u2)

	req := info.MustNewInro(info.Narrow, 99, 0, 6)
	req.Cause |= info.TestFlag
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
