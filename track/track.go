// Package track follows address states.
package track

import "sync"

import "github.com/pascaldekloe/part5/info"

// Head is a database with the latest values.
type Head struct {
	db sync.Map
}

// Latest is a Head entry.
type latest struct {
	Type   info.TypeID
	Serial []byte
}

// Add merges all measured values from u into Head.
func (h *Head) Add(u *info.ASDU) {
	// store only measured values
	if u.Type >= info.C_SC_NA_1 {
		return
	}

	objSize := info.ObjSize[u.Type]
	if objSize == 0 {
		return
	}

	addr := u.ObjAddr(u.Info)
	i := u.ObjAddrSize
	for {
		end := i + objSize
		h.db.Store(addr, &latest{
			Type:   u.Type,
			Serial: u.Info[i:end],
		})
		if end >= len(u.Info) {
			break
		}

		if u.Struct&info.Sequence != 0 {
			addr++
			i = end
		} else {
			addr = u.ObjAddr(u.Info[end:])
			i = end + u.ObjAddrSize
		}
	}
}

// Inro responds to an interrogation request C_IC_NA_1.
func (h *Head) Inro(req *info.ASDU, resp chan<- *info.ASDU) {
	if req.Type != info.C_IC_NA_1 {
		panic("not an interrogation request")
	}

	// check cause of transmission
	if req.Cause&127 != info.Act {
		resp <- req.Reply(info.UnkCause)
		return
	}

	// check payload
	if len(req.Info) != req.ObjAddrSize+1 || req.ObjAddr(req.Info) != info.IrrelevantAddr {
		resp <- req.Reply(info.UnkInfo)
		return
	}

	// Qualifier of interrogation command numeric value matches cause
	// of transmission value for generic interrogation and group 1‥16.
	cause := info.Cause(req.Info[req.ObjAddrSize])
	if cause < info.Inrogen || cause > info.Inro16 {
		resp <- req.Reply(info.Actcon | info.NegFlag)
		return
	}

	// confirm
	resp <- req.Reply(info.Actcon)

	h.db.Range(func(key, value interface{}) bool {
		addr := key.(info.ObjAddr)
		l := value.(*latest)

		u := req.Respond(l.Type, cause)
		if err := u.AddObjAddr(addr); err != nil {
			panic(err)
		}
		u.Info = append(u.Info, l.Serial...)

		resp <- u
		return true
	})

	// terminate
	resp <- req.Reply(info.Actterm)
}
