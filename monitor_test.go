package part5

import (
	"testing"

	"github.com/pascaldekloe/part5/info"
)

// Be resillient against malicious and faulty ASDU.
func FuzzMonitorWideASDU(f *testing.F) {
	var sys info.Params[info.COT16, info.ComAddr16, info.ObjAddr24]
	mon := NewMonitor(sys)
	f.Fuzz(func(t *testing.T, bytes []byte) {
		u := sys.NewDataUnit()
		u.Adopt(bytes)
		mon.OnDataUnit(u)
	})
}
