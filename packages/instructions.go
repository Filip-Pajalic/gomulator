package gameboypackage

type addrMode int

const (
	AM_IMP addrMode = iota + 1
	AM_R_D16
	AM_R_R
	AM_MR_R
	AM_R
	AM_R_D8
	AM_R_MR
	AM_R_HLI
	AM_R_HLD
	AM_HLI_R
	AM_HLD_R
	AM_R_A8
	AM_A8_R
	AM_HL_SPR
	AM_D16
	AM_D8
	AM_D16_R
	AM_MR_D8
	AM_MR
	AM_A16_R
	AM_R_A16
)

type regTypes int

const (
	RT_NONE regTypes = iota + 1
	RT_A
	RT_F
	RT_B
	RT_C
	RT_D
	RT_E
	RT_H
	RT_L
	RT_AF
	RT_BC
	RT_DE
	RT_HL
	RT_SP
	RT_PC
)

type inType int

const (
	IN_NONE inType = iota + 1
	IN_NOP
	IN_LD
	IN_INC
	IN_DEC
	IN_RLCA
	IN_ADD
	IN_RRCA
	IN_STOP
	IN_RLA
	IN_JR
	IN_RRA
	IN_DAA
	IN_CPL
	IN_SCF
	IN_CCF
	IN_HALT
	IN_ADC
	IN_SUB
	IN_SBC
	IN_AND
	IN_XOR
	IN_OR
	IN_CP
	IN_POP
	IN_JP
	IN_PUSH
	IN_RET
	IN_CB
	IN_CALL
	IN_RETI
	IN_LDH
	IN_JPHL
	IN_DI
	IN_EI
	IN_RST
	IN_ERR
	//CB instructions...
	IN_RLC
	IN_RRC
	IN_RL
	IN_RR
	IN_SLA
	IN_SRA
	IN_SWAP
	IN_SRL
	IN_BIT
	IN_RES
	IN_SET
)

type conditionTypes int

const (
	CT_NONE conditionTypes = iota
	CT_NZ
	CT_Z
	CT_NC
	CT_C
)

var instLookup = []string{
	"<NONE",
	"NOP",
	"LD",
	"INC",
	"DEC",
	"RLCA",
	"ADD",
	"RRCA",
	"STOP",
	"RLA",
	"JR",
	"RRA",
	"DAA",
	"CPL",
	"SCF",
	"CCF",
	"HALT",
	"ADC",
	"SUB",
	"SBC",
	"AND",
	"XOR",
	"OR",
	"CP",
	"POP",
	"JP",
	"PUSH",
	"RET",
	"CB",
	"CALL",
	"RETI",
	"LDH",
	"JPHL",
	"DI",
	"EI",
	"RST",
	"IN_ERR",
	"IN_RLC",
	"IN_RRC",
	"IN_RL",
	"IN_RR",
	"IN_SLA",
	"IN_SRA",
	"IN_SWAP",
	"IN_SRL",
	"IN_BIT",
	"IN_RES",
	"IN_SET",
}

type Instruction struct {
	Type      inType
	Mode      addrMode
	Reg1      regTypes
	Reg2      regTypes
	Condition conditionTypes
	Param     byte
}

var instructions [0x100]Instruction

func initInstructions() {
	//DO NOP , adressing mode implied , does nothing
	instructions[0x00].addInstructions(IN_NOP, AM_IMP, 0, 0, 0, 0)
	instructions[0x05].addInstructions(IN_DEC, AM_IMP, RT_C, 0, 0, 0)
	instructions[0xAF].addInstructions(IN_XOR, AM_R, RT_C, RT_A, 0, 0)
	instructions[0xC3].addInstructions(IN_JP, AM_D16, 0, 0, 0, 0)
	instructions[0xF3].addInstructions(IN_DI, 0, 0, 0, 0, 0)
}

func (instruction *Instruction) addInstructions(
	intype inType,
	mode addrMode,
	rega1 regTypes,
	reg2 regTypes,
	condition conditionTypes,
	param byte,
) {
	instruction.Type = intype
	instruction.Mode = mode
	instruction.Reg1 = rega1
	instruction.Reg2 = reg2
	instruction.Condition = condition
	instruction.Param = param
}

func instructionByOpcode(opcode byte) (instruction *Instruction) {
	return &instructions[opcode]
}

func getInstructionName(t inType) *[]byte {
	b := []byte(instLookup[t])
	return &b
}
