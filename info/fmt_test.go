package info

import (
	"fmt"
	"testing"
)

var goldenFormats = []struct {
	value  any
	format string
	want   string
}{
	// numeric value
	{Addr8{1}, "%d", "1"},
	{Addr16{2, 1}, "%d", "258"},
	{Addr24{3, 2, 1}, "%d", "66051"},
	{&Addr8{1}, "% d", "  1"},
	{Addr16{2, 1}, "% d", "  258"},
	{Addr24{3, 2, 1}, "% d", "   66051"},
	{Addr8{0xa}, "%x", "0a"},
	{&Addr16{0xb, 0xa}, "%x", "0a0b"},
	{Addr24{0xc, 0xb, 0xa}, "%x", "0a0b0c"},
	{Addr8{0xa}, "%X", "0A"},
	{Addr16{0xb, 0xa}, "%X", "0A0B"},
	{&Addr24{0xc, 0xb, 0xa}, "%X", "0A0B0C"},

	// alternate format
	{Addr8{1}, "%#d", "1"},
	{Addr16{1, 2}, "%#d", "2.1"},
	{Addr24{1, 2, 3}, "%#d", "3.2.1"},
	{Addr8{0xa}, "%#x", "0a"},
	{Addr16{0xa, 0xb}, "%#x", "0b:0a"},
	{Addr24{0xa, 0xb, 0xc}, "%#x", "0c:0b:0a"},
	{Addr8{0xa}, "%#X", "0A"},
	{Addr16{0xa, 0xb}, "%#X", "0B:0A"},
	{Addr24{0xa, 0xb, 0xc}, "%#X", "0C:0B:0A"},
}

func TestGoldenFormats(t *testing.T) {
	for _, gold := range goldenFormats {
		got := fmt.Sprintf(gold.format, gold.value)
		if got != gold.want {
			t.Errorf("%q of %T got %q, want %q",
				gold.format, gold.value, got, gold.want)
		}
	}
}
