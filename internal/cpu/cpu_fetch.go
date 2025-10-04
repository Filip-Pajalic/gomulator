package cpu

import (
	logger "app/internal/logger"
	"app/internal/memory"
)

/*
AM - Addressing mode, LD , JR, ADD , RET, POP ,PUSH ..
MR	- Memory address for register
R	- Register
D8	- means immediate 8-bit data
D16	- means immediate 16-bit data
A8	- means 8-bit unsigned data, which are added to $FF00 in certain instructions (replacement for missing IN and OUT instructions)
A16	- means 16-bit address
R8	- means 8-bit signed data, which are added to program counter
*/

func FetchData() {
	cpuInstance.MemDest = 0
	cpuInstance.DestIsMem = false

	if cpuInstance.currentInst == nil {
		return
	}

	switch cpuInstance.currentInst.Mode {
	case AM_IMP:
		// Implied mode: No data to fetch. OK.
		return
	case AM_R:
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg1)
		return
	case AM_R_R:
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg2)
		return
	case AM_R_D8, AM_D8:
		// Immediate 8-bit data. Correct, but should check for signedness in JR r8 (signed offset).
		cpuInstance.FetchedData = uint16(memory.BusCtx().BusRead(cpuInstance.Regs.Pc)) & 0xFF
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc++
		return
	case AM_R_D16, AM_D16:
		// Immediate 16-bit data. Correct for LD r,nn and similar.
		var lo = uint16(memory.BusCtx().BusRead(cpuInstance.Regs.Pc))
		Cm.IncreaseCycle(1)
		var hi = uint16(memory.BusCtx().BusRead(cpuInstance.Regs.Pc + 1))
		Cm.IncreaseCycle(1)
		cpuInstance.FetchedData = lo | (hi << 8)
		cpuInstance.Regs.Pc += 2
		return
	case AM_MR_R:
		// LD (reg),r. FetchedData = r, MemDest = reg. OK for most cases.
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg2)
		cpuInstance.MemDest = CpuRegRead(cpuInstance.currentInst.Reg1)
		cpuInstance.DestIsMem = true
		if cpuInstance.currentInst.Reg1 == RT_C {
			cpuInstance.MemDest |= 0xFF00 // For LD (C),A and similar. OK.
		}
		return
	case AM_R_MR:
		// LD r,(reg). FetchedData = (reg). OK for most cases.
		addr := CpuRegRead(cpuInstance.currentInst.Reg2)
		if cpuInstance.currentInst.Reg2 == RT_C {
			addr |= 0xFF00 // For LD A,(C). OK.
		}
		cpuInstance.FetchedData = uint16(memory.BusCtx().BusRead(addr)) & 0xFF
		Cm.IncreaseCycle(1)
		return
	case AM_R_HLI:
		// LD r,(HL+). FetchedData = (HL), then HL++.
		addr := CpuRegRead(RT_HL)
		cpuInstance.FetchedData = uint16(memory.BusCtx().BusRead(addr)) & 0xFF
		Cm.IncreaseCycle(1)
		CpuSetReg(RT_HL, addr+1)
		return
	case AM_R_HLD:
		// LD r,(HL-). FetchedData = (HL), then HL--.
		addr := CpuRegRead(RT_HL)
		cpuInstance.FetchedData = uint16(memory.BusCtx().BusRead(addr)) & 0xFF
		Cm.IncreaseCycle(1)
		CpuSetReg(RT_HL, addr-1)
		return
	case AM_HLI_R:
		// LD (HL+),r. FetchedData = r, MemDest = HL, then HL++.
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg2) & 0xFF
		cpuInstance.MemDest = CpuRegRead(RT_HL)
		cpuInstance.DestIsMem = true
		CpuSetReg(RT_HL, cpuInstance.MemDest+1)
		return
	case AM_HLD_R:
		// LD (HL-),r. FetchedData = r, MemDest = HL, then HL--.
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg2) & 0xFF
		cpuInstance.MemDest = CpuRegRead(RT_HL)
		cpuInstance.DestIsMem = true
		CpuSetReg(RT_HL, cpuInstance.MemDest-1)
		return
	case AM_R_A8:
		cpuInstance.FetchedData = uint16(memory.BusCtx().BusRead(cpuInstance.Regs.Pc))
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc++
		return
	case AM_A8_R:
		cpuInstance.MemDest = uint16(memory.BusCtx().BusRead(cpuInstance.Regs.Pc)) | 0xFF00
		cpuInstance.DestIsMem = true
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc++
		return
	case AM_HL_SPR:
		// LD HL,SP+e8. FetchedData = SP+signed offset. OK, but must ensure e8 is signed.
		e8 := int8(memory.BusCtx().BusRead(cpuInstance.Regs.Pc))
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc++
		cpuInstance.FetchedData = uint16(int32(cpuInstance.Regs.Sp) + int32(e8))
		return
	case AM_A16_R, AM_D16_R:
		// LD (a16),r. FetchedData = r, MemDest = a16.
		var lo = uint16(memory.BusCtx().BusRead(cpuInstance.Regs.Pc))
		Cm.IncreaseCycle(1)
		var hi = uint16(memory.BusCtx().BusRead(cpuInstance.Regs.Pc + 1))
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc += 2
		cpuInstance.MemDest = lo | (hi << 8)
		cpuInstance.DestIsMem = true
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg2) & 0xFF
		return
	case AM_MR_D8:
		// LD (reg),d8. FetchedData = d8, MemDest = reg.
		cpuInstance.FetchedData = uint16(memory.BusCtx().BusRead(cpuInstance.Regs.Pc)) & 0xFF
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc++
		cpuInstance.MemDest = CpuRegRead(cpuInstance.currentInst.Reg1)
		cpuInstance.DestIsMem = true
		return
	case AM_MR:
		// INC/DEC (reg). FetchedData = (reg), MemDest = reg.
		cpuInstance.MemDest = CpuRegRead(cpuInstance.currentInst.Reg1)
		cpuInstance.DestIsMem = true
		cpuInstance.FetchedData = uint16(memory.BusCtx().BusRead(cpuInstance.MemDest)) & 0xFF
		Cm.IncreaseCycle(1)
		return
	case AM_R_A16:
		// LD r,(a16). FetchedData = (a16).
		var lo = uint16(memory.BusCtx().BusRead(cpuInstance.Regs.Pc))
		Cm.IncreaseCycle(1)
		var hi = uint16(memory.BusCtx().BusRead(cpuInstance.Regs.Pc + 1))
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc += 2
		addr := lo | (hi << 8)
		cpuInstance.FetchedData = uint16(memory.BusCtx().BusRead(addr)) & 0xFF
		Cm.IncreaseCycle(1)
		return
	default:
		// Fault: Unknown addressing mode. Should not happen if instruction table is correct.
		logger.Warn("Unknown Addressing Mode! %d (%02X)\n", cpuInstance.currentInst.Mode, cpuInstance.CurOpCode)
		return
	}
}
