package info

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

var goldenAddrFormats = []struct {
	value  any
	format string
	want   string
}{
	// numeric value
	{ObjAddr8{1}, "%d", "1"},
	{ObjAddr16{2, 1}, "%d", "258"},
	{ObjAddr24{3, 2, 1}, "%d", "66051"},
	{&ObjAddr8{1}, "% d", "  1"},
	{ObjAddr16{2, 1}, "% d", "  258"},
	{ObjAddr24{3, 2, 1}, "% d", "   66051"},
	{ObjAddr8{0xa}, "%x", "0a"},
	{&ObjAddr16{0xb, 0xa}, "%x", "0a0b"},
	{ObjAddr24{0xc, 0xb, 0xa}, "%x", "0a0b0c"},
	{ObjAddr8{0xa}, "%X", "0A"},
	{ObjAddr16{0xb, 0xa}, "%X", "0A0B"},
	{&ObjAddr24{0xc, 0xb, 0xa}, "%X", "0A0B0C"},

	// alternate format
	{ObjAddr8{1}, "%#d", "1"},
	{ObjAddr16{1, 2}, "%#d", "2.1"},
	{ObjAddr24{1, 2, 3}, "%#d", "3.2.1"},
	{ObjAddr8{0xa}, "%#x", "0a"},
	{ObjAddr16{0xa, 0xb}, "%#x", "0b:0a"},
	{ObjAddr24{0xa, 0xb, 0xc}, "%#x", "0c:0b:0a"},
	{ObjAddr8{0xa}, "%#X", "0A"},
	{ObjAddr16{0xa, 0xb}, "%#X", "0B:0A"},
	{ObjAddr24{0xa, 0xb, 0xc}, "%#X", "0C:0B:0A"},
}

func TestGoldenAddrFormats(t *testing.T) {
	for _, gold := range goldenAddrFormats {
		got := fmt.Sprintf(gold.format, gold.value)
		if got != gold.want {
			t.Errorf("%q of %T got %q, want %q",
				gold.format, gold.value, got, gold.want)
		}
	}
}

func FuzzDataUnitFormat(f *testing.F) {
	f.Fuzz(func(t *testing.T, typeID byte, enc byte, cause byte, origAddr byte, comAddr byte, info []byte) {
		u := DataUnit[OrigAddr8, ComAddr8, ObjAddr8]{
			Type:  TypeID(typeID),
			Enc:   Enc(enc),
			Cause: Cause(cause),
			Orig:  OrigAddr8{origAddr},
			Addr:  ComAddr8{comAddr},
			Info:  info,
		}
		got := fmt.Sprintf("%s\n%#s", &u, u)
		a, b, _ := strings.Cut(got, "\n")
		if a == "" || b == "" {
			t.Fatalf("got %q and %q", a, b)
		}

		// Check the ending first to be reasonable sure of a complete formatting.
		switch a[len(a)-1] {
		case '.', '?', '!':
			if a[len(a)-1] != b[len(b)-1] {
				t.Errorf("last character differs between %q and %q", a, b)
			}
		default:
			t.Errorf("last character not '.', '?' or '!' of %q (with %q)", a, b)
		}

		// ASDUs have a fixed data-unit identifier.
		switch DUI, _, _ := strings.Cut(a, ":"); {
		case DUI == "":
			t.Errorf("want data-unit identifier terminated by ':' in %q", a)
		case strings.Count(DUI, " ") != 3:
			t.Errorf("data-uint identifier %q of %q does not have 4 fields", DUI, a)
		}
	})
}

func TestCP24Time2aFormat(t *testing.T) {
	// timezone should be ignored
	timestamp := time.Date(2006, 01, 02, 15, 04, 05, 987000000, time.Local)

	want := func(t *testing.T, t2a CP24Time2a, format, want string) {
		t.Helper()
		got := fmt.Sprintf(format, t2a)
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	t.Run("Plain", func(t *testing.T) {
		var t2a CP24Time2a
		t2a.Set(timestamp)
		want(t, t2a, "%s", ":04:05.987")
	})

	t.Run("Invalid", func(t *testing.T) {
		var t2a CP24Time2a
		t2a.Set(timestamp)
		t2a[2] |= 0x80 // invalid
		want(t, t2a, "%s", ":04:05.987,IV")
	})

	t.Run("NoRES1", func(t *testing.T) {
		var t2a CP24Time2a
		t2a.Set(timestamp)
		want(t, t2a, "%.1s", ":04:05.987,RES1=0")
	})

	t.Run("InvalidAndRES1", func(t *testing.T) {
		var t2a CP24Time2a
		t2a.Set(timestamp)
		t2a[2] |= 0x80 // invalid
		t2a.FlagReserve1()
		want(t, t2a, "%.1s", ":04:05.987,IV,RES1=1")
	})
}

func TestCP56Time2aFormat(t *testing.T) {
	// timezone should be ignored
	timestamp := time.Date(2006, 01, 02, 15, 04, 05, 987000000, time.Local)

	want := func(t *testing.T, t2a CP56Time2a, format, want string) {
		t.Helper()
		got := fmt.Sprintf(format, t2a)
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	t.Run("Plain", func(t *testing.T) {
		var t2a CP56Time2a
		t2a.Set(timestamp)
		want(t, t2a, "%s", "06-01-02T15:04:05.987")
		want(t, t2a, "%.0s", "06-01-02T15:04:05.987")
		want(t, t2a, "%.0s", "06-01-02T15:04:05.987")
		want(t, t2a, "%.0s", "06-01-02T15:04:05.987")
		want(t, t2a, "%.1s", "06-01-02T15:04:05.987,RES1=0")
		want(t, t2a, "%.2s", "06-01-02T15:04:05.987,RES1=0,RES2=00")
		want(t, t2a, "%.3s", "06-01-02T15:04:05.987,RES1=0,RES2=00,RES3=0000")
		want(t, t2a, "%.4s", "06-01-02T15:04:05.987,RES1=0,RES2=00,RES3=0000")
	})

	t.Run("Reserves", func(t *testing.T) {
		var t2a CP56Time2a
		t2a.Set(timestamp)
		t2a.FlagReserve1()
		t2a.FlagReserve2(2)
		t2a.FlagReserve3(5)
		want(t, t2a, "%s", "06-01-02T15:04:05.987")
		want(t, t2a, "%.1s", "06-01-02T15:04:05.987,RES1=1")
		want(t, t2a, "%.2s", "06-01-02T15:04:05.987,RES1=1,RES2=10")
		want(t, t2a, "%.3s", "06-01-02T15:04:05.987,RES1=1,RES2=10,RES3=0101")
	})

	t.Run("Invalid", func(t *testing.T) {
		var t2a CP56Time2a
		t2a.Set(timestamp)
		t2a.FlagReserve2(1)
		t2a.FlagReserve3(10)
		t2a[2] |= 0x80 // invalid
		want(t, t2a, "%s", "06-01-02T15:04:05.987,IV")
		want(t, t2a, "%.3s", "06-01-02T15:04:05.987,IV,RES1=0,RES2=01,RES3=1010")
	})
}
