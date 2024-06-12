package cpu

import (
	"fmt"
	"strconv"
)

func BitSet(a byte, n byte, on *bool) byte {
	if *on {
		a |= 1 << n
	} else {
		a &= ^(1 << n)
	}
	return a
}

// BIT(a, n) ((a & (1 << n)) ? 1 : 0)
func Bit(a byte, n byte) bool {
	//om > 0 om detta ej fungerar
	if (a & (1 << n)) == (1 << n) {
		return true
	}
	return false
}

func Between(a byte, b byte, c byte) bool {
	if (a >= b) && (a <= c) {
		return true
	}
	return false
}

func Between16(a uint16, b uint16, c uint16) bool {
	if (a >= b) && (a <= c) {
		return true
	}
	return false
}

func Reverse(n uint16) uint16 {
	return ((n & 0xFF00) >> 8) | ((n & 0x00FF) << 8)
}

func IntToHex(n uint16) uint16 {
	h, _ := strconv.ParseUint(fmt.Sprintf("%02x", n), 0, 64)
	return uint16(h)
}
