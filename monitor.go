package part5

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/pascaldekloe/part5/info"
)

// Monitor listens for measured value updates.
type Monitor[COT info.Cause, Common info.ComAddr, Object info.Addr] struct {
	SinglePt         func(info.DataUnit[COT, Common, Object], Object, info.SinglePtQual)
	SinglePtWithTime func(info.DataUnit[COT, Common, Object], Object, info.SinglePtQual, CP24Time2a)
	SinglePtWithDate func(info.DataUnit[COT, Common, Object], Object, info.SinglePtQual, CP56Time2a)

	DoublePt         func(info.DataUnit[COT, Common, Object], Object, info.DoublePtQual)
	DoublePtWithTime func(info.DataUnit[COT, Common, Object], Object, info.DoublePtQual, CP24Time2a)
	DoublePtWithDate func(info.DataUnit[COT, Common, Object], Object, info.DoublePtQual, CP56Time2a)

	Step         func(info.DataUnit[COT, Common, Object], Object, info.StepQual)
	StepWithTime func(info.DataUnit[COT, Common, Object], Object, info.StepQual, CP24Time2a)
	StepWithDate func(info.DataUnit[COT, Common, Object], Object, info.StepQual, CP56Time2a)

	Bits         func(info.DataUnit[COT, Common, Object], Object, info.BitsQual)
	BitsWithTime func(info.DataUnit[COT, Common, Object], Object, info.BitsQual, CP24Time2a)
	BitsWithDate func(info.DataUnit[COT, Common, Object], Object, info.BitsQual, CP56Time2a)

	NormUnqual   func(info.DataUnit[COT, Common, Object], Object, info.Norm)
	Norm         func(info.DataUnit[COT, Common, Object], Object, info.NormQual)
	NormWithTime func(info.DataUnit[COT, Common, Object], Object, info.NormQual, CP24Time2a)
	NormWithDate func(info.DataUnit[COT, Common, Object], Object, info.NormQual, CP56Time2a)

	Scaled         func(info.DataUnit[COT, Common, Object], Object, int16, info.Qual)
	ScaledWithTime func(info.DataUnit[COT, Common, Object], Object, int16, info.Qual, CP24Time2a)
	ScaledWithDate func(info.DataUnit[COT, Common, Object], Object, int16, info.Qual, CP56Time2a)

	Float         func(info.DataUnit[COT, Common, Object], Object, float32, info.Qual)
	FloatWithTime func(info.DataUnit[COT, Common, Object], Object, float32, info.Qual, CP24Time2a)
	FloatWithDate func(info.DataUnit[COT, Common, Object], Object, float32, info.Qual, CP56Time2a)
}

// NextAddr gets the followup in an info.Sequence.
func nextAddr[T info.Addr](addr T) T {
	// overflow ignored (as unspecified behaviour)
	sum, _ := info.NewAddrN[T](addr.N() + 1)
	return sum
}

var errNotMonitor = errors.New("part5: ASDU type is monitor information")

var errInfoSize = errors.New("part5: ASDU payload size does not match the count of variable structure qualifier")

// OnDataUnit applies the packet to respective listener. Data is not retained.
func (m Monitor[COT, Common, Object]) OnDataUnit(u info.DataUnit[COT, Common, Object]) error {
	if u.Type >= 45 {
		return errNotMonitor
	}

	// NOTE: Go can't get the array length from a generic as a constant yet.
	var addr Object

	switch u.Type {
	case info.M_SP_NA_1: // single-point
		if u.Struct&info.Sequence != 0 {
			if len(u.Info) != len(addr)+u.Struct.Count() {
				return errInfoSize
			}
			addr := Object(u.Info[:len(addr)])
			for _, b := range u.Info[len(addr):] {
				m.DoublePt(u, addr, info.DoublePtQual(b))
				addr = nextAddr(addr)
			}
		} else {
			if len(u.Info) != u.Struct.Count()*(len(addr)+1) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+1 {
				m.SinglePt(u, Object(u.Info[:len(addr)]),
					info.SinglePtQual(u.Info[len(addr)]),
				)
				u.Info = u.Info[len(addr)+1:]
			}
		}

	case info.M_SP_TA_1: // single-point with 3 octet time-tag
		if u.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_SP_TA_1 not allowed")
		}
		if len(u.Info) != u.Struct.Count()*(len(addr)+4) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+4 {
			m.SinglePtWithTime(u, Object(u.Info[:len(addr)]),
				info.SinglePtQual(u.Info[len(addr)]),
				CP24Time2a(u.Info[len(addr)+1:len(addr)+4]),
			)
			u.Info = u.Info[len(addr)+4:]
		}

	case info.M_SP_TB_1: // single-point with 7 octet time-tag
		if u.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_SP_TB_1 not allowed")
		}
		if len(u.Info) != u.Struct.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for len(u.Info) >= (len(addr) + 8) {
			m.SinglePtWithDate(u, Object(u.Info[:len(addr)]),
				info.SinglePtQual(u.Info[len(addr)]),
				CP56Time2a(u.Info[len(addr)+1:len(addr)+8]),
			)
			u.Info = u.Info[len(addr)+8:]
		}

	case info.M_DP_NA_1: // double-point
		if u.Struct&info.Sequence != 0 {
			if len(u.Info) != len(addr)+u.Struct.Count() {
				return errInfoSize
			}
			addr := Object(u.Info[:len(addr)])
			for _, b := range u.Info[len(addr):] {
				m.DoublePt(u, addr, info.DoublePtQual(b))
				addr = nextAddr(addr)
			}
		} else {
			if len(u.Info) != u.Struct.Count()*(len(addr)+1) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+1 {
				m.DoublePt(u, Object(u.Info[:len(addr)]),
					info.DoublePtQual(u.Info[len(addr)]),
				)
				u.Info = u.Info[len(addr)+1:]
			}
		}

	case info.M_DP_TA_1: // double-point with 3 octet time-tag
		if u.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_DP_TA_1 not allowed")
		}
		if len(u.Info) != u.Struct.Count()*(len(addr)+4) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+4 {
			m.DoublePtWithTime(u, Object(u.Info[:len(addr)]),
				info.DoublePtQual(u.Info[len(addr)]),
				CP24Time2a(u.Info[len(addr)+1:len(addr)+4]),
			)
			u.Info = u.Info[len(addr)+4:]
		}

	case info.M_DP_TB_1: // double-point with 7 octet time-tag
		if u.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_DP_TB_1 not allowed")
		}
		if len(u.Info) != u.Struct.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+8 {
			m.DoublePtWithDate(u, Object(u.Info[:len(addr)]),
				info.DoublePtQual(u.Info[len(addr)]),
				CP56Time2a(u.Info[len(addr)+1:len(addr)+8]),
			)
			u.Info = u.Info[len(addr)+8:]
		}

	case info.M_ST_NA_1: // step position
		if u.Struct&info.Sequence != 0 {
			if len(u.Info) != len(addr)+u.Struct.Count()*2 {
				return errInfoSize
			}
			addr := Object(u.Info)
			for i := len(addr); i+2 <= len(u.Info); i += 2 {
				m.Step(u, addr, info.StepQual(u.Info[i:i+2]))
				addr = nextAddr(addr)
			}
		} else {
			if len(u.Info) != u.Struct.Count()*(len(addr)+2) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+2 {
				m.Step(u, Object(u.Info[:len(addr)]),
					info.StepQual(u.Info[len(addr):len(addr)+2]),
				)
				u.Info = u.Info[len(addr)+2:]
			}
		}

	case info.M_ST_TA_1: // step position with 3 octet time-tag
		if u.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ST_TA_1 not allowed")
		}
		if len(u.Info) != u.Struct.Count()*(len(addr)+5) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+5 {
			m.StepWithTime(u, Object(u.Info[:len(addr)]),
				info.StepQual(u.Info[len(addr):len(addr)+2]),
				CP24Time2a(u.Info[len(addr)+2:len(addr)+5]),
			)
			u.Info = u.Info[len(addr)+5:]
		}

	case info.M_ST_TB_1: // step position with 7 octet time-tag
		if u.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ST_TB_1 not allowed")
		}
		if len(u.Info) != u.Struct.Count()*(len(addr)+9) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+9 {
			m.StepWithDate(u, Object(u.Info[:len(addr)]),
				info.StepQual(u.Info[len(addr):len(addr)+2]),
				CP56Time2a(u.Info[len(addr)+2:len(addr)+9]),
			)
			u.Info = u.Info[len(addr)+9:]
		}

	case info.M_BO_NA_1: // bitstring
		if u.Struct&info.Sequence != 0 {
			if len(u.Info) != len(addr)+u.Struct.Count()*5 {
				return errInfoSize
			}
			addr := Object(u.Info[:len(addr)])
			for i := len(addr); i+5 <= len(u.Info); i += 5 {
				m.Bits(u, addr, info.BitsQual(u.Info[i:i+5]))
				addr = nextAddr(addr)
			}
		} else {
			if len(u.Info) != u.Struct.Count()*(len(addr)+5) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+5 {
				m.Bits(u, Object(u.Info),
					info.BitsQual(u.Info[len(addr):len(addr)+5]),
				)
				u.Info = u.Info[len(addr)+5:]
			}
		}

	case info.M_BO_TA_1: // bit string with 3 octet time-tag
		if u.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_BO_TA_1 not allowed")
		}
		if len(u.Info) != u.Struct.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+8 {
			m.BitsWithTime(u, Object(u.Info[:len(addr)]),
				info.BitsQual(u.Info[len(addr):len(addr)+5]),
				CP24Time2a(u.Info[len(addr)+5:len(addr)+8]),
			)
			u.Info = u.Info[len(addr)+8:]
		}

	case info.M_BO_TB_1: // bit string with 7 octet time-tag
		if u.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_BO_TB_1 not allowed")
		}
		if len(u.Info) != u.Struct.Count()*(len(addr)+12) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+12 {
			m.BitsWithDate(u, Object(u.Info[:len(addr)]),
				info.BitsQual(u.Info[len(addr):len(addr)+5]),
				CP56Time2a(u.Info[len(addr)+5:len(addr)+12]),
			)
			u.Info = u.Info[len(addr)+12:]
		}

	case info.M_ME_ND_1: // normalized value without quality descriptor
		if u.Struct&info.Sequence != 0 {
			if len(u.Info) != len(addr)+u.Struct.Count()*2 {
				return errInfoSize
			}
			addr := Object(u.Info[:len(addr)])
			for i := len(addr); i+1 < len(u.Info); i += 2 {
				m.NormUnqual(u, addr, info.Norm(u.Info[i:i+2]))
				addr = nextAddr(addr)
			}
		} else {
			if len(u.Info) != u.Struct.Count()*(len(addr)+2) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+2 {
				m.NormUnqual(u, Object(u.Info[:len(addr)]),
					info.Norm(u.Info[len(addr):len(addr)+2]),
				)
				u.Info = u.Info[len(addr)+2:]
			}
		}

	case info.M_ME_NA_1: // normalized value
		if u.Struct&info.Sequence != 0 {
			if len(u.Info) != len(addr)+u.Struct.Count()*3 {
				return errInfoSize
			}
			addr := Object(u.Info)
			for i := len(addr); i+3 <= len(u.Info); i += 3 {
				m.Norm(u, addr, info.NormQual(u.Info[i:i+3]))
				addr = nextAddr(addr)
			}
		} else {
			if len(u.Info) != u.Struct.Count()*(len(addr)+3) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+3 {
				m.Norm(u, Object(u.Info[:len(addr)]),
					info.NormQual(u.Info[len(addr):len(addr)+3]),
				)
				u.Info = u.Info[len(addr)+3:]
			}
		}

	case info.M_ME_TA_1: // normalized value with 3 octet time-tag
		if u.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ME_TA_1 not allowed")
		}
		if len(u.Info) != u.Struct.Count()*(len(addr)+6) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+6 {
			m.NormWithTime(u, Object(u.Info[:len(addr)]),
				info.NormQual(u.Info[len(addr):len(addr)+3]),
				CP24Time2a(u.Info[len(addr)+3:len(addr)+6]),
			)
			u.Info = u.Info[len(addr)+6:]
		}

	case info.M_ME_TD_1: // normalized value with 7 octet time-tag
		if u.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ME_TD_1 not allowed")
		}
		if len(u.Info) != u.Struct.Count()*(len(addr)+10) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+10 {
			m.NormWithDate(u, Object(u.Info[:len(addr)]),
				info.NormQual(u.Info[len(addr):len(addr)+3]),
				CP56Time2a(u.Info[len(addr)+3:len(addr)+10]),
			)
			u.Info = u.Info[len(addr)+10:]
		}

	case info.M_ME_NB_1: // scaled value
		if u.Struct&info.Sequence != 0 {
			if len(u.Info) != len(addr)+u.Struct.Count()*3 {
				return errInfoSize
			}
			addr := Object(u.Info)
			for i := len(addr); i+3 <= len(u.Info); i += 3 {
				m.Scaled(u, addr,
					int16(binary.LittleEndian.Uint16(u.Info[i:i+2])),
					info.Qual(u.Info[i+2]),
				)
				addr = nextAddr(addr)
			}
		} else {
			if len(u.Info) != u.Struct.Count()*(len(addr)+3) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+3 {
				m.Scaled(u, Object(u.Info[:len(addr)]),
					int16(binary.LittleEndian.Uint16(u.Info[len(addr):len(addr)+2])),
					info.Qual(u.Info[len(addr)+2]),
				)
				u.Info = u.Info[len(addr)+3:]
			}
		}

	case info.M_ME_TB_1: // scaled value with 3 octet time-tag
		if u.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ME_TB_1 not allowed")
		}
		if len(u.Info) != u.Struct.Count()*(len(addr)+6) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+6 {
			m.ScaledWithTime(u, Object(u.Info[:len(addr)]),
				int16(binary.LittleEndian.Uint16(u.Info[len(addr):len(addr)+2])),
				info.Qual(u.Info[len(addr)+2]),
				CP24Time2a(u.Info[len(addr)+3:len(addr)+6]),
			)
			u.Info = u.Info[len(addr)+6:]
		}

	case info.M_ME_TE_1: // scaled value with 7 octet time-tag
		if u.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ME_TE_1 not allowed")
		}
		if len(u.Info) != u.Struct.Count()*(len(addr)+10) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+10 {
			m.ScaledWithDate(u, Object(u.Info[:len(addr)]),
				int16(binary.LittleEndian.Uint16(u.Info[len(addr):len(addr)+2])),
				info.Qual(u.Info[len(addr)+2]),
				CP56Time2a(u.Info[len(addr)+3:len(addr)+10]),
			)
			u.Info = u.Info[len(addr)+10:]
		}

	case info.M_ME_NC_1: // floating-point
		if u.Struct&info.Sequence != 0 {
			if len(u.Info) != len(addr)+u.Struct.Count()*5 {
				return errInfoSize
			}
			addr := Object(u.Info)
			for i := len(addr); i+5 <= len(u.Info); i += 5 {
				m.Float(u, addr,
					math.Float32frombits(binary.LittleEndian.Uint32(u.Info[i:i+4])),
					info.Qual(u.Info[i+4]),
				)
				addr = nextAddr(addr)
			}
		} else {
			if len(u.Info) != u.Struct.Count()*(len(addr)+5) {
				return errInfoSize
			}
			for len(u.Info) >= len(addr)+5 {
				m.Float(u, Object(u.Info[:len(addr)]),
					math.Float32frombits(
						binary.LittleEndian.Uint32(
							u.Info[len(addr):len(addr)+4],
						),
					),
					info.Qual(u.Info[len(addr)+4]),
				)
				u.Info = u.Info[len(addr)+5:]
			}
		}

	case info.M_ME_TC_1: // floating point with 3 octet time-tag
		if u.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ME_TC_1 not allowed")
		}
		if len(u.Info) != u.Struct.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+8 {
			m.FloatWithTime(u, Object(u.Info[:len(addr)]),
				math.Float32frombits(
					binary.LittleEndian.Uint32(
						u.Info[len(addr):len(addr)+4],
					),
				),
				info.Qual(u.Info[len(addr)+4]),
				CP24Time2a(u.Info[len(addr)+5:len(addr)+8]),
			)
			u.Info = u.Info[len(addr)+8:]
		}

	case info.M_ME_TF_1: // floating point with 7 octet time-tag
		if u.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ME_TF_1 not allowed")
		}
		if len(u.Info) != u.Struct.Count()*(len(addr)+12) {
			return errInfoSize
		}
		for len(u.Info) >= len(addr)+12 {
			m.FloatWithDate(u, Object(u.Info[:len(addr)]),
				math.Float32frombits(
					binary.LittleEndian.Uint32(
						u.Info[len(addr):len(addr)+4],
					),
				),
				info.Qual(u.Info[len(addr)+4]),
				CP56Time2a(u.Info[len(addr)+5:len(addr)+12]),
			)
			u.Info = u.Info[len(addr)+12:]
		}

	}
	return nil
}
