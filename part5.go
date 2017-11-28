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

	// Flags wheather to execute with "select" protection.
	// See README.md and section 5, subclause 6.8.
	InSelect bool
}

// CmdTerm gets called when the respective command terminates.
// This step is optional. See info.Actterm.
type CmdTerm func()

// SingleCmd gets called for type C_SC_NC_1 and C_SC_TA_1.
type SingleCmd func(info.ID, info.SinglePoint, ExecAttrs, CmdTerm) (accept bool)

// DoubleCmd gets called for type C_DC_NA_1 and C_DC_TA_1.
type DoubleCmd func(info.ID, info.DoublePoint, ExecAttrs, CmdTerm) (accept bool)

// RegulCmd gets called for type C_RC_NA_1 and C_RC_TA_1.
type RegulCmd func(info.ID, info.DoublePoint, ExecAttrs, CmdTerm) (accept bool)

// NormalSetpoint gets called for type C_SE_NA_1 and C_SE_TA_1.
type NormalSetpoint func(info.ID, info.Normal, ExecAttrs, CmdTerm) (accept bool)

// ScaledSetpoint gets called for type C_SE_NB_1 and C_SE_TB_1.
type ScaledSetpoint func(info.ID, int16, ExecAttrs, CmdTerm) (accept bool)

// FloatSetpoint gets called for type C_SE_NC_1 and C_SE_TC_1.
type FloatSetpoint func(info.ID, float32, ExecAttrs, CmdTerm) (accept bool)

// BitsCmd gets called for type C_BO_NA_1 and C_BO_TA_1.
// The timestamp is nil when the data is invalid or
// when the type does not include timing at all.
type BitsCmd func(info.ID, uint32, info.ObjAddr, *time.Time, CmdTerm) (accept bool)
