package media

import (
	"bytes"
	"encoding/hex"
	"io"
	"testing"
	"testing/iotest"
)

type GoldenFT struct {
	Impl FT
	Feed string // hex serial

	Addr Addr
	Ctrl Ctrl
	Data string // hex
}

var GoldenFT12s = []GoldenFT{
	// Examples borrowed from Beckhoff Information System:
	{NewFT12(2, 20, Ack, Nack), "100b0c001716", 12, OK, ""},
	{NewFT12(2, 20, Ack, Nack), "680b0b68080c0065010a0c000000059516", 12, Data, "65010a0c00000005"},
}

func TestGoldenFT12Encodes(t *testing.T) {
	for _, gold := range GoldenFT12s {
		data, err := hex.DecodeString(gold.Data)
		if err != nil {
			t.Fatalf("%s: broken test data: %s", gold.Feed, err)
		}

		var buf bytes.Buffer
		err = gold.Impl.Encode(&buf, Packet{gold.Addr, gold.Ctrl, data})
		if err != nil {
			t.Errorf("%s: write error: %s", gold.Feed, err)
			continue
		}
		got := hex.EncodeToString(buf.Bytes())

		if got != gold.Feed {
			t.Errorf("%s: got data 0x%s", gold.Feed, got)
		}
	}
}

func TestGoldenFT12Decodes(t *testing.T) {
	for _, gold := range GoldenFT12s {
		feed, err := hex.DecodeString(gold.Feed)
		if err != nil {
			t.Fatalf("%s: broken test feed: %s", gold.Feed, err)
		}
		got, err := gold.Impl.Decode(iotest.OneByteReader(bytes.NewReader(feed)))
		if err != nil {
			t.Errorf("%s: read error: %s", gold.Feed, err)
			continue
		}

		if got.Addr != gold.Addr {
			t.Errorf("%s: got link station address %#x, want %#x", gold.Feed, got.Addr, gold.Addr)
		}
		if got.Ctrl != gold.Ctrl {
			t.Errorf("%s: got control field %#x, want %#x", gold.Feed, got.Ctrl, gold.Ctrl)
		}

		if s := hex.EncodeToString(got.Data); s != gold.Data {
			t.Errorf("%s: got data 0x%s, want 0x%s", gold.Feed, s, gold.Data)
		}
	}
}

func TestGoldenFT12Corruptions(t *testing.T) {
	for _, gold := range GoldenFT12s {
		feed, err := hex.DecodeString(gold.Feed)
		if err != nil {
			t.Fatalf("%s: broken test feed: %s", gold.Feed, err)
		}

		for i := range feed {
			broken := make([]byte, len(feed))
			copy(broken, feed)

			// toggle one byte
			broken[i]++

			got, err := gold.Impl.Decode(bytes.NewReader(broken))
			switch err {
			case ErrCheck:
				break // pass
			case nil:
				t.Errorf("%x: decode succes: %+v", broken, got)
			default:
				t.Errorf("%x: wrong error: %v", broken, err)
			}
		}
	}
}

func TestGoldenFT12EOF(t *testing.T) {
	for _, gold := range GoldenFT12s {
		feed, err := hex.DecodeString(gold.Feed)
		if err != nil {
			t.Fatalf("%s: broken test feed: %s", gold.Feed, err)
		}

		for i := range feed {
			_, err := gold.Impl.Decode(bytes.NewReader(feed[:i]))
			if i == 0 && err != io.EOF {
				t.Errorf(`%x: got error %q, want EOF`, feed[:i], err)
			}
			if i != 0 && err != io.ErrUnexpectedEOF {
				t.Errorf(`%x: got error %q, want unexpected EOF`, feed[:i], err)
			}
		}
	}
}
