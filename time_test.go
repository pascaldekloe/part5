package part5

import (
	"testing"
	"time"
)

var goldenCP24Time2as = []struct {
	serial         []byte
	min, sec, nsec int
}{
	{[]byte{1, 2, 3}, 3, 0, 513000000},
	{[]byte{11, 12, 13}, 13, 3, 83000000},
}

var brokenCP24Time2as = [][]byte{
	[]byte{1, 2, 131},
}

func TestGetCP24Time2a(t *testing.T) {
	for _, gold := range goldenCP24Time2as {
		got := getCP24Time2a(gold.serial, time.UTC)
		switch {
		case got.Nanosecond() != gold.nsec,
			got.Second() != gold.sec,
			got.Minute() != gold.min:
			t.Errorf("%#x got %s, want …:%02d:%02d.%09d", gold.serial, got, gold.min, gold.sec, gold.nsec)
		}
	}

	for _, serial := range brokenCP24Time2as {
		got := getCP24Time2a(serial, time.UTC)
		if !got.IsZero() {
			t.Errorf("%#x got %s, want the zero time", serial, got)
		}
	}
}

var goldenCP56Time2as = []struct {
	serial []byte
	time   time.Time
}{
	{[]byte{1, 2, 3, 4, 5, 6, 7}, time.Date(2007, 6, 5, 4, 3, 0, 513000000, time.UTC)},
	{[]byte{1, 2, 131, 4, 5, 6, 7}, time.Time{}},
	{[]byte{11, 12, 13, 14, 15, 16, 17}, time.Date(2016, 12, 15, 14, 13, 3, 83000000, time.UTC)},
}

func TestGetCP56Time2a(t *testing.T) {
	for _, gold := range goldenCP56Time2as {
		got := getCP56Time2a(gold.serial, time.UTC)
		if !got.Equal(gold.time) {
			t.Errorf("%#x got %s, want %s", gold.serial, got, gold.time)
		}
	}
}
