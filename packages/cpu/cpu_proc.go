package cpu

import (
	"log"

	"pajalic.go.emulator/packages/common"
	"pajalic.go.emulator/packages/emulator"
)

func procNone(ctx *CpuContext) {
	log.Fatal("Invalid Instruction!")
}

func procLd(ctx *CpuContext) {
	if ctx.DestIsMem {
		//LD (BC), A for instance...

		if ctx.currentInst.Reg2 >= RT_AF {
			//if 16 bit register...
			emulator.EmuCycles(1)
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

func procDi(ctx *CpuContext) {
	ctx.IntMasterEnabled = false
}

func procPop(ctx *CpuContext) {
	var lo = uint16(StackPop())
	emulator.EmuCycles(1)
	var hi = uint16(StackPop())
	emulator.EmuCycles(1)
	var n = (hi << 8) | lo
	CpuSetReg(ctx.currentInst.Reg1, n)

	if ctx.currentInst.Reg1 == RT_AF {
		CpuSetReg(ctx.currentInst.Reg1, n&0xFFF0)
	}
}

func procPush(ctx *CpuContext) {
	var hi = (CpuRegRead(ctx.currentInst.Reg1) >> 8) & 0xFF
	emulator.EmuCycles(1)
	StackPush(byte(hi))
	var lo = (CpuRegRead(ctx.currentInst.Reg2)) & 0xFF
	emulator.EmuCycles(1)
	StackPush(byte(lo))
	emulator.EmuCycles(1)
}

func goToAddr(ctx *CpuContext, addr uint16, pushpc bool) {
	if CheckCondition(ctx) {
		if pushpc {
			emulator.EmuCycles(2)
			StackPush16(ctx.Regs.pc)
		}
		ctx.Regs.pc = addr
		emulator.EmuCycles(1)
	}
}

func procJp(ctx *CpuContext) {
	goToAddr(ctx, ctx.FetchedData, false)
}

//Jump relative
func procJr(ctx *CpuContext) {
	var rel = byte(ctx.FetchedData & 0xFF) //casting cause it might be negative
	var addr = ctx.Regs.pc + uint16(rel)
	goToAddr(ctx, addr, false)
}

func procCall(ctx *CpuContext) {
	goToAddr(ctx, ctx.FetchedData, true)
}

func procRet(ctx *CpuContext) {
	if ctx.currentInst.Condition != CT_NONE {
		emulator.EmuCycles(1)
	}

	if CheckCondition(ctx) {
		var lo = uint16(StackPop())
		emulator.EmuCycles(1)

		var hi = uint16(StackPop())
		emulator.EmuCycles(1)

		var n = (hi << 8) | lo
		ctx.Regs.pc = n

		emulator.EmuCycles(1)
	}
}

func procRst(ctx *CpuContext) {
	goToAddr(ctx, uint16(ctx.currentInst.Param), true)
}

func procReti(ctx *CpuContext) {
	ctx.IntMasterEnabled = true
	procRet(ctx)
}

func procLdh(ctx *CpuContext) {
	//Ensure this is proper
	if ctx.currentInst.Reg1 == RT_A {
		CpuSetReg(ctx.currentInst.Reg1, uint16(BusRead(0xFF00|uint16(ctx.Regs.c))))
	} else {
		BusWrite(ctx.MemDest, ctx.Regs.a)
	}
	emulator.EmuCycles(1)
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
	processors[IN_JP] = procJp
	processors[IN_JR] = procJr
	processors[IN_CALL] = procCall
	processors[IN_XOR] = procXor
	processors[IN_POP] = procPop
	processors[IN_PUSH] = procPush
	processors[IN_LDH] = procLdh
	processors[IN_RET] = procRet
	processors[IN_RETI] = procReti
	processors[IN_DI] = procDi
	processors[IN_RST] = procRst
}

//fix this to return properly
func InstGetProccessor(intype InType) InProc {

	return processors[intype]

}

func CpuSetFlags(ctx CpuContext, z *bool, n *bool, h *bool, c *bool) {
	if z != nil {
		common.BitSet(&ctx.Regs.f, 7, z)
	}

	if n != nil {
		common.BitSet(&ctx.Regs.f, 6, n)
	}

	if h != nil {
		common.BitSet(&ctx.Regs.f, 5, h)
	}

	if c != nil {
		common.BitSet(&ctx.Regs.f, 4, c)
	}
}
