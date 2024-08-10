// Package part5 provides the OSI application layer of IEC 60870-5.
//
// # Monitor
//
// The Monitor interface holds listeners for measured values, i.e., the type
// codes with an "M_" prefix. Subinterfaces of Monitor organise per information
// type. Proxies of the subinterfaces abstract notification even further into a
// single method, regardless of the time tag. Such hyrachical setup allows users
// to write for the applicable types only. MonitorDelegate can default to a
// Logger or some equivalent to detect unsupported types.
package part5

import (
	"errors"
	"fmt"

	"github.com/pascaldekloe/part5/info"
)

// CmdUnk signals command rejection in control direction due to an unknown type
// identifier, or due to an unknown cause of transmission, or due to an unknown
// common address, or due to an unknown information-object address.
type CmdUnk struct {
	Type info.TypeID // command identifier

	// The cause of transmission from the response has the info.NegFlag set.
	// It may also have the info.TestFlag set.
	info.Cause
}

// Error implements the builtin.error interface.
func (unk CmdUnk) Error() string {
	return fmt.Sprintf("part5: %s %d: request denied", unk.Type, unk.Cause)
}

// CauseMis signals an illegal cause in response.
type CauseMis struct {
	Type info.TypeID // command identifier

	Req info.Cause // from request
	Res info.Cause // from response
}

// Error implements the builtin.error interface.
func (mis CauseMis) Error() string {
	return fmt.Sprintf("part5: %s %s response on %s request: response mismatch",
		mis.Type, mis.Res, mis.Req)
}

// ErrNotCmd rejects an info.DataUnit based on its type identifier.
var ErrNotCmd = errors.New("part5: ASDU type identifier not in command range 45..69")

func isCommand(t info.TypeID) bool {
	return t-45 < 25
}

// Response alternatives to info.Actcon are propagated as errors for simplicity.
var (
	// A negative confirm denies without justification.
	ErrConNeg = errors.New("part5: command not activated [actcon, P/N := <1>]")
	// Termination notification is optional, depending on the implementation.
	ErrTerm = errors.New("part5: command terminated [actterm]")
)

// ErrOtherCmd may indicate an orphaned command from earlier activity.
var ErrOtherCmd = errors.New("part5: response for other command")

// ConOf returns a specific error if the in(bound) packet is not a positive
// confirmation of an activation or a deactivation req(uest). Valid alternatives
// may include ErrNotCmd, ErrTerm or ErrConNeg.
func ConOf[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](in, req info.DataUnit[Orig, Com, Obj]) error {
	// command match
	if !in.Mirrors(req) {
		if !isCommand(in.Type) {
			return ErrNotCmd
		}
		return ErrOtherCmd
	}

	// sanity check
	switch req.Cause &^ info.TestFlag {
	case info.Act, info.Deact:
		break // pass
	default:
		// incorrect API use
		return fmt.Errorf("part5: cause %s of request not act nor deact", req.Cause)
	}

	// Match cause of transmission (or not) excluding test flag.
	// Note that Mirrors above checked the test flag for equality.
	switch in.Cause &^ info.TestFlag {
	case info.Actcon:
		if req.Cause&^info.TestFlag == info.Act {
			// activation confirmed
			return nil
		}
	case info.Deactcon:
		if req.Cause&^info.TestFlag == info.Deact {
			// deactivation confirmed
			return nil // OK
		}

	case info.Actcon | info.NegFlag:
		return ErrConNeg

	case info.Actterm:
		return ErrTerm

	// “ASDUs in the control direction with undefined values in the data
	// unit identifier (except the variable structure qualifier) and the
	// information object address are mirrored by the controlled station
	// with bit “P/N := <1> negative confirm” …”
	// — companion standard 101, subsection 7.2.3
	case info.UnkType | info.NegFlag,
		info.UnkCause | info.NegFlag,
		info.UnkAddr | info.NegFlag,
		info.UnkInfo | info.NegFlag:
		return CmdUnk{in.Type, in.Cause}

	}
	return CauseMis{Type: req.Type, Req: req.Cause, Res: in.Cause}
}
