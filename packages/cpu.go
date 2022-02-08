package gameboypackage

import (
	"os"
)

/*
PC - Program counter, where in memory(address) the processor should read from
SP - In 8086, the main "stack register" is called stack pointer. Tracks the operations of the stack and
stores address of the last program request.
F - 8Bit register to store flags, indicates outcome of last operation performed
--- z zero flag, is set if result of operation is zero
--- n subtraction flag, is set if the last operation was a subtraction
--- h half carry flag, carrying result of lower 4 bits
--- c carry flag, When the result of an 8-bit addition is higher than $FF.
				  When the result of a 16-bit addition is higher than $FFFF.
				  When the result of a subtraction or comparison is lower than zero (like in Z80 and 80x86 CPUs, but unlike in 65XX and ARM CPUs).
				  When a rotate/shift operation shifts out a “1” bit.
A - 8bit register , can be combined with F to store 16bits - Accumulator, to contain values or
store results. Can be shifted with a one byte instruction, can be complemented, adjusted, negated with
single byte instruction.  Number you want to add should be in A and other register, result should be in A.
B - 8bit register , can be combined with C to store 16bits, Generally used as a counter whe moving data, can be used for operations
C - 8bit register , can be combined with B to store 16bits, Generally used as a counter whe moving data, can be used for operations
D - 8bit register , can be combined with E to store 16bits, Generally used with E to store 16 bit destination
addresses when moving data. Can be used to other operations.
E - 8bit register , can be combined with D to store 16bits, Generally used with D to store 16 bit destination
addresses when moving data. Can be used to other operations.
H - 8bit register , can be combined with L to store 16bits, Special registers, used as pair with HL for indirect addressing,
instead of specifying an address in an operation you can use HL as the destination
L - 8bit register , can be combined with H to store 16bits Special registers, used as pair with HL for indirect addressing,
instead of specifying an address in an operation you can use HL as the destination

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

	IntMasterEnabled bool

	IERegister byte
}

var CpuCtx CpuContext

func CpuInit() {
	CpuCtx.Regs.pc = 0x100
	CpuCtx.Regs.a = 0x01
	CpuCtx.Halted = false
	InitProcessors()
}

func fetchInstruction() {

	CpuCtx.CurOpCode = BusRead(CpuCtx.Regs.pc)
	CpuCtx.currentInst = instructionByOpcode(CpuCtx.CurOpCode)
	CpuCtx.Regs.pc++

}

func execute() {
	var proc = InstGetProccessor(CpuCtx.currentInst.Type)
	if proc == nil {
		Logger.Warnf("No processor for this execution!")
		return
	}
	proc(&CpuCtx)
}

func CpuStep() bool {
	if !CpuCtx.Halted {
		pc := CpuCtx.Regs.pc
		fetchInstruction()
		FetchData()

		Logger.Debugf("%04X: %-7s (%02X %02X %02X) A: %02X BC: %02X%02X DE: %02X%02X HL: %02X%02X\n",
			pc, getInstructionName(CpuCtx.currentInst.Type), CpuCtx.CurOpCode,
			BusRead(pc+1), BusRead(pc+2), CpuCtx.Regs.a, CpuCtx.Regs.b, CpuCtx.Regs.c,
			CpuCtx.Regs.d, CpuCtx.Regs.e, CpuCtx.Regs.h, CpuCtx.Regs.l)

		if CpuCtx.currentInst == nil {
			Logger.Warnf("Unknown instruction! %02X\n", CpuCtx.CurOpCode)
			os.Exit(1)
		}
		execute()
	}
	return true
}

func CpuGetIERegister() byte {
	return CpuCtx.IERegister
}

//Interupt enable register
func CpuSetIERegister(n byte) {
	CpuCtx.IERegister = n
}
