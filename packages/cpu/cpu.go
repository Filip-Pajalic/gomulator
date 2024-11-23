package cpu

import (
	"os"
	log "pajalic.go.emulator/packages/logger"
	"pajalic.go.emulator/packages/pubsub"
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

type CPU interface {
	Fetch()
	Step() bool
	Execute()
	setIERegister(b byte)
	getIERegister() byte
}

type ExternalPins interface {
	RequestInterrupt(t InterruptType)
}

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
	iERegister       byte
	IntFlags         byte
}

var cpuInstance *CpuContext

func NewCpuContext() *CpuContext {

	//GetTimerContext().div = 0xABCC TODO
	InitInstructions()
	InitProcessors()
	return &CpuContext{
		Regs: CpuRegisters{
			A:  0x01,
			F:  0xB0,
			B:  0x00,
			C:  0x13,
			D:  0x00,
			E:  0xD8,
			H:  0x01,
			L:  0x4D,
			Pc: 0x100,
			Sp: 0xFFFE,
		},
		FetchedData:      0,
		MemDest:          0,
		DestIsMem:        false,
		CurOpCode:        0,
		currentInst:      nil,
		Halted:           false,
		Stepping:         false,
		IntMasterEnabled: false,
		enablingIme:      false,
		iERegister:       0,
		IntFlags:         0,
	}
}

func CpuCtx() *CpuContext {
	if cpuInstance == nil {

		cpuInstance = NewCpuContext()
	}
	return cpuInstance
}

func (c *CpuContext) Fetch() {
	//pubsub.GetPubSubManager().Subscribe(pubsub.PPUVramReadEvent)
	c.CurOpCode = pubsub.BusCtx().BusRead(c.Regs.Pc)
	c.Regs.Pc++
	c.currentInst = instructionByOpcode(c.CurOpCode)
}

func (c *CpuContext) Execute() {
	var proc = InstGetProccessor(c.currentInst.Type)
	if proc == nil {
		log.Warn("No processor for this execution!")
		return
	}
	proc(c)
}

// This should probably not call the emulator
func (c *CpuContext) Step() bool {
	if !c.Halted {
		pc := c.Regs.Pc
		c.Fetch()
		Cm.IncreaseCycle(1)
		FetchData()

		var zf = "-"
		var nf = "-"
		var hf = "-"
		var cf = "-"
		if (c.Regs.F & (1 << 7)) != 0 {
			zf = "Z"
		}

		if c.Regs.F&(1<<6) != 0 {
			nf = "N"
		}

		if c.Regs.F&(1<<5) != 0 {
			cf = "H"
		}

		if c.Regs.F&(1<<4) != 0 {
			cf = "C"
		}

		var inst string
		instToStr(c, &inst)
		log.Info("%08X - %04X: %-12s (%02X %02X %02X) A: %02X  F: %s%s%s%s BC: %02X%02X DE: %02X%02X HL: %02X%02X\n",
			Cm.GetCycleTicks(),
			pc, inst, c.CurOpCode,
			pubsub.BusCtx().BusRead(pc+1), pubsub.BusCtx().BusRead(pc+2), c.Regs.A, zf, nf, hf, cf, c.Regs.B, c.Regs.C,
			c.Regs.D, c.Regs.E, c.Regs.H, c.Regs.L)

		if c.currentInst == nil {
			log.Warn("Unknown instruction! %02X\n", c.CurOpCode)
			os.Exit(1)
		}

		//DbgUpdate()
		/*	if !DbgPrint() {
				return false
			}
		*/c.Execute()
	} else {
		Cm.IncreaseCycle(1)

		if c.IntFlags != 0 {
			c.Halted = false
		}

	}
	if c.IntMasterEnabled {
		CpuHandleInterrupts(c)
		c.enablingIme = false
	}
	if c.enablingIme {
		c.IntMasterEnabled = true
	}

	return true
}

func (c *CpuContext) getIERegister() byte {
	return c.iERegister
}

// Interupt enable register
func (c *CpuContext) setIERegister(n byte) {
	c.iERegister = n
}

func (c *CpuContext) RequestInterrupt(t InterruptType) {
	c.IntFlags |= byte(t)
}
