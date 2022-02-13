package gameboypackage

func StackPush(data byte) {
	CpuGetRegs().sp--
	BusWrite(CpuGetRegs().sp, data)

}

func StackPush16(data uint16) {
	StackPush(byte((data >> 8) & 0xFF))
	StackPush(byte(data & 0xFF))
}

func StackPop() byte {
	CpuGetRegs().sp++
	return BusRead(CpuGetRegs().sp)
}

func StackPop16() uint16 {
	var lo = uint16(StackPop())
	var hi = uint16(StackPop())

	return (hi << 8) | lo
}
