package cpu

import (
	log "pajalic.go.emulator/packages/logger"
	"pajalic.go.emulator/packages/pubsub"
)

func CpuFlagZ() bool {
	return Bit(cpuInstance.Regs.F, 7)
}

func CpuFlagN() bool {
	return Bit(cpuInstance.Regs.F, 6)
}

func CpuFlagH() bool {
	return Bit(cpuInstance.Regs.F, 5)
}

func CpuFlagC() bool {
	return Bit(cpuInstance.Regs.F, 4)
}

// Broken here
func CpuRegRead(regType regTypes) uint16 {
	switch regType {
	case RT_A:
		return uint16(cpuInstance.Regs.A)
	case RT_F:
		return uint16(cpuInstance.Regs.F)
	case RT_B:
		return uint16(cpuInstance.Regs.B)
	case RT_C:
		return uint16(cpuInstance.Regs.C)
	case RT_D:
		return uint16(cpuInstance.Regs.D)
	case RT_E:
		return uint16(cpuInstance.Regs.E)
	case RT_H:
		return uint16(cpuInstance.Regs.H)
	case RT_L:
		return uint16(cpuInstance.Regs.L)
	case RT_AF:

		return Reverse(uint16(cpuInstance.Regs.F)<<8 | uint16(cpuInstance.Regs.A))
	case RT_BC:
		return Reverse(uint16(cpuInstance.Regs.C)<<8 | uint16(cpuInstance.Regs.B))
	case RT_DE:
		return Reverse(uint16(cpuInstance.Regs.E)<<8 | uint16(cpuInstance.Regs.D))
	case RT_HL:
		return Reverse(uint16(cpuInstance.Regs.L)<<8 | uint16(cpuInstance.Regs.H))
	case RT_PC:
		return cpuInstance.Regs.Pc
	case RT_SP:
		return cpuInstance.Regs.Sp
	default:
		return 0
	}
}

// could be problem with cast here
func CpuSetReg(regType regTypes, val uint16) {
	switch regType {
	case RT_A:
		cpuInstance.Regs.A = byte(val & 0xFF)
		return
	case RT_F:
		cpuInstance.Regs.F = byte(val & 0xFF)
		return
	case RT_B:
		cpuInstance.Regs.B = byte(val & 0xFF)
		return
	case RT_C:
		cpuInstance.Regs.C = byte(val & 0xFF)
		return
	case RT_D:
		cpuInstance.Regs.D = byte(val & 0xFF)
		return
	case RT_E:
		cpuInstance.Regs.E = byte(val & 0xFF)
		return
	case RT_H:
		cpuInstance.Regs.H = byte(val & 0xFF)
		return
	case RT_L:
		cpuInstance.Regs.L = byte(val & 0xFF)
		return
	case RT_AF:
		result := Reverse(val)
		cpuInstance.Regs.F = byte((result >> 8) & 0xFF)
		cpuInstance.Regs.A = byte((result) & 0xFF)
		return
	case RT_BC:
		result := Reverse(val)
		cpuInstance.Regs.C = byte((result >> 8) & 0xFF)
		cpuInstance.Regs.B = byte((result) & 0xFF)
		return
	case RT_DE:
		result := Reverse(val)
		cpuInstance.Regs.E = byte((result >> 8) & 0xFF)
		cpuInstance.Regs.D = byte((result) & 0xFF)
		return
	case RT_HL:
		result := Reverse(val)
		cpuInstance.Regs.L = byte((result >> 8) & 0xFF)
		cpuInstance.Regs.H = byte((result) & 0xFF)
		return
	case RT_PC:
		cpuInstance.Regs.Pc = val
		return
	case RT_SP:
		cpuInstance.Regs.Sp = val
		return
	case RT_NONE:
		return
	}
}

func CpuRegRead8(rt regTypes) byte {
	switch rt {
	case RT_A:
		return cpuInstance.Regs.A
	case RT_F:
		return cpuInstance.Regs.F
	case RT_B:
		return cpuInstance.Regs.B
	case RT_C:
		return cpuInstance.Regs.C
	case RT_D:
		return cpuInstance.Regs.D
	case RT_E:
		return cpuInstance.Regs.E
	case RT_H:
		return cpuInstance.Regs.H
	case RT_L:
		return cpuInstance.Regs.L
	case RT_HL:
		return pubsub.BusCtx().BusRead(CpuRegRead(RT_HL))
	default:
		log.Fatal("**ERR INVALID REG8: %d\n", rt)

	}
	return 0
}

func CpuSetReg8(rt regTypes, val byte) {
	switch rt {
	case RT_A:
		cpuInstance.Regs.A = val & 0xFF
		return
	case RT_F:
		cpuInstance.Regs.F = val & 0xFF
		return
	case RT_B:
		cpuInstance.Regs.B = val & 0xFF
		return
	case RT_C:
		cpuInstance.Regs.C = val & 0xFF
		return
	case RT_D:
		cpuInstance.Regs.D = val & 0xFF
		return
	case RT_E:
		cpuInstance.Regs.E = val & 0xFF
		return
	case RT_H:
		cpuInstance.Regs.H = val & 0xFF
		return
	case RT_L:
		cpuInstance.Regs.L = val & 0xFF
		return
	case RT_HL:
		pubsub.BusCtx().BusWrite(CpuRegRead(RT_HL), val)
		return
	default:
		log.Fatal("**ERR INVALID REG8: %d\n", rt)
	}
}

func CpuGetRegs() *CpuRegisters {
	return &cpuInstance.Regs
}

func CpuGetIntFlags() byte {
	return cpuInstance.IntFlags
}

func CpuSetIntFlags(value byte) {
	cpuInstance.IntFlags = value
}
