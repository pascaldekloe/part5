package part5

import "time"

// GetCP56Time2a reads 7 octets and returns the value or nil when invalid.
// The year is assumed to be in the 20th century.
// See IEC 60870-5-4 § 6.8 and IEC 60870-5-101 second edition § 7.2.6.18.
func getCP56Time2a(bytes []byte, loc *time.Location) *time.Time {
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
		return nil
	}

	hour := int(bytes[3] & 31)
	day := int(bytes[4] & 31)
	month := time.Month(bytes[5] & 15)
	year := 2000 + int(bytes[6]&127)

	nsec := msec * 1000000
	val := time.Date(year, month, day, hour, min, sec, nsec, loc)
	return &val
}

// GetCP24Time2a reads 3 octets and returns the value or nil when invalid.
// The moment is assumed to be in the recent present.
// See IEC 60870-5-4 § 6.8 and IEC 60870-5-101 second edition § 7.2.6.19.
func getCP24Time2a(bytes []byte, loc *time.Location) *time.Time {
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
		return nil
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

	return &val
}
