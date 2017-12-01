package media

import "errors"
import "io"

// Upon detecting an error, a minimum interval of line idle bits is required
// between packets, depending on the type.
const (
	// format class 1.1, section 1, subclause 6.2.4.1, rule R4
	FT11IdleBits = 22
	// format class 1.2, section 1, subclause 6.2.4.2.1, rule R4
	FT12IdleBits = 33
	// format class 2, section 1, subclause 6.2.4.3.1, rule R5
	FT2IdleBits = 48
	// format class 3, section 1, subclause 6.2.4.4.1, rule R5
	FT3IdleBits = 54
)

// ErrCheck signals integrity verification failure.
var ErrCheck = errors.New("part5: checksum, constant or redundant value mismatch in packet")

// ErrDataFit signals user data out of bounds.
var ErrDataFit = errors.New("part5: user data size exceeds packet capacity or system limit")

// NewFT12 returns a new codec for format class FT 1.2,
// integrity class Ⅱ, Hamming distance 4.
//
// The number of octets for station addresses must be in [0, 2].
// When disabled with 0, packets use GlobalAddr exclusively.
//
// The fuction of single (control) character Ⅰ (0xE5) and Ⅱ (0xA2) depends on the
// system. Character Ⅰ is usually Ack. Don't forget to set FromPrimaryFlag where
// applicable. Both characters are decoded as a packet with GlobalAddr in place.
//
// WARNING! The implementation reuses one data buffer which makes operation
// efficient. The user data in packets can only be used until the next call.
//
// See section 2, subclause 3.2 and companion standard 101, annex 1.2.
func NewFT12(addrSize, maxSize int, singleChar1, singleChar2 Ctrl) FT {
	if addrSize < 0 || addrSize > 2 {
		panic("station address octet size not in [0, 2]")
	}
	if maxSize < 0 || 1+addrSize+maxSize > 255 {
		panic("station address + data size limit exceeds 254 octets")
	}
	return &ft12{
		addrSize:    addrSize,
		maxSize:     maxSize,
		singleChar1: singleChar1,
		singleChar2: singleChar2,
	}
}

type ft12 struct {
	addrSize    int
	maxSize     int
	singleChar1 Ctrl
	singleChar2 Ctrl
	buf         [257]byte // max user data size 255 + checksum & end byte
}

// Encode honors the FT interface.
func (ft *ft12) Encode(w io.Writer, p Packet) error {
	buf := &ft.buf

	// start user data at byte 5
	buf[4] = byte(p.Ctrl)

	switch ft.addrSize {
	case 2:
		// little-endian
		buf[5] = byte(p.Addr)
		buf[6] = byte(p.Addr >> 8)

	case 1:
		a := p.Addr
		if a == GlobalAddr {
			a = 255
		} else if a >= 255 {
			return ErrAddrFit
		}
		buf[5] = byte(a)

	case 0:
		if p.Addr != GlobalAddr {
			return ErrAddrFit
		}
	}
	end := 5 + ft.addrSize

	if len(p.Data) == 0 {
		buf[3] = 0x10
		if _, err := w.Write(buf[3:end]); err != nil {
			return err
		}
	} else {
		size := len(p.Data) + 1 + ft.addrSize
		if size > 255 {
			return ErrDataFit
		}
		buf[0] = 0x68
		buf[1] = byte(size)
		buf[2] = byte(size)
		buf[3] = 0x68
		if _, err := w.Write(buf[:end]); err != nil {
			return err
		}
	}

	// checksum conform section 1, clause 6.2.4.2.1, rule R5
	var sum byte
	for _, c := range buf[4:end] {
		sum += c
	}
	for _, c := range p.Data {
		sum += c
	}

	if _, err := w.Write(p.Data); err != nil {
		return err
	}

	buf[0] = sum
	buf[1] = 0x16
	if _, err := w.Write(buf[:2]); err != nil {
		return err
	}

	return nil
}

// Decode honors the FT interface.
func (ft *ft12) Decode(r io.Reader) (Packet, error) {
	buf := &ft.buf
	var size int // octet count of user data

	// check start character
	if _, err := io.ReadFull(r, buf[:1]); err != nil {
		return Packet{}, err
	}
	switch buf[0] {
	default:
		return Packet{}, ErrCheck

	case 0xe5: // single (control) character Ⅰ
		return Packet{Addr: GlobalAddr, Ctrl: ft.singleChar1}, nil

	case 0xa2: // single (control) character Ⅱ
		return Packet{Addr: GlobalAddr, Ctrl: ft.singleChar2}, nil

	case 0x10: // fixed length
		size = ft.addrSize + 1

	case 0x68: // variable length
		switch _, err := io.ReadFull(r, buf[:3]); err {
		case nil:
			break
		case io.EOF:
			return Packet{}, io.ErrUnexpectedEOF
		default:
			return Packet{}, err
		}

		n := buf[0]
		if n != buf[1] || buf[2] != 0x68 {
			return Packet{}, ErrCheck
		}
		size = int(n)
		if size > ft.maxSize+ft.addrSize+1 {
			return Packet{}, ErrDataFit
		}
	}

	switch _, err := io.ReadFull(r, buf[:size+2]); err {
	case nil:
		break
	case io.EOF:
		return Packet{}, io.ErrUnexpectedEOF
	default:
		return Packet{}, err
	}

	// check end character
	if buf[size+1] != 0x16 {
		return Packet{}, ErrCheck
	}

	// checksum conform section 1, clause 6.2.4.2.1, rule R5
	var sum byte
	for i := size - 1; i >= 0; i-- {
		sum += buf[i]
	}
	if sum != buf[size] {
		return Packet{}, ErrCheck
	}

	// decode little-endian address
	var addr Addr
	switch ft.addrSize {
	case 1:
		addr = Addr(buf[1])
	case 2:
		addr = Addr(buf[1]) | Addr(buf[2])<<8
	}

	return Packet{addr, Ctrl(buf[0]), buf[1+ft.addrSize : size]}, nil
}
