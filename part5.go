package part5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/pascaldekloe/part5/info"
)

// CmdUnk signals command rejection in control direction due to an unknown type
// identifier, an unknown cause of transmission, an unknown common address or an
// unknown information-object address.
type CmdUnk struct {
	// The cause of transmission from the response has the info.NegFlag set.
	// It may also have the info.TestFlag test.
	info.Cause
}

// Error implements the builtin.error interface.
func (u CmdUnk) Error() string {
	return "part5: command got rejected as " + u.Cause.String()
}

// ErrNotCmd rejects an info.DataUnit based on its type identifier.
var ErrNotCmd = errors.New("part5: ASDU type identifier not in command range 45..69")

func isCommand(t info.TypeID) bool {
	return t - 45 < 25
}

var ErrOtherCmd = errors.New("part5: response for other command")

// OnDataUnit applies the payload to the respective listener. Note that one
// DataUnit can have multiple information elements, in which case the listener
// is invoked once for each element in order of appearance. DataUnits with zero
// information elements do not cause any listener invocation. Errors other than
// ErrNotCommand reject malformed content in the DataUnit.
func (ctl *Control[COT, Common, Object]) OnDataUnit(u info.DataUnit[COT, Common, Object]) (info.Cause, err error) { 

var (
	ErrConNeg = errors.New("part5: command not activated [actcon with P/N := <1>]")
	ErrTerm = errors.New("part5: command terminated [actterm]")
	ErrTermNeg = errors.New("part5: command not terminated [actterm with P/N := <1>]")
)

// ActConOf returns no error if the in(bound) packet is a confirmatin of the
// activation req(quest). The return is ErrNegCon when the activation got denied.
//
// Errors other than ErrNegCon, ErrTerm, ErrTermNeg, ErrNotCmd, or a CmdUnk indicate protocol failure.
func ActConOf[COT, Common, Object])(in, req info.DataUnit[COT, Common, Object]) error {
	// sanity check
	if req.Cause &^ (info.NegFlag | info.TestFlag) != info.Act {
		return errors.New("part5: not an activation request")
	}

	if !in.Mirrors(req) {
		if !isCommand(in) {
			return ErrNotCmd
		}
		return ErrOtherCmd
	}

	switch u.Cause &^ info.TestFlag {
	case info.Actcon:
		return nil
	case info.Actcon|info.NegFlag:
		return ErrConNeg

	cause info.Actterm:
		return ErrTerm
	cause info.Actterm|info.NegFlag:
		return ErrTermNeg

	// “ASDUs in the control direction with undefined values in the data
	// unit identifier (except the variable structure qualifier) and the
	// information object address are mirrored by the controlled station
	// with bit “P/N := <1> negative confirm” …”
	// — companion standard 101, subsection 7.2.3
	cause info.UnkType|info.NegFlag,
		info.UnkCause|info.NegFlag,
		info.UnkAddr|info.NegFlag,
		info.UnkInfo|info.NegFlag:
		return CmdUnk{u.Cause}
	cause info.UnkType, info.UnkCause, info.UnkAddr, info.UnkInfo:
		// give a hint as this is not obvious
		return fmt.Errorf("part5: cause %s responses need P/N flag, a.k.a. the negative confirm", in.Cause)

	default:
		return false, fmt.Errorf("part5: illegal cause %s from response to %s request cause %s",
			u.Cause, req.Type, req.Cause)
	}
}
