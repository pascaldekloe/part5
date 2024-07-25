package info

import "errors"

//go:generate stringer -type TypeID -output code_string.go code.go

// ErrTypeZero denies the zero value as type identification. Use is explicitly
// prohibited in subsection 7.2.1.1 from companion standard 101.
var errTypeZero = errors.New("part5: type identification <0> is not used")

// TypeID classifies ASDU with a code.
type TypeID uint8

// PrivateTypeFlag takes half of the TypeID in the private range.
// Codes [128..135] are reserved for routing messages, and codes
// [136..255] is for “special use”.
const PrivateTypeFlag TypeID = 128

// The compatible codes are defined in table 8 from companion standard 101.
// More compatible codes are defined in table 1 from companion standard 104.
// The security extensions are defined in table 2 from section 7.
// Prefix M_ is for monitored information, C_ is for control information,
// P_ is for parameter information, F_ is for file transfer, and S_ is for
// security.
const (
	_         TypeID = iota // 0: not used
	M_SP_NA_1               // single-point information
	M_SP_TA_1               // single-point information with time tag
	M_DP_NA_1               // double-point information
	M_DP_TA_1               // double-point information with time tag
	M_ST_NA_1               // step position information
	M_ST_TA_1               // step position information with time tag
	M_BO_NA_1               // bitstring of 32 bit
	M_BO_TA_1               // bitstring of 32 bit with time tag
	M_ME_NA_1               // measured value, normalized value

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
	_         // 71: reserved for further compatible definitions
	_         // 72: reserved for further compatible definitions
	_         // 73: reserved for further compatible definitions
	_         // 74: reserved for further compatible definitions
	_         // 75: reserved for further compatible definitions
	_         // 76: reserved for further compatible definitions
	_         // 77: reserved for further compatible definitions
	_         // 78: reserved for further compatible definitions
	_         // 79: reserved for further compatible definitions

	_         // 80: reserved for further compatible definitions
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

// String returns the label.
func (ID TypeID) String() string { return typeIDLabels[ID] }

var typeIDLabels = [256]string{
	"notused<0>",
	"M_SP_NA_1",
	"M_SP_TA_1",
	"M_DP_NA_1",
	"M_DP_TA_1",
	"M_ST_NA_1",
	"M_ST_TA_1",
	"M_BO_NA_1",
	"M_BO_TA_1",
	"M_ME_NA_1",

	"M_ME_TA_1",
	"M_ME_NB_1",
	"M_ME_TB_1",
	"M_ME_NC_1",
	"M_ME_TC_1",
	"M_IT_NA_1",
	"M_IT_TA_1",
	"M_EP_TA_1",
	"M_EP_TB_1",
	"M_EP_TC_1",

	"M_PS_NA_1",
	"M_ME_ND_1",
	"reserve<22>",
	"reserve<23>",
	"reserve<24>",
	"reserve<25>",
	"reserve<26>",
	"reserve<27>",
	"reserve<28>",
	"reserve<29>",

	"M_SP_TB_1",
	"M_DP_TB_1",
	"M_ST_TB_1",
	"M_BO_TB_1",
	"M_ME_TD_1",
	"M_ME_TE_1",
	"M_ME_TF_1",
	"M_IT_TB_1",
	"M_EP_TD_1",
	"M_EP_TE_1",

	"M_EP_TF_1",
	"S_IT_TC_1",
	"reserve<42>",
	"reserve<43>",
	"reserve<44>",
	"C_SC_NA_1",
	"C_DC_NA_1",
	"C_RC_NA_1",
	"C_SE_NA_1",
	"C_SE_NB_1",

	"C_SE_NC_1",
	"C_BO_NA_1",
	"reserve<52>",
	"reserve<53>",
	"reserve<54>",
	"reserve<55>",
	"reserve<56>",
	"reserve<57>",
	"C_SC_TA_1",
	"C_DC_TA_1",

	"C_RC_TA_1",
	"C_SE_TA_1",
	"C_SE_TB_1",
	"C_SE_TC_1",
	"C_BO_TA_1",
	"reserve<65>",
	"reserve<66>",
	"reserve<67>",
	"reserve<68>",
	"reserve<69>",

	"M_EI_NA_1",
	"reserve<71>",
	"reserve<72>",
	"reserve<73>",
	"reserve<74>",
	"reserve<75>",
	"reserve<76>",
	"reserve<77>",
	"reserve<78>",
	"reserve<79>",

	"reserve<80>",
	"S_CH_NA_1",
	"S_RP_NA_1",
	"S_AR_NA_1",
	"S_KR_NA_1",
	"S_KS_NA_1",
	"S_KC_NA_1",
	"S_ER_NA_1",
	"reserve<86>",
	"reserve<89>",

	"S_US_NA_1",
	"S_UQ_NA_1",
	"S_UR_NA_1",
	"S_UK_NA_1",
	"S_UA_NA_1",
	"S_UC_NA_1",
	"reserve<96>",
	"reserve<97>",
	"reserve<98>",
	"reserve<99>",

	"C_IC_NA_1",
	"C_CI_NA_1",
	"C_RD_NA_1",
	"C_CS_NA_1",
	"C_TS_NA_1",
	"C_RP_NA_1",
	"C_CD_NA_1",
	"C_TS_TA_1",
	"reserve<108>",
	"reserve<109>",

	"P_ME_NA_1",
	"P_ME_NB_1",
	"P_ME_NC_1",
	"P_AC_NA_1",
	"reserve<114>",
	"reserve<115>",
	"reserve<116>",
	"reserve<117>",
	"reserve<118>",
	"reserve<119>",

	"F_FR_NA_1",
	"F_SR_NA_1",
	"F_SC_NA_1",
	"F_LS_NA_1",
	"F_AF_NA_1",
	"F_SG_NA_1",
	"F_DR_TA_1",

	"F_SC_NB_1",
	"routing<128>",
	"routing<129>",

	"routing<130>",
	"routing<131>",
	"routing<132>",
	"routing<133>",
	"routing<134>",
	"routing<135>",
	"special<136>",
	"special<137>",
	"special<138>",
	"special<139>",

	"special<140>",
	"special<141>",
	"special<142>",
	"special<143>",
	"special<144>",
	"special<145>",
	"special<146>",
	"special<147>",
	"special<148>",
	"special<149>",

	"special<150>",
	"special<151>",
	"special<152>",
	"special<153>",
	"special<154>",
	"special<155>",
	"special<156>",
	"special<157>",
	"special<158>",
	"special<159>",

	"special<160>",
	"special<161>",
	"special<162>",
	"special<163>",
	"special<164>",
	"special<165>",
	"special<166>",
	"special<167>",
	"special<168>",
	"special<169>",

	"special<170>",
	"special<171>",
	"special<172>",
	"special<173>",
	"special<174>",
	"special<175>",
	"special<176>",
	"special<177>",
	"special<178>",
	"special<179>",

	"special<180>",
	"special<181>",
	"special<182>",
	"special<183>",
	"special<184>",
	"special<185>",
	"special<186>",
	"special<187>",
	"special<188>",
	"special<189>",

	"special<190>",
	"special<191>",
	"special<192>",
	"special<193>",
	"special<194>",
	"special<195>",
	"special<196>",
	"special<197>",
	"special<198>",
	"special<199>",

	"special<200>",
	"special<201>",
	"special<202>",
	"special<203>",
	"special<204>",
	"special<205>",
	"special<206>",
	"special<207>",
	"special<208>",
	"special<209>",

	"special<210>",
	"special<211>",
	"special<212>",
	"special<213>",
	"special<214>",
	"special<215>",
	"special<216>",
	"special<217>",
	"special<218>",
	"special<219>",

	"special<220>",
	"special<221>",
	"special<222>",
	"special<223>",
	"special<224>",
	"special<225>",
	"special<226>",
	"special<227>",
	"special<228>",
	"special<229>",

	"special<230>",
	"special<231>",
	"special<232>",
	"special<233>",
	"special<234>",
	"special<235>",
	"special<236>",
	"special<237>",
	"special<238>",
	"special<239>",

	"special<240>",
	"special<241>",
	"special<242>",
	"special<243>",
	"special<244>",
	"special<245>",
	"special<246>",
	"special<247>",
	"special<248>",
	"special<249>",

	"special<250>",
	"special<251>",
	"special<252>",
	"special<253>",
	"special<254>",
	"special<255>",
}

// Cause of transmission includes the NegFlag and TestFlag.
type Cause uint8

// ErrCauseZero denies Cause zero.

// ErrCauseZero denies the zero value as a cause code. Use is explicitly
// prohibited in table 14 from companion standard 101.
var errCauseZero = errors.New("part5: cause of transmission <0> is not used")

// The cause of transmission codes are defined in table 14 from companion
// standard 101. Auth, Seskey and Usrkey are defined in table 1 from section 7,
// “Additional cause of transmission”.
const (
	_        Cause = iota // 0: not defined
	Percyc                // periodic, cyclic
	Back                  // background scan
	Spont                 // spontaneous
	Init                  // initialized
	Req                   // request or requested
	Act                   // activation
	Actcon                // activation confirmation
	Deact                 // deactivation
	Deactcon              // deactivation confirmation

	Actterm // activation termination
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
	Inro1   // interrogated by group 1 interrogation
	Inro2   // interrogated by group 2 interrogation
	Inro3   // interrogated by group 3 interrogation
	Inro4   // interrogated by group 4 interrogation
	Inro5   // interrogated by group 5 interrogation
	Inro6   // interrogated by group 6 interrogation
	Inro7   // interrogated by group 7 interrogation
	Inro8   // interrogated by group 8 interrogation
	Inro9   // interrogated by group 9 interrogation

	Inro10   // interrogated by group 10 interrogation
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

	_ // 50: for special use (private range)
	_ // 51: for special use (private range)
	_ // 52: for special use (private range)
	_ // 53: for special use (private range)
	_ // 54: for special use (private range)
	_ // 55: for special use (private range)
	_ // 55: for special use (private range)
	_ // 57: for special use (private range)
	_ // 58: for special use (private range)
	_ // 59: for special use (private range)

	_ // 60: for special use (private range)
	_ // 61: for special use (private range)
	_ // 62: for special use (private range)
	_ // 63: for special use (private range)

	// The P/N flag negates confirmation when set.
	// In case of irrelevance, the bit should be zero.
	NegFlag Cause = 0x40

	// The T flag classifies the unit for testing when set.
	TestFlag Cause = 0x80
)

// String returns the label plus ",neg" when NegFlag plus ",test" when TestFlag.
func (c Cause) String() string {
	s := causeLabels[c&0x3f]
	switch c & (NegFlag | TestFlag) {
	case NegFlag:
		s += ",neg"
	case TestFlag:
		s += ",test"
	case NegFlag | TestFlag:
		s += ",neg,test"
	}
	return s
}

var causeLabels = [64]string{
	"notdef",
	"per/cyc",
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
