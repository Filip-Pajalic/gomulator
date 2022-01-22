package gameboypackage

func BitSet(a byte, n byte, on *byte) {
	if on != nil {
		a |= 1 << n
	} else {
		a &= ^(1 << n)
	}
}
