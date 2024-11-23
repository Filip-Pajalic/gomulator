package cpu

import (
	log "pajalic.go.emulator/packages/logger"
	"pajalic.go.emulator/packages/pubsub"
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
		return
	case AM_R:
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg1) & 0xFF
		return
	case AM_R_R:
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg2) & 0xFF
		return
	case AM_R_D8, AM_D8:
		cpuInstance.FetchedData = uint16(pubsub.BusCtx().BusRead(cpuInstance.Regs.Pc)) & 0xFF
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc++
		return
	case AM_R_D16, AM_D16:
		var lo = uint16(pubsub.BusCtx().BusRead(cpuInstance.Regs.Pc))
		Cm.IncreaseCycle(1)
		var hi = uint16(pubsub.BusCtx().BusRead(cpuInstance.Regs.Pc + 1))
		Cm.IncreaseCycle(1)
		cpuInstance.FetchedData = lo | (hi << 8)
		cpuInstance.Regs.Pc += 2
		return
	case AM_MR_R:
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg2)
		cpuInstance.MemDest = CpuRegRead(cpuInstance.currentInst.Reg1)
		cpuInstance.DestIsMem = true
		if cpuInstance.currentInst.Reg1 == RT_C {
			cpuInstance.MemDest |= 0xFF00
		}

		return
	case AM_R_MR:
		addr := CpuRegRead(cpuInstance.currentInst.Reg2)
		if cpuInstance.currentInst.Reg2 == RT_C {
			addr |= 0xFF00
		}
		cpuInstance.FetchedData = uint16(pubsub.BusCtx().BusRead(addr)) & 0xFF
		Cm.IncreaseCycle(1)
		return
	case AM_R_HLI:
		addr := CpuRegRead(RT_HL)
		cpuInstance.FetchedData = uint16(pubsub.BusCtx().BusRead(addr)) & 0xFF
		Cm.IncreaseCycle(1)
		CpuSetReg(RT_HL, addr+1)
		return

	case AM_R_HLD:
		addr := CpuRegRead(RT_HL)
		cpuInstance.FetchedData = uint16(pubsub.BusCtx().BusRead(addr)) & 0xFF
		Cm.IncreaseCycle(1)
		CpuSetReg(RT_HL, addr-1)
		return

	case AM_HLI_R:
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg2) & 0xFF
		cpuInstance.MemDest = CpuRegRead(RT_HL)
		cpuInstance.DestIsMem = true
		CpuSetReg(RT_HL, cpuInstance.MemDest+1)
		return

	case AM_HLD_R:
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg2) & 0xFF
		cpuInstance.MemDest = CpuRegRead(RT_HL)
		cpuInstance.DestIsMem = true
		CpuSetReg(RT_HL, cpuInstance.MemDest-1)
		return

	case AM_R_A8:
		offset := uint16(pubsub.BusCtx().BusRead(cpuInstance.Regs.Pc))
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc++
		addr := 0xFF00 | offset
		cpuInstance.FetchedData = uint16(pubsub.BusCtx().BusRead(addr)) & 0xFF
		Cm.IncreaseCycle(1)
		return

	case AM_A8_R:
		offset := uint16(pubsub.BusCtx().BusRead(cpuInstance.Regs.Pc))
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc++
		cpuInstance.MemDest = 0xFF00 | offset
		cpuInstance.DestIsMem = true
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg2) & 0xFF
		return
	case AM_HL_SPR:
		e8 := int8(pubsub.BusCtx().BusRead(cpuInstance.Regs.Pc))
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc++
		cpuInstance.FetchedData = uint16(int32(cpuInstance.Regs.Sp) + int32(e8))
		return
	case AM_A16_R, AM_D16_R:
		var lo = uint16(pubsub.BusCtx().BusRead(cpuInstance.Regs.Pc))
		Cm.IncreaseCycle(1)
		var hi = uint16(pubsub.BusCtx().BusRead(cpuInstance.Regs.Pc + 1))
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc += 2
		cpuInstance.MemDest = lo | (hi << 8)
		cpuInstance.DestIsMem = true
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg2) & 0xFF
		return
	case AM_MR_D8:
		cpuInstance.FetchedData = uint16(pubsub.BusCtx().BusRead(cpuInstance.Regs.Pc)) & 0xFF
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc++
		cpuInstance.MemDest = CpuRegRead(cpuInstance.currentInst.Reg1)
		cpuInstance.DestIsMem = true
		return
	case AM_MR:
		cpuInstance.MemDest = CpuRegRead(cpuInstance.currentInst.Reg1)
		cpuInstance.DestIsMem = true
		cpuInstance.FetchedData = uint16(pubsub.BusCtx().BusRead(cpuInstance.MemDest)) & 0xFF
		Cm.IncreaseCycle(1)
		return
	case AM_R_A16:
		var lo = uint16(pubsub.BusCtx().BusRead(cpuInstance.Regs.Pc))
		Cm.IncreaseCycle(1)
		var hi = uint16(pubsub.BusCtx().BusRead(cpuInstance.Regs.Pc + 1))
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc += 2
		addr := lo | (hi << 8)
		cpuInstance.FetchedData = uint16(pubsub.BusCtx().BusRead(addr)) & 0xFF
		Cm.IncreaseCycle(1)
		return
	default:
		log.Warn("Unknown Addressing Mode! %d (%02X)\n", cpuInstance.currentInst.Mode, cpuInstance.CurOpCode)
		return
	}
}
