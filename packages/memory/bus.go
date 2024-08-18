package memory

import (
	"pajalic.go.emulator/packages/pubsub"
)

/*
0x0000	0x3FFF	16 KiB ROM bank 00	From cartridge, usually a fixed bank
0x4000	0x7FFF	16 KiB ROM Bank 01~NN	From cartridge, switchable bank via mapper (if any)
0x8000	0x9FFF	8 KiB Video RAM (VRAM)	In CGB mode, switchable bank 0/1
0xA000	0xBFFF	8 KiB External RAM	From cartridge, switchable bank if any
0xC000	0xCFFF	4 KiB Work RAM (WRAM) Ram bank 0 Cartridge
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

type Bus struct {
	psManager *pubsub.Manager
}

func NewBus(psManager *pubsub.Manager) *Bus {
	return &Bus{
		psManager: psManager,
	}
}

func (m *Bus) BusRead(address uint16) byte {
	BusRead(address)

	// Publish a memory read event
	m.psManager.Publish(pubsub.Event{
		Type: pubsub.MemoryReadEvent,
		Data: address,
	})

	// Continue with existing logic
	return 0 // Replace with actual return value
}

func (m *Bus) BusWrite(address uint16, data byte) {
	BusWrite(address, data)

	// Publish a memory write event
	m.psManager.Publish(pubsub.Event{
		Type: pubsub.MemoryWriteEvent,
		Data: struct {
			Address uint16
			Data    byte
		}{address, data},
	})

	// Continue with existing logic
}

func BusRead(address uint16) pubsub.EventType {
	if address < 0x8000 {
		return pubsub.MemoryReadEvent
		//return CartRead(address)
	} else if address < 0xA000 {
		//Char/Map Data
		return pubsub.PPUWramReadEvent
	} else if address < 0xC000 {
		//Cartridge ram //not working
		return pubsub.MemoryReadEvent
	} else if address < 0xE000 {
		//WRAM Working ram
		return pubsub.WramReadEvent
	} else if address < 0xFE00 {
		// Reserved eco ram, not used
		return 0
	} else if address < 0xFEA0 {
		//Object attribute memory (OAM)
		//TODO
		/*		if cpu.GetDMAContext().DMATransferring() {
				return 0xFF
			}*/
		return pubsub.PPUOamReadEvent

	} else if address < 0xFF00 {
		// Reserved not used
		return 0
	} else if address < 0xFF80 {
		//Io registers
		return pubsub.IoReadEvent
	} else if address == 0xFFFF {
		//TODO
		//CPU interupt enable register
		//return CpuGetIERegister()
	}

	return pubsub.HramReadEvent
}

/*
Writes data from the cartridge, memory locations above represent what the different adresses access
@Param address - where to write
@Param byte - what to write
*/

func BusWrite(address uint16, data byte) pubsub.EventType {

	if address < 0x8000 {
		CartWrite(address, data)
	} else if address < 0xA000 {
		//Char/Map Data
		return pubsub.PPUWramWriteEvent
		//ppu.PpuWramWrite(address, data)
	} else if address < 0xC000 {
		//EXT-RAM
		return pubsub.MemoryWriteEvent
		//CartWrite(address, data)
	} else if address < 0xE000 {
		//WRAM
		return pubsub.WramWriteEvent
		//WramWrite(address, data)
	} else if address < 0xFE00 {

		//reserved echo ram
	} else if address < 0xFEA0 {
		//OAM
		return pubsub.PPUOamWriteEvent
		//TODO
		/*if cpu.GetDMAContext().DMATransferring() {
			//return
		}*/
		//	ppu.PpuOamWrite(address, data)
	} else if address < 0xFF00 {
		//unusable reserved
	} else if address < 0xFF80 {
		//IO Registers
		return pubsub.IoWriteEvent
		//cpu.IoWrite(address, data)
	} else if address == 0xFFFF {
		//CPU SET ENABLE REGISTER
		//TODO
		//CpuSetIERegister(data)
	} else {
		return pubsub.HramWriteEvent
		//HramWrite(address, data)
	}
	//TODO
	return 0

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
