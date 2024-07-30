// Package part5 provides the OSI application layer of IEC 60870-5.
package part5

import (
	"errors"
	"fmt"

	"github.com/pascaldekloe/part5/info"
)

// CmdUnk signals command rejection in control direction due to an unknown type
// identifier, an unknown cause of transmission, an unknown common address or an
// unknown information-object address.
type CmdUnk struct {
	// The cause of transmission from the response has the info.NegFlag set.
	// It may also have the info.TestFlag set.
	info.Cause
}

// Error implements the builtin.error interface.
func (u CmdUnk) Error() string {
	return "part5: command rejected as " + u.Cause.String()
}

// CauseMis signals an incorrect response to either an activation or a
// deactivation request.
type CauseMis struct {
	Type info.TypeID // command identifier
	Req  info.Cause  // from request
	Res  info.Cause  // from response
}

// Error implements the builtin.error interface.
func (mis CauseMis) Error() string {
	return fmt.Sprintf("part5: illegal response %s to %s %s",
		mis.Res, mis.Type, mis.Req)
}

// ErrNotCmd rejects an info.DataUnit based on its type identifier.
var ErrNotCmd = errors.New("part5: ASDU type identifier not in command range 45..69")

func isCommand(t info.TypeID) bool {
	return t-45 < 25
}

// Response alternatives to info.ActCon are propagated as errors for simplicity.
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

	// match cause of transmission, or not
	switch in.Cause &^ info.TestFlag {
	case info.ActCon:
		if req.Cause&^info.TestFlag == info.Act {
			return nil // OK
		}
	case info.DeactCon:
		if req.Cause&^info.TestFlag == info.Deact {
			return nil // OK
		}

	case info.ActCon | info.NegFlag:
		return ErrConNeg
	case info.ActTerm:
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
		return CmdUnk{in.Cause}

	}
	return CauseMis{Type: req.Type, Req: req.Cause, Res: in.Cause}
}
