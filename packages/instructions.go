package gameboypackage

/* Different instructionmodes needed
D8  means immediate 8 bit data
D16 means immediate 16 bit data
A8  means 8 bit unsigned data, which are added to $FF00 in certain instructions (replacement for missing IN and OUT instructions)
A16 means 16 bit address
R8  means 8 bit signed data, which are added to program counter

*/
type addrMode byte

const (
	AM_IMP addrMode = iota
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

type regTypes byte

const (
	RT_NONE regTypes = iota
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

type InType byte

const (
	IN_NONE InType = iota
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

type conditionTypes byte

const (
	CT_NONE conditionTypes = iota
	CT_NZ
	CT_Z
	CT_NC
	CT_C
)

// Better with map parhaps, stupid to have two
var instLookup = []string{
	"<NONE>",
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
	Type      InType
	Mode      addrMode
	Reg1      regTypes
	Reg2      regTypes
	Condition conditionTypes
	Param     byte
}

type InstPointer struct {
	Type      *InType
	Mode      *addrMode
	Reg1      *regTypes
	Reg2      *regTypes
	Condition *conditionTypes
	Param     *byte
}

var inst [0x100]Instruction

func initInstructions() {
	//DO NOP , adressing mode implied , does nothing
	/*instructions[0x00].addInstructions(IN_NOP, AM_IMP, nil, nil, 0, 0)
	instructions[0x05].addInstructions(IN_DEC, AM_IMP, RT_C, 0, 0, 0)
	instructions[0xAF].addInstructions(IN_XOR, AM_R, RT_C, RT_A, 0, 0)
	instructions[0xC3].addInstructions(IN_JP, AM_D16, 0, 0, 0, 0)
	instructions[0xF3].addInstructions(IN_DI, 0, 0, 0, 0, 0)

	*/
	//0x1X
	//Helper functions for const pointer conversion
	in := func(in InType) *InType { return &in }
	ad := func(ad addrMode) *addrMode { return &ad }
	re := func(re regTypes) *regTypes { return &re }
	co := func(co conditionTypes) *conditionTypes { return &co }
	//pa := func(pa byte) *byte { return &pa }

	inst[0x00].addInst(InstPointer{Type: in(IN_NOP), Mode: ad(AM_IMP)})
	inst[0x01].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_D16), Reg1: re(RT_BC)})

	inst[0x02].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_MR_R), Reg1: re(RT_BC), Reg2: re(RT_A)})

	inst[0x05].addInst(InstPointer{Type: in(IN_DEC), Mode: ad(AM_R), Reg1: re(RT_B)})
	inst[0x06].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_D8), Reg1: re(RT_B)})

	inst[0x08].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_A16_R), Reg1: re(RT_NONE), Reg2: re(RT_SP)})

	inst[0x0A].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_MR), Reg1: re(RT_A), Reg2: re(RT_BC)})

	inst[0x0E].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_D8), Reg1: re(RT_C)})
	//0x1X
	inst[0x11].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_D16), Reg1: re(RT_DE)})
	inst[0x12].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_MR_R), Reg1: re(RT_DE), Reg2: re(RT_A)})
	inst[0x15].addInst(InstPointer{Type: in(IN_DEC), Mode: ad(AM_R), Reg1: re(RT_D)})
	inst[0x16].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_D8), Reg1: re(RT_D)})
	inst[0x18].addInst(InstPointer{Type: in(IN_JR), Mode: ad(AM_D8)})
	inst[0x1A].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_MR), Reg1: re(RT_A), Reg2: re(RT_DE)})
	inst[0x1E].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_D8), Reg1: re(RT_E)})

	//0x2X
	inst[0x20].addInst(InstPointer{Type: in(IN_JR), Mode: ad(AM_D8), Condition: co(CT_NZ)})
	inst[0x21].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_D16), Reg1: re(RT_HL)})
	inst[0x22].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_HLI_R), Reg1: re(RT_HL), Reg2: re(RT_A)})
	inst[0x25].addInst(InstPointer{Type: in(IN_DEC), Mode: ad(AM_R), Reg1: re(RT_H)})
	inst[0x26].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_D8), Reg1: re(RT_H)})
	inst[0x28].addInst(InstPointer{Type: in(IN_JR), Mode: ad(AM_D8), Condition: co(CT_Z)})
	inst[0x2A].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_HLI), Reg1: re(RT_A), Reg2: re(RT_HL)})
	inst[0x2E].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_D8), Reg1: re(RT_L)})

	//0x3X
	inst[0x30].addInst(InstPointer{Type: in(IN_JR), Mode: ad(AM_D8), Condition: co(CT_NC)})
	inst[0x31].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_D16), Reg1: re(RT_SP)})
	inst[0x32].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_HLD_R), Reg1: re(RT_HL), Reg2: re(RT_A)})
	inst[0x35].addInst(InstPointer{Type: in(IN_DEC), Mode: ad(AM_R), Reg1: re(RT_HL)})
	inst[0x36].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_MR_D8), Reg1: re(RT_HL)})
	inst[0x38].addInst(InstPointer{Type: in(IN_JR), Mode: ad(AM_D8), Condition: co(CT_C)})
	inst[0x3A].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_HLD), Reg1: re(RT_A), Reg2: re(RT_HL)})
	inst[0x3E].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_D8), Reg1: re(RT_A)})

	//0x4X
	inst[0x40].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_B), Reg2: re(RT_B)})
	inst[0x41].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_B), Reg2: re(RT_C)})
	inst[0x42].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_B), Reg2: re(RT_D)})
	inst[0x43].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_B), Reg2: re(RT_E)})
	inst[0x44].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_B), Reg2: re(RT_H)})
	inst[0x45].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_B), Reg2: re(RT_L)})
	inst[0x46].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_MR), Reg1: re(RT_B), Reg2: re(RT_HL)})
	inst[0x47].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_B), Reg2: re(RT_A)})
	inst[0x48].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_C), Reg2: re(RT_B)})
	inst[0x49].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_C), Reg2: re(RT_C)})
	inst[0x4A].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_C), Reg2: re(RT_D)})
	inst[0x4B].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_C), Reg2: re(RT_E)})
	inst[0x4C].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_C), Reg2: re(RT_H)})
	inst[0x4D].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_C), Reg2: re(RT_L)})
	inst[0x4E].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_MR), Reg1: re(RT_C), Reg2: re(RT_HL)})
	inst[0x4F].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_C), Reg2: re(RT_A)})

	//0x5X
	inst[0x50].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_D), Reg2: re(RT_B)})
	inst[0x51].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_D), Reg2: re(RT_C)})
	inst[0x52].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_D), Reg2: re(RT_D)})
	inst[0x53].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_D), Reg2: re(RT_E)})
	inst[0x54].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_D), Reg2: re(RT_H)})
	inst[0x55].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_D), Reg2: re(RT_L)})
	inst[0x56].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_MR), Reg1: re(RT_D), Reg2: re(RT_HL)})
	inst[0x57].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_D), Reg2: re(RT_A)})
	inst[0x58].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_E), Reg2: re(RT_B)})
	inst[0x59].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_E), Reg2: re(RT_C)})
	inst[0x5A].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_E), Reg2: re(RT_D)})
	inst[0x5B].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_E), Reg2: re(RT_E)})
	inst[0x5C].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_E), Reg2: re(RT_H)})
	inst[0x5D].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_E), Reg2: re(RT_L)})
	inst[0x5E].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_MR), Reg1: re(RT_E), Reg2: re(RT_HL)})
	inst[0x5F].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_E), Reg2: re(RT_A)})

	//0x6X
	inst[0x60].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_H), Reg2: re(RT_B)})
	inst[0x61].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_H), Reg2: re(RT_C)})
	inst[0x62].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_H), Reg2: re(RT_D)})
	inst[0x63].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_H), Reg2: re(RT_E)})
	inst[0x64].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_H), Reg2: re(RT_H)})
	inst[0x65].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_H), Reg2: re(RT_L)})
	inst[0x66].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_MR), Reg1: re(RT_H), Reg2: re(RT_HL)})
	inst[0x67].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_H), Reg2: re(RT_A)})
	inst[0x68].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_L), Reg2: re(RT_B)})
	inst[0x69].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_L), Reg2: re(RT_C)})
	inst[0x6A].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_L), Reg2: re(RT_D)})
	inst[0x6B].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_L), Reg2: re(RT_E)})
	inst[0x6C].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_L), Reg2: re(RT_H)})
	inst[0x6D].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_L), Reg2: re(RT_L)})
	inst[0x6E].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_MR), Reg1: re(RT_L), Reg2: re(RT_HL)})
	inst[0x6F].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_L), Reg2: re(RT_A)})

	//0x7X
	inst[0x70].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_MR_R), Reg1: re(RT_HL), Reg2: re(RT_B)})
	inst[0x71].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_MR_R), Reg1: re(RT_HL), Reg2: re(RT_C)})
	inst[0x72].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_MR_R), Reg1: re(RT_HL), Reg2: re(RT_D)})
	inst[0x73].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_MR_R), Reg1: re(RT_HL), Reg2: re(RT_E)})
	inst[0x74].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_MR_R), Reg1: re(RT_HL), Reg2: re(RT_H)})
	inst[0x75].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_MR_R), Reg1: re(RT_HL), Reg2: re(RT_L)})
	inst[0x76].addInst(InstPointer{Type: in(IN_HALT)})
	inst[0x77].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_MR_R), Reg1: re(RT_HL), Reg2: re(RT_A)})
	inst[0x78].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_A), Reg2: re(RT_B)})
	inst[0x79].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_A), Reg2: re(RT_C)})
	inst[0x7A].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_A), Reg2: re(RT_D)})
	inst[0x7B].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_A), Reg2: re(RT_E)})
	inst[0x7C].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_A), Reg2: re(RT_H)})
	inst[0x7D].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_A), Reg2: re(RT_L)})
	inst[0x7E].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_MR), Reg1: re(RT_A), Reg2: re(RT_HL)})
	inst[0x7F].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_R), Reg1: re(RT_A), Reg2: re(RT_A)})

	inst[0xAF].addInst(InstPointer{Type: in(IN_XOR), Mode: ad(AM_R), Reg1: re(RT_A)})

	inst[0xC0].addInst(InstPointer{Type: in(IN_RET), Mode: ad(AM_IMP), Condition: co(CT_NZ)})
	inst[0xC1].addInst(InstPointer{Type: in(IN_POP), Mode: ad(AM_R), Reg1: re(RT_BC)})
	inst[0xC2].addInst(InstPointer{Type: in(IN_JP), Mode: ad(AM_D16), Condition: co(CT_NZ)})
	inst[0xC3].addInst(InstPointer{Type: in(IN_JP), Mode: ad(AM_D16)})
	inst[0xC4].addInst(InstPointer{Type: in(IN_CALL), Mode: ad(AM_D16), Condition: co(CT_NZ)})
	inst[0xC5].addInst(InstPointer{Type: in(IN_PUSH), Mode: ad(AM_R), Reg1: re(RT_BC)})
	inst[0xC8].addInst(InstPointer{Type: in(IN_RET), Mode: ad(AM_IMP), Condition: co(CT_Z)})
	inst[0xC9].addInst(InstPointer{Type: in(IN_RET)})
	inst[0xCA].addInst(InstPointer{Type: in(IN_JP), Mode: ad(AM_D16), Condition: co(CT_Z)})
	inst[0xCC].addInst(InstPointer{Type: in(IN_CALL), Mode: ad(AM_D16), Condition: co(CT_Z)})
	inst[0xCC].addInst(InstPointer{Type: in(IN_CALL), Mode: ad(AM_D16)})

	inst[0xD0].addInst(InstPointer{Type: in(IN_RET), Mode: ad(AM_IMP), Condition: co(CT_NC)})
	inst[0xD1].addInst(InstPointer{Type: in(IN_POP), Mode: ad(AM_R), Reg1: re(RT_DE)})
	inst[0xD2].addInst(InstPointer{Type: in(IN_JP), Mode: ad(AM_D16), Condition: co(CT_NC)})
	inst[0xD4].addInst(InstPointer{Type: in(IN_CALL), Mode: ad(AM_D16), Condition: co(CT_NC)})
	inst[0xD5].addInst(InstPointer{Type: in(IN_PUSH), Mode: ad(AM_R), Reg1: re(RT_DE)})
	inst[0xD8].addInst(InstPointer{Type: in(IN_RET), Mode: ad(AM_IMP), Condition: co(CT_C)})
	inst[0xD9].addInst(InstPointer{Type: in(IN_RETI)})
	inst[0xDA].addInst(InstPointer{Type: in(IN_JP), Mode: ad(AM_D16), Condition: co(CT_C)})
	inst[0xDC].addInst(InstPointer{Type: in(IN_CALL), Mode: ad(AM_D16), Condition: co(CT_C)})

	//0xEX
	inst[0xE0].addInst(InstPointer{Type: in(IN_LDH), Mode: ad(AM_A8_R), Reg1: re(RT_NONE), Reg2: re(RT_A)})
	inst[0xE1].addInst(InstPointer{Type: in(IN_POP), Mode: ad(AM_R), Reg1: re(RT_HL)})
	inst[0xE2].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_MR_R), Reg1: re(RT_C), Reg2: re(RT_A)})
	inst[0xE5].addInst(InstPointer{Type: in(IN_PUSH), Mode: ad(AM_R), Reg1: re(RT_HL)})
	inst[0xE9].addInst(InstPointer{Type: in(IN_JP), Mode: ad(AM_MR), Reg1: re(RT_HL)})
	inst[0xEA].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_A16_R), Reg1: re(RT_NONE), Reg2: re(RT_A)})

	//0xFX
	inst[0xF0].addInst(InstPointer{Type: in(IN_LDH), Mode: ad(AM_R_D8), Reg1: re(RT_A)})
	inst[0xF1].addInst(InstPointer{Type: in(IN_POP), Mode: ad(AM_IMP), Reg1: re(RT_AF)})
	inst[0xF2].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_MR), Reg1: re(RT_A), Reg2: re(RT_C)})
	inst[0xF3].addInst(InstPointer{Type: in(IN_DI)})
	inst[0xF5].addInst(InstPointer{Type: in(IN_PUSH), Mode: ad(AM_R), Reg1: re(RT_AF)})

	inst[0xFA].addInst(InstPointer{Type: in(IN_LD), Mode: ad(AM_R_A16), Reg1: re(RT_A)})
}

func (instruction *Instruction) addInst(ip InstPointer) {

	if ip.Type != nil {
		instruction.Type = *ip.Type
	}

	if ip.Mode != nil {
		instruction.Mode = *ip.Mode
	}

	if ip.Reg1 != nil {
		instruction.Reg1 = *ip.Reg1
	}

	if ip.Reg2 != nil {
		instruction.Reg2 = *ip.Reg2
	}

	if ip.Condition != nil {
		instruction.Condition = *ip.Condition
	}

	if ip.Param != nil {
		instruction.Param = *ip.Param
	}
}

func instructionByOpcode(opcode byte) (instruction *Instruction) {
	return &inst[opcode]
}

func getInstructionName(t InType) []byte {
	return []byte(instLookup[t])
}
