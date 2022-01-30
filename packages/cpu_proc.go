package gameboypackage

func procNone(cpucontext *CpuContext) {
	Logger.Fatalf("Invalid Instruction!")
}

func procLd(cpucontext *CpuContext) {
	Logger.Warnf("Not implemented Instruction!")
}

func procNop(cpucontext *CpuContext) {

}

func ProcJp(ctx *CpuContext) {
	if CheckCondition(ctx) {
		ctx.Regs.pc = ctx.FetchedData
		EmuCycles(1)
	}
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

func CpuSetFlags(ctx CpuContext, z *byte, n *byte, h *byte, c *byte) {
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
