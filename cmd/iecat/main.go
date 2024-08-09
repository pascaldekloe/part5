package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pascaldekloe/part5"
	"github.com/pascaldekloe/part5/info"
	"github.com/pascaldekloe/part5/session"
)

var CmdLog = log.New(os.Stderr, filepath.Base(os.Args[0])+": ", 0)

var (
	hostFlag = flag.String("host", "localhost", "Set the host name or IP number to connect with.")
	portFlag = flag.Uint("port", session.TCPPort, "Set the TCP port-`number` to connect with.")
)

// System Parameters
var (
	cOTAddrFlag = flag.Uint("cot-addr", 1, "The width for the cause of transmission is either 1 or 2 `octets`."+
		"\nValue 2 enables the originator address.")
	comAddrFlag = flag.Uint("com-addr", 2, "The width for the common address is either 1 or 2 `octets`.")
	objAddrFlag = flag.Uint("obj-addr", 2, "The width for the information object address is either 1, 2 or 3"+
		"\n`octets`.")
)

// The TCP configuration parameters are defined in companion standard 104,
// subsection 9.6, “Basic application functions”.
var (
	t0Flag = flag.Uint("t0", 30, "Time-out t₀ of connection establishment must be in range 1 to"+
		"\n255 `seconds`.")
	t1Flag = flag.Uint("t1", 15, "Time-out t₁ of send or test APDUs must be in range 1 to 255"+
		"\n`seconds`.")
	t2Flag = flag.Uint("t2", 10, "Time-out t₂ for acknowledges in case of no data messages must be"+
		"\nrange of 1 to 255 `seconds`.")
	t3Flag = flag.Uint("t3", 20, "Time-out t₃ for sending test frames in case of a long idle state"+
		"\nis limited to 32727 `seconds` for interoperatibility reasons.")
	kFlag = flag.Uint("k", 12, "The maximum number of outstanding I-frames [𝑘 `amount` without"+
		"\nacknowledge] must be in range 1 to 32767.")
	wFlag = flag.Uint("w", 8, "The latest acknowledge (after receiving 𝑤 `amount` of I-frames)"+
		"\nmust be in range 1 to 32767, and it should not exceed two thirds"+
		"\nof 𝑘.")
)

// Execution Options
var (
	inCapFlag = flag.Uint64("in-cap", 0, "Terminate the connection after receiving the `number` of ASDU"+
		"\nwith zero for unbound.")
	inroFlag = flag.String("inro", "<none>", "Send an interrogation activation request to the common `address`."+
		"\nUse the 0x prefix for a hexadecimal interpretation.")
)

func main() {
	log.SetFlags(0) // none
	flag.Parse()

	// listen to signals right away
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)

	// validated configariton
	stream := mustPacketStream()
	config := mustTCPConfig()

	addr := net.JoinHostPort(*hostFlag, strconv.FormatUint(uint64(*portFlag), 10))
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	client := session.TCP(config, conn)

	go stream.streamInbound(client)
	go stream.streamOutbound(client)

	var exitCode int
	defer os.Exit(exitCode)

	for {
		select {
		case sig := <-signals:
			CmdLog.Printf("got signal %s", sig)
			switch sig {
			case syscall.SIGINT:
				exitCode = 130
				client.Target <- session.Exit
			}

		case l := <-client.Level:
			CmdLog.Printf("connection level %s", l)
			if l == session.Exit {
				return
			}

		case err, ok := <-client.Err:
			if !ok {
				return
			}

			if s := err.Error(); strings.HasPrefix(s, "part5: ") {
				CmdLog.Print(s[7:])
			} else {
				log.Print(err)
			}
		}
	}
}

type packetStream interface {
	streamInbound(*session.Station)
	streamOutbound(*session.Station)
}

func mustPacketStream() packetStream {
	switch {
	case *cOTAddrFlag != 1 && *cOTAddrFlag != 2:
		CmdLog.Fatal("width of cause of transmission is neither 1 nor 2 octets")
	case *comAddrFlag != 1 && *comAddrFlag != 2:
		CmdLog.Fatal("width of common address is neither 1 nor 2 octets")
	case *objAddrFlag != 1 && *objAddrFlag != 2 && *objAddrFlag != 3:
		CmdLog.Fatal("width of information-object address is neither 1 nor 2 nor 3 octets")

	case *cOTAddrFlag == 1 && *comAddrFlag == 1 && *objAddrFlag == 1:
		return system[info.OrigAddr0, info.ComAddr8, info.ObjAddr8]{}
	case *cOTAddrFlag == 1 && *comAddrFlag == 1 && *objAddrFlag == 2:
		return system[info.OrigAddr0, info.ComAddr8, info.ObjAddr16]{}
	case *cOTAddrFlag == 1 && *comAddrFlag == 1 && *objAddrFlag == 3:
		return system[info.OrigAddr0, info.ComAddr8, info.ObjAddr24]{}

	case *cOTAddrFlag == 1 && *comAddrFlag == 2 && *objAddrFlag == 1:
		return system[info.OrigAddr0, info.ComAddr16, info.ObjAddr8]{}
	case *cOTAddrFlag == 1 && *comAddrFlag == 2 && *objAddrFlag == 2:
		return system[info.OrigAddr0, info.ComAddr16, info.ObjAddr16]{}
	case *cOTAddrFlag == 1 && *comAddrFlag == 2 && *objAddrFlag == 3:
		return system[info.OrigAddr0, info.ComAddr16, info.ObjAddr24]{}

	case *cOTAddrFlag == 2 && *comAddrFlag == 1 && *objAddrFlag == 1:
		return system[info.OrigAddr8, info.ComAddr8, info.ObjAddr8]{}
	case *cOTAddrFlag == 2 && *comAddrFlag == 1 && *objAddrFlag == 2:
		return system[info.OrigAddr8, info.ComAddr8, info.ObjAddr16]{}
	case *cOTAddrFlag == 2 && *comAddrFlag == 1 && *objAddrFlag == 3:
		return system[info.OrigAddr8, info.ComAddr8, info.ObjAddr24]{}

	case *cOTAddrFlag == 2 && *comAddrFlag == 2 && *objAddrFlag == 1:
		return system[info.OrigAddr8, info.ComAddr16, info.ObjAddr8]{}
	case *cOTAddrFlag == 2 && *comAddrFlag == 2 && *objAddrFlag == 2:
		return system[info.OrigAddr8, info.ComAddr16, info.ObjAddr16]{}
	case *cOTAddrFlag == 2 && *comAddrFlag == 2 && *objAddrFlag == 3:
		return system[info.OrigAddr8, info.ComAddr16, info.ObjAddr24]{}
	}

	panic("unreachable")
}

// System has a network setup.
type system[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	info.Params[Orig, Com, Obj]
}

// StreamInbound implements the packetStream interface.
func (sys system[Orig, Com, Obj]) streamInbound(client *session.Station) {
	mon := part5.NewLogger(sys.Params, os.Stdout)

	u := sys.NewDataUnit() // reusable
	var n uint64
	for p := range client.In {
		n++

		// parse
		err := u.Adopt(p)
		if err != nil {
			CmdLog.Print("payload from inbound APDU dropped: ",
				strings.TrimPrefix(err.Error(), "part5: "))
		} else {
			err := part5.MonitorDataUnit(mon, u)
			switch err {
			case nil:
				break // printed with payload aware formatting

			case part5.ErrNotMonitor, part5.ErrMonitorReserve:
				// fallback to print with default formatting,
				// unaware of its payload structure
				fmt.Printf("%s", u)

			default:
				CmdLog.Print("APDU dropped: ",
					strings.TrimPrefix(err.Error(), "part5: "))
			}
		}

		if n == *inCapFlag {
			CmdLog.Printf("reached %d inbound messages", n)
			client.Target <- session.Exit
		}
	}
}

// StreamOutbound implements the packetStream interface.
func (sys system[Orig, Com, Obj]) streamOutbound(client *session.Station) {
	client.Target <- session.Up

	if *inroFlag != "<none>" {
		// send interrogation activation
		// TODO: allow #<n> suffix for an origating address
		addrN, err := strconv.ParseUint(*inroFlag, 0, 32)
		if err != nil {
			CmdLog.Fatal("illegal interrogation address: ", err)
		}
		addr, ok := sys.ComAddrN(uint(addrN))
		if !ok {
			CmdLog.Fatalf("common address %d for interrogation exceeds %d-octet width of system",
				addrN, len(addr))
		}

		cmd := part5.Exchange[Orig, Com, Obj]{Com: addr}.Command()

		client.Class2 <- session.NewOutbound(cmd.Inro().Append(nil))
	}

	// TODO: Read standard input for messages to send.
}

// Reads an IEC 104 configuration from flags.
func mustTCPConfig() session.TCPConfig {
	switch {
	case *t0Flag == 0:
		CmdLog.Fatal("t₀ is zero")
	case *t1Flag == 0:
		CmdLog.Fatal("t₁ is zero")
	case *t2Flag == 0:
		CmdLog.Fatal("t₂ is zero")
	case *t3Flag == 0:
		CmdLog.Fatal("t₃ is zero")
	case *kFlag == 0:
		CmdLog.Fatal("𝑘 is zero")
	case *wFlag == 0:
		CmdLog.Fatal("𝑤 is zero")

	case *t0Flag > 255:
		CmdLog.Fatal("t₀ exceeds 255")
	case *t1Flag > 255:
		CmdLog.Fatal("t₁ exceeds 255")
	case *t2Flag > 255:
		CmdLog.Fatal("t₂ exceeds 255")
	case *t3Flag > 32767:
		CmdLog.Fatal("t₃ exceeds 32767")
	case *kFlag > 32767:
		CmdLog.Fatal("𝑘 exceeds 32767")
	case *wFlag > 32767:
		CmdLog.Fatal("𝑤 exceeds 32767")
	}

	if *wFlag > *kFlag-*kFlag/3 {
		// just a waring
		CmdLog.Print("𝑤 exceeds ⅔ of 𝑘")
	}

	return session.TCPConfig{
		ConnectTimeout:   time.Duration(*t0Flag) * time.Second,
		SendUnackTimeout: time.Duration(*t1Flag) * time.Second,
		RecvUnackTimeout: time.Duration(*t2Flag) * time.Second,
		IdleTimeout:      time.Duration(*t3Flag) * time.Second,

		SendUnackMax: *kFlag,
		RecvUnackMax: *wFlag,
	}
}
