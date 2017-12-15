package session

import (
	"bytes"
	"encoding/hex"
	"io"
	"math/rand"
	"strings"
	"testing"
	"testing/iotest"
)

// TestUFormat tests all 6 function compositions.
func TestUFormat(t *testing.T) {
	var golden = []struct {
		f            function
		name, serial string
	}{
		{bringUp, "STARTDT_ACT", "680407000000"},
		{bringUpOK, "STARTDT_CON", "68040b000000"},
		{bringDown, "STOPDT_ACT", "680413000000"},
		{bringDownOK, "STOPDT_CON", "680423000000"},
		{keepAlive, "TESTFR_ACT", "680443000000"},
		{keepAliveOK, "TESTFR_CON", "680483000000"},
		{bringUp | bringDown, "<illegal 0x14>", "680417000000"},
	}

	for _, gold := range golden {
		u := newFunc(gold.f)

		if got := u.Format(); got != uFrame {
			t.Errorf("%s(%s): got %c-frame", gold.serial, gold.name, got)
		}
		if got := u.Function().String(); got != gold.name {
			t.Errorf("%s(%s): got function %q", gold.serial, gold.name, got)
		}
		if got := hex.EncodeToString(u[:]); !strings.HasPrefix(got, gold.serial) {
			t.Errorf("%s(%s): got serial 0x%s", gold.serial, gold.name, got)
		}

		var buf bytes.Buffer
		if n, err := u.Marshal(&buf, 0); n != 6 || err != nil {
			t.Errorf("%s(%s): marshall returned (%d, %#v)", gold.serial, gold.name, n, err)
		}
		if got := hex.EncodeToString(buf.Bytes()); got != gold.serial {
			t.Errorf("%s(%s): marshalled 0x%s", gold.serial, gold.name, got)
		}
	}
}

// TestSFormat tests all sequence number acknowledges.
func TestSFormat(t *testing.T) {
	for seqNo := uint(0); seqNo < 1<<15; seqNo++ {
		u := newAck(seqNo)

		if got := u.RecvSeqNo(); got != seqNo {
			t.Fatalf("got sequence number %d, want %d", got, seqNo)
		}
		if got := u.Format(); got != sFrame {
			t.Fatalf("acknowledge %d %c-frame", seqNo, got)
		}

		var buf bytes.Buffer
		if n, err := u.Marshal(&buf, 0); n != 6 || err != nil {
			t.Fatalf("acknowledge %d marshall returned (%d, %#v)", seqNo, n, err)
		}
		if got := hex.EncodeToString(buf.Bytes()); !strings.HasPrefix(got, "68040100") {
			t.Fatalf("acknowledge %d marshalled 0x%s", seqNo, got)
		}
	}
}

// TestIFormat test all sequence numbers with ASDU payloads.
func TestIFormat(t *testing.T) {
	var u apdu
	rand.Read(u[:]) // populate with junk

	var feed [249]byte
	for i := range feed {
		feed[i] = byte(i + 1)
	}

	for seqNo := uint(0); seqNo < 1<<15; seqNo++ {
		seqNo2 := (seqNo + 1) % (1 << 15)
		asduLen := int(seqNo) % len(feed)

		u, err := packASDU(feed[:asduLen], seqNo, seqNo2)
		if err != nil {
			t.Fatal("ASDU wrap error:", err)
		}

		if got := u.SendSeqNo(); got != seqNo {
			t.Fatalf("got send sequence number %d, want %d", got, seqNo)
		}
		if got := u.RecvSeqNo(); got != seqNo2 {
			t.Fatalf("got receive sequence number %d, want %d", got, seqNo2)
		}
		got := u.Payload()
		if len(got) != asduLen || (len(got) != 0 && int(got[len(got)-1]) != asduLen) {
			t.Fatalf("want %d byte payload, got %#x", asduLen, got)
		}

		if got := u.Format(); got != iFrame {
			t.Error("got ", got)
		}
	}
}

var goldenUnmarshals = []struct {
	hex string
	err error
}{
	{"", io.EOF},
	{"79", errStart},
	{"68", io.ErrUnexpectedEOF},
	{"6800", errLength},
	{"6801", errLength},
	{"68fe", errLength},
	{"680100", errLength},
	{"680200", errLength},
	{"68020000", errLength},
	{"68030000", errLength},
	{"6803000000", errLength},
	{"6804", io.ErrUnexpectedEOF},
	{"680400", io.ErrUnexpectedEOF},
	{"68040000", io.ErrUnexpectedEOF},
	{"6804000000", io.ErrUnexpectedEOF},
	{"680400000000", nil},
	{"68040100ffff", nil},
	{"68050100ffffee", nil},
	{"680407000000", nil},
	{"570400000000", errStart},
	{"68fd" + strings.Repeat("00", 253), nil},
}

func TestUnmarshals(t *testing.T) {
	for _, gold := range goldenUnmarshals {
		serial, err := hex.DecodeString(gold.hex)
		if err != nil {
			t.Fatal(err)
		}

		var u apdu
		n, err := u.Unmarshal(iotest.OneByteReader(bytes.NewReader(serial)), 0)

		if gold.err != nil {
			if err != gold.err {
				t.Errorf("%s: got error %#v, want %#v", gold.hex, err, gold.err)
			}

			continue
		}

		if n != len(serial) {
			t.Errorf("%s: unmarshaled %d bytes", gold.hex, n)
		}
		if err != nil {
			t.Errorf("%s: unmarshal error: %s", gold.hex, err)
		}
	}
}
