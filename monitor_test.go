package part5

import (
	"bytes"
	"testing"

	"github.com/pascaldekloe/part5/info"
)

// Be resillient against malicious and faulty ASDU.
func FuzzMonitorWideASDU(f *testing.F) {
	var sys info.Params[info.OrigAddr8, info.ComAddr16, info.ObjAddr24]

	f.Fuzz(func(t *testing.T, asdu []byte) {
		var buf bytes.Buffer
		mon := NewMonitorDelegateDefault(NewLogger(sys, &buf))

		u := sys.NewDataUnit()
		u.Adopt(asdu)
		MonitorDataUnit(mon, u)

		if bytes.IndexByte(buf.Bytes(), '!') >= 0 {
			t.Error("got formatting error in logger output:", buf)
		}
	})
}
