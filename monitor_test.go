package part5

import (
	"io"
	"testing"

	"github.com/pascaldekloe/part5/info"
)

// Be resillient against malicious and faulty ASDU.
func FuzzMonitorWideASDU(f *testing.F) {
	var sys info.Params[info.OrigAddr8, info.ComAddr16, info.ObjAddr24]
	mon := NewMonitorDelegateDefault(NewLogger(sys, io.Discard))

	f.Fuzz(func(t *testing.T, bytes []byte) {
		u := sys.NewDataUnit()
		u.Adopt(bytes)
		ApplyDataUnit(mon, u)
	})
}
