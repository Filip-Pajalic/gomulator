package pubsub

// Read-only 8-bit configuration
type Read8BitConfig struct {
	config ChannelConfig[uint16, byte]
}

func NewRead8BitConfig(readType OperationType, processRead Read8BitFunc) *Read8BitConfig {
	return &Read8BitConfig{
		config: ChannelConfig[uint16, byte]{
			ReadType:    readType,
			ProcessRead: processRead,
		},
	}
}

func (c *Read8BitConfig) ProcessReadEvent(address uint16) byte {
	if c.config.ProcessRead != nil {
		return c.config.ProcessRead(address)
	}
	// Handle nil processRead gracefully if needed
	return 0
}

func (c *Read8BitConfig) GetChannelConfig() ChannelConfig[uint16, byte] {
	return c.config
}

// Write-only 8-bit configuration
type Write8BitConfig struct {
	config ChannelConfig[uint16, byte]
}

func NewWrite8BitConfig(writeType OperationType, processWrite Write8BitFunc) *Write8BitConfig {
	return &Write8BitConfig{
		config: ChannelConfig[uint16, byte]{
			WriteType:    writeType,
			ProcessWrite: processWrite,
		},
	}
}

func (c *Write8BitConfig) ProcessWriteEvent(address uint16, data byte) {
	if c.config.ProcessWrite != nil {
		c.config.ProcessWrite(address, data)
	}
	// Handle nil processWrite gracefully if needed
}

func (c *Write8BitConfig) GetChannelConfig() ChannelConfig[uint16, byte] {
	return c.config
}

// Read-write 8-bit configuration
type ReadWrite8BitConfig struct {
	config ChannelConfig[uint16, byte]
}

func NewReadWrite8BitConfig(readType OperationType, writeType OperationType, processRead Read8BitFunc, processWrite Write8BitFunc) *ReadWrite8BitConfig {
	return &ReadWrite8BitConfig{
		config: ChannelConfig[uint16, byte]{
			ReadType:     readType,
			WriteType:    writeType,
			ProcessRead:  processRead,
			ProcessWrite: processWrite,
		},
	}
}

func (c *ReadWrite8BitConfig) ProcessReadEvent(address uint16) byte {
	if c.config.ProcessRead != nil {
		return c.config.ProcessRead(address)
	}
	return 0
}

func (c *ReadWrite8BitConfig) ProcessWriteEvent(address uint16, data byte) {
	if c.config.ProcessWrite != nil {
		c.config.ProcessWrite(address, data)
	}
}

func (c *ReadWrite8BitConfig) GetChannelConfig() ChannelConfig[uint16, byte] {
	return c.config
}

// Function types for read and write operations
type Read8BitFunc func(uint16) byte
type Write8BitFunc func(uint16, byte)
