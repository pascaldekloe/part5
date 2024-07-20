package info

import (
	"fmt"
	"io"
)

// Format implements the fmt.Formatter interface.
// See the package documentation of info for options.
func (c COT8) Format(f fmt.State, verb rune) {
	formatCauseFlagOrig(f, verb, c[0], 0)
}

// Format implements the fmt.Formatter interface.
// See the package documentation of info for options.
func (c COT16) Format(f fmt.State, verb rune) {
	formatCauseFlagOrig(f, verb, c[0], c[1])
}

func formatCauseFlagOrig(f fmt.State, verb rune, c, orig uint8) {
	if verb != 's' {
		fmt.Fprintf(f, "%%!%c(BADVERB)", verb)
		return
	}

	io.WriteString(f, causeLabels[c&63])

	switch c & (negFlag | testFlag) {
	case negFlag:
		io.WriteString(f, ",neg")
	case testFlag:
		io.WriteString(f, ",test")
	case negFlag | testFlag:
		io.WriteString(f, ",neg,test")
	}

	if orig != 0 {
		format := " #%d"
		if f.Flag('#') {
			format = " #%02x"
		}
		fmt.Fprintf(f, format, orig)
	}
}

// Format implements the fmt.Formatter interface.
// See the package documentation of info for options.
func (addr ComAddr8) Format(f fmt.State, verb rune) { format8(addr, f, verb) }

// Format implements the fmt.Formatter interface.
// See the package documentation of info for options.
func (addr ComAddr16) Format(f fmt.State, verb rune) { format16(addr, f, verb) }

// Format implements the fmt.Formatter interface.
// See the package documentation of info for options.
func (addr Addr8) Format(f fmt.State, verb rune) { format8(addr, f, verb) }

// Format implements the fmt.Formatter interface.
// See the package documentation of info for options.
func (addr Addr16) Format(f fmt.State, verb rune) { format16(addr, f, verb) }

// Format implements the fmt.Formatter interface.
// See the package documentation of info for options.
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

// Format implements the fmt.Formatter interface.
// See the package documentation of info for details.
func (u DataUnit[COT, Common, Object]) Format(f fmt.State, verb rune) {
	if verb != 's' {
		fmt.Fprintf(f, "%%!%c(BADVERB)", verb)
		return
	}

	format := "%s @%d %s:"
	if f.Flag('#') {
		format = "%s @%#x %s:"
	}
	fmt.Fprintf(f, format, u.Type, u.Addr, u.COT)

	dataSize := ObjSize[u.Type]
	switch {
	case dataSize == 0:
		// structure unknown
		fmt.Fprintf(f, " %#x ?", u.Info)

	case !u.Var.IsSeq():
		// objects paired with an address each
		var i int // read index
		for obj_n := 0; ; obj_n++ {
			var addr Object
			if i+len(addr)+dataSize > len(u.Info) {
				if i < len(u.Info) {
					fmt.Fprintf(f, " %#x<EOF>", u.Info[i:])
					obj_n++
				}

				diff := obj_n - u.Var.Count()
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
		var addr Object
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

				diff := obj_n - u.Var.Count()
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
			addr, _ = u.AddrOf(addr.N() + 1)
		}
	}
}
