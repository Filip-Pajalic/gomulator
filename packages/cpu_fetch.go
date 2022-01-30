package gameboypackage

/*
AM - Adressing mode, LD , JR, ADD , RET, POP ,PUSH ..
MR	- Memory address for register
R	- Register
D8	- means immediate 8 bit data
D16	- means immediate 16 bit data
A8	- means 8 bit unsigned data, which are added to $FF00 in certain instructions (replacement for missing IN and OUT instructions)
A16	- means 16 bit address
R8	- means 8 bit signed data, which are added to program counter
*/

func FetchData() {
	CpuCtx.MemDest = 0
	CpuCtx.DestIsMem = false

	switch CpuCtx.currentInst.Mode {
	case AM_IMP:
		return
	case AM_R:
		CpuCtx.FetchedData = CpuRegRead(CpuCtx.currentInst.Reg1)
		return

	case AM_R_R:
		CpuCtx.FetchedData = CpuRegRead(CpuCtx.currentInst.Reg2)
		return
	case AM_R_D8:
		CpuCtx.FetchedData = uint16(BusRead(CpuCtx.Regs.pc))
		EmuCycles(1)
		CpuCtx.Regs.pc++
		return
	case AM_R_D16, AM_D16:
		/*
			Example:
				lo = (0000000)01010000 (8 bit loaded into 16)
				hi = (0000000)00000110 (8 bit loaded into 16)
				hi << 8 = 000001100000000
				lo | hi = 000001101010000

		*/
		var lo = uint16(BusRead(CpuCtx.Regs.pc))
		EmuCycles(1)
		var hi = uint16(BusRead(CpuCtx.Regs.pc + 1))
		EmuCycles(1)
		CpuCtx.FetchedData = lo | (hi << 8)
		CpuCtx.Regs.pc += 2
		return

	/*
		LD (DE),A
		LD  REG1 REG2
	*/
	case AM_MR_R:
		CpuCtx.FetchedData = CpuRegRead(CpuCtx.currentInst.Reg2)
		CpuCtx.MemDest = CpuRegRead(CpuCtx.currentInst.Reg1)
		CpuCtx.DestIsMem = true
		/*
			opcode = read(PC++)
			if opcode == 0xE2: //RT_C?
			write(unsigned_16(lsb=C, msb=0xFF), A)
		*/
		if CpuCtx.currentInst.Reg1 == RT_C {
			CpuCtx.MemDest |= 0xFF00
		}

		return

	case AM_R_MR:
		addr := CpuRegRead(CpuCtx.currentInst.Reg2)
		if CpuCtx.currentInst.Reg1 == RT_C {
			addr |= 0xFF00
		}

		CpuCtx.FetchedData = uint16(BusRead(addr))
		EmuCycles(1)

	case AM_R_HLI:
		CpuCtx.FetchedData = uint16(BusRead(CpuRegRead(CpuCtx.currentInst.Reg2)))
		EmuCycles(1)
		CpuSetReg(RT_HL, CpuRegRead(RT_HL)+1)
		return

	case AM_R_HLD:
		CpuCtx.FetchedData = uint16(BusRead(CpuRegRead(CpuCtx.currentInst.Reg2)))
		EmuCycles(1)
		CpuSetReg(RT_HL, CpuRegRead(RT_HL)-1)
		return

	case AM_HLI_R:
		CpuCtx.FetchedData = CpuRegRead(CpuCtx.currentInst.Reg2)
		CpuCtx.MemDest = CpuRegRead(CpuCtx.currentInst.Reg1)
		CpuCtx.DestIsMem = true
		CpuSetReg(RT_HL, CpuRegRead(RT_HL)+1)
		return

	case AM_HLD_R:
		CpuCtx.FetchedData = CpuRegRead(CpuCtx.currentInst.Reg2)
		CpuCtx.MemDest = CpuRegRead(CpuCtx.currentInst.Reg1)
		CpuCtx.DestIsMem = true
		CpuSetReg(RT_HL, CpuRegRead(RT_HL)-1)
		return

	case AM_R_A8:
		CpuCtx.FetchedData = uint16(BusRead(CpuCtx.Regs.pc))
		EmuCycles(1)
		CpuCtx.Regs.pc++
		return

	case AM_A8_R:
		CpuCtx.MemDest = uint16(BusRead(CpuCtx.Regs.pc)) | 0xFF00
		CpuCtx.DestIsMem = true
		EmuCycles(1)
		CpuCtx.Regs.pc++
		return

	case AM_HL_SPR:
		CpuCtx.FetchedData = uint16(BusRead(CpuCtx.Regs.pc))
		EmuCycles(1)
		CpuCtx.Regs.pc++
		return

	case AM_D8:
		CpuCtx.FetchedData = uint16(BusRead(CpuCtx.Regs.pc))
		EmuCycles(1)
		CpuCtx.Regs.pc++
		return

	case AM_A16_R, AM_D16_R:
		var lo = uint16(BusRead(CpuCtx.Regs.pc))
		EmuCycles(1)
		var hi = uint16(BusRead(CpuCtx.Regs.pc + 1))
		EmuCycles(1)
		CpuCtx.MemDest = lo | (hi << 8)
		CpuCtx.DestIsMem = true
		CpuCtx.Regs.pc += 2
		CpuCtx.FetchedData = CpuRegRead(CpuCtx.currentInst.Reg2)
		return

	case AM_MR_D8:
		CpuCtx.FetchedData = uint16(BusRead(CpuCtx.Regs.pc))
		EmuCycles(1)
		CpuCtx.Regs.pc++
		CpuCtx.MemDest = CpuRegRead(CpuCtx.currentInst.Reg1)
		CpuCtx.DestIsMem = true
		return

	case AM_MR:
		CpuCtx.MemDest = CpuRegRead(CpuCtx.currentInst.Reg1)
		CpuCtx.DestIsMem = true
		CpuCtx.FetchedData = uint16(BusRead(CpuRegRead(CpuCtx.currentInst.Reg1)))
		EmuCycles(1)
		return

	case AM_R_A16:
		var lo = uint16(BusRead(CpuCtx.Regs.pc))
		EmuCycles(1)
		var hi = uint16(BusRead(CpuCtx.Regs.pc + 1))
		EmuCycles(1)
		var addr = lo | (hi << 8)
		CpuCtx.Regs.pc += 2
		CpuCtx.FetchedData = uint16(BusRead(addr))
		EmuCycles(1)
		return
	default:
		Logger.Warnf("Unknown adressing mode! %d\n", CpuCtx.currentInst.Mode)
		//os.Exit(1)
		return
	}
}
