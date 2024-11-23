package pubsub

// Event represents an operation and its exchange type
type Event struct {
	Operation OperationType
	Exchange  ExchangeType
}

// MemoryArchitecture represents the types used in the channels
type MemoryArchitecture interface {
	~uint16 | ~byte
}

// EventChannelBase is the base structure for events sent through channels
type EventChannelBase struct {
	Event        Event
	Data         interface{}
	Address      interface{}
	ResponseChan interface{}
}
