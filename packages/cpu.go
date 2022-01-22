package gameboypackage

import (
	"fmt"
	"os"
)

/*
PC - Program counter, where in memory(address) the processor should read from
SP - In 8086, the main "stack register" is called stack pointer. Tracks the operations of the stack and
stores address of the last program request.
F - Can be used to store results of math operations
A - 8bit register , can be combined with F to store 16bits - AF
B - 8bit register , can be combined with C to store 16bits - CD
C - 8bit register , can be combined with B to store 16bits - CD
D - 8bit register , can be combined with E to store 16bits - DE
E - 8bit register , can be combined with D to store 16bits - DE
H - 8bit register , can be combined with L to store 16bits - HL
L - 8bit register , can be combined with H to store 16bits - HL
*/

type CpuRegisters struct {
	a  byte
	f  byte
	b  byte
	c  byte
	d  byte
	e  byte
	h  byte
	l  byte
	pc uint16
	sp uint16
}

type CpuContext struct {
	Regs CpuRegisters

	//Current fetch
	FetchedData uint16
	MemDest     uint16
	DestIsMem   bool
	CurOpCode   byte
	currentInst *Instruction

	Halted   bool
	Stepping bool
}

var CpuCtx CpuContext

func CpuInit() {
	CpuCtx.Regs.pc = 0x100
	CpuCtx.Halted = false

}

func fetchInstruction() {
	CpuCtx.CurOpCode = byte(BusRead(CpuCtx.Regs.pc))
	CpuCtx.Regs.pc++
	CpuCtx.currentInst = instructionByOpcode(CpuCtx.CurOpCode)
	//possible reflection here instead if needed
	if CpuCtx.currentInst.Type == 0 {
		fmt.Printf("Unknown instruction! %02X\n", CpuCtx.CurOpCode)
		os.Exit(7)
	}

}

func fetchData() {
	CpuCtx.MemDest = 0
	CpuCtx.DestIsMem = false

	switch CpuCtx.currentInst.Mode {
	case AM_IMP:
		return
	case AM_R:
		CpuCtx.FetchedData = CpuRegRead(CpuCtx.currentInst.Reg1)
		return
	case AM_R_D8:
		CpuCtx.FetchedData = uint16(BusRead(CpuCtx.Regs.pc))
		EmuCycles(1)
		CpuCtx.Regs.pc++
		return
	case AM_D16:
		var lo = uint16(BusRead(CpuCtx.Regs.pc))
		EmuCycles(1)
		var hi = uint16(BusRead(CpuCtx.Regs.pc + 1))
		EmuCycles(1)
		CpuCtx.FetchedData = lo | (hi << 8) // how does this work
		CpuCtx.Regs.pc += 2
		return
	default:
		fmt.Printf("Unknown adressing mode! %d\n", CpuCtx.currentInst.Mode)
		os.Exit(7)
		return
	}
}

func execute() {

}

func CpuStep() bool {
	if !CpuCtx.Halted {
		pc := CpuCtx.Regs.pc
		fetchInstruction()
		fetchData()
		fmt.Printf("Executing instruction: %02X   PC: %04X\n", CpuCtx.CurOpCode, pc)
		execute()
	}
	return true
}
