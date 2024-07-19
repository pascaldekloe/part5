package part5

import (
	"testing"

	"github.com/pascaldekloe/part5/info"
)

// Be resillient against malicious and faulty ASDU.
func FuzzMonitorWideASDU(f *testing.F) {
	mon := NewMonitor[info.Cause16, info.ComAddr16, info.Addr24]()
	f.Fuzz(func(t *testing.T, bytes []byte) {
		var u info.DataUnit[info.Cause16, info.ComAddr16, info.Addr24]
		u.Adopt(bytes)
		mon.OnDataUnit(u)
	})
}
