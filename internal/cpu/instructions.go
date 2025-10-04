package cpu

import (
	"app/internal/logger"
	"app/internal/memory"
	"fmt"
)

/*
	Different instructionmodes needed

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

var inst [0x100]Instruction

func InitInstructions() {
	// --- Begin review and comments for spec compliance ---
	// NOTE: This table should match the official Game Boy CPU instruction set (see Pandocs)
	// Check for missing, incorrect, or mis-assigned instructions below:

	// 0x08: LD (a16),SP is correct (Reg2: RT_SP)
	// 0x10: STOP is correct
	// 0x18: JR r8 (should be signed offset, AM_D8 is used, but AM_D8 is unsigned; see fetch/execute)
	// 0x20, 0x28, 0x30, 0x38: JR cc,r8 (should be signed offset, AM_D8)
	// 0x22, 0x2A, 0x32, 0x3A: HL+ and HL- modes are correct
	// 0x36: LD (HL),d8 is correct
	// 0x76: HALT is correct
	// 0xC3: JP a16 (unconditional)
	// 0xC9: RET (unconditional)
	// 0xCB: CB prefix (handled in procCb)
	// 0xCD: CALL a16 (unconditional)
	// 0xE0: LDH (a8),A (should use AM_A8_R, Reg2: RT_A)
	// 0xE2: LD (C),A (should use AM_MR_R, Reg1: RT_C, Reg2: RT_A)
	// 0xEA: LD (a16),A (should use AM_A16_R, Reg2: RT_A)
	// 0xF0: LDH A,(a8) (should use AM_R_A8, Reg1: RT_A)
	// 0xF2: LD A,(C) (should use AM_R_MR, Reg1: RT_A, Reg2: RT_C)
	// 0xF8: LD HL,SP+e8 (should use AM_HL_SPR, Reg1: RT_HL, Reg2: RT_SP)
	// 0xF9: LD SP,HL (should use AM_R_R, Reg1: RT_SP, Reg2: RT_HL)
	// 0xFA: LD A,(a16) (should use AM_R_A16, Reg1: RT_A)
	// 0xFB: EI is correct
	// 0xFE: CP d8 (should use AM_R_D8, Reg1: RT_A)
	// 0xFF: RST 38h is correct

	// --- End review ---
	// If you add new instructions, ensure all 0x00-0xFF opcodes are covered and match the spec.
	// For CB-prefixed instructions, see procCb and CB decode logic.

	inst[0x00] = Instruction{
		Type: IN_NOP, Mode: AM_IMP,
	}
	inst[0x01] = Instruction{
		Type: IN_LD, Mode: AM_R_D16, Reg1: RT_BC,
	}
	inst[0x02] = Instruction{
		Type: IN_LD, Mode: AM_MR_R, Reg1: RT_BC, Reg2: RT_A,
	}
	inst[0x03] = Instruction{
		Type: IN_INC, Mode: AM_R, Reg1: RT_BC,
	}
	inst[0x04] = Instruction{
		Type: IN_INC, Mode: AM_R, Reg1: RT_B,
	}
	inst[0x05] = Instruction{
		Type: IN_DEC, Mode: AM_R, Reg1: RT_B,
	}
	inst[0x06] = Instruction{
		Type: IN_LD, Mode: AM_R_D8, Reg1: RT_B,
	}
	inst[0x07] = Instruction{
		Type: IN_RLCA,
	}
	inst[0x08] = Instruction{
		Type: IN_LD, Mode: AM_A16_R, Reg1: RT_NONE, Reg2: RT_SP,
	}
	inst[0x09] = Instruction{
		Type: IN_ADD, Mode: AM_R_R, Reg1: RT_HL, Reg2: RT_BC,
	}
	inst[0x0A] = Instruction{
		Type: IN_LD, Mode: AM_R_MR, Reg1: RT_A, Reg2: RT_BC,
	}
	inst[0x0B] = Instruction{
		Type: IN_DEC, Mode: AM_R, Reg1: RT_BC,
	}
	inst[0x0C] = Instruction{
		Type: IN_INC, Mode: AM_R, Reg1: RT_C,
	}
	inst[0x0D] = Instruction{
		Type: IN_DEC, Mode: AM_R, Reg1: RT_C,
	}
	inst[0x0E] = Instruction{
		Type: IN_LD, Mode: AM_R_D8, Reg1: RT_C,
	}
	inst[0x0F] = Instruction{
		Type: IN_RRCA,
	}

	//0x1X
	inst[0x10] = Instruction{
		Type: IN_STOP,
	}
	inst[0x11] = Instruction{
		Type: IN_LD, Mode: AM_R_D16, Reg1: RT_DE,
	}
	inst[0x12] = Instruction{
		Type: IN_LD, Mode: AM_MR_R, Reg1: RT_DE, Reg2: RT_A,
	}
	inst[0x13] = Instruction{
		Type: IN_INC, Mode: AM_R, Reg1: RT_DE,
	}
	inst[0x14] = Instruction{
		Type: IN_INC, Mode: AM_R, Reg1: RT_D,
	}
	inst[0x15] = Instruction{
		Type: IN_DEC, Mode: AM_R, Reg1: RT_D,
	}
	inst[0x16] = Instruction{
		Type: IN_LD, Mode: AM_R_D8, Reg1: RT_D,
	}
	inst[0x17] = Instruction{
		Type: IN_RLA,
	}
	inst[0x18] = Instruction{
		Type: IN_JR, Mode: AM_D8,
	}
	inst[0x19] = Instruction{
		Type: IN_ADD, Mode: AM_R_R, Reg1: RT_HL, Reg2: RT_DE,
	}
	inst[0x1A] = Instruction{
		Type: IN_LD, Mode: AM_R_MR, Reg1: RT_A, Reg2: RT_DE,
	}
	inst[0x1B] = Instruction{
		Type: IN_DEC, Mode: AM_R, Reg1: RT_DE,
	}
	inst[0x1C] = Instruction{
		Type: IN_INC, Mode: AM_R, Reg1: RT_E,
	}
	inst[0x1D] = Instruction{
		Type: IN_DEC, Mode: AM_R, Reg1: RT_E,
	}
	inst[0x1E] = Instruction{
		Type: IN_LD, Mode: AM_R_D8, Reg1: RT_E,
	}
	inst[0x1F] = Instruction{
		Type: IN_RRA,
	}

	//0x2X

	inst[0x20] = Instruction{
		Type: IN_JR, Mode: AM_D8, Condition: CT_NZ,
	}
	inst[0x21] = Instruction{
		Type: IN_LD, Mode: AM_R_D16, Reg1: RT_HL,
	}
	inst[0x22] = Instruction{
		Type: IN_LD, Mode: AM_HLI_R, Reg1: RT_HL, Reg2: RT_A,
	}
	inst[0x23] = Instruction{
		Type: IN_INC, Mode: AM_R, Reg1: RT_HL,
	}
	inst[0x24] = Instruction{
		Type: IN_INC, Mode: AM_R, Reg1: RT_H,
	}
	inst[0x25] = Instruction{
		Type: IN_DEC, Mode: AM_R, Reg1: RT_H,
	}
	inst[0x26] = Instruction{
		Type: IN_LD, Mode: AM_R_D8, Reg1: RT_H,
	}
	inst[0x27] = Instruction{
		Type: IN_DAA,
	}
	inst[0x28] = Instruction{
		Type: IN_JR, Mode: AM_D8, Condition: CT_Z,
	}
	inst[0x29] = Instruction{
		Type: IN_ADD, Mode: AM_R_R, Reg1: RT_HL, Reg2: RT_HL,
	}
	inst[0x2A] = Instruction{
		Type: IN_LD, Mode: AM_R_HLI, Reg1: RT_A, Reg2: RT_HL,
	}
	inst[0x2B] = Instruction{
		Type: IN_DEC, Mode: AM_R, Reg1: RT_HL,
	}
	inst[0x2C] = Instruction{
		Type: IN_INC, Mode: AM_R, Reg1: RT_L,
	}
	inst[0x2D] = Instruction{
		Type: IN_DEC, Mode: AM_R, Reg1: RT_L,
	}
	inst[0x2E] = Instruction{
		Type: IN_LD, Mode: AM_R_D8, Reg1: RT_L,
	}
	inst[0x2F] = Instruction{
		Type: IN_CPL,
	}

	//0x3X
	inst[0x30] = Instruction{
		Type: IN_JR, Mode: AM_D8, Condition: CT_NC,
	}
	inst[0x31] = Instruction{
		Type: IN_LD, Mode: AM_R_D16, Reg1: RT_SP,
	}
	inst[0x32] = Instruction{
		Type: IN_LD, Mode: AM_HLD_R, Reg1: RT_HL, Reg2: RT_A,
	}
	inst[0x33] = Instruction{
		Type: IN_INC, Mode: AM_R, Reg1: RT_SP,
	}
	inst[0x34] = Instruction{
		Type: IN_INC, Mode: AM_MR, Reg1: RT_HL,
	}
	inst[0x35] = Instruction{
		Type: IN_DEC, Mode: AM_MR, Reg1: RT_HL,
	}
	inst[0x36] = Instruction{
		Type: IN_LD, Mode: AM_MR_D8, Reg1: RT_HL,
	}
	inst[0x37] = Instruction{
		Type: IN_SCF,
	}
	inst[0x38] = Instruction{
		Type: IN_JR, Mode: AM_D8, Condition: CT_C,
	}
	inst[0x39] = Instruction{
		Type: IN_ADD, Mode: AM_R_R, Reg1: RT_HL, Reg2: RT_SP,
	}
	inst[0x3A] = Instruction{
		Type: IN_LD, Mode: AM_R_HLD, Reg1: RT_A, Reg2: RT_HL,
	}
	inst[0x3B] = Instruction{
		Type: IN_DEC, Mode: AM_R, Reg1: RT_SP,
	}
	inst[0x3C] = Instruction{
		Type: IN_INC, Mode: AM_R, Reg1: RT_A,
	}
	inst[0x3D] = Instruction{
		Type: IN_DEC, Mode: AM_R, Reg1: RT_A,
	}
	inst[0x3E] = Instruction{
		Type: IN_LD, Mode: AM_R_D8, Reg1: RT_A,
	}
	inst[0x3F] = Instruction{
		Type: IN_CCF,
	}

	//0x4X
	inst[0x40] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_B, Reg2: RT_B,
	}
	inst[0x41] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_B, Reg2: RT_C,
	}
	inst[0x42] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_B, Reg2: RT_D,
	}
	inst[0x43] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_B, Reg2: RT_E,
	}
	inst[0x44] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_B, Reg2: RT_H,
	}
	inst[0x45] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_B, Reg2: RT_L,
	}
	inst[0x46] = Instruction{
		Type: IN_LD, Mode: AM_R_MR, Reg1: RT_B, Reg2: RT_HL,
	}
	inst[0x47] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_B, Reg2: RT_A,
	}
	inst[0x48] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_C, Reg2: RT_B,
	}
	inst[0x49] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_C, Reg2: RT_C,
	}
	inst[0x4A] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_C, Reg2: RT_D,
	}
	inst[0x4B] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_C, Reg2: RT_E,
	}
	inst[0x4C] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_C, Reg2: RT_H,
	}
	inst[0x4D] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_C, Reg2: RT_L,
	}
	inst[0x4E] = Instruction{
		Type: IN_LD, Mode: AM_R_MR, Reg1: RT_C, Reg2: RT_HL,
	}
	inst[0x4F] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_C, Reg2: RT_A,
	}

	//0x5X
	inst[0x50] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_D, Reg2: RT_B,
	}
	inst[0x51] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_D, Reg2: RT_C,
	}
	inst[0x52] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_D, Reg2: RT_D,
	}
	inst[0x53] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_D, Reg2: RT_E,
	}
	inst[0x54] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_D, Reg2: RT_H,
	}
	inst[0x55] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_D, Reg2: RT_L,
	}
	inst[0x56] = Instruction{
		Type: IN_LD, Mode: AM_R_MR, Reg1: RT_D, Reg2: RT_HL,
	}
	inst[0x57] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_D, Reg2: RT_A,
	}
	inst[0x58] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_E, Reg2: RT_B,
	}
	inst[0x59] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_E, Reg2: RT_C,
	}
	inst[0x5A] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_E, Reg2: RT_D,
	}
	inst[0x5B] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_E, Reg2: RT_E,
	}
	inst[0x5C] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_E, Reg2: RT_H,
	}
	inst[0x5D] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_E, Reg2: RT_L,
	}
	inst[0x5E] = Instruction{
		Type: IN_LD, Mode: AM_R_MR, Reg1: RT_E, Reg2: RT_HL,
	}
	inst[0x5F] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_E, Reg2: RT_A,
	}

	//0x6X
	inst[0x60] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_H, Reg2: RT_B,
	}
	inst[0x61] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_H, Reg2: RT_C,
	}
	inst[0x62] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_H, Reg2: RT_D,
	}
	inst[0x63] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_H, Reg2: RT_E,
	}
	inst[0x64] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_H, Reg2: RT_H,
	}
	inst[0x65] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_H, Reg2: RT_L,
	}
	inst[0x66] = Instruction{
		Type: IN_LD, Mode: AM_R_MR, Reg1: RT_H, Reg2: RT_HL,
	}
	inst[0x67] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_H, Reg2: RT_A,
	}
	inst[0x68] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_L, Reg2: RT_B,
	}
	inst[0x69] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_L, Reg2: RT_C,
	}
	inst[0x6A] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_L, Reg2: RT_D,
	}
	inst[0x6B] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_L, Reg2: RT_E,
	}
	inst[0x6C] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_L, Reg2: RT_H,
	}
	inst[0x6D] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_L, Reg2: RT_L,
	}
	inst[0x6E] = Instruction{
		Type: IN_LD, Mode: AM_R_MR, Reg1: RT_L, Reg2: RT_HL,
	}
	inst[0x6F] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_L, Reg2: RT_A,
	}

	//0x7X
	inst[0x70] = Instruction{
		Type: IN_LD, Mode: AM_MR_R, Reg1: RT_HL, Reg2: RT_B,
	}
	inst[0x71] = Instruction{
		Type: IN_LD, Mode: AM_MR_R, Reg1: RT_HL, Reg2: RT_C,
	}
	inst[0x72] = Instruction{
		Type: IN_LD, Mode: AM_MR_R, Reg1: RT_HL, Reg2: RT_D,
	}
	inst[0x73] = Instruction{
		Type: IN_LD, Mode: AM_MR_R, Reg1: RT_HL, Reg2: RT_E,
	}
	inst[0x74] = Instruction{
		Type: IN_LD, Mode: AM_MR_R, Reg1: RT_HL, Reg2: RT_H,
	}
	inst[0x75] = Instruction{
		Type: IN_LD, Mode: AM_MR_R, Reg1: RT_HL, Reg2: RT_L,
	}
	inst[0x76] = Instruction{
		Type: IN_HALT,
	}
	inst[0x77] = Instruction{
		Type: IN_LD, Mode: AM_MR_R, Reg1: RT_HL, Reg2: RT_A,
	}
	inst[0x78] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_B,
	}
	inst[0x79] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_C,
	}
	inst[0x7A] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_D,
	}
	inst[0x7B] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_E,
	}
	inst[0x7C] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_H,
	}
	inst[0x7D] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_L,
	}
	inst[0x7E] = Instruction{
		Type: IN_LD, Mode: AM_R_MR, Reg1: RT_A, Reg2: RT_HL,
	}
	inst[0x7F] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_A,
	}

	//0x8X
	inst[0x80] = Instruction{
		Type: IN_ADD, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_B,
	}
	inst[0x81] = Instruction{
		Type: IN_ADD, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_C,
	}
	inst[0x82] = Instruction{
		Type: IN_ADD, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_D,
	}
	inst[0x83] = Instruction{
		Type: IN_ADD, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_E,
	}
	inst[0x84] = Instruction{
		Type: IN_ADD, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_H,
	}
	inst[0x85] = Instruction{
		Type: IN_ADD, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_L,
	}
	inst[0x86] = Instruction{
		Type: IN_ADD, Mode: AM_R_MR, Reg1: RT_A, Reg2: RT_HL,
	}
	inst[0x87] = Instruction{
		Type: IN_ADD, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_A,
	}
	inst[0x88] = Instruction{
		Type: IN_ADC, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_B,
	}
	inst[0x89] = Instruction{
		Type: IN_ADC, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_C,
	}
	inst[0x8A] = Instruction{
		Type: IN_ADC, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_D,
	}
	inst[0x8B] = Instruction{
		Type: IN_ADC, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_E,
	}
	inst[0x8C] = Instruction{
		Type: IN_ADC, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_H,
	}
	inst[0x8D] = Instruction{
		Type: IN_ADC, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_L,
	}
	inst[0x8E] = Instruction{
		Type: IN_ADC, Mode: AM_R_MR, Reg1: RT_A, Reg2: RT_HL,
	}
	inst[0x8F] = Instruction{
		Type: IN_ADC, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_A,
	}

	//0x9X
	inst[0x90] = Instruction{
		Type: IN_SUB, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_B,
	}
	inst[0x91] = Instruction{
		Type: IN_SUB, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_C,
	}
	inst[0x92] = Instruction{
		Type: IN_SUB, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_D,
	}
	inst[0x93] = Instruction{
		Type: IN_SUB, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_E,
	}
	inst[0x94] = Instruction{
		Type: IN_SUB, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_H,
	}
	inst[0x95] = Instruction{
		Type: IN_SUB, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_L,
	}
	inst[0x96] = Instruction{
		Type: IN_SUB, Mode: AM_R_MR, Reg1: RT_A, Reg2: RT_HL,
	}
	inst[0x97] = Instruction{
		Type: IN_SUB, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_A,
	}
	inst[0x98] = Instruction{
		Type: IN_SBC, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_B,
	}
	inst[0x99] = Instruction{
		Type: IN_SBC, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_C,
	}
	inst[0x9A] = Instruction{
		Type: IN_SBC, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_D,
	}
	inst[0x9B] = Instruction{
		Type: IN_SBC, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_E,
	}
	inst[0x9C] = Instruction{
		Type: IN_SBC, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_H,
	}
	inst[0x9D] = Instruction{
		Type: IN_SBC, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_L,
	}
	inst[0x9E] = Instruction{
		Type: IN_SBC, Mode: AM_R_MR, Reg1: RT_A, Reg2: RT_HL,
	}
	inst[0x9F] = Instruction{
		Type: IN_SBC, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_A,
	}

	//0xAX
	inst[0xA0] = Instruction{
		Type: IN_AND, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_B,
	}
	inst[0xA1] = Instruction{
		Type: IN_AND, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_C,
	}
	inst[0xA2] = Instruction{
		Type: IN_AND, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_D,
	}
	inst[0xA3] = Instruction{
		Type: IN_AND, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_E,
	}
	inst[0xA4] = Instruction{
		Type: IN_AND, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_H,
	}
	inst[0xA5] = Instruction{
		Type: IN_AND, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_L,
	}
	inst[0xA6] = Instruction{
		Type: IN_AND, Mode: AM_R_MR, Reg1: RT_A, Reg2: RT_HL,
	}
	inst[0xA7] = Instruction{
		Type: IN_AND, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_A,
	}
	inst[0xA8] = Instruction{
		Type: IN_XOR, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_B,
	}
	inst[0xA9] = Instruction{
		Type: IN_XOR, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_C,
	}
	inst[0xAA] = Instruction{
		Type: IN_XOR, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_D,
	}
	inst[0xAB] = Instruction{
		Type: IN_XOR, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_E,
	}
	inst[0xAC] = Instruction{
		Type: IN_XOR, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_H,
	}
	inst[0xAD] = Instruction{
		Type: IN_XOR, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_L,
	}
	inst[0xAE] = Instruction{
		Type: IN_XOR, Mode: AM_R_MR, Reg1: RT_A, Reg2: RT_HL,
	}
	inst[0xAF] = Instruction{
		Type: IN_XOR, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_A,
	}

	//0xBX
	inst[0xB0] = Instruction{
		Type: IN_OR, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_B,
	}
	inst[0xB1] = Instruction{
		Type: IN_OR, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_C,
	}
	inst[0xB2] = Instruction{
		Type: IN_OR, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_D,
	}
	inst[0xB3] = Instruction{
		Type: IN_OR, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_E,
	}
	inst[0xB4] = Instruction{
		Type: IN_OR, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_H,
	}
	inst[0xB5] = Instruction{
		Type: IN_OR, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_L,
	}
	inst[0xB6] = Instruction{
		Type: IN_OR, Mode: AM_R_MR, Reg1: RT_A, Reg2: RT_HL,
	}
	inst[0xB7] = Instruction{
		Type: IN_OR, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_A,
	}
	inst[0xB8] = Instruction{
		Type: IN_CP, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_B,
	}
	inst[0xB9] = Instruction{
		Type: IN_CP, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_C,
	}
	inst[0xBA] = Instruction{
		Type: IN_CP, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_D,
	}
	inst[0xBB] = Instruction{
		Type: IN_CP, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_E,
	}
	inst[0xBC] = Instruction{
		Type: IN_CP, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_H,
	}
	inst[0xBD] = Instruction{
		Type: IN_CP, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_L,
	}
	inst[0xBE] = Instruction{
		Type: IN_CP, Mode: AM_R_MR, Reg1: RT_A, Reg2: RT_HL,
	}
	inst[0xBF] = Instruction{
		Type: IN_CP, Mode: AM_R_R, Reg1: RT_A, Reg2: RT_A,
	}

	//0xCX
	inst[0xC0] = Instruction{
		Type: IN_RET, Mode: AM_IMP, Condition: CT_NZ,
	}
	inst[0xC1] = Instruction{
		Type: IN_POP, Mode: AM_R, Reg1: RT_BC,
	}
	inst[0xC2] = Instruction{
		Type: IN_JP, Mode: AM_D16, Condition: CT_NZ,
	}
	inst[0xC3] = Instruction{
		Type: IN_JP, Mode: AM_D16,
	}
	inst[0xC4] = Instruction{
		Type: IN_CALL, Mode: AM_D16, Condition: CT_NZ,
	}
	inst[0xC5] = Instruction{
		Type: IN_PUSH, Mode: AM_R, Reg1: RT_BC,
	}
	inst[0xC6] = Instruction{
		Type: IN_ADD, Mode: AM_R_D8, Reg1: RT_A,
	}
	inst[0xC7] = Instruction{
		Type: IN_RST, Mode: AM_IMP, Param: 0x00,
	}
	inst[0xC8] = Instruction{
		Type: IN_RET, Mode: AM_IMP, Condition: CT_Z,
	}
	inst[0xC9] = Instruction{
		Type: IN_RET,
	}
	inst[0xCA] = Instruction{
		Type: IN_JP, Mode: AM_D16, Condition: CT_Z,
	}
	inst[0xCB] = Instruction{
		Type: IN_CB, Mode: AM_D8,
	}
	inst[0xCC] = Instruction{
		Type: IN_CALL, Mode: AM_D16, Condition: CT_Z,
	}
	inst[0xCD] = Instruction{
		Type: IN_CALL, Mode: AM_D16,
	}
	inst[0xCE] = Instruction{
		Type: IN_ADC, Mode: AM_R_D8, Reg1: RT_A,
	}
	inst[0xCF] = Instruction{
		Type: IN_RST, Mode: AM_IMP, Param: 0x08,
	}

	inst[0xD0] = Instruction{
		Type: IN_RET, Mode: AM_IMP, Condition: CT_NC,
	}
	inst[0xD1] = Instruction{
		Type: IN_POP, Mode: AM_R, Reg1: RT_DE,
	}
	inst[0xD2] = Instruction{
		Type: IN_JP, Mode: AM_D16, Condition: CT_NC,
	}
	inst[0xD4] = Instruction{
		Type: IN_CALL, Mode: AM_D16, Condition: CT_NC,
	}
	inst[0xD5] = Instruction{
		Type: IN_PUSH, Mode: AM_R, Reg1: RT_DE,
	}
	inst[0xD6] = Instruction{
		Type: IN_SUB, Mode: AM_R_D8, Reg1: RT_A,
	}
	inst[0xD7] = Instruction{
		Type: IN_RST, Mode: AM_IMP, Param: 0x10,
	}
	inst[0xD8] = Instruction{
		Type: IN_RET, Mode: AM_IMP, Condition: CT_C,
	}
	inst[0xD9] = Instruction{
		Type: IN_RETI,
	}
	inst[0xDA] = Instruction{
		Type: IN_JP, Mode: AM_D16, Condition: CT_C,
	}
	inst[0xDC] = Instruction{
		Type: IN_CALL, Mode: AM_D16, Condition: CT_C,
	}
	inst[0xDE] = Instruction{
		Type: IN_SBC, Mode: AM_R_D8, Reg1: RT_A,
	}
	inst[0xDF] = Instruction{
		Type: IN_RST, Mode: AM_IMP, Param: 0x18,
	}

	//0xEX
	inst[0xE0] = Instruction{
		Type: IN_LDH, Mode: AM_A8_R, Reg2: RT_A,
	}
	inst[0xE1] = Instruction{
		Type: IN_POP, Mode: AM_R, Reg1: RT_HL,
	}
	inst[0xE2] = Instruction{
		Type: IN_LD, Mode: AM_MR_R, Reg1: RT_C, Reg2: RT_A,
	}
	inst[0xE5] = Instruction{
		Type: IN_PUSH, Mode: AM_R, Reg1: RT_HL,
	}
	inst[0xE6] = Instruction{
		Type: IN_AND, Mode: AM_R_D8, Reg1: RT_A,
	}
	inst[0xE7] = Instruction{
		Type: IN_RST, Mode: AM_IMP, Param: 0x20,
	}
	inst[0xE8] = Instruction{
		Type: IN_ADD, Mode: AM_R_D8, Reg1: RT_SP,
	}
	inst[0xE9] = Instruction{
		Type: IN_JP, Mode: AM_R, Reg1: RT_HL,
	}
	inst[0xEA] = Instruction{
		Type: IN_LD, Mode: AM_A16_R, Reg1: RT_NONE, Reg2: RT_A,
	}
	inst[0xEE] = Instruction{
		Type: IN_XOR, Mode: AM_R_D8, Reg1: RT_A,
	}
	inst[0xEF] = Instruction{
		Type: IN_RST, Mode: AM_IMP, Param: 0x28,
	}

	//0xFX
	inst[0xF0] = Instruction{
		Type: IN_LDH, Mode: AM_R_A8, Reg1: RT_A,
	}
	inst[0xF1] = Instruction{
		Type: IN_POP, Mode: AM_R, Reg1: RT_AF,
	}
	inst[0xF2] = Instruction{
		Type: IN_LD, Mode: AM_R_MR, Reg1: RT_A, Reg2: RT_C,
	}
	inst[0xF3] = Instruction{
		Type: IN_DI,
	}
	inst[0xF5] = Instruction{
		Type: IN_PUSH, Mode: AM_R, Reg1: RT_AF,
	}
	inst[0xF6] = Instruction{
		Type: IN_OR, Mode: AM_R_D8, Reg1: RT_A,
	}
	inst[0xF7] = Instruction{
		Type: IN_RST, Mode: AM_IMP, Param: 0x30,
	}
	inst[0xF8] = Instruction{
		Type: IN_LD, Mode: AM_HL_SPR, Reg1: RT_HL, Reg2: RT_SP,
	}
	inst[0xF9] = Instruction{
		Type: IN_LD, Mode: AM_R_R, Reg1: RT_SP, Reg2: RT_HL,
	}
	inst[0xFA] = Instruction{
		Type: IN_LD, Mode: AM_R_A16, Reg1: RT_A,
	}
	inst[0xFB] = Instruction{
		Type: IN_EI,
	}
	inst[0xFE] = Instruction{
		Type: IN_CP, Mode: AM_R_D8, Reg1: RT_A,
	}
	inst[0xFF] = Instruction{
		Type: IN_RST, Mode: AM_IMP, Param: 0x38,
	}
}

func instructionByOpcode(opcode byte) (instruction *Instruction) {
	return &inst[opcode]
}

func getInstructionName(t InType) string {
	return instLookup[t]
}

var rtLookupString = []string{
	"<NONE>",
	"A",
	"F",
	"B",
	"C",
	"D",
	"E",
	"H",
	"L",
	"AF",
	"BC",
	"DE",
	"HL",
	"SP",
	"PC",
}

func instName(t InType) string {
	return instLookup[t]
}

func instToStr(ctx *CpuContext, s *string) {
	var inst = ctx.currentInst
	*s = fmt.Sprintf("%s ", instName(inst.Type))

	switch inst.Mode {
	case AM_IMP:
		return

	case AM_R_D16, AM_R_A16:
		*s += fmt.Sprintf("%s,$%04X", rtLookupString[inst.Reg1], ctx.FetchedData)
		return

	case AM_R:
		*s += fmt.Sprintf("%s", rtLookupString[inst.Reg1])
		return

	case AM_R_R:
		*s += fmt.Sprintf("%s,%s", rtLookupString[inst.Reg1], rtLookupString[inst.Reg2])
		return

	case AM_MR_R:
		*s += fmt.Sprintf("(%s),%s", rtLookupString[inst.Reg1], rtLookupString[inst.Reg2])
		return

	case AM_MR:
		*s += fmt.Sprintf("(%s)", rtLookupString[inst.Reg1])
		return

	case AM_R_MR:
		*s += fmt.Sprintf("%s,(%s)", rtLookupString[inst.Reg1], rtLookupString[inst.Reg2])
		return

	case AM_R_D8, AM_R_A8:
		*s += fmt.Sprintf("%s,$%02X", rtLookupString[inst.Reg1], ctx.FetchedData&0xFF)
		return

	case AM_R_HLI:
		*s += fmt.Sprintf("%s,(%s+)", rtLookupString[inst.Reg1], rtLookupString[inst.Reg2])
		return

	case AM_R_HLD:
		*s += fmt.Sprintf("%s,(%s-)", rtLookupString[inst.Reg1], rtLookupString[inst.Reg2])
		return

	case AM_HLI_R:
		*s += fmt.Sprintf("(%s+),%s", rtLookupString[inst.Reg1], rtLookupString[inst.Reg2])
		return

	case AM_HLD_R:
		*s += fmt.Sprintf("(%s-),%s", rtLookupString[inst.Reg1], rtLookupString[inst.Reg2])
		return

	case AM_A8_R:
		fetchedValue := memory.BusCtx().BusRead(ctx.Regs.Pc - 1)
		*s += fmt.Sprintf("$%02X,%s", fetchedValue, rtLookupString[inst.Reg2])
		return

	case AM_HL_SPR:
		*s += fmt.Sprintf("(%s),SP+%d", rtLookupString[inst.Reg1], ctx.FetchedData&0xFF)
		return

	case AM_D8:
		*s += fmt.Sprintf("$%02X", ctx.FetchedData&0xFF)
		return

	case AM_D16:
		*s += fmt.Sprintf("$%04X", ctx.FetchedData)
		return

	case AM_MR_D8:
		*s += fmt.Sprintf("(%s),$%02X", rtLookupString[inst.Reg1], ctx.FetchedData&0xFF)
		return

	case AM_A16_R:
		*s += fmt.Sprintf("($%04X),%s", ctx.FetchedData, rtLookupString[inst.Reg2])
		return

	default:
		logger.Fatal("INVALID Addressing Mode: %d", inst.Mode)
	}
}
