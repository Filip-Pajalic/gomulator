package cpu

import (
	log "pajalic.go.emulator/packages/logger"
	"pajalic.go.emulator/packages/memory"
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
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg1)
		return
	//Problem on 2D , fetched data is 57339 instead of 2
	case AM_R_R:
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg2)
		return
	case AM_R_D8:
		cpuInstance.FetchedData = uint16(memory.BusRead(cpuInstance.Regs.Pc))
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc++
		return
	case AM_R_D16, AM_D16:
		/*
			Example:
				lo = (0000000)01010000 (8 bit loaded into 16)
				hi = (0000000)00000110 (8 bit loaded into 16)
				hi << 8 = 000001100000000
				lo | hi = 000001101010000

		*/
		var lo = uint16(memory.BusRead(cpuInstance.Regs.Pc))

		Cm.IncreaseCycle(1)
		var hi = uint16(memory.BusRead(cpuInstance.Regs.Pc + 1))

		Cm.IncreaseCycle(1)
		cpuInstance.FetchedData = lo | (hi << 8)

		cpuInstance.Regs.Pc += 2
		return

	/*
		LD (DE),A
		LD  REG1 REG2
	*/
	case AM_MR_R:
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg2)
		cpuInstance.MemDest = CpuRegRead(cpuInstance.currentInst.Reg1)
		cpuInstance.DestIsMem = true
		/*
			opcode = read(PC++)
			if opcode == 0xE2: //RT_C?
			write(unsigned_16(lsb=C, msb=0xFF), A)
		*/
		if cpuInstance.currentInst.Reg1 == RT_C {
			cpuInstance.MemDest |= 0xFF00
		}

		return

	case AM_R_MR:
		addr := CpuRegRead(cpuInstance.currentInst.Reg2)

		if cpuInstance.currentInst.Reg2 == RT_C {
			addr |= 0xFF00
		}

		cpuInstance.FetchedData = uint16(memory.BusRead(addr))
		Cm.IncreaseCycle(1)
		return
	case AM_R_HLI:
		cpuInstance.FetchedData = uint16(memory.BusRead(CpuRegRead(cpuInstance.currentInst.Reg2)))
		Cm.IncreaseCycle(1)
		CpuSetReg(RT_HL, CpuRegRead(RT_HL)+1)
		return

	case AM_R_HLD:
		cpuInstance.FetchedData = uint16(memory.BusRead(CpuRegRead(cpuInstance.currentInst.Reg2)))
		Cm.IncreaseCycle(1)
		CpuSetReg(RT_HL, CpuRegRead(RT_HL)-1)
		return

	case AM_HLI_R:
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg2)
		cpuInstance.MemDest = CpuRegRead(cpuInstance.currentInst.Reg1)
		cpuInstance.DestIsMem = true
		CpuSetReg(RT_HL, CpuRegRead(RT_HL)+1)
		return

	case AM_HLD_R:
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg2)
		cpuInstance.MemDest = CpuRegRead(cpuInstance.currentInst.Reg1)
		cpuInstance.DestIsMem = true
		CpuSetReg(RT_HL, CpuRegRead(RT_HL)-1)
		return

	case AM_R_A8:
		cpuInstance.FetchedData = uint16(memory.BusRead(cpuInstance.Regs.Pc))
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc++
		return

	case AM_A8_R:
		cpuInstance.MemDest = uint16(memory.BusRead(cpuInstance.Regs.Pc)) | 0xFF00
		cpuInstance.DestIsMem = true
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc++
		return

	case AM_HL_SPR:
		cpuInstance.FetchedData = uint16(memory.BusRead(cpuInstance.Regs.Pc))
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc++
		return

	case AM_D8:
		cpuInstance.FetchedData = uint16(memory.BusRead(cpuInstance.Regs.Pc))
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc++
		return

	case AM_A16_R, AM_D16_R:
		var lo = uint16(memory.BusRead(cpuInstance.Regs.Pc))
		Cm.IncreaseCycle(1)
		var hi = uint16(memory.BusRead(cpuInstance.Regs.Pc + 1))
		Cm.IncreaseCycle(1)
		cpuInstance.MemDest = lo | (hi << 8)
		cpuInstance.DestIsMem = true
		cpuInstance.Regs.Pc += 2
		cpuInstance.FetchedData = CpuRegRead(cpuInstance.currentInst.Reg2)
		return

	case AM_MR_D8:
		cpuInstance.FetchedData = uint16(memory.BusRead(cpuInstance.Regs.Pc))
		Cm.IncreaseCycle(1)
		cpuInstance.Regs.Pc++
		cpuInstance.MemDest = CpuRegRead(cpuInstance.currentInst.Reg1)
		cpuInstance.DestIsMem = true
		return

	case AM_MR:
		cpuInstance.MemDest = CpuRegRead(cpuInstance.currentInst.Reg1)
		cpuInstance.DestIsMem = true
		cpuInstance.FetchedData = uint16(memory.BusRead(CpuRegRead(cpuInstance.currentInst.Reg1)))
		Cm.IncreaseCycle(1)
		return

	case AM_R_A16:
		var lo = uint16(memory.BusRead(cpuInstance.Regs.Pc))
		Cm.IncreaseCycle(1)
		var hi = uint16(memory.BusRead(cpuInstance.Regs.Pc + 1))
		Cm.IncreaseCycle(1)
		var addr = lo | (hi << 8)
		cpuInstance.Regs.Pc += 2
		cpuInstance.FetchedData = uint16(memory.BusRead(addr))
		Cm.IncreaseCycle(1)
		return
	default:
		log.Warn("Unknown Addressing Mode! %d (%02X)\n", cpuInstance.currentInst.Mode, cpuInstance.CurOpCode)
		//os.Exit(1)
		return
	}
}
