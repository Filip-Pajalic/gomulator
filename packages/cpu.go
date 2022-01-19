package gameboypackage

/*
PC - Program counter, where in memory(address) the processor should read from
SP - In 8086, the main "stack register" is called stack pointer. Tracks the operations of the stack and
stores address of the last program request.

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
	FetchData uint16
	MemDest   uint16
	CurOpCode byte

	Halted   bool
	Stepping bool
}

func CpuInit() {

}

func CpuStep() bool {
	return false
}
