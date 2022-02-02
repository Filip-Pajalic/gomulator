package gameboypackage

type RamContext struct {
	Wram [0x2000]byte
	Hram [0x80]byte
}

var RamCtx RamContext

func WramRead(address uint16) byte {
	address -= 0xC000

	if address >= 0x2000 {
		Logger.Errorf("INVALID WRAM ADDR %08X\n", address+0xC000)
	}

	return RamCtx.Wram[address]
}

func WramWrite(address uint16, value byte) {
	address -= 0xC000

	RamCtx.Wram[address] = value
}

func HramRead(address uint16) byte {
	address -= 0xFF80

	return RamCtx.Hram[address]
}

func HramWrite(address uint16, value byte) {
	address -= 0xFF80

	RamCtx.Hram[address] = value
}
