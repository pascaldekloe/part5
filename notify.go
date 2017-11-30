package part5

import (
	"encoding/binary"
	"math"

	"github.com/pascaldekloe/part5/info"
)

// Notify informs Delegate about the ASDU.
func (d *Delegate) Notify(u *info.ASDU) {
	for _, f := range d.All {
		if !f(u) {
			return
		}
	}

	impl := notifyFunc[u.Type]
	if impl == nil {
		return
	}

	serial := u.Info
	objSize := info.ObjSize[u.Type]
	if u.InfoSeq {
		addr := u.GetObjAddrAt(0)
		for i := u.ObjAddrSize; i < len(serial); addr++ {
			offset := i
			i += objSize
			impl(serial[offset:i], addr, d, u)
		}
	} else {
		addrSize := u.ObjAddrSize
		for i := 0; i < len(serial); {
			addr := u.GetObjAddrAt(i)
			offset := i + addrSize
			i = offset + objSize
			impl(serial[offset:i], addr, d, u)
		}
	}
}

var notifyFunc = [256]func([]byte, info.ObjAddr, *Delegate, *info.ASDU){
	info.M_SP_NA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.1.
		p := uint(bytes[0])
		d.single(u.ID, info.SinglePoint(p&1), MeasureAttrs{
			Addr: addr,
			QualDesc: p & 0x70,
		})
	},

	info.M_SP_TA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.2.
		p := uint(bytes[0])
		d.single(u.ID, info.SinglePoint(p&1), MeasureAttrs{
			Addr: addr,
			QualDesc: p & 0x70,
			Time: getCP24Time2a(bytes[1:], u.TimeZone),
		})
	},

	info.M_DP_NA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.3.
		p := uint(bytes[0])
		d.double(u.ID, info.DoublePoint(p&3), MeasureAttrs{
			Addr: addr,
			QualDesc: p & 0x70,
		})
	},

	info.M_DP_TA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.4.
		p := uint(bytes[0])
		d.double(u.ID, info.DoublePoint(p&3), MeasureAttrs{
			Addr: addr,
			QualDesc: p & 0x70,
			Time: getCP24Time2a(bytes[1:], u.TimeZone),
		})
	},

	info.M_ST_NA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.5.
		d.step(u.ID, info.StepPos(bytes[0]), MeasureAttrs{
			Addr: addr,
			QualDesc: uint(bytes[1]),
		})
	},

	info.M_ST_TA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.6.
		d.step(u.ID, info.StepPos(bytes[0]), MeasureAttrs{
			Addr: addr,
			QualDesc: uint(bytes[1]),
			Time: getCP24Time2a(bytes[2:], u.TimeZone),
		})
	},

	info.M_BO_NA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.7.
		d.bits(u.ID, binary.BigEndian.Uint32(bytes), MeasureAttrs{
			Addr: addr,
			QualDesc: uint(bytes[4]),
		})
	},

	info.M_BO_TA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.8.
		d.bits(u.ID, binary.BigEndian.Uint32(bytes), MeasureAttrs{
			Addr: addr,
			QualDesc: uint(bytes[4]),
			Time: getCP24Time2a(bytes[5:], u.TimeZone),
		})
	},

	info.M_ME_NA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.9.
		d.normal(u.ID, info.Normal(binary.LittleEndian.Uint16(bytes)), MeasureAttrs{
			Addr: addr,
			QualDesc: uint(bytes[2]),
		})
	},

	info.M_ME_TA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.10.
		d.normal(u.ID, info.Normal(binary.LittleEndian.Uint16(bytes)), MeasureAttrs{
			Addr: addr,
			QualDesc: uint(bytes[2]),
			Time: getCP24Time2a(bytes[3:], u.TimeZone),
		})
	},

	info.M_ME_NB_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.11.
		d.scaled(u.ID, int16(binary.LittleEndian.Uint16(bytes)), MeasureAttrs{
			Addr: addr,
			QualDesc: uint(bytes[2]),
		})
	},

	info.M_ME_TB_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.12.
		d.scaled(u.ID, int16(binary.LittleEndian.Uint16(bytes)), MeasureAttrs{
			Addr: addr,
			QualDesc: uint(bytes[2]),
			Time: getCP24Time2a(bytes[3:], u.TimeZone),
		})
	},

	info.M_ME_NC_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.13.
		d.float(u.ID, math.Float32frombits(binary.LittleEndian.Uint32(bytes)), MeasureAttrs{
			Addr: addr,
			QualDesc: uint(bytes[4]),
		})
	},

	info.M_ME_TC_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.14.
		d.float(u.ID, math.Float32frombits(binary.LittleEndian.Uint32(bytes)), MeasureAttrs{
			Addr: addr,
			QualDesc: uint(bytes[4]),
			Time: getCP24Time2a(bytes[5:], u.TimeZone),
		})
	},

	info.M_ME_ND_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.21.
		d.normal(u.ID, info.Normal(binary.LittleEndian.Uint16(bytes)), MeasureAttrs{
			Addr: addr,
		})
	},

	info.M_SP_TB_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.22.
		p := uint(bytes[0])
		d.single(u.ID, info.SinglePoint(p&1), MeasureAttrs{
			Addr: addr,
			QualDesc: p & 0x70,
			Time: getCP56Time2a(bytes[1:], u.TimeZone),
		})
	},

	info.M_DP_TB_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.23.
		p := uint(bytes[0])
		d.double(u.ID, info.DoublePoint(p&3), MeasureAttrs{
			Addr: addr,
			QualDesc: p & 0x70,
			Time: getCP56Time2a(bytes[1:], u.TimeZone),
		})
	},

	info.M_ST_TB_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.24.
		d.step(u.ID, info.StepPos(bytes[0]), MeasureAttrs{
			Addr: addr,
			QualDesc: uint(bytes[1]),
			Time: getCP56Time2a(bytes[2:], u.TimeZone),
		})
	},

	info.M_BO_TB_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.25.
		d.bits(u.ID, binary.BigEndian.Uint32(bytes), MeasureAttrs{
			Addr: addr,
			QualDesc: uint(bytes[4]),
			Time: getCP56Time2a(bytes[5:], u.TimeZone),
		})
	},

	info.M_ME_TD_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.26.

		d.normal(u.ID, info.Normal(binary.LittleEndian.Uint16(bytes)), MeasureAttrs{
			Addr: addr,
			QualDesc: uint(bytes[2]),
			Time: getCP56Time2a(bytes[3:], u.TimeZone),
		})
	},

	info.M_ME_TE_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.27.
		d.scaled(u.ID, int16(binary.LittleEndian.Uint16(bytes)), MeasureAttrs{
			Addr: addr,
			QualDesc: uint(bytes[2]),
			Time: getCP56Time2a(bytes[3:], u.TimeZone),
		})
	},

	info.M_ME_TF_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.28.
		d.float(u.ID, math.Float32frombits(binary.LittleEndian.Uint32(bytes)), MeasureAttrs{
			Addr: addr,
			QualDesc: uint(bytes[4]),
			Time: getCP56Time2a(bytes[5:], u.TimeZone),
		})
	},
}
