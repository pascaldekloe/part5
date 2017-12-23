package part5

import (
	"encoding/binary"
	"math"

	"github.com/pascaldekloe/part5/info"
	"github.com/pascaldekloe/part5/session"
)

// MustLaunch starts a setup which runs until session.Exit.
// If the configuration is invalid, MustLaunch will panic.
func MustLaunch(t *session.Transport, p *info.Params, delegates map[info.CommonAddr]*Delegate) (*Caller, <-chan error) {
	// validate
	if err := p.Valid(); err != nil {
		panic(err)
	}
	for addr, d := range delegates {
		if err := p.ValidAddr(addr); err != nil {
			panic(err)
		}
		if err := d.Valid(); err != nil {
			panic(err)
		}
	}

	errCh := make(chan error)

	c := &Caller{
		params:    p,
		transport: t,
	}

	go func() {
		defer close(errCh)

	ReadLoop:
		for bytes := range t.In {
			u := info.NewASDU(p, info.ID{})
			if err := u.UnmarshalBinary(bytes); err != nil {
				errCh <- err
				continue
			}

			if callback, ok := c.confirm.Load(cmdKey(u)); ok {
				callback.(chan *info.ASDU) <- u
				continue
			}

			d := delegates[u.Addr]
			if d == nil {
				if u.Cause&^info.TestFlag == info.Act {
					sendNB(u.Reply(info.UnkAddr), c, errCh)
				}
				continue
			}

			for _, f := range d.All {
				if !f(u) {
					continue ReadLoop
				}
			}

			d.notify(u, c, errCh)
		}
	}()

	return c, errCh
}

// Notify informs Delegate about the ASDU.
func (d *Delegate) notify(u *info.ASDU, c *Caller, errCh chan<- error) {
	if u.Type >= info.TypeID(len(notifyMeasureFunc)) {
		switch u.Type {
		case info.C_SC_NA_1, info.C_SC_TA_1:
			d.singleCmd(u, c, errCh)
		case info.C_DC_NA_1, info.C_DC_TA_1:
			d.doubleCmd(u, c, errCh)
		// RegulCmd gets called for type C_RC_NA_1 and C_RC_TA_1.
		// NormalSetpoint gets called for type C_SE_NA_1 and C_SE_TA_1.
		// ScaledSetpoint gets called for type C_SE_NB_1 and C_SE_TB_1.
		// FloatSetpoint gets called for type
		case info.C_SE_NC_1, info.C_SE_TC_1:
			d.floatSetpoint(u, c, errCh)
		// BitsCmd gets called for type C_BO_NA_1 and C_BO_TA_1.
		default:
			sendNB(u.Reply(info.UnkType), c, errCh)
		}

		return
	}

	impl := notifyMeasureFunc[u.Type]
	if impl == nil {
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

		d.Add(1)
		go func(bytes []byte) {
			impl(bytes, addr, d, u)
			d.Done()
		}(u.Info[i:end])

		if end >= len(u.Info) {
			return
		}

		if u.InfoSeq {
			addr++
			i = end
		} else {
			addr = u.ObjAddr(u.Info[end:])
			i = end + u.ObjAddrSize
		}
	}
}

var notifyMeasureFunc = [45]func([]byte, info.ObjAddr, *Delegate, *info.ASDU){
	info.M_SP_NA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.1.
		p := uint(bytes[0])
		value := info.SinglePoint(p & 1)
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: p & 0x70,
		}
		d.single(u.ID, value, attrs)
	},
	info.M_SP_TA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.2.
		p := uint(bytes[0])
		value := info.SinglePoint(p & 1)
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: p & 0x70,
			Time:     getCP24Time2a(bytes[1:], u.TimeZone),
		}
		d.single(u.ID, value, attrs)
	},
	info.M_DP_NA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.3.
		p := uint(bytes[0])
		value := info.DoublePoint(p & 3)
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: p & 0x70,
		}
		d.double(u.ID, value, attrs)
	},
	info.M_DP_TA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.4.
		p := uint(bytes[0])
		value := info.DoublePoint(p & 3)
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: p & 0x70,
			Time:     getCP24Time2a(bytes[1:], u.TimeZone),
		}
		d.double(u.ID, value, attrs)
	},
	info.M_ST_NA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.5.
		value := info.StepPos(bytes[0])
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: uint(bytes[1]),
		}
		d.step(u.ID, value, attrs)
	},
	info.M_ST_TA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.6.
		value := info.StepPos(bytes[0])
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: uint(bytes[1]),
			Time:     getCP24Time2a(bytes[2:], u.TimeZone),
		}
		d.step(u.ID, value, attrs)
	},
	info.M_BO_NA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.7.
		value := binary.BigEndian.Uint32(bytes)
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: uint(bytes[4]),
		}
		d.bits(u.ID, value, attrs)
	},
	info.M_BO_TA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.8.
		value := binary.BigEndian.Uint32(bytes)
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: uint(bytes[4]),
			Time:     getCP24Time2a(bytes[5:], u.TimeZone),
		}
		d.bits(u.ID, value, attrs)
	},
	info.M_ME_NA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.9.
		value := info.Normal(binary.LittleEndian.Uint16(bytes))
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: uint(bytes[2]),
		}
		d.normal(u.ID, value, attrs)
	},
	info.M_ME_TA_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.10.
		value := info.Normal(binary.LittleEndian.Uint16(bytes))
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: uint(bytes[2]),
			Time:     getCP24Time2a(bytes[3:], u.TimeZone),
		}
		d.normal(u.ID, value, attrs)
	},
	info.M_ME_NB_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.11.
		value := int16(binary.LittleEndian.Uint16(bytes))
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: uint(bytes[2]),
		}
		d.scaled(u.ID, value, attrs)
	},
	info.M_ME_TB_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.12.
		value := int16(binary.LittleEndian.Uint16(bytes))
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: uint(bytes[2]),
			Time:     getCP24Time2a(bytes[3:], u.TimeZone),
		}
		d.scaled(u.ID, value, attrs)
	},
	info.M_ME_NC_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.13.
		value := math.Float32frombits(binary.LittleEndian.Uint32(bytes))
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: uint(bytes[4]),
		}
		d.float(u.ID, value, attrs)
	},
	info.M_ME_TC_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.14.
		value := math.Float32frombits(binary.LittleEndian.Uint32(bytes))
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: uint(bytes[4]),
			Time:     getCP24Time2a(bytes[5:], u.TimeZone),
		}
		d.float(u.ID, value, attrs)
	},
	info.M_ME_ND_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.21.
		value := info.Normal(binary.LittleEndian.Uint16(bytes))
		attrs := MeasureAttrs{
			Addr: addr,
		}
		d.normal(u.ID, value, attrs)
	},
	info.M_SP_TB_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.22.
		p := uint(bytes[0])
		value := info.SinglePoint(p & 1)
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: p & 0x70,
			Time:     getCP56Time2a(bytes[1:], u.TimeZone),
		}
		d.single(u.ID, value, attrs)
	},
	info.M_DP_TB_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.23.
		p := uint(bytes[0])
		value := info.DoublePoint(p & 3)
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: p & 0x70,
			Time:     getCP56Time2a(bytes[1:], u.TimeZone),
		}
		d.double(u.ID, value, attrs)
	},
	info.M_ST_TB_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.24.
		value := info.StepPos(bytes[0])
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: uint(bytes[1]),
			Time:     getCP56Time2a(bytes[2:], u.TimeZone),
		}
		d.step(u.ID, value, attrs)
	},
	info.M_BO_TB_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.25.
		value := binary.BigEndian.Uint32(bytes)
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: uint(bytes[4]),
			Time:     getCP56Time2a(bytes[5:], u.TimeZone),
		}
		d.bits(u.ID, value, attrs)
	},
	info.M_ME_TD_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.26.
		value := info.Normal(binary.LittleEndian.Uint16(bytes))
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: uint(bytes[2]),
			Time:     getCP56Time2a(bytes[3:], u.TimeZone),
		}
		d.normal(u.ID, value, attrs)
	},
	info.M_ME_TE_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.27.
		value := int16(binary.LittleEndian.Uint16(bytes))
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: uint(bytes[2]),
			Time:     getCP56Time2a(bytes[3:], u.TimeZone),
		}
		d.scaled(u.ID, value, attrs)
	},
	info.M_ME_TF_1: func(bytes []byte, addr info.ObjAddr, d *Delegate, u *info.ASDU) {
		// companion standard 101, subclause 7.3.1.28.
		value := math.Float32frombits(binary.LittleEndian.Uint32(bytes))
		attrs := MeasureAttrs{
			Addr:     addr,
			QualDesc: uint(bytes[4]),
			Time:     getCP56Time2a(bytes[5:], u.TimeZone),
		}
		d.float(u.ID, value, attrs)
	},
}
