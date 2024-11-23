package pubsub

func toReadEvent(address uint16) OperationType {
	if address < 0x8000 {
		return MemoryReadEvent
	} else if address < 0xA000 {
		return PPUVramReadEvent
	} else if address < 0xC000 {
		return MemoryReadEvent
	} else if address < 0xE000 {
		return WramReadEvent
	} else if address < 0xFE00 {
		// Reserved echo RAM, not used
		return NoEvent
	} else if address < 0xFEA0 {
		return PPUOamReadEvent
	} else if address < 0xFF00 {
		// Reserved, not used
		return NoEvent
	} else if address < 0xFF80 {
		return IoReadEvent
	} else if address == 0xFFFF {
		// Interrupt Enable Register
		return IoReadEvent
	}
	return HramReadEvent
}

func toWriteEvent(address uint16) OperationType {
	if address < 0x8000 {
		return MemoryWriteEvent
	} else if address < 0xA000 {
		return PPUWramWriteEvent
	} else if address < 0xC000 {
		return MemoryWriteEvent
	} else if address < 0xE000 {
		return WramWriteEvent
	} else if address < 0xFE00 {
		// Reserved echo RAM, not used
		return NoEvent
	} else if address < 0xFEA0 {
		return PPUOamWriteEvent
	} else if address < 0xFF00 {
		// Reserved, not used
		return NoEvent
	} else if address < 0xFF80 {
		return IoWriteEvent
	} else if address == 0xFFFF {
		// Interrupt Enable Register
		return IoWriteEvent
	}
	return HramWriteEvent
}
