package part5

import "time"

// CP56Time2a, a.k.a. the seven octet “Binary Time 2a”, provides millisecond
// precission while omitting both the century and the time-zone.
type CP56Time2a [7]byte

// Set marshals the time. Reserved bits can be set afterwards.
// The zero value marks the time as invalid (with the IV flag).
// Digits beyond milliseconds are dropped (without rounding).
// ⚠️ Encoding loses the time-zone context.
func (c *CP56Time2a) SetTo(t time.Time) {
	c.setAll(t, true)
}

// SetNotAll skips both the day-of-the-week field and the summer-time flag.
func (c *CP56Time2a) SetNotAll(t time.Time) {
	c.setAll(t, false)
}

func (c *CP56Time2a) setAll(t time.Time, all bool) {
	if t.IsZero() {
		c[0] = 0
		c[1] = 0
		c[2] = 0x80 // invalid flag [IV]
		c[3] = 0
		c[4] = 0
		c[5] = 0
		c[6] = 0
		return
	}

	year, month, day := t.Date()
	hour, minute, second := t.Clock()

	if all {
		if t.IsDST() {
			// summer-time flag [SU]
			hour |= 0x80
		}
		day |= int(t.Weekday()+1) << 5
	}

	millis := uint(second*1000) + uint(t.Nanosecond()/1e6)
	c[0] = byte(millis)
	c[1] = byte(millis >> 8)

	c[2] = byte(minute)
	c[3] = byte(hour)
	c[4] = byte(day)
	c[5] = byte(month)
	c[6] = byte(year % 100)
}

// In20thCentury unmarshals the time without any validation, and it then
// reconstructs the timestamp under the assumption that the moment was in the
// 20th century, and that encoding applied an equivalent of the time.Location.
// The return is zero when the invalid flag [IV] is set. Otherwise the return is
// in range [2000-01-01 00:00:00.000, 2099-12-31 23:59:59.999].
func (c *CP56Time2a) In20thCentury(loc *time.Location) time.Time {
	// check invalid flag
	if c[2]&0x80 != 0 {
		return time.Time{}
	}

	yearInCentury, month, day := c.Calendar()
	year := yearInCentury + 2000

	hour, min, secInMilli := c.ClockAndMillis()
	// split millisecond count
	sec := secInMilli / 1000
	nanos := (secInMilli % 1000) * 1e6

	return time.Date(year, month, day, hour, min, sec, nanos, loc)
}

// Calendar unmarshals the day counts without any validation.
// Note that the year count range [0..99] excludes the century.
func (c *CP56Time2a) Calendar() (yearInCentury int, _ time.Month, day int) {
	return int(uint(c[6]) & 0x7f),
		time.Month(uint(c[5]) & 0x0f),
		int(uint(c[4]) & 0x1f)
}

// ClockAndMillis unmarshals the time of the day without any validation.
// Note that the seconds and the milliseconds are combined in one count,
// range [0..59999].
func (c *CP56Time2a) ClockAndMillis() (hour, min, secInMilli int) {
	return int(uint(c[3]) & 0x1f),
		int(uint(c[2]) & 0x3f),
		int(uint(c[1])<<8 | uint(c[0]))
}

// Reserve1 returns the RES1 bit.
// The flag may indicate genuine time, with false for substituted time.
//
// “The RES1-bit may be used in the monitor direction to indicate whether the
// time tag was added to the information object when it was acquired by the RTU
// (genuine time) or the time tag was substituted by intermediate equipment such
// as concentrator stations or by the controlling station itself (substituted
// time).”
// — companion standard 101, subsection 7.2.6.18
func (c *CP56Time2a) Reserve1() bool {
	return c[2]&0x40 != 0
}

// FlagReserve1 sets the RES1 bit.
func (c *CP56Time2a) FlagReserve1() {
	c[2] |= 0x40
}

// Reserve2 returns the RES2 bits, range [0x00..0x03].
func (c *CP56Time2a) Reserve2() uint {
	return uint(c[3]>>5) & 0x03
}

// FlagReserve2 sets [logical OR] the RES2 bits from any of the two
// least-significant bits given.
func (c *CP56Time2a) FlagReserve2(bits uint) {
	c[3] |= byte(bits&0x03) << 5
}

// Reserve3 returns the RES3 bits, range [0x00..0x0f].
func (c *CP56Time2a) Reserve3() uint {
	return uint(c[5]>>4) & 0x0f
}

// FlagReserve3 sets [logical OR] the RES3 bits from any of the four
// least-significant bits given.
func (c *CP56Time2a) FlagReserve3(bits uint) {
	c[5] |= byte(bits) << 4
}

// CP24Time2a, a.k.a. the three octet “Binary Time 2a”, provides millisecond
// precission while omitting both the hour and the time-zone context.
type CP24Time2a [3]byte

// Set marshals the time. Reserved bits can be set afterwards.
// The zero value marks the time as invalid (with the IV flag).
// Digits beyond milliseconds are dropped (without rounding).
// ⚠️ Note that encoding loses the time-zone context.
func (c *CP24Time2a) Set(t time.Time) {
	if t.IsZero() {
		c[0] = 0
		c[1] = 0
		c[2] = 0x80 // invalid flag [IV]

		return
	}

	_, minute, second := t.Clock()
	millis := uint(second*1000) + uint(t.Nanosecond()/1e6)

	c[0] = byte(millis)
	c[1] = byte(millis >> 8)
	c[2] = byte(minute)
}

// MinuteAndMillis unmarshals the time without any validation.
// Note that the seconds and the milliseconds are combined in one
// count, range [0..59999].
func (c *CP24Time2a) MinuteAndMillis() (min, secInMilli int) {
	return int(uint(c[2]) & 0x3f),
		int(uint(c[1])<<8 | uint(c[0]))
}

// InLessThanHourBefore unmarshals the time without any validation, and it then
// reconstructs the timestamp under the assumption that the moment is less than
// one hour before t, and that the encoding applied an equivalent time.Location.
// The return is zero when the invalid flag [IV] is set. Otherwise the return is
// in range [t − 00:59:59.999, t].
func (c *CP24Time2a) InLessThanHourBefore(t time.Time) time.Time {
	if c.Invalid() {
		return time.Time{}
	}

	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	secInMilli := sec*1000 + t.Nanosecond()/1e6

	minEnc, secInMilliEnc := c.MinuteAndMillis()

	// could be in previous hour
	if min < minEnc || (min == minEnc && secInMilli < secInMilliEnc) {
		hour--
		// Note that Go applies a negative hour count correctly.
	}

	return time.Date(year, month, day, hour, minEnc,
		secInMilliEnc/1000, (secInMilliEnc%1000)*1e6,
		t.Location())
}

// Invalid returns the IV flag.
func (c *CP24Time2a) Invalid() bool {
	return c[2]&0x80 != 0
}

// Reserve1 returns the RES1 bit.
// The flag may indicate genuine time, with false for substituted time.
//
// “The RES1-bit may be used in the monitor direction to indicate whether the
// time tag was added to the information object when it was acquired by the RTU
// (genuine time) or the time tag was substituted by intermediate equipment such
// as concentrator stations or by the controlling station itself (substituted
// time).”
// — companion standard 101, subsection 7.2.6.18
func (c *CP24Time2a) Reserve1() bool {
	return c[2]&0x40 != 0
}

// FlagReserve1 sets the RES1 bit.
func (c *CP24Time2a) FlagReserve1() {
	c[2] |= 0x40
}

// GetCP56Time2a reads 7 octets and returns the value or zero when invalid.
// The year is assumed to be in the 20th century.
// See IEC 60870-5-4 § 6.8 and IEC 60870-5-101 second edition § 7.2.6.18.
func getCP56Time2a(bytes []byte, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}

	x := int(bytes[0])
	x |= int(bytes[1]) << 8
	msec := x % 1000
	sec := (x / 1000)

	o := bytes[2]
	min := int(o & 63)
	if o > 127 {
		return time.Time{}
	}

	hour := int(bytes[3] & 31)
	day := int(bytes[4] & 31)
	month := time.Month(bytes[5] & 15)
	year := 2000 + int(bytes[6]&127)

	nsec := msec * 1000000
	return time.Date(year, month, day, hour, min, sec, nsec, loc)
}

// GetCP24Time2a reads 3 octets and returns the value or zero when invalid.
// The moment is assumed to be in the recent present.
// See IEC 60870-5-4 § 6.8 and IEC 60870-5-101 second edition § 7.2.6.19.
func getCP24Time2a(bytes []byte, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}

	x := int(bytes[0])
	x |= int(bytes[1]) << 8
	msec := x % 1000
	sec := (x / 1000)

	o := bytes[2]
	min := int(o & 63)
	if o > 127 {
		return time.Time{}
	}

	now := time.Now()
	year, month, day := now.Date()
	hour, currentMin, _ := now.Clock()

	nsec := msec * 1000000
	val := time.Date(year, month, day, hour, min, sec, nsec, loc)

	// 5 minute rounding - 55 minute span
	if min > currentMin+5 {
		val = val.Add(-time.Hour)
	}

	return val
}
