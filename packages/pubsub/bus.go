package pubsub

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
}

var busInstance *Bus

func NewBus() *Bus {
	return &Bus{}
}

func BusCtx() *Bus {
	once.Do(func() {
		busInstance = NewBus()
	})
	return busInstance
}

func (m *Bus) BusRead(address uint16) byte {

	config := &Read8BitConfig{
		ReadType: toReadEvent(address),
	}

	PublishChannelRequest(config)

	config2 := &Write8BitConfig{
		WriteType: toWriteEvent(address),
	}

	eventData := <-SubscribeChannelRequest(config2)
	data := eventData.Data
	return data

}

func (m *Bus) BusWrite(address uint16, data byte) {
	/*	writeData := Write8BitData{Address: address, Data: data}
		event := WriteEvent[Write8BitData]{eventType: MemoryWriteEvent, eventData: writeData}
		pbManager := GetPubSubManager()
		pbManager.Publish(event)*/
	// Publish a memory write event
	/*	m.psManager.Publish(EventChannel{
		Operation: toWriteEvent(address),
		Data: struct {
			Address uint16
			Data    byte
		}{address, data},
	})*/

}

func toWriteEvent(address uint16) OperationType {
	if address < 0x8000 {
		return MemoryWriteEvent
		//return CartRead(address)
	} else if address < 0xA000 {
		//Char/Map Data
		return PPUWramWriteEvent
	} else if address < 0xC000 {
		//Cartridge ram //not working
		return MemoryWriteEvent
	} else if address < 0xE000 {
		//WRAM Working ram
		return WramWriteEvent
	} else if address < 0xFE00 {
		// Reserved eco ram, not used
		return 0
	} else if address < 0xFEA0 {
		//Object attribute memory (OAM)
		//TODO
		/*		if cpu.GetDMAContext().DMATransferring() {
				return 0xFF
			}*/
		return PPUOamWriteEvent

	} else if address < 0xFF00 {
		// Reserved not used
		return 0
	} else if address < 0xFF80 {
		//Io registers
		return IoWriteEvent
	} else if address == 0xFFFF {
		//TODO
		//CPU interupt enable register
		//return CpuGetIERegister()
	}

	return HramWriteEvent
}

func toReadEvent(address uint16) OperationType {
	if address < 0x8000 {
		return MemoryReadEvent
		//return CartRead(address)
	} else if address < 0xA000 {
		//Char/Map Data
		return PPUVramReadEvent
	} else if address < 0xC000 {
		//Cartridge ram //not working
		return MemoryReadEvent
	} else if address < 0xE000 {
		//WRAM Working ram
		return WramReadEvent
	} else if address < 0xFE00 {
		// Reserved eco ram, not used
		return 0
	} else if address < 0xFEA0 {
		//Object attribute memory (OAM)
		//TODO
		/*		if cpu.GetDMAContext().DMATransferring() {
				return 0xFF
			}*/
		return PPUOamReadEvent

	} else if address < 0xFF00 {
		// Reserved not used
		return 0
	} else if address < 0xFF80 {
		//Io registers
		return IoReadEvent
	} else if address == 0xFFFF {
		//TODO
		//CPU interupt enable register
		//return CpuGetIERegister()
	}

	return HramReadEvent
}

/*
Writes data from the cartridge, memory locations above represent what the different adresses access
@Param address - where to write
@Param byte - what to write
*/

func oldBusWrite(address uint16, data byte) OperationType {

	if address < 0x8000 {
		//CartWrite(address, data)
	} else if address < 0xA000 {
		//Char/Map Data
		return PPUWramWriteEvent
		//ppu.PpuWramWrite(address, data)
	} else if address < 0xC000 {
		//EXT-RAM
		return MemoryWriteEvent
		//CartWrite(address, data)
	} else if address < 0xE000 {
		//WRAM
		return WramWriteEvent
		//WramWrite(address, data)
	} else if address < 0xFE00 {

		//reserved echo ram
	} else if address < 0xFEA0 {
		//OAM
		return PPUOamWriteEvent
		//TODO
		/*if cpu.GetDMAContext().DMATransferring() {
			//return
		}*/
		//	ppu.PpuOamWrite(address, data)
	} else if address < 0xFF00 {
		//unusable reserved
	} else if address < 0xFF80 {
		//IO Registers
		return IoWriteEvent
		//cpu.IoWrite(address, data)
	} else if address == 0xFFFF {
		//CPU SET ENABLE REGISTER
		//TODO
		//CpuSetIERegister(data)
	} else {
		return HramWriteEvent
		//HramWrite(address, data)
	}
	//TODO
	return 0

}

func (m *Bus) BusWrite16(address uint16, data uint16) {

	m.BusWrite(address+1, byte((data>>8)&0xFF))
	m.BusWrite(address, byte(data&0xFF))

}

func (m *Bus) BusRead16(address uint16) uint16 {
	var lo = uint16(m.BusRead(address))
	var hi = uint16(m.BusRead(address + 1))

	return lo | (hi << 8)
}

// BusManager is responsible for handling channel transactions and events
type BusManager struct {
	Processors []ChannelProcessor[MemoryArchitecture, MemoryArchitecture, MemoryArchitecture, MemoryArchitecture]
}

// AddProcessor adds a ChannelProcessor to the BusManager
func (bm *BusManager) AddProcessor(processor ChannelProcessor[MemoryArchitecture, MemoryArchitecture, MemoryArchitecture, MemoryArchitecture]) {
	bm.Processors = append(bm.Processors, processor)
}

// ProcessAllChannels processes all registered channels and handles incoming events
func (bm *BusManager) ProcessAllChannels() {
	for _, processor := range bm.Processors {
		go ProcessChannelTransactions(processor)
	}
}
