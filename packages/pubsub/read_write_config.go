package pubsub

type ReadConfig[AddressType MemoryArchitecture, DataType MemoryArchitecture] struct {
	config ChannelConfig[AddressType, DataType]
}

func NewReadConfig[AddressType MemoryArchitecture, DataType MemoryArchitecture](readType OperationType, processRead ReadFunc[AddressType, DataType]) *ReadConfig[AddressType, DataType] {
	return &ReadConfig[AddressType, DataType]{
		config: ChannelConfig[AddressType, DataType]{
			ReadType:    readType,
			ProcessRead: processRead,
		},
	}
}

func (c *ReadConfig[AddressType, DataType]) processReadEvent(address AddressType) DataType {
	return c.config.ProcessRead(address)
}

func (c *ReadConfig[AddressType, DataType]) processWriteEvent(address AddressType, data DataType) {
	// Not used
}

func (c *ReadConfig[AddressType, DataType]) getChannelConfig() ChannelConfig[AddressType, DataType] {
	return c.config
}

// Write-only configuration
type WriteConfig[AddressType MemoryArchitecture, DataType MemoryArchitecture] struct {
	config ChannelConfig[AddressType, DataType]
}

func NewWriteConfig[AddressType MemoryArchitecture, DataType MemoryArchitecture](writeType OperationType, processWrite WriteFunc[AddressType, DataType]) *WriteConfig[AddressType, DataType] {
	return &WriteConfig[AddressType, DataType]{
		config: ChannelConfig[AddressType, DataType]{
			WriteType:    writeType,
			ProcessWrite: processWrite,
		},
	}
}

func (c *WriteConfig[AddressType, DataType]) processReadEvent(address AddressType) DataType {
	// Not used
	var zero DataType
	return zero
}

func (c *WriteConfig[AddressType, DataType]) processWriteEvent(address AddressType, data DataType) {
	c.config.ProcessWrite(address, data)
}

func (c *WriteConfig[AddressType, DataType]) getChannelConfig() ChannelConfig[AddressType, DataType] {
	return c.config
}

// Read-write configuration
type ReadWriteConfig[AddressType MemoryArchitecture, DataType MemoryArchitecture] struct {
	config ChannelConfig[AddressType, DataType]
}

func NewReadWriteConfig[AddressType MemoryArchitecture, DataType MemoryArchitecture](readType OperationType, writeType OperationType, processRead ReadFunc[AddressType, DataType], processWrite WriteFunc[AddressType, DataType]) *ReadWriteConfig[AddressType, DataType] {
	return &ReadWriteConfig[AddressType, DataType]{
		config: ChannelConfig[AddressType, DataType]{
			ReadType:     readType,
			WriteType:    writeType,
			ProcessRead:  processRead,
			ProcessWrite: processWrite,
		},
	}
}

func (c *ReadWriteConfig[AddressType, DataType]) processReadEvent(address AddressType) DataType {
	return c.config.ProcessRead(address)
}

func (c *ReadWriteConfig[AddressType, DataType]) processWriteEvent(address AddressType, data DataType) {
	c.config.ProcessWrite(address, data)
}

func (c *ReadWriteConfig[AddressType, DataType]) getChannelConfig() ChannelConfig[AddressType, DataType] {
	return c.config
}
