package info

import (
	"fmt"
	"io"
)

// Format implements the fmt.Formatter interface.
// See the documentation of the info package for options.
func (addr OrigAddr0) Format(f fmt.State, verb rune) {
	switch verb {
	case 'X', 'x':
		io.WriteString(f, "00")

	case 'd':
		if f.Flag(' ') {
			io.WriteString(f, "  0")
		} else {
			io.WriteString(f, "0")
		}

	default:
		fmt.Fprintf(f, "%%!%c(BADVERB)", verb)
	}
}

// Format implements the fmt.Formatter interface.
// See the documentation of the info package for options.
func (addr OrigAddr8) Format(f fmt.State, verb rune) { formatAddr8(addr, f, verb) }

// Format implements the fmt.Formatter interface.
// See the documentation of the info package for options.
func (addr ComAddr8) Format(f fmt.State, verb rune) { formatAddr8(addr, f, verb) }

// Format implements the fmt.Formatter interface.
// See the documentation of the info package for options.
func (addr ComAddr16) Format(f fmt.State, verb rune) { formatAddr16(addr, f, verb) }

// Format implements the fmt.Formatter interface.
// See the documentation of the info package for options.
func (addr ObjAddr8) Format(f fmt.State, verb rune) { formatAddr8(addr, f, verb) }

// Format implements the fmt.Formatter interface.
// See the documentation of the info package for options.
func (addr ObjAddr16) Format(f fmt.State, verb rune) { formatAddr16(addr, f, verb) }

// Format implements the fmt.Formatter interface.
// See the documentation of the info package for options.
func (addr ObjAddr24) Format(f fmt.State, verb rune) { formatAddr24(addr, f, verb) }

func formatAddr8(addr [1]uint8, f fmt.State, verb rune) {
	switch verb {
	case 'X':
		fmt.Fprintf(f, "%02X", addr[0])

	case 'x':
		fmt.Fprintf(f, "%02x", addr[0])

	case 'd':
		switch {
		case f.Flag(' ') && !f.Flag('#'):
			fmt.Fprintf(f, "%3d", addr[0])
		default:
			fmt.Fprintf(f, "%d", addr[0])
		}

	default:
		fmt.Fprintf(f, "%%!%c(BADVERB)", verb)
	}
}

func formatAddr16(addr [2]uint8, f fmt.State, verb rune) {
	switch verb {
	case 'X':
		if f.Flag('#') {
			fmt.Fprintf(f, "%02X:%02X", addr[1], addr[0])
		} else {
			fmt.Fprintf(f, "%04X", uint(addr[0])|uint(addr[1])<<8)
		}

	case 'x':
		if f.Flag('#') {
			fmt.Fprintf(f, "%02x:%02x", addr[1], addr[0])
		} else {
			fmt.Fprintf(f, "%04x", uint(addr[0])|uint(addr[1])<<8)
		}

	case 'd':
		switch {
		case f.Flag('#'):
			fmt.Fprintf(f, "%d.%d", addr[1], addr[0])
		case f.Flag(' '):
			fmt.Fprintf(f, "%5d", uint(addr[0])|uint(addr[1])<<8)
		default:
			fmt.Fprintf(f, "%d", uint(addr[0])|uint(addr[1])<<8)
		}

	default:
		fmt.Fprintf(f, "%%!%c(BADVERB)", verb)
	}
}

func formatAddr24(addr [3]uint8, f fmt.State, verb rune) {
	switch verb {
	case 'X':
		if f.Flag('#') {
			fmt.Fprintf(f, "%02X:%02X:%02X", addr[2], addr[1], addr[0])
		} else {
			fmt.Fprintf(f, "%06X", uint(addr[0])|uint(addr[1])<<8|uint(addr[2])<<16)
		}

	case 'x':
		if f.Flag('#') {
			fmt.Fprintf(f, "%02x:%02x:%02x", addr[2], addr[1], addr[0])
		} else {
			fmt.Fprintf(f, "%06x", uint(addr[0])|uint(addr[1])<<8|uint(addr[2])<<16)
		}

	case 'd':
		switch {
		case f.Flag('#'):
			fmt.Fprintf(f, "%d.%d.%d", addr[2], addr[1], addr[0])
		case f.Flag(' '):
			fmt.Fprintf(f, "%8d", uint(addr[0])|uint(addr[1])<<8|uint(addr[2])<<16)
		default:
			fmt.Fprintf(f, "%d", uint(addr[0])|uint(addr[1])<<8|uint(addr[2])<<16)
		}

	default:
		fmt.Fprintf(f, "%%!%c(BADVERB)", verb)
	}
}

// Format implements the fmt.Formatter interface. A "%s" describes the ASDU with
// addresses as decimal numbers. Use the “alternated format” "%#s" for addresses
// in hexadecimal, i.e., the "%#x" as described in the documentation of the info
// package.
func (u DataUnit[Orig, Com, Obj]) Format(f fmt.State, verb rune) {
	if verb != 's' {
		fmt.Fprintf(f, "%%!%c(BADVERB)", verb)
		return
	}

	format := "%s %s %d %d:"
	if f.Flag('#') {
		format = "%s %s %x %#x:"
	}
	fmt.Fprintf(f, format, u.Type, u.Cause, u.Orig, u.Addr)

	var addr Obj

	n := u.Enc.Count()
	switch {
	case u.Enc.AddrSeq():
		// only the first address is encoded
		if len(addr) > len(u.Info) {
			fmt.Fprintf(f, " SQ @ %#x<EOF> ~%d !", u.Info, n)
			return
		}
		firstAddr := Obj(u.Info[:len(addr)])

		// print sequence start
		addrFmt := " SQ@%d"
		if f.Flag('#') {
			addrFmt = " SQ@%#x"
		}
		fmt.Fprintf(f, addrFmt, firstAddr)

		// payload with n fixed-size elements
		p := u.Info[len(addr):]
		if n == 0 {
			if len(p) != 0 {
				fmt.Fprintf(f, " %#x ~0 ?", p)
				return
			}
			break // orphan address allowed
		}

		// size protection
		if len(p) > 250 {
			fmt.Fprintf(f, " %#x... (%d B) ~%d ?", p[:80], len(p), n)
			return
		}

		size := len(p) / n
		if len(p)%n != 0 || (size == 0 && len(p) != 0) {
			fmt.Fprintf(f, " %#x ~%d ?", p, n)
			return
		}
		if size != 0 {
			// print elements
			for i := 0; i+size <= len(p); i += size {
				fmt.Fprintf(f, " %#x", p[i:i+size])
			}
		}

		// overflow check
		lastAddr := firstAddr.N() + uint(n) - 1
		if _, ok := u.ObjAddrN(lastAddr); !ok {
			io.WriteString(f, " @^ !")
			return
		}

	default:
		// n addresses are followed by a fixed-size element (in pairs)
		if n == 0 {
			if len(u.Info) != 0 {
				fmt.Fprintf(f, " %#x ~0 ?", u.Info)
				return
			}
			break // not explicitly prohibited
		}
		// size protection
		if len(u.Info) > 250 {
			fmt.Fprintf(f, " %#x... (%d B) ~%d ?", u.Info[:80], len(u.Info), n)
			return
		}
		size := len(u.Info) / n
		if size < len(addr) || len(u.Info)%n != 0 {
			fmt.Fprintf(f, " %#x ~%d ?", u.Info, n)
			return
		}

		format := " %#x@%d"
		if f.Flag('#') {
			format = " %#x@%#x"
		}
		for i := 0; i+size <= len(u.Info); i += size {
			fmt.Fprintf(f, format,
				u.Info[i+len(addr):i+size],
				Obj(u.Info[i:i+len(addr)]),
			)
		}
	}

	// OK
	io.WriteString(f, " .")
}
