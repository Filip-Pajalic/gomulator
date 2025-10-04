package common

func BitSet(a byte, n byte, on bool) byte {
	if on {
		a |= 1 << n
	} else {
		a &= ^(1 << n)
	}
	return a
}

func Bit(a byte, n byte) bool {
	return (a & (1 << n)) != 0
}

func Between(a byte, b byte, c byte) bool {
	return a >= b && a <= c
}

func Between16(a uint16, b uint16, c uint16) bool {
	return a >= b && a <= c
}

func Reverse(n uint16) uint16 {
	return ((n & 0xFF00) >> 8) | ((n & 0x00FF) << 8)
}
