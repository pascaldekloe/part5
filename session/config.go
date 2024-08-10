package session

import "time"

const (
	// TCPPort is the IANA registered port number for unsecure connection.
	TCPPort = 2404

	// TCPPortSecure is the IANA registered port number for secure connection.
	TCPPortSecure = 19998
)

// TCPConfig defines an IEC 60870-5-104 configuration.
// The default is applied for each unspecified value.
type TCPConfig struct {
	// Maximum amount of time for TCP connection establishment. The standard
	// specifies "t₀" in [1, 255] seconds with a default of 30.
	ConnectTimeout time.Duration

	// Upper limit for the number of I-frames send without reception of a
	// confiramation. Transmission stops once this number has been reached.
	// The standard specifies "𝑘" in [1, 32767] with a default of 12.
	// See chapter 5.5 of companion standard 104.
	SendUnackMax uint

	// Maximum amount of time for frame reception confirmation. On expiry
	// the connection is closed immediately. The standard specifies "t₁" in
	// [1, 255] seconds with a default of 15.
	// See figure 18 of companion standard 104.
	SendUnackTimeout time.Duration

	// Upper limit for the number of I-frames received without sending a
	// receival confirmation. It is recommended that RecvUnackMax should not
	// exceed two thirds of SendUnackMax. The standard specifies "𝑤" in [1,
	// 32767] with a default of 8.
	// See chapter 5.5 of companion standard 104.
	RecvUnackMax uint

	// Maximum amount of time allowed for sending a receival confirmation.
	// In practice this framework will send such acknowledgement within a
	// second. The standard specifies "t₂" in [1, 255] seconds with a
	// default of 10.
	// See figure 10 of companion standard 104.
	RecvUnackTimeout time.Duration

	// Amount of idle time needed to trigger "TESTFR" keep-alives. The
	// standard recommends "t₃" in [1 second, 48 hours] and the default is
	// set to 20 seconds.
	// See chapter 5.2 of companion standard 104.
	IdleTimeout time.Duration
}

// Check applies the default (defined by IEC) for each unspecified value.
// A panic is raised for values out of range.
func (c *TCPConfig) check() *TCPConfig {
	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = 30 * time.Second
	} else if c.ConnectTimeout < 1*time.Second || c.ConnectTimeout > 255*time.Second {
		panic(`ConnectTimeout "t₀" not in [1, 255]s`)
	}

	if c.SendUnackMax == 0 {
		c.SendUnackMax = 12
	} else if c.SendUnackMax < 1 || c.SendUnackMax > 32767 {
		panic(`SendUnackMax "k" not in [1, 32767]`)
	}

	if c.SendUnackTimeout == 0 {
		c.SendUnackTimeout = 15 * time.Second
	} else if c.SendUnackTimeout < 1*time.Second || c.SendUnackTimeout > 255*time.Second {
		panic(`SendUnackTimeout "t₁" not in [1, 255]s`)
	}

	if c.RecvUnackMax == 0 {
		c.RecvUnackMax = 8
	} else if c.RecvUnackMax < 1 || c.RecvUnackMax > 32767 {
		panic(`RecvUnackMax "w" not in [1, 32767]`)
	}

	if c.RecvUnackTimeout == 0 {
		c.RecvUnackTimeout = 10 * time.Second
	} else if c.RecvUnackTimeout < 1*time.Second || c.RecvUnackTimeout > 255*time.Second {
		panic(`RecvUnackTimeout "t₂" not in [1, 255]s`)
	}

	if c.IdleTimeout == 0 {
		c.IdleTimeout = 20 * time.Second
	}

	return c
}
