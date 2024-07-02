package part5

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/pascaldekloe/part5/info"
)

// Monitor listens for measured value updates.
type Monitor[Addr info.Addr] struct {
	SinglePoint         func(info.ID, Addr, info.SinglePointQual)
	SinglePointWithTime func(info.ID, Addr, info.SinglePointQual, CP24Time2a)
	SinglePointWithDate func(info.ID, Addr, info.SinglePointQual, CP56Time2a)

	DoublePoint         func(info.ID, Addr, info.DoublePointQual)
	DoublePointWithTime func(info.ID, Addr, info.DoublePointQual, CP24Time2a)
	DoublePointWithDate func(info.ID, Addr, info.DoublePointQual, CP56Time2a)

	StepPos         func(info.ID, Addr, info.StepPosQual)
	StepPosWithTime func(info.ID, Addr, info.StepPosQual, CP24Time2a)
	StepPosWithDate func(info.ID, Addr, info.StepPosQual, CP56Time2a)

	Bits         func(info.ID, Addr, info.BitsQual)
	BitsWithTime func(info.ID, Addr, info.BitsQual, CP24Time2a)
	BitsWithDate func(info.ID, Addr, info.BitsQual, CP56Time2a)

	NormUnqual   func(info.ID, Addr, info.Norm)
	Norm         func(info.ID, Addr, info.NormQual)
	NormWithTime func(info.ID, Addr, info.NormQual, CP24Time2a)
	NormWithDate func(info.ID, Addr, info.NormQual, CP56Time2a)

	Scaled         func(info.ID, Addr, int16, info.Qual)
	ScaledWithTime func(info.ID, Addr, int16, info.Qual, CP24Time2a)
	ScaledWithDate func(info.ID, Addr, int16, info.Qual, CP56Time2a)

	Float         func(info.ID, Addr, float32, info.Qual)
	FloatWithTime func(info.ID, Addr, float32, info.Qual, CP24Time2a)
	FloatWithDate func(info.ID, Addr, float32, info.Qual, CP56Time2a)
}

// NextAddr gets the followup in an info.Sequence.
func nextAddr[T info.Addr](addr T) T {
	addr.SetN(addr.N() + 1)
	return addr
}

var errNotMonitor = errors.New("part5: ASDU type is monitor information")

var errInfoSize = errors.New("part5: ASDU payload size does not match the count of variable structure qualifier")

// OnASDU applies the packet to respective listener.
// Payload is not retained.
func (m Monitor[Addr]) OnASDU(header info.ID, payload []byte) error {
	if header.Type >= 45 {
		return errNotMonitor
	}

	// NOTE: Go can't get the array length from a generic as a constant yet.
	var addr Addr

	switch header.Type {
	case info.M_SP_NA_1: // single-point
		if header.Struct&info.Sequence != 0 {
			if len(payload) != len(addr)+header.Struct.Count() {
				return errInfoSize
			}
			addr := Addr(payload[:len(addr)])
			for _, b := range payload[len(addr):] {
				m.DoublePoint(header, addr, info.DoublePointQual(b))
				addr = nextAddr(addr)
			}
		} else {
			if len(payload) != header.Struct.Count()*(len(addr)+1) {
				return errInfoSize
			}
			for len(payload) >= len(addr)+1 {
				m.SinglePoint(header, Addr(payload[:len(addr)]),
					info.SinglePointQual(payload[len(addr)]),
				)
				payload = payload[len(addr)+1:]
			}
		}

	case info.M_SP_TA_1: // single-point with 3 octet time-tag
		if header.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_SP_TA_1 not allowed")
		}
		if len(payload) != header.Struct.Count()*(len(addr)+4) {
			return errInfoSize
		}
		for len(payload) >= len(addr)+4 {
			m.SinglePointWithTime(header, Addr(payload[:len(addr)]),
				info.SinglePointQual(payload[len(addr)]),
				CP24Time2a(payload[len(addr)+1:len(addr)+4]),
			)
			payload = payload[len(addr)+4:]
		}

	case info.M_SP_TB_1: // single-point with 7 octet time-tag
		if header.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_SP_TB_1 not allowed")
		}
		if len(payload) != header.Struct.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for len(payload) >= (len(addr) + 8) {
			m.SinglePointWithDate(header, Addr(payload[:len(addr)]),
				info.SinglePointQual(payload[len(addr)]),
				CP56Time2a(payload[len(addr)+1:len(addr)+8]),
			)
			payload = payload[len(addr)+8:]
		}

	case info.M_DP_NA_1: // double-point
		if header.Struct&info.Sequence != 0 {
			if len(payload) != len(addr)+header.Struct.Count() {
				return errInfoSize
			}
			addr := Addr(payload[:len(addr)])
			for _, b := range payload[len(addr):] {
				m.DoublePoint(header, addr, info.DoublePointQual(b))
				addr = nextAddr(addr)
			}
		} else {
			if len(payload) != header.Struct.Count()*(len(addr)+1) {
				return errInfoSize
			}
			for len(payload) >= len(addr)+1 {
				m.DoublePoint(header, Addr(payload[:len(addr)]),
					info.DoublePointQual(payload[len(addr)]),
				)
				payload = payload[len(addr)+1:]
			}
		}

	case info.M_DP_TA_1: // double-point with 3 octet time-tag
		if header.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_DP_TA_1 not allowed")
		}
		if len(payload) != header.Struct.Count()*(len(addr)+4) {
			return errInfoSize
		}
		for len(payload) >= len(addr)+4 {
			m.DoublePointWithTime(header, Addr(payload[:len(addr)]),
				info.DoublePointQual(payload[len(addr)]),
				CP24Time2a(payload[len(addr)+1:len(addr)+4]),
			)
			payload = payload[len(addr)+4:]
		}

	case info.M_DP_TB_1: // double-point with 7 octet time-tag
		if header.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_DP_TB_1 not allowed")
		}
		if len(payload) != header.Struct.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for len(payload) >= len(addr)+8 {
			m.DoublePointWithDate(header, Addr(payload[:len(addr)]),
				info.DoublePointQual(payload[len(addr)]),
				CP56Time2a(payload[len(addr)+1:len(addr)+8]),
			)
			payload = payload[len(addr)+8:]
		}

	case info.M_ST_NA_1: // step position
		if header.Struct&info.Sequence != 0 {
			if len(payload) != len(addr)+header.Struct.Count()*2 {
				return errInfoSize
			}
			addr := Addr(payload)
			for i := len(addr); i+2 <= len(payload); i += 2 {
				m.StepPos(header, addr, info.StepPosQual(payload[i:i+2]))
				addr = nextAddr(addr)
			}
		} else {
			if len(payload) != header.Struct.Count()*(len(addr)+2) {
				return errInfoSize
			}
			for len(payload) >= len(addr)+2 {
				m.StepPos(header, Addr(payload[:len(addr)]),
					info.StepPosQual(payload[len(addr):len(addr)+2]),
				)
				payload = payload[len(addr)+2:]
			}
		}

	case info.M_ST_TA_1: // step position with 3 octet time-tag
		if header.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ST_TA_1 not allowed")
		}
		if len(payload) != header.Struct.Count()*(len(addr)+5) {
			return errInfoSize
		}
		for len(payload) >= len(addr)+5 {
			m.StepPosWithTime(header, Addr(payload[:len(addr)]),
				info.StepPosQual(payload[len(addr):len(addr)+2]),
				CP24Time2a(payload[len(addr)+2:len(addr)+5]),
			)
			payload = payload[len(addr)+5:]
		}

	case info.M_ST_TB_1: // step position with 7 octet time-tag
		if header.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ST_TB_1 not allowed")
		}
		if len(payload) != header.Struct.Count()*(len(addr)+9) {
			return errInfoSize
		}
		for len(payload) >= len(addr)+9 {
			m.StepPosWithDate(header, Addr(payload[:len(addr)]),
				info.StepPosQual(payload[len(addr):len(addr)+2]),
				CP56Time2a(payload[len(addr)+2:len(addr)+9]),
			)
			payload = payload[len(addr)+9:]
		}

	case info.M_BO_NA_1: // bitstring
		if header.Struct&info.Sequence != 0 {
			if len(payload) != len(addr)+header.Struct.Count()*5 {
				return errInfoSize
			}
			addr := Addr(payload[:len(addr)])
			for i := len(addr); i+5 <= len(payload); i += 5 {
				m.Bits(header, addr, info.BitsQual(payload[i:i+5]))
				addr = nextAddr(addr)
			}
		} else {
			if len(payload) != header.Struct.Count()*(len(addr)+5) {
				return errInfoSize
			}
			for len(payload) >= len(addr)+5 {
				m.Bits(header, Addr(payload),
					info.BitsQual(payload[len(addr):len(addr)+5]),
				)
				payload = payload[len(addr)+5:]
			}
		}

	case info.M_BO_TA_1: // bit string with 3 octet time-tag
		if header.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_BO_TA_1 not allowed")
		}
		if len(payload) != header.Struct.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for len(payload) >= len(addr)+8 {
			m.BitsWithTime(header, Addr(payload[:len(addr)]),
				info.BitsQual(payload[len(addr):len(addr)+5]),
				CP24Time2a(payload[len(addr)+5:len(addr)+8]),
			)
			payload = payload[len(addr)+8:]
		}

	case info.M_BO_TB_1: // bit string with 7 octet time-tag
		if header.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_BO_TB_1 not allowed")
		}
		if len(payload) != header.Struct.Count()*(len(addr)+12) {
			return errInfoSize
		}
		for len(payload) >= len(addr)+12 {
			m.BitsWithDate(header, Addr(payload[:len(addr)]),
				info.BitsQual(payload[len(addr):len(addr)+5]),
				CP56Time2a(payload[len(addr)+5:len(addr)+12]),
			)
			payload = payload[len(addr)+12:]
		}

	case info.M_ME_ND_1: // normalized value without quality descriptor
		if header.Struct&info.Sequence != 0 {
			if len(payload) != len(addr)+header.Struct.Count()*2 {
				return errInfoSize
			}
			addr := Addr(payload[:len(addr)])
			for i := len(addr); i+1 < len(payload); i += 2 {
				m.NormUnqual(header, addr, info.Norm(payload[i:i+2]))
				addr = nextAddr(addr)
			}
		} else {
			if len(payload) != header.Struct.Count()*(len(addr)+2) {
				return errInfoSize
			}
			for len(payload) >= len(addr)+2 {
				m.NormUnqual(header, Addr(payload[:len(addr)]),
					info.Norm(payload[len(addr):len(addr)+2]),
				)
				payload = payload[len(addr)+2:]
			}
		}

	case info.M_ME_NA_1: // normalized value
		if header.Struct&info.Sequence != 0 {
			if len(payload) != len(addr)+header.Struct.Count()*3 {
				return errInfoSize
			}
			addr := Addr(payload)
			for i := len(addr); i+3 <= len(payload); i += 3 {
				m.Norm(header, addr, info.NormQual(payload[i:i+3]))
				addr = nextAddr(addr)
			}
		} else {
			if len(payload) != header.Struct.Count()*(len(addr)+3) {
				return errInfoSize
			}
			for len(payload) >= len(addr)+3 {
				m.Norm(header, Addr(payload[:len(addr)]),
					info.NormQual(payload[len(addr):len(addr)+3]),
				)
				payload = payload[len(addr)+3:]
			}
		}

	case info.M_ME_TA_1: // normalized value with 3 octet time-tag
		if header.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ME_TA_1 not allowed")
		}
		if len(payload) != header.Struct.Count()*(len(addr)+6) {
			return errInfoSize
		}
		for len(payload) >= len(addr)+6 {
			m.NormWithTime(header, Addr(payload[:len(addr)]),
				info.NormQual(payload[len(addr):len(addr)+3]),
				CP24Time2a(payload[len(addr)+3:len(addr)+6]),
			)
			payload = payload[len(addr)+6:]
		}

	case info.M_ME_TD_1: // normalized value with 7 octet time-tag
		if header.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ME_TD_1 not allowed")
		}
		if len(payload) != header.Struct.Count()*(len(addr)+10) {
			return errInfoSize
		}
		for len(payload) >= len(addr)+10 {
			m.NormWithDate(header, Addr(payload[:len(addr)]),
				info.NormQual(payload[len(addr):len(addr)+3]),
				CP56Time2a(payload[len(addr)+3:len(addr)+10]),
			)
			payload = payload[len(addr)+10:]
		}

	case info.M_ME_NB_1: // scaled value
		if header.Struct&info.Sequence != 0 {
			if len(payload) != len(addr)+header.Struct.Count()*3 {
				return errInfoSize
			}
			addr := Addr(payload)
			for i := len(addr); i+3 <= len(payload); i += 3 {
				m.Scaled(header, addr,
					int16(binary.LittleEndian.Uint16(payload[i:i+2])),
					info.Qual(payload[i+2]),
				)
				addr = nextAddr(addr)
			}
		} else {
			if len(payload) != header.Struct.Count()*(len(addr)+3) {
				return errInfoSize
			}
			for len(payload) >= len(addr)+3 {
				m.Scaled(header, Addr(payload[:len(addr)]),
					int16(binary.LittleEndian.Uint16(payload[len(addr):len(addr)+2])),
					info.Qual(payload[len(addr)+2]),
				)
				payload = payload[len(addr)+3:]
			}
		}

	case info.M_ME_TB_1: // scaled value with 3 octet time-tag
		if header.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ME_TB_1 not allowed")
		}
		if len(payload) != header.Struct.Count()*(len(addr)+6) {
			return errInfoSize
		}
		for len(payload) >= len(addr)+6 {
			m.ScaledWithTime(header, Addr(payload[:len(addr)]),
				int16(binary.LittleEndian.Uint16(payload[len(addr):len(addr)+2])),
				info.Qual(payload[len(addr)+2]),
				CP24Time2a(payload[len(addr)+3:len(addr)+6]),
			)
			payload = payload[len(addr)+6:]
		}

	case info.M_ME_TE_1: // scaled value with 7 octet time-tag
		if header.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ME_TE_1 not allowed")
		}
		if len(payload) != header.Struct.Count()*(len(addr)+10) {
			return errInfoSize
		}
		for len(payload) >= len(addr)+10 {
			m.ScaledWithDate(header, Addr(payload[:len(addr)]),
				int16(binary.LittleEndian.Uint16(payload[len(addr):len(addr)+2])),
				info.Qual(payload[len(addr)+2]),
				CP56Time2a(payload[len(addr)+3:len(addr)+10]),
			)
			payload = payload[len(addr)+10:]
		}

	case info.M_ME_NC_1: // floating-point
		if header.Struct&info.Sequence != 0 {
			if len(payload) != len(addr)+header.Struct.Count()*5 {
				return errInfoSize
			}
			addr := Addr(payload)
			for i := len(addr); i+5 <= len(payload); i += 5 {
				m.Float(header, addr,
					math.Float32frombits(binary.LittleEndian.Uint32(payload[i:i+4])),
					info.Qual(payload[i+4]),
				)
				addr = nextAddr(addr)
			}
		} else {
			if len(payload) != header.Struct.Count()*(len(addr)+5) {
				return errInfoSize
			}
			for len(payload) >= len(addr)+5 {
				m.Float(header, Addr(payload[:len(addr)]),
					math.Float32frombits(
						binary.LittleEndian.Uint32(
							payload[len(addr):len(addr)+4],
						),
					),
					info.Qual(payload[len(addr)+4]),
				)
				payload = payload[len(addr)+5:]
			}
		}

	case info.M_ME_TC_1: // floating point with 3 octet time-tag
		if header.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ME_TC_1 not allowed")
		}
		if len(payload) != header.Struct.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for len(payload) >= len(addr)+8 {
			m.FloatWithTime(header, Addr(payload[:len(addr)]),
				math.Float32frombits(
					binary.LittleEndian.Uint32(
						payload[len(addr):len(addr)+4],
					),
				),
				info.Qual(payload[len(addr)+4]),
				CP24Time2a(payload[len(addr)+5:len(addr)+8]),
			)
			payload = payload[len(addr)+8:]
		}

	case info.M_ME_TF_1: // floating point with 7 octet time-tag
		if header.Struct&info.Sequence != 0 {
			return errors.New("part5: ASDU address sequence with M_ME_TF_1 not allowed")
		}
		if len(payload) != header.Struct.Count()*(len(addr)+12) {
			return errInfoSize
		}
		for len(payload) >= len(addr)+12 {
			m.FloatWithDate(header, Addr(payload[:len(addr)]),
				math.Float32frombits(
					binary.LittleEndian.Uint32(
						payload[len(addr):len(addr)+4],
					),
				),
				info.Qual(payload[len(addr)+4]),
				CP56Time2a(payload[len(addr)+5:len(addr)+12]),
			)
			payload = payload[len(addr)+12:]
		}

	}
	return nil
}
