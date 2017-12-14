// Package part5 provides the OSI application layer with asynchronous I/O.
package part5

import "time"
import "github.com/pascaldekloe/part5/info"

// MeasureAttrs are the measured value attributes.
type MeasureAttrs struct {
	Addr info.ObjAddr

	// Quality descriptor info.OK means no remarks.
	QualDesc uint

	// The timestamp is nil when the data is invalid or
	// when the type does not include timing at all.
	Time *time.Time
}

// Single gets called for type M_SP_NA_1, M_SP_TA_1 and M_SP_TB_1.
type Single func(info.ID, info.SinglePoint, MeasureAttrs)

// Double gets called for type M_DP_NA_1, M_DP_TA_1 and M_DP_TB_1.
type Double func(info.ID, info.DoublePoint, MeasureAttrs)

// Step gets called for type M_ST_NA_1, M_ST_TA_1 and M_ST_TB_1.
type Step func(info.ID, info.StepPos, MeasureAttrs)

// Bits gets called for type M_BO_NA_1, M_BO_TA_1 and M_BO_TB_1.
type Bits func(info.ID, uint32, MeasureAttrs)

// Normal gets called for type M_ME_NA_1, M_ME_TA_1, M_ME_TD_1 and M_ME_ND_1.
// The quality descriptor defaults to info.OK for type M_ME_ND_1.
type Normal func(info.ID, info.Normal, MeasureAttrs)

// Scaled gets called for type M_ME_NB_1, M_ME_TB_1 and M_ME_TE_1.
type Scaled func(info.ID, int16, MeasureAttrs)

// Float gets called for type M_ME_NC_1, M_ME_TC_1 and M_ME_TF_1.
type Float func(info.ID, float32, MeasureAttrs)

// ExecAttrs are the command execution attributes.
type ExecAttrs struct {
	Addr info.ObjAddr

	// Use constants for the qualifier of command.
	// See info.Cmd.Qual and info.SetpointCmd.Qual.
	Qual uint

	// The timestamp is nil when the data is invalid or
	// when the type does not include timing at all.
	Time *time.Time

	// Flags whether to execute with "select" protection.
	// See README.md and section 5, subclause 6.8.
	InSelect bool
}

// ExecTerm causes a notification message to be send about the termination
// of a command execution. Note that the outcome, whether success or failure,
// is out of scope and error handling is unspecified.
//
// This notification step is optional depending on the vendor.
type ExecTerm func()

// CmdTerm gets called for notification messages about the termination of a
// command execution. Note that the outcome, whether success or failure, is
// out of scope and error handling is unspecified.
//
// This notification step is optional depending on the vendor. When ignored
// (with a nil CmdTerm) then it must be done consistently so per address or
// risk false hits.
//
// Go routines may safely wait for a call thanks to timeout protection(s).
type CmdTerm func(error)

// SingleCmd gets called for type C_SC_NA_1 and C_SC_TA_1.
// Return ok causes a positive [info.Actcon] or negative [info.NegFlag]
// confirmation message to be send about whether the specified control
// action is about to commence—hence the actual execution should proceed
// in a separate routine and ExecTerm may be called once on completion.
type SingleCmd func(info.ID, info.SinglePoint, ExecAttrs, ExecTerm) (ok bool)

// DoubleCmd gets called for type C_DC_NA_1 and C_DC_TA_1.
// Return ok causes a positive [info.Actcon] or negative [info.NegFlag]
// confirmation message to be send about whether the specified control
// action is about to commence—hence the actual execution should proceed
// in a separate routine and ExecTerm may be called once on completion.
type DoubleCmd func(info.ID, info.DoublePoint, ExecAttrs, ExecTerm) (ok bool)

// RegulCmd gets called for type C_RC_NA_1 and C_RC_TA_1.
// Return ok causes a positive [info.Actcon] or negative [info.NegFlag]
// confirmation message to be send about whether the specified control
// action is about to commence—hence the actual execution should proceed
// in a separate routine and ExecTerm may be called once on completion.
type RegulCmd func(info.ID, info.DoublePoint, ExecAttrs, ExecTerm) (ok bool)

// NormalSetpoint gets called for type C_SE_NA_1 and C_SE_TA_1.
// Return ok causes a positive [info.Actcon] or negative [info.NegFlag]
// confirmation message to be send about whether the specified control
// action is about to commence—hence the actual execution should proceed
// in a separate routine and ExecTerm may be called once on completion.
type NormalSetpoint func(info.ID, info.Normal, ExecAttrs, ExecTerm) (ok bool)

// ScaledSetpoint gets called for type C_SE_NB_1 and C_SE_TB_1.
// Return ok causes a positive [info.Actcon] or negative [info.NegFlag]
// confirmation message to be send about whether the specified control
// action is about to commence—hence the actual execution should proceed
// in a separate routine and ExecTerm may be called once on completion.
type ScaledSetpoint func(info.ID, int16, ExecAttrs, ExecTerm) (ok bool)

// FloatSetpoint gets called for type C_SE_NC_1 and C_SE_TC_1.
// Return ok causes a positive [info.Actcon] or negative [info.NegFlag]
// confirmation message to be send about whether the specified control
// action is about to commence—hence the actual execution should proceed
// in a separate routine and ExecTerm may be called once on completion.
type FloatSetpoint func(info.ID, float32, ExecAttrs, ExecTerm) (ok bool)

// BitsCmd gets called for type C_BO_NA_1 and C_BO_TA_1.
// The timestamp is nil when the data is invalid or
// when the type does not include timing at all.
// Return ok causes a positive [info.Actcon] or negative [info.NegFlag]
// confirmation message to be send about whether the specified control
// action is about to commence—hence the actual execution should proceed
// in a separate routine and ExecTerm may be called once on completion.
type BitsCmd func(info.ID, uint32, info.ObjAddr, *time.Time, ExecTerm) (ok bool)
