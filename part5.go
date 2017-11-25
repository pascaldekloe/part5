// Package part5 provides the OSI application layer with asynchronous I/O.
package part5

import "time"
import "github.com/pascaldekloe/part5/info"

// MeasureAttr annotates measured values.
type MeasureAttr struct {
	// Quality descriptor info.OK means no remarks.
	QualDesc uint

	// The timestamp is nil when the data is invalid or
	// when the type does not include timing at all.
	Tag *time.Time
}

// Single gets called for type M_SP_NA_1, M_SP_TA_1 and M_SP_TB_1.
type Single func(info.ID, info.ObjAddr, info.SinglePoint, MeasureAttr)

// Double gets called for type M_DP_NA_1, M_DP_TA_1 and M_DP_TB_1.
type Double func(info.ID, info.ObjAddr, info.DoublePoint, MeasureAttr)

// Step gets called for type M_ST_NA_1, M_ST_TA_1 and M_ST_TB_1.
type Step func(info.ID, info.ObjAddr, info.StepPos, MeasureAttr)

// Bits gets called for type M_BO_NA_1, M_BO_TA_1 and M_BO_TB_1.
type Bits func(info.ID, info.ObjAddr, uint32, MeasureAttr)

// Normal gets called for type M_ME_NA_1, M_ME_TA_1, M_ME_TD_1 and M_ME_ND_1.
// The quality descriptor defaults to info.OK for type M_ME_ND_1.
type Normal func(info.ID, info.ObjAddr, info.Normal, MeasureAttr)

// Scaled gets called for type M_ME_NB_1, M_ME_TB_1 and M_ME_TE_1.
type Scaled func(info.ID, info.ObjAddr, int16, MeasureAttr)

// Float gets called for type M_ME_NC_1, M_ME_TC_1 and M_ME_TF_1.
type Float func(info.ID, info.ObjAddr, float32, MeasureAttr)

// ExecAttr annotates command execution.
type ExecAttr struct {
	// The qualifier of command.
	Qual uint

	// Can be nil when marked as invalid or when not present for the type.
	Tag *time.Time

	// Flags wheather to execute with "select" protection.
	// See README and section 5, subclause 6.8.
	InSelect bool
}

// SingleCmd gets called for type C_SC_NC_1 and C_SC_TA_1.
type SingleCmd func(info.ID, info.ObjAddr, info.SinglePoint, ExecAttr) (accept bool)

// DoubleCmd gets called for type C_DC_NA_1 and C_DC_TA_1.
type DoubleCmd func(info.ID, info.ObjAddr, info.DoublePoint, ExecAttr) (accept bool)

// RegulCmd gets called for type C_RC_NA_1 and C_RC_TA_1.
type RegulCmd func(info.ID, info.ObjAddr, info.DoublePoint, ExecAttr) (accept bool)

// NormalSetpoint gets called for type C_SE_NA_1 and C_SE_TA_1.
type NormalSetpoint func(info.ID, info.ObjAddr, info.Normal, ExecAttr) (accept bool)

// ScaledSetpoint gets called for type C_SE_NB_1 and C_SE_TB_1.
type ScaledSetpoint func(info.ID, info.ObjAddr, int16, ExecAttr) (accept bool)

// FloatSetpoint gets called for type C_SE_NC_1 and C_SE_TC_1.
type FloatSetpoint func(info.ID, info.ObjAddr, float32, ExecAttr) (accept bool)

// BitsCmd gets called for type C_BO_NA_1 and C_BO_TA_1.
// The timestamp is nil when the data is invalid or
// when the type does not include timing at all.
type BitsCmd func(info.ID, info.ObjAddr, uint32, *time.Time) (accept bool)
