package info

import (
	"errors"
	"fmt"
)

//go:generate stringer -type TypeID -output code_string.go code.go

// TypeID classifies ASDU with a code.
type TypeID uint8

// PrivateTypeFlag takes half of the TypeID in the private range.
// Codes [128..135] are reserved for routing messages, and codes
// [136..255] is for “special use”.
const PrivateTypeFlag = 128

// ErrTypeZero denies zero as a type identification.
var errTypeZero = errors.New("part5: type identification 0 is not used")

// The compatible codes are defined in table 8 from companion standard 101.
// More compatible codes are defined in table 1 from companion standard 104.
// The security extension are defined in table 2 from section 7.
// Prefix M_ is for monitored information, C_ is for control information,
// P_ is for parameter information, F_ is for file transfer, and S_ is for
// security.
const (
	_ TypeID = iota // 0: not used

	M_SP_NA_1 // single-point information
	M_SP_TA_1 // single-point information with time tag
	M_DP_NA_1 // double-point information
	M_DP_TA_1 // double-point information with time tag
	M_ST_NA_1 // step position information
	M_ST_TA_1 // step position information with time tag
	M_BO_NA_1 // bitstring of 32 bit
	M_BO_TA_1 // bitstring of 32 bit with time tag
	M_ME_NA_1 // measured value, normalized value
	M_ME_TA_1 // measured value, normalized value with time tag

	M_ME_NB_1 // measured value, scaled value
	M_ME_TB_1 // measured value, scaled value with time tag
	M_ME_NC_1 // measured value, short floating point number
	M_ME_TC_1 // measured value, short floating point number with time tag
	M_IT_NA_1 // integrated totals
	M_IT_TA_1 // integrated totals with time tag
	M_EP_TA_1 // event of protection equipment with time tag
	M_EP_TB_1 // packed start events of protection equipment with time tag
	M_EP_TC_1 // packed output circuit information of protection equipment with time tag
	M_PS_NA_1 // packed single-point information with status change detection

	M_ME_ND_1 // measured value, normalized value without quality descriptor
	_         // 22: reserved for further compatible definitions
	_         // 23: reserved for further compatible definitions
	_         // 24: reserved for further compatible definitions
	_         // 25: reserved for further compatible definitions
	_         // 26: reserved for further compatible definitions
	_         // 27: reserved for further compatible definitions
	_         // 28: reserved for further compatible definitions
	_         // 29: reserved for further compatible definitions
	M_SP_TB_1 // single-point information with time tag CP56Time2a

	M_DP_TB_1 // double-point information with time tag CP56Time2a
	M_ST_TB_1 // step position information with time tag CP56Time2a
	M_BO_TB_1 // bitstring of 32 bits with time tag CP56Time2a
	M_ME_TD_1 // measured value, normalized value with time tag CP56Time2a
	M_ME_TE_1 // measured value, scaled value with time tag CP56Time2a
	M_ME_TF_1 // measured value, short floating point number with time tag CP56Time2a
	M_IT_TB_1 // integrated totals with time tag CP56Time2a
	M_EP_TD_1 // event of protection equipment with time tag CP56Time2a
	M_EP_TE_1 // packed start events of protection equipment with time tag CP56Time2a
	M_EP_TF_1 // packed output circuit information of protection equipment with time tag CP56Time2a

	S_IT_TC_1 // integrated totals containing security statistics with time tag CP56Time2a
	_         // 42: reserved for further compatible definitions
	_         // 43: reserved for further compatible definitions
	_         // 44: reserved for further compatible definitions
	C_SC_NA_1 // single command
	C_DC_NA_1 // double command
	C_RC_NA_1 // regulating step command
	C_SE_NA_1 // set-point command, normalized value
	C_SE_NB_1 // set-point command, scaled value
	C_SE_NC_1 // set-point command, short floating point number

	C_BO_NA_1 // bitstring of 32 bits
	_         // 52: reserved for further compatible definitions
	_         // 53: reserved for further compatible definitions
	_         // 54: reserved for further compatible definitions
	_         // 55: reserved for further compatible definitions
	_         // 56: reserved for further compatible definitions
	_         // 57: reserved for further compatible definitions
	C_SC_TA_1 // single command with time tag CP56Time2a
	C_DC_TA_1 // double command with time tag CP56Time2a
	C_RC_TA_1 // regulating step command with time tag CP56Time2a

	C_SE_TA_1 // set-point command with time tag CP56Time2a, normalized value
	C_SE_TB_1 // set-point command with time tag CP56Time2a, scaled value
	C_SE_TC_1 // set-point command with time tag CP56Time2a, short floating point number
	C_BO_TA_1 // bitstring of 32-bit with time tag CP56Time2a
	_         // 65: reserved for further compatible definitions
	_         // 66: reserved for further compatible definitions
	_         // 67: reserved for further compatible definitions
	_         // 68: reserved for further compatible definitions
	_         // 69: reserved for further compatible definitions
	M_EI_NA_1 // end of initialization

	_ // 71: reserved for further compatible definitions
	_ // 72: reserved for further compatible definitions
	_ // 73: reserved for further compatible definitions
	_ // 74: reserved for further compatible definitions
	_ // 75: reserved for further compatible definitions
	_ // 76: reserved for further compatible definitions
	_ // 77: reserved for further compatible definitions
	_ // 78: reserved for further compatible definitions
	_ // 79: reserved for further compatible definitions
	_ // 80: reserved for further compatible definitions

	S_CH_NA_1 // authentication challenge
	S_RP_NA_1 // authentication reply
	S_AR_NA_1 // aggressive mode authentication request
	S_KR_NA_1 // session key status request
	S_KS_NA_1 // session key status
	S_KC_NA_1 // session key change
	S_ER_NA_1 // authentication error
	_         // 88: reserved for further compatible definitions
	_         // 89: reserved for further compatible definitions
	S_US_NA_1 // user status change

	S_UQ_NA_1 // update key change request
	S_UR_NA_1 // update key change reply
	S_UK_NA_1 // update key change symetric
	S_UA_NA_1 // update key change asymetric
	S_UC_NA_1 // update key change confirmation
	_         // 96: reserved for further compatible definitions
	_         // 97: reserved for further compatible definitions
	_         // 98: reserved for further compatible definitions
	_         // 99: reserved for further compatible definitions
	C_IC_NA_1 // interrogation command

	C_CI_NA_1 // counter interrogation command
	C_RD_NA_1 // read command
	C_CS_NA_1 // clock synchronization command
	C_TS_NA_1 // test command
	C_RP_NA_1 // reset process command
	C_CD_NA_1 // delay acquisition command
	C_TS_TA_1 // test command with time tag CP56Time2a
	_         // 108: reserved for further compatible definitions
	_         // 109: reserved for further compatible definitions
	P_ME_NA_1 // parameter of measured value, normalized value

	P_ME_NB_1 // parameter of measured value, scaled value
	P_ME_NC_1 // parameter of measured value, short floating point number
	P_AC_NA_1 // parameter activation
	_         // 114: reserved for further compatible definitions
	_         // 115: reserved for further compatible definitions
	_         // 116: reserved for further compatible definitions
	_         // 117: reserved for further compatible definitions
	_         // 118: reserved for further compatible definitions
	_         // 119: reserved for further compatible definitions
	F_FR_NA_1 // file ready

	F_SR_NA_1 // section ready
	F_SC_NA_1 // call directory, select file, call file, call section
	F_LS_NA_1 // last section, last segment
	F_AF_NA_1 // ack file, ack section
	F_SG_NA_1 // segment
	F_DR_TA_1 // directory
	F_SC_NB_1 // QueryLog - request archive file (section 104)
)

// ObjSize maps the type identification (TypeID) to the serial octet size.
// Type extensions must register here.
var ObjSize [256]int = builtins

// don't polute public API documentation with this content
var builtins = [256]int{
	M_SP_NA_1: 1,
	M_SP_TA_1: 4,
	M_DP_NA_1: 1,
	M_DP_TA_1: 4,
	M_ST_NA_1: 2,
	M_ST_TA_1: 5,
	M_BO_NA_1: 5,
	M_BO_TA_1: 8,
	M_ME_NA_1: 3,
	M_ME_TA_1: 6,
	M_ME_NB_1: 3,
	M_ME_TB_1: 6,
	M_ME_NC_1: 5,
	M_ME_TC_1: 8,
	M_IT_NA_1: 5,
	M_IT_TA_1: 8,
	M_EP_TA_1: 6,
	M_EP_TB_1: 7,
	M_EP_TC_1: 7,
	M_PS_NA_1: 5,
	M_ME_ND_1: 2,

	M_SP_TB_1: 8,
	M_DP_TB_1: 8,
	M_ST_TB_1: 9,
	M_BO_TB_1: 12,
	M_ME_TD_1: 10,
	M_ME_TE_1: 10,
	M_ME_TF_1: 12,
	M_IT_TB_1: 12,
	M_EP_TD_1: 11,
	M_EP_TE_1: 11,
	M_EP_TF_1: 11,

	C_SC_NA_1: 1,
	C_DC_NA_1: 1,
	C_RC_NA_1: 1,
	C_SE_NA_1: 3,
	C_SE_NB_1: 3,
	C_SE_NC_1: 5,
	C_BO_NA_1: 4,

	M_EI_NA_1: 1,

	C_IC_NA_1: 1,
	C_CI_NA_1: 1,
	C_RD_NA_1: 0,
	C_CS_NA_1: 7,
	C_TS_NA_1: 2,
	C_RP_NA_1: 1,
	C_CD_NA_1: 2,

	P_ME_NA_1: 3,
	P_ME_NB_1: 3,
	P_ME_NC_1: 5,
	P_AC_NA_1: 1,

	F_FR_NA_1: 6,
	F_SR_NA_1: 7,
	F_SC_NA_1: 4,
	F_LS_NA_1: 5,
	F_AF_NA_1: 4,
	// F_SG_NA_1: 4 + variable,
	F_DR_TA_1: 13,
}

// ErrCauseZero denies zero as a cause of transmission.
var errCauseZero = errors.New("part5: cause of transmission 0 is not used")

// The cause of transmission codes are defined in table 14 from companion
// standard 101. Auth, Seskey and Usrkey are defined in section 7, table 1,
// “Additional cause of transmission”.
const (
	_ = iota // 0: not used

	Percyc   // periodic, cyclic
	Back     // background scan
	Spont    // spontaneous
	Init     // initialized
	Req      // request or requested
	Act      // activation
	Actcon   // activation confirmation
	Deact    // deactivation
	Deactcon // deactivation confirmation
	Actterm  // activation termination

	Retrem  // return information caused by a remote command
	Retloc  // return information caused by a local command
	File    // file transfer
	Auth    // authentication
	Seskey  // maintenance of authentication session key
	Usrkey  // maintenance of user role and update key
	_       // 17: reserved for further compatible definitions
	_       // 18: reserved for further compatible definitions
	_       // 19: reserved for further compatible definitions
	Inrogen // interrogated by station interrogation

	Inro1  // interrogated by group 1 interrogation
	Inro2  // interrogated by group 2 interrogation
	Inro3  // interrogated by group 3 interrogation
	Inro4  // interrogated by group 4 interrogation
	Inro5  // interrogated by group 5 interrogation
	Inro6  // interrogated by group 6 interrogation
	Inro7  // interrogated by group 7 interrogation
	Inro8  // interrogated by group 8 interrogation
	Inro9  // interrogated by group 9 interrogation
	Inro10 // interrogated by group 10 interrogation

	Inro11   // interrogated by group 11 interrogation
	Inro12   // interrogated by group 12 interrogation
	Inro13   // interrogated by group 13 interrogation
	Inro14   // interrogated by group 14 interrogation
	Inro15   // interrogated by group 15 interrogation
	Inro16   // interrogated by group 16 interrogation
	Reqcogen // requested by general counter request
	Reqco1   // requested by group 1 counter request
	Reqco2   // requested by group 2 counter request
	Reqco3   // requested by group 3 counter request

	Reqco4   // requested by group 4 counter request
	_        // 41: reserved for further compatible definitions
	_        // 42: reserved for further compatible definitions
	UnkType  // unknown type identification
	UnkCause // unknown cause of transmission
	UnkAddr  // unknown common address of ASDU
	UnkInfo  // unknown information object address
	_        // 48: for special use (private range)
	_        // 49: for special use (private range)
	_        // 50: for special use (private range)

	_ // 51: for special use (private range)
	_ // 52: for special use (private range)
	_ // 53: for special use (private range)
	_ // 54: for special use (private range)
	_ // 55: for special use (private range)
	_ // 55: for special use (private range)
	_ // 57: for special use (private range)
	_ // 58: for special use (private range)
	_ // 59: for special use (private range)

	_ // 61: for special use (private range)
	_ // 62: for special use (private range)
	_ // 63: for special use (private range)

	// NegFlag indicates the negative (or positive) confirmation
	// of activation requested by the primary application function.
	// In case of irrelevance, the bit is zero.
	NegFlag = 0x40

	// TestFlag marks the cause of transmission for testing.
	TestFlag = 0x80
)

// Cause of transmission is a code to justify emission. The 2-octet extension
// includes an originator address, which is optional per system.
// See companion standard 101, subclause 7.2.3.
type (
	// The Cause size is fixed per system.
	Cause interface {
		Cause8 | Cause16

		// Value gets the cause of transmission together
		// with the NegFlag and TestFlag.
		Value() (codeAndFlags uint8)
	}

	Cause8  [1]uint8
	Cause16 [2]uint8
)

// Value implements the Cause interface.
func (cause Cause8) Value() uint8 { return cause[0] }

// Value implements the Cause interface.
func (cause Cause16) Value() uint8 { return cause[0] }

// OrigAddr returns either the originator address in range [1..255], or it
// returns zero for default.
func (cause Cause16) OrigAddr() uint8 { return cause[1] }

// String returns a compact description.
func (cause Cause8) String() string {
	return causeLabel(cause[0])
}

// String returns a compact description.
func (cause Cause16) String() string {
	return fmt.Sprintf("%s #%d", causeLabel(cause[0]), cause[1])
}

func causeLabel(cause uint8) string {
	s := causeLabels[cause&^(NegFlag|TestFlag)]

	switch cause & (NegFlag | TestFlag) {
	case NegFlag:
		s += ",neg"
	case TestFlag:
		s += ",test"
	case NegFlag | TestFlag:
		s += ",neg,test"
	}

	return s
}

// Can't use stringer: need lowercase, init reserved name & flags.
var causeLabels = [64]string{
	"nocause",
	"percyc",
	"back",
	"spont",
	"init",
	"req",
	"act",
	"actcon",
	"deact",
	"deactcon",
	"actterm",
	"retrem",
	"retloc",
	"file",
	"auth",
	"seskey",
	"usrkey",
	"reserved17",
	"reserved18",
	"reserved19",
	"inrogen",
	"inro1",
	"inro2",
	"inro3",
	"inro4",
	"inro5",
	"inro6",
	"inro7",
	"inro8",
	"inro9",
	"inro10",
	"inro11",
	"inro12",
	"inro13",
	"inro14",
	"inro15",
	"inro16",
	"reqcogen",
	"reqco1",
	"reqco2",
	"reqco3",
	"reqco4",
	"reserved42",
	"reserved43",
	"unktype",
	"unkcause",
	"unkaddr",
	"unkinfo",
	"special48",
	"special49",
	"special50",
	"special51",
	"special52",
	"special53",
	"special54",
	"special55",
	"special56",
	"special57",
	"special58",
	"special59",
	"special60",
	"special61",
	"special62",
	"special63",
}
