package cpu

import (
	logger "app/internal/logger"
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

type Bus interface {
	BusRead(address uint16) byte
	BusWrite(address uint16, data byte)
}

type CPU interface {
	Fetch()
	Step() bool
	Execute()
	SetIERegister(b byte)
	GetIERegister() byte
	IsStopped() bool
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
	Stopped  bool
	Stepping bool

	IntMasterEnabled bool
	enablingIme      bool
	iERegister       byte
	IntFlags         byte
	memoryBus        Bus
}

var cpuInstance *CpuContext

func NewCpuContext(memoryBus Bus) *CpuContext {

	//TimerCtx().div = 0xABCC TODO
	InitInstructions()
	InitProcessors()
	cpuInstance = &CpuContext{
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
		Stopped:          false,
		Stepping:         false,
		IntMasterEnabled: false,
		enablingIme:      false,
		iERegister:       0,
		IntFlags:         0,
		memoryBus:        memoryBus,
	}
	return cpuInstance
}

func CpuCtx() *CpuContext {
	if cpuInstance == nil {
		logger.Fatal("Create new Cpu with NewCpuContext")
		os.Exit(1)
	}
	return cpuInstance
}

func (c *CpuContext) Fetch() {
	c.CurOpCode = c.memoryBus.BusRead(c.Regs.Pc)
	c.Regs.Pc++
	c.currentInst = instructionByOpcode(c.CurOpCode)
}

func (c *CpuContext) Execute() {
	var proc = InstGetProccessor(c.currentInst.Type)
	if proc == nil {
		logger.Warn("No processor for this execution!")
		return
	}
	proc(c)
}

// This should probably not call the emulator
func (c *CpuContext) Step() bool {
	if c.Stopped {
		return false
	}

	if !c.Halted {
		c.Fetch()
		Cm.IncreaseCycle(1)
		FetchData()

		var inst string
		instToStr(c, &inst)

		if c.currentInst == nil {
			logger.Warn("Unknown instruction! %02X\n", c.CurOpCode)
			os.Exit(1)
		}

		DbgUpdate()
		if !DbgPrint() {
			return false
		}

		c.Execute()
	} else {
		Cm.IncreaseCycle(1)
		if c.IntFlags != 0 {
			c.Halted = false
		}
	}

	if c.Stopped {
		return false
	}

	// Handle interrupts AFTER instruction execution (reference implementation order)
	if c.IntMasterEnabled {
		CpuHandleInterrupts(c)
		c.enablingIme = false
	}

	// Handle EI instruction: enable interrupts now if EI was executed
	if c.enablingIme {
		logger.Debug("IME enabled at PC=%04X", c.Regs.Pc)
		c.IntMasterEnabled = true
	}

	return true
}

func (c *CpuContext) GetIERegister() byte {
	return c.iERegister
}

// Interupt enable register
func (c *CpuContext) SetIERegister(n byte) {
	c.iERegister = n
}

func (c *CpuContext) RequestInterrupt(t InterruptType) {
	c.IntFlags |= byte(t)
}

func (c *CpuContext) IsStopped() bool {
	return c.Stopped
}
