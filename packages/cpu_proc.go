package gameboypackage

func procNone(ctx *CpuContext) {
	Logger.Fatalf("Invalid Instruction!")
}

func procLd(ctx *CpuContext) {
	if ctx.DestIsMem {
		//LD (BC), A for instance...

		if ctx.currentInst.Reg2 >= RT_AF {
			//if 16 bit register...
			EmuCycles(1)
			BusWrite16(ctx.MemDest, ctx.FetchedData)
		} else {
			BusWrite(ctx.MemDest, byte(ctx.FetchedData))
		}

		return
	}

	if ctx.currentInst.Mode == AM_HL_SPR {
		var hflag = (CpuRegRead(ctx.currentInst.Reg2)&0xF)+
			(ctx.FetchedData&0xF) >= 0x10

		var cflag = (CpuRegRead(ctx.currentInst.Reg2)&0xFF)+
			(ctx.FetchedData&0xFF) >= 0x100

		CpuSetFlags(*ctx, nil, nil, &hflag, &cflag)
		CpuSetReg(ctx.currentInst.Reg1,
			CpuRegRead(ctx.currentInst.Reg2)+ctx.FetchedData)

		return
	}

	CpuSetReg(ctx.currentInst.Reg1, ctx.FetchedData)
}

func procNop(cpucontext *CpuContext) {

}

func ProcDi(ctx *CpuContext) {
	ctx.IntMasterEnabled = false
}

func ProcJp(ctx *CpuContext) {
	if CheckCondition(ctx) {
		ctx.Regs.pc = ctx.FetchedData
		EmuCycles(1)
	}
}

func procXor(ctx *CpuContext) {
	ctx.Regs.a ^= byte(ctx.FetchedData & 0xFF)
	var zflag = false
	if ctx.Regs.a == 0 {
		zflag = true
	}
	CpuSetFlags(*ctx, &zflag, nil, nil, nil)

}

func CheckCondition(ctx *CpuContext) bool {
	z := CpuFlagZ()
	c := CpuFlagC()

	switch CpuCtx.currentInst.Condition {
	case CT_NONE:
		return true
	case CT_C:
		return c
	case CT_NC:
		return !c
	case CT_Z:
		return z
	case CT_NZ:
		return !z

	}
	return false

}

//Function pointer MAP
type InProc func(ctx *CpuContext)

var processors = make(map[InType]InProc)

func InitProcessors() {
	processors[IN_NONE] = procNone
	processors[IN_NOP] = procNop
	processors[IN_LD] = procLd
	processors[IN_JP] = ProcJp
}

func InstGetProccessor(intype InType) InProc {

	if val, ok := processors[intype]; ok {
		return val
	} else {
		return processors[IN_NONE]
	}

}

func CpuSetFlags(ctx CpuContext, z *bool, n *bool, h *bool, c *bool) {
	if z != nil {
		BitSet(&ctx.Regs.f, 7, z)
	}

	if n != nil {
		BitSet(&ctx.Regs.f, 6, n)
	}

	if h != nil {
		BitSet(&ctx.Regs.f, 5, h)
	}

	if c != nil {
		BitSet(&ctx.Regs.f, 4, c)
	}
}
