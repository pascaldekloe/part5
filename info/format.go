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

	dataSize := ObjSize[u.Type]
	switch {
	case dataSize == 0:
		// structure unknown
		fmt.Fprintf(f, " %#x ?", u.Info)

	case !u.Enc.AddrSeq():
		// objects paired with an address each
		var i int // read index
		for obj_n := 0; ; obj_n++ {
			var addr Obj
			if i+len(addr)+dataSize > len(u.Info) {
				if i < len(u.Info) {
					fmt.Fprintf(f, " %#x<EOF>", u.Info[i:])
					obj_n++
				}

				diff := obj_n - u.Enc.Count()
				switch {
				case diff < 0:
					fmt.Fprintf(f, " 𝚫 −%d !", -diff)
				case diff > 0:
					fmt.Fprintf(f, " 𝚫 +%d !", diff)
				case i < len(u.Info):
					io.WriteString(f, " !")
				default:
					io.WriteString(f, " .")
				}

				break
			}

			for j := 0; j < len(addr); j++ {
				addr[j] = u.Info[i+j]
			}
			i += len(addr)
			info := u.Info[i : i+dataSize]
			i += dataSize

			if f.Flag('#') {
				fmt.Fprintf(f, " %#x@%#x ", info, addr)
			} else {
				fmt.Fprintf(f, " %#x@%d", info, addr)
			}
		}

	default:
		// offset address in Sequence
		var addr Obj
		if len(addr) > len(u.Info) {
			fmt.Fprintf(f, " %#x<EOF> !", u.Info)
			break
		}
		for i := 0; i < len(addr); i++ {
			addr[i] = u.Info[i]
		}
		i := len(addr) // read index

		for obj_n := 0; ; obj_n++ {
			if i+dataSize > len(u.Info) {
				if i < len(u.Info) {
					info := u.Info[i:]
					if f.Flag('#') {
						fmt.Fprintf(f, " %#x<EOF>@%#x ", info, addr)
					} else {
						fmt.Fprintf(f, " %#x<EOF>@%d", info, addr)
					}
					obj_n++
				}

				diff := obj_n - u.Enc.Count()
				switch {
				case diff < 0:
					fmt.Fprintf(f, " 𝚫 −%d !", -diff)
				case diff > 0:
					fmt.Fprintf(f, " 𝚫 +%d !", diff)
				case i < len(u.Info):
					io.WriteString(f, " !")
				default:
					io.WriteString(f, " .")
				}

				break
			}

			info := u.Info[i : i+dataSize]
			i += dataSize

			if f.Flag('#') {
				fmt.Fprintf(f, " %#x@%#x ", info, addr)
			} else {
				fmt.Fprintf(f, " %#x@%d", info, addr)
			}

			// silent overflow
			addr, _ = u.ObjAddrN(addr.N() + 1)
		}
	}
}
