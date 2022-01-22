package gameboypackage

import (
	"fmt"
	"os"
)

func procNone(cpucontext *CpuContext) {
	fmt.Printf("Invalid Instruction!")
	os.Exit(7)
}

func procLd(cpucontext *CpuContext) {
	fmt.Printf("Invalid Instruction!")
	os.Exit(7)
}

func procNop(cpucontext *CpuContext) {

}

func procName(ctx *CpuContext) {
	fmt.Println("iproc")

}

//Function pointer MAP
type InProc func(ctx *CpuContext)

var processors = make(map[string]InProc)

func initProcessors() {
	processors["PROC_NONE"] = procNone
	processors["PROC_NOP"] = procNop
	processors["PROC_LD"] = procLd
	processors["PROC_NAME"] = procName
}

func CpuSetFlags(ctx *CpuContext, z *byte, n *byte, h *byte, c *byte) {
	if z != nil {
		BitSet(ctx.Regs.f, 7, z)
	}

	if n != nil {
		BitSet(ctx.Regs.f, 6, n)
	}

	if h != nil {
		BitSet(ctx.Regs.f, 5, h)
	}

	if c != nil {
		BitSet(ctx.Regs.f, 4, c)
	}
}
