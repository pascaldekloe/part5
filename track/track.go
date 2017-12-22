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

	serial := u.Info
	if u.InfoSeq {
		addr := u.GetObjAddrAt(0)
		for i := u.ObjAddrSize; i < len(serial); addr++ {
			offset := i
			i += objSize

			h.db.Store(addr, &latest{
				Type:   u.Type,
				Serial: serial[offset:i],
			})
		}
	} else {
		addrSize := u.ObjAddrSize
		for i := 0; i < len(serial); {
			addr := u.GetObjAddrAt(i)
			offset := i + addrSize
			i = offset + objSize

			h.db.Store(addr, &latest{
				Type:   u.Type,
				Serial: serial[offset:i],
			})
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
	if len(req.Info) != req.ObjAddrSize+1 || req.GetObjAddrAt(0) != 0 {
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
		if err := u.AppendAddr(addr); err != nil {
			panic(err)
		}
		u.Info = append(u.Info, l.Serial...)

		resp <- u
		return true
	})

	// terminate
	resp <- req.Reply(info.Actterm)
}
