package info

import (
	"encoding/binary"
	"errors"
	"math"
)

// AddSingle sets Type to M_SP_NA_1 and appends the measured value to Info.
func (u *ASDU) AddSingle(addr ObjAddr, v SinglePoint) error {
	u.Type = M_SP_NA_1
	bytes, err := u.add(addr, 1)
	if err != nil {
		return err
	}
	bytes[0] = byte(v)
	return nil
}

// SetSingles sets Type to M_SP_NA_1 and encodes a sequence of measured values
// into Info, whereby the n-th entry has an address of addr + n - 1.
func (u *ASDU) SetSingles(addr ObjAddr, seq ...SinglePoint) error {
	u.Type = M_SP_NA_1
	bytes, err := u.setSeq(len(seq), 1, addr)
	if err != nil {
		return err
	}
	for i, v := range seq {
		bytes[i] = byte(v)
	}
	return nil
}

// AddDouble sets Type to M_DP_NA_1 and appends the measured value to Info.
func (u *ASDU) AddDouble(addr ObjAddr, v DoublePoint) error {
	u.Type = M_SP_NA_1
	bytes, err := u.add(addr, 1)
	if err != nil {
		return err
	}
	bytes[0] = byte(v)
	return nil
}

// SetDoubles sets Type to M_DP_NA_1 and encodes a sequence of measured values
// into Info, whereby the n-th entry has an address of addr + n - 1.
func (u *ASDU) SetDoubles(addr ObjAddr, seq ...DoublePoint) error {
	u.Type = M_DP_NA_1
	bytes, err := u.setSeq(len(seq), 1, addr)
	if err != nil {
		return err
	}
	for i, v := range seq {
		bytes[i] = byte(v)
	}
	return nil
}

// MeasuredFloat is a floating-point with a quality descriptor.
type MeasuredFloat struct {
	Value float32
	// quality descriptor flags
	QualDesc uint8
}

// AddFloat sets Type to M_ME_NC_1 and appends the measured value to Info.
func (u *ASDU) AddFloat(addr ObjAddr, m MeasuredFloat) error {
	u.Type = M_ME_NC_1
	bytes, err := u.add(addr, 5)
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint32(bytes, math.Float32bits(m.Value))
	bytes[4] = m.QualDesc
	return nil
}

// SetFloats sets Type to M_ME_NC_1 and encodes a sequence of measured values
// into Info, whereby the n-th entry has an address of addr + n - 1.
func (u *ASDU) SetFloats(addr ObjAddr, seq ...MeasuredFloat) error {
	u.Type = M_ME_NC_1
	bytes, err := u.setSeq(len(seq), 5, addr)
	if err != nil {
		return err
	}
	for i, m := range seq {
		offset := i * 5
		binary.LittleEndian.PutUint32(bytes[offset:], math.Float32bits(m.Value))
		bytes[offset+4] = m.QualDesc
	}
	return nil
}

// AddFloatSetpoint sets Type to C_SE_NC_1 and appends the setpoint to Info.
func (u *ASDU) AddFloatSetpoint(addr ObjAddr, cmd SetpointCmd, v float32) error {
	u.Type = C_SE_NC_1
	bytes, err := u.add(addr, 5)
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint32(bytes, math.Float32bits(v))
	bytes[4] = byte(cmd)
	return nil
}

var errInroZero = errors.New("IEC60870-5-101: qualifier of interrogation 0 is not used")

// SetInro applies the qualifier of interrogation Inrogen or Inro1‥Inro16.
func (u *ASDU) SetInro(qual byte) error {
	u.Type = C_IC_NA_1
	if qual == 0 {
		return errInroZero
	}
	u.Info = make([]byte, u.CommonAddrSize+1)
	u.Info[u.CommonAddrSize] = qual
	return nil
}

func (u *ASDU) add(addr ObjAddr, n int) ([]byte, error) {
	bytes := u.Info
	i := len(bytes)

	// allocate & register when needed
	if l := i + u.ObjAddrSize + n; cap(bytes) < l {
		bytes = make([]byte, l)
		copy(bytes, u.Info)
	} else {
		bytes = bytes[:l]
	}
	u.Info = bytes
	u.InfoSeq = false

	if err := u.PutObjAddrAt(i, addr); err != nil {
		return nil, err
	}

	return bytes[i+u.ObjAddrSize:], nil
}

func (u *ASDU) setSeq(n, size int, addr ObjAddr) ([]byte, error) {
	if n < 0 || n > 127 {
		return nil, errObjFit
	}

	u.InfoSeq = true

	// allocate
	if l := u.ObjAddrSize + (n * size); cap(u.Info) >= l {
		u.Info = u.Info[:l]
	} else {
		u.Info = make([]byte, l)
	}

	if err := u.PutObjAddrAt(0, addr); err != nil {
		return nil, err
	}

	return u.Info[u.ObjAddrSize:], nil
}
