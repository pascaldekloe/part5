package info_test

import (
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/pascaldekloe/part5/info"
)

func TestMain(m *testing.M) {
	// Logging documents intend in examples.
	// Not interested in the output though.
	log.SetOutput(io.Discard)

	os.Exit(m.Run())
}

var goldenCP24Time2as = []struct {
	enc            info.CP24Time2a
	min, sec, nsec int
}{
	{info.CP24Time2a{1, 2, 3}, 3, 0, 513000000},
	{info.CP24Time2a{11, 12, 13}, 13, 3, 83000000},
}

var brokenCP24Time2as = []info.CP24Time2a{
	{1, 2, 131},
}

func TestWithinHourBefore(t *testing.T) {
	now := time.Now()

	for _, gold := range goldenCP24Time2as {
		got := gold.enc.WithinHourBefore(now)
		if got.After(now) || now.Sub(got) >= time.Hour {
			t.Errorf("%#x got %s, want less than hour before %s",
				gold.enc, got, now)
		}

		min, sec, nsec := got.Minute(), got.Second(), got.Nanosecond()
		if min != gold.min || sec != gold.sec || nsec != gold.nsec {
			t.Errorf("%#x got %02d:%02d.%09d, want %02d:%02d.%09d",
				gold.enc, min, sec, nsec, gold.min, gold.sec, gold.nsec)
		}
	}

	for _, enc := range brokenCP24Time2as {
		got := enc.WithinHourBefore(now)
		if !got.IsZero() {
			t.Errorf("%#x got %s, want zero", enc, got)
		}
	}
}

var goldenCP56Time2as = []struct {
	enc  info.CP56Time2a
	time time.Time
}{
	{info.CP56Time2a{1, 2, 3, 4, 5, 6, 7},
		time.Date(2007, 6, 5, 4, 3, 0, 513000000, time.UTC)},
	{info.CP56Time2a{1, 2, 131, 4, 5, 6, 7},
		time.Time{}},
	{info.CP56Time2a{11, 12, 13, 14, 15, 16, 17},
		time.Date(2016, 12, 15, 14, 13, 3, 83000000, time.UTC)},
}

func TestWithin20thCentury(t *testing.T) {
	for _, gold := range goldenCP56Time2as {
		got := gold.enc.Within20thCentury(time.UTC)
		if !got.Equal(gold.time) {
			t.Errorf("%#x got %s, want %s", gold.enc, got, gold.time)
		}
	}
}

// Time tag reconstruction with some leeway comes recommended.
func ExampleCP24Time2a_WithinHourBefore() {
	// Suppose a measured value with time information.
	var tag info.CP24Time2a

	// The time tag is relative to the now.
	received := time.Now()

	// Allow up to 5 seconds into the future to
	// account for time synchronisation issues.
	const leeway = 5 * time.Second

	t := tag.WithinHourBefore(received.Add(leeway))
	log.Println("timestamp reconstructed to", t)
}
