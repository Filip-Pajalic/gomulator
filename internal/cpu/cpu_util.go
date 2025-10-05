package cpu

import (
	"app/internal/common"
	logger "app/internal/logger"
	"app/internal/memory"
)

func CpuFlagZ() bool {
	return common.Bit(cpuInstance.Regs.F, 7)
}

func CpuFlagN() bool {
	return common.Bit(cpuInstance.Regs.F, 6)
}

func CpuFlagH() bool {
	return common.Bit(cpuInstance.Regs.F, 5)
}

func CpuFlagC() bool {
	return common.Bit(cpuInstance.Regs.F, 4)
}

// CpuRegRead: Reads 8/16-bit register values. For F, only upper nibble is valid. For AF, lower nibble of F is always zero.
func CpuRegRead(regType regTypes) uint16 {
	switch regType {
	case RT_A:
		return uint16(cpuInstance.Regs.A)
	case RT_F:
		return uint16(cpuInstance.Regs.F & 0xF0) // Only upper nibble is valid
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
		// Lower nibble of F is always zero
		return (uint16(cpuInstance.Regs.A) << 8) | uint16(cpuInstance.Regs.F&0xF0)
	case RT_BC:
		return (uint16(cpuInstance.Regs.B) << 8) | uint16(cpuInstance.Regs.C)
	case RT_DE:
		return (uint16(cpuInstance.Regs.D) << 8) | uint16(cpuInstance.Regs.E)
	case RT_HL:
		return (uint16(cpuInstance.Regs.H) << 8) | uint16(cpuInstance.Regs.L)
	case RT_PC:
		return cpuInstance.Regs.Pc
	case RT_SP:
		return cpuInstance.Regs.Sp
	default:
		return 0
	}
}

// CpuSetReg: Sets 8/16-bit register values. For F, only upper nibble is set. For AF, lower nibble of F is always zero.
func CpuSetReg(regType regTypes, val uint16) {
	switch regType {
	case RT_A:
		cpuInstance.Regs.A = byte(val & 0xFF)
	case RT_F:
		cpuInstance.Regs.F = byte(val & 0xF0) // Only upper nibble is set
	case RT_B:
		cpuInstance.Regs.B = byte(val & 0xFF)
	case RT_C:
		cpuInstance.Regs.C = byte(val & 0xFF)
	case RT_D:
		cpuInstance.Regs.D = byte(val & 0xFF)
	case RT_E:
		cpuInstance.Regs.E = byte(val & 0xFF)
	case RT_H:
		cpuInstance.Regs.H = byte(val & 0xFF)
	case RT_L:
		cpuInstance.Regs.L = byte(val & 0xFF)
	case RT_AF:
		cpuInstance.Regs.A = byte((val >> 8) & 0xFF)
		cpuInstance.Regs.F = byte(val & 0xF0) // Lower nibble always zero
	case RT_BC:
		cpuInstance.Regs.B = byte((val >> 8) & 0xFF)
		cpuInstance.Regs.C = byte(val & 0xFF)
	case RT_DE:
		cpuInstance.Regs.D = byte((val >> 8) & 0xFF)
		cpuInstance.Regs.E = byte(val & 0xFF)
	case RT_HL:
		cpuInstance.Regs.H = byte((val >> 8) & 0xFF)
		cpuInstance.Regs.L = byte(val & 0xFF)
	case RT_PC:
		cpuInstance.Regs.Pc = val
	case RT_SP:
		cpuInstance.Regs.Sp = val
	case RT_NONE:
		// Do nothing
	}
}

// CpuRegRead8: Reads 8-bit register or memory at HL. For F, only upper nibble is valid.
func CpuRegRead8(rt regTypes) byte {
	switch rt {
	case RT_A:
		return cpuInstance.Regs.A
	case RT_F:
		return cpuInstance.Regs.F & 0xF0
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
		addr := CpuRegRead(RT_HL)
		return memory.BusCtx().BusRead(addr)
	case RT_NONE:
		// Reference implementation allows RT_NONE but doesn't return a value
		return 0
	default:
		logger.Error("**ERR INVALID REG8: %d\n", rt)
		return 0
	}
}

// CpuSetReg8: Sets 8-bit register or memory at HL. For F, only upper nibble is set.
func CpuSetReg8(rt regTypes, val byte) {
	switch rt {
	case RT_A:
		cpuInstance.Regs.A = val
	case RT_F:
		cpuInstance.Regs.F = val & 0xF0
	case RT_B:
		cpuInstance.Regs.B = val
	case RT_C:
		cpuInstance.Regs.C = val
	case RT_D:
		cpuInstance.Regs.D = val
	case RT_E:
		cpuInstance.Regs.E = val
	case RT_H:
		cpuInstance.Regs.H = val
	case RT_L:
		cpuInstance.Regs.L = val
	case RT_HL:
		addr := CpuRegRead(RT_HL)
		memory.BusCtx().BusWrite(addr, val)
	case RT_NONE:
		// Reference implementation allows RT_NONE but doesn't set anything
		// Just ignore the operation
	default:
		logger.Error("**ERR INVALID REG8: %d\n", rt)
	}
}

func CpuGetRegs() *CpuRegisters {
	return &cpuInstance.Regs
}

func CpuGetIntFlags() byte {
	return cpuInstance.IntFlags
}

func CpuSetIntFlags(value byte) {
	logger.Debug("CpuSetIntFlags: value=%02X", value)
	cpuInstance.IntFlags = value
}

func CpuRequestInterrupt(it InterruptType) {
	cpuInstance.RequestInterrupt(it)
}
