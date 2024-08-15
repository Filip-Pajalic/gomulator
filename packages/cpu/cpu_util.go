package cpu

import (
	log "pajalic.go.emulator/packages/logger"
	"pajalic.go.emulator/packages/memory"
)

func CpuFlagZ() bool {
	return Bit(CpuCtx.Regs.F, 7)
}

func CpuFlagN() bool {
	return Bit(CpuCtx.Regs.F, 6)
}

func CpuFlagH() bool {
	return Bit(CpuCtx.Regs.F, 5)
}

func CpuFlagC() bool {
	return Bit(CpuCtx.Regs.F, 4)
}

// Broken here
func CpuRegRead(regType regTypes) uint16 {
	switch regType {
	case RT_A:
		return uint16(CpuCtx.Regs.A)
	case RT_F:
		return uint16(CpuCtx.Regs.F)
	case RT_B:
		return uint16(CpuCtx.Regs.B)
	case RT_C:
		return uint16(CpuCtx.Regs.C)
	case RT_D:
		return uint16(CpuCtx.Regs.D)
	case RT_E:
		return uint16(CpuCtx.Regs.E)
	case RT_H:
		return uint16(CpuCtx.Regs.H)
	case RT_L:
		return uint16(CpuCtx.Regs.L)
	case RT_AF:

		return Reverse(uint16(CpuCtx.Regs.F)<<8 | uint16(CpuCtx.Regs.A))
	case RT_BC:
		return Reverse(uint16(CpuCtx.Regs.C)<<8 | uint16(CpuCtx.Regs.B))
	case RT_DE:
		return Reverse(uint16(CpuCtx.Regs.E)<<8 | uint16(CpuCtx.Regs.D))
	case RT_HL:
		return Reverse(uint16(CpuCtx.Regs.L)<<8 | uint16(CpuCtx.Regs.H))
	case RT_PC:
		return CpuCtx.Regs.Pc
	case RT_SP:
		return CpuCtx.Regs.Sp
	default:
		return 0
	}
}

// could be problem with cast here
func CpuSetReg(regType regTypes, val uint16) {
	switch regType {
	case RT_A:
		CpuCtx.Regs.A = byte(val & 0xFF)
		return
	case RT_F:
		CpuCtx.Regs.F = byte(val & 0xFF)
		return
	case RT_B:
		CpuCtx.Regs.B = byte(val & 0xFF)
		return
	case RT_C:
		CpuCtx.Regs.C = byte(val & 0xFF)
		return
	case RT_D:
		CpuCtx.Regs.D = byte(val & 0xFF)
		return
	case RT_E:
		CpuCtx.Regs.E = byte(val & 0xFF)
		return
	case RT_H:
		CpuCtx.Regs.H = byte(val & 0xFF)
		return
	case RT_L:
		CpuCtx.Regs.L = byte(val & 0xFF)
		return
	case RT_AF:
		result := Reverse(val)
		CpuCtx.Regs.F = byte((result >> 8) & 0xFF)
		CpuCtx.Regs.A = byte((result) & 0xFF)
		return
	case RT_BC:
		result := Reverse(val)
		CpuCtx.Regs.C = byte((result >> 8) & 0xFF)
		CpuCtx.Regs.B = byte((result) & 0xFF)
		return
	case RT_DE:
		result := Reverse(val)
		CpuCtx.Regs.E = byte((result >> 8) & 0xFF)
		CpuCtx.Regs.D = byte((result) & 0xFF)
		return
	case RT_HL:
		result := Reverse(val)
		CpuCtx.Regs.L = byte((result >> 8) & 0xFF)
		CpuCtx.Regs.H = byte((result) & 0xFF)
		return
	case RT_PC:
		CpuCtx.Regs.Pc = val
		return
	case RT_SP:
		CpuCtx.Regs.Sp = val
		return
	case RT_NONE:
		return
	}
}

func CpuRegRead8(rt regTypes) byte {
	switch rt {
	case RT_A:
		return CpuCtx.Regs.A
	case RT_F:
		return CpuCtx.Regs.F
	case RT_B:
		return CpuCtx.Regs.B
	case RT_C:
		return CpuCtx.Regs.C
	case RT_D:
		return CpuCtx.Regs.D
	case RT_E:
		return CpuCtx.Regs.E
	case RT_H:
		return CpuCtx.Regs.H
	case RT_L:
		return CpuCtx.Regs.L
	case RT_HL:
		return memory.BusRead(CpuRegRead(RT_HL))
	default:
		log.Fatal("**ERR INVALID REG8: %d\n", rt)

	}
	return 0
}

func CpuSetReg8(rt regTypes, val byte) {
	switch rt {
	case RT_A:
		CpuCtx.Regs.A = val & 0xFF
		return
	case RT_F:
		CpuCtx.Regs.F = val & 0xFF
		return
	case RT_B:
		CpuCtx.Regs.B = val & 0xFF
		return
	case RT_C:
		CpuCtx.Regs.C = val & 0xFF
		return
	case RT_D:
		CpuCtx.Regs.D = val & 0xFF
		return
	case RT_E:
		CpuCtx.Regs.E = val & 0xFF
		return
	case RT_H:
		CpuCtx.Regs.H = val & 0xFF
		return
	case RT_L:
		CpuCtx.Regs.L = val & 0xFF
		return
	case RT_HL:
		memory.BusWrite(CpuRegRead(RT_HL), val)
		return
	default:
		log.Fatal("**ERR INVALID REG8: %d\n", rt)
	}
}

func CpuGetRegs() *CpuRegisters {
	return &CpuCtx.Regs
}

func CpuGetIntFlags() byte {
	return CpuCtx.IntFlags
}

func CpuSetIntFlags(value byte) {
	CpuCtx.IntFlags = value
}
