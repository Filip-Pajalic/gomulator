package pubsub

type ChannelProcessor[AddressType MemoryArchitecture, DataType MemoryArchitecture] interface {
	processReadEvent(AddressType) DataType
	processWriteEvent(AddressType, DataType)
	getChannelConfig() ChannelConfig[AddressType, DataType]
}
