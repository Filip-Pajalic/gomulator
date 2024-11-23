package pubsub

type ReadFunc[AddressType MemoryArchitecture, DataType MemoryArchitecture] func(AddressType) DataType
type WriteFunc[AddressType MemoryArchitecture, DataType MemoryArchitecture] func(AddressType, DataType)

// ChannelConfig holds the configuration for a channel processor
type ChannelConfig[AddressType MemoryArchitecture, DataType MemoryArchitecture] struct {
	ReadType     OperationType
	WriteType    OperationType
	ProcessRead  ReadFunc[AddressType, DataType]
	ProcessWrite WriteFunc[AddressType, DataType]
}
