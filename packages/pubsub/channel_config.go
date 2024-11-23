package pubsub

// For Subscribers
type Read8BitConfig struct {
	ReadType    OperationType
	processRead Read8BitFunc
}

func (c *Read8BitConfig) processReadEvent(address uint16) byte {
	//Dummy
	return c.processRead(address)
}

func (c *Read8BitConfig) processWriteEvent(address uint16, data byte) {
	//Dummy
}

func (c *Read8BitConfig) getChannelConfig() ChannelConfig[uint16, byte, uint16, byte] {
	return ChannelConfig[uint16, byte, uint16, byte]{
		SubscribeType: c.ReadType,
		PublishType:   NoEvent,
	}
}

// For Subscribers
type Read16BitConfig struct {
	ReadType    OperationType
	processRead Read8BitFunc
}

type Write8BitConfig struct {
	WriteType    OperationType
	ProcessWrite Write8BitFunc
}

func (c *Write8BitConfig) processReadEvent(address uint16) byte {
	//Dummy
	return 0
}

func (c *Write8BitConfig) processWriteEvent(address uint16, data byte) {
	//Dummy
	c.ProcessWrite(address, data)
}

func (c *Write8BitConfig) getChannelConfig() ChannelConfig[uint16, byte, uint16, byte] {
	return ChannelConfig[uint16, byte, uint16, byte]{
		SubscribeType: c.WriteType,
		PublishType:   NoEvent,
	}
}

type Write16BitConfig struct {
	WriteType    OperationType
	ProcessWrite Write8BitFunc
}

type Read8BitFunc func(uint16) byte
type Write8BitFunc func(uint16, byte)

type ReadWrite8BitConfig struct {
	ReadType     OperationType
	WriteType    OperationType
	ProcessRead  Read8BitFunc
	ProcessWrite Write8BitFunc
}

func (c *ReadWrite8BitConfig) processReadEvent(address uint16) byte {
	return c.ProcessRead(address)
}

func (c *ReadWrite8BitConfig) processWriteEvent(address uint16, data byte) {
	c.ProcessWrite(address, data)
}

func (c *ReadWrite8BitConfig) getChannelConfig() ChannelConfig[uint16, byte, uint16, byte] {
	return ChannelConfig[uint16, byte, uint16, byte]{
		SubscribeType: c.ReadType,
		PublishType:   c.WriteType,
	}
}

type Read16BitFunc func(uint16) byte
type Write16BitFunc func(uint16, byte)

type ReadWrite16BitConfig struct {
	ReadType     OperationType
	WriteType    OperationType
	ProcessRead  Read16BitFunc
	ProcessWrite Write16BitFunc
}

func (c *ReadWrite16BitConfig) GetChannelConfig() ChannelConfig[uint16, uint16, uint16, uint16] {
	return ChannelConfig[uint16, uint16, uint16, uint16]{
		SubscribeType: c.ReadType,
		PublishType:   c.WriteType,
	}
}
