package info

import "fmt"

// Format implements the fmt.Formatter interface.
func (addr ComAddr8) Format(f fmt.State, verb rune) { format8(addr, f, verb) }

// Format implements the fmt.Formatter interface.
func (addr ComAddr16) Format(f fmt.State, verb rune) { format16(addr, f, verb) }

// Format implements the fmt.Formatter interface.
func (addr Addr8) Format(f fmt.State, verb rune) { format8(addr, f, verb) }

// Format implements the fmt.Formatter interface.
func (addr Addr16) Format(f fmt.State, verb rune) { format16(addr, f, verb) }

// Format implements the fmt.Formatter interface.
func (addr Addr24) Format(f fmt.State, verb rune) { format24(addr, f, verb) }

func format8(addr [1]uint8, f fmt.State, verb rune) {
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

func format16(addr [2]uint8, f fmt.State, verb rune) {
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

func format24(addr [3]uint8, f fmt.State, verb rune) {
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
