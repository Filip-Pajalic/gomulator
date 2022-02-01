package gameboypackage

/*
0x0000	0x3FFF	16 KiB ROM bank 00	From cartridge, usually a fixed bank
0x4000	0x7FFF	16 KiB ROM Bank 01~NN	From cartridge, switchable bank via mapper (if any)
0x8000	0x9FFF	8 KiB Video RAM (VRAM)	In CGB mode, switchable bank 0/1
0xA000	0xBFFF	8 KiB External RAM	From cartridge, switchable bank if any
0xC000	0xCFFF	4 KiB Work RAM (WRAM)
0xD000	0xDFFF	4 KiB Work RAM (WRAM)	In CGB mode, switchable bank 1~7
0xE000	0xFDFF	Mirror of C000~DDFF (ECHO RAM)	Nintendo says use of this area is prohibited.
0xFE00	0xFE9F	Sprite attribute table (OAM)
0xFEA0	0xFEFF	Not Usable	Nintendo says use of this area is prohibited
0xFF00	0xFF7F	I/O Registers
0xFF80	0xFFFE	High RAM (HRAM)
0xFFFF	0xFFFF	Interrupt Enable register (IE)
*/

/*
Reads data from the cartridge, memory locations above represent what the different adresses access
@Param address - where to read
@return byte with value read from memory
*/

func BusRead(address uint16) byte {
	if address < 0x8000 {
		return CartRead(address)
	}
	// Further development needed for other memory locations
	Logger.Warnf("UNSUPPORTED BusRead(%04X)\n", address)

	return byte(0)
}

/*
Writes data from the cartridge, memory locations above represent what the different adresses access
@Param address - where to write
@Param byte - what to write
*/

func BusWrite(address uint16, data byte) {

	if address < 0x8000 {
		CartWrite(address, data)
	}

	Logger.Warnf("UNSUPPORTED BusWrite(%04X)\n", address)

}

func BusWrite16(address uint16, data uint16) {

	BusWrite(address+1, byte((data>>8)&0xFF))
	BusWrite(address, byte(data&0xFF))

}

func BusRead16(address uint16) uint16 {
	var lo = uint16(BusRead(address))
	var hi = uint16(BusRead(address + 1))

	return lo | (hi << 8)
}
