package cpu

import (
	"os"

	log "pajalic.go.emulator/packages/logger"
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
	A  byte
	F  byte
	B  byte
	C  byte
	D  byte
	E  byte
	H  byte
	L  byte
	Pc uint16
	Sp uint16
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
	enablingIme      bool
	IERegister       byte
	IntFlags         byte
}

var CpuCtx CpuContext

func CpuInit() {
	CpuCtx.Regs.Pc = 0x100
	CpuCtx.Halted = false

	CpuCtx.Regs.Sp = 0xFFFE
	CpuCtx.Regs.F = 0xB0
	CpuCtx.Regs.A = 0x01
	CpuCtx.Regs.C = 0x13
	CpuCtx.Regs.B = 0x00
	CpuCtx.Regs.E = 0xD8
	CpuCtx.Regs.D = 0x00
	CpuCtx.Regs.L = 0x4D
	CpuCtx.Regs.H = 0x01
	CpuCtx.IERegister = 0
	CpuCtx.IntFlags = 0
	CpuCtx.IntMasterEnabled = false
	CpuCtx.enablingIme = false

	GetTimerContext().div = 0xABCC

	InitProcessors()
}

func NewCpu(cpu CpuContext) {
	CpuCtx = cpu

	GetTimerContext().div = 0xABCC

	InitProcessors()
}

func fetchInstruction() {
	CpuCtx.CurOpCode = BusRead(CpuCtx.Regs.Pc)
	CpuCtx.Regs.Pc++
	CpuCtx.currentInst = instructionByOpcode(CpuCtx.CurOpCode)
}

func Execute() {
	var proc = InstGetProccessor(CpuCtx.currentInst.Type)
	if proc == nil {
		log.Warn("No processor for this execution!")
		return
	}
	proc(&CpuCtx)
}

func CpuStep() bool {
	if !CpuCtx.Halted {
		pc := CpuCtx.Regs.Pc
		fetchInstruction()
		EmuCycles(1)
		FetchData()

		var z = "-"
		var n = "-"
		var h = "-"
		var c = "-"
		if (CpuCtx.Regs.F & (1 << 7)) >= 1 {
			z = "Z"
		}

		if CpuCtx.Regs.F&(1<<6) >= 1 {
			n = "N"
		}

		if CpuCtx.Regs.F&(1<<5) >= 1 {
			h = "H"
		}

		if CpuCtx.Regs.F&(1<<4) >= 1 {
			c = "C"
		}

		var inst string
		instToStr(&CpuCtx, &inst)
		temp := GetEmuContext().Ticks
		log.Info("%08X - %04X: %-12s (%02X %02X %02X) A: %02X  F: %s%s%s%s BC: %02X%02X DE: %02X%02X HL: %02X%02X\n",
			temp,
			pc, inst, CpuCtx.CurOpCode,
			BusRead(pc+1), BusRead(pc+2), CpuCtx.Regs.A, z, n, h, c, CpuCtx.Regs.B, CpuCtx.Regs.C,
			CpuCtx.Regs.D, CpuCtx.Regs.E, CpuCtx.Regs.H, CpuCtx.Regs.L)

		if CpuCtx.currentInst == nil {
			log.Warn("Unknown instruction! %02X\n", CpuCtx.CurOpCode)
			os.Exit(1)
		}

		DbgUpdate()
		if !DbgPrint() {
			return false
		}
		Execute()
	} else {
		EmuCycles(1)

		if CpuCtx.IntFlags == 1 {
			CpuCtx.Halted = false
		}

	}
	if CpuCtx.IntMasterEnabled {
		CpuHandleInterrupts(&CpuCtx)
		CpuCtx.enablingIme = false
	}
	if CpuCtx.enablingIme {
		CpuCtx.IntMasterEnabled = true
	}

	return true
}

func CpuGetIERegister() byte {
	return CpuCtx.IERegister
}

// Interupt enable register
func CpuSetIERegister(n byte) {
	CpuCtx.IERegister = n
}

func CpuRequestInterrupt(t InterruptType) {
	CpuCtx.IntFlags |= byte(t)
}
