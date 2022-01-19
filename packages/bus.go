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

func BusRead(address uint16) byte {

	if address < 0x8000 {
		return CartRead(address)
	}

	var temporary byte
	return temporary

}
func BusWrite(address uint16, data byte) {

	if address < 0x8000 {
		CartWrite(address, data)
	}

}
