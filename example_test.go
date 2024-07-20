package part5_test

import (
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/pascaldekloe/part5"
)

func TestMain(m *testing.M) {
	// Logging documents intend in examples.
	// Not interested in the output though.
	log.SetOutput(io.Discard)

	os.Exit(m.Run())
}

// Time tag reconstruction with some leeway comes recommended.
func ExampleCP24Time2a_WithinHourBefore() {
	// Suppose a measured value with time information.
	var tag part5.CP24Time2a

	// The time tag is relative to the now.
	received := time.Now()

	// Allow up to 5 seconds into the future to
	// account for time synchronisation issues.
	const leeway = 5 * time.Second

	t := tag.WithinHourBefore(received.Add(leeway))
	log.Println("timestamp reconstructed to", t)
}
