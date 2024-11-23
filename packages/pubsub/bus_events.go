package pubsub

// Define event types as needed
type OperationType int

const (
	MemoryReadEvent OperationType = iota
	MemoryWriteEvent
	MemoryDataBusEvent
	DMATransferEvent
	PPUVramReadEvent
	PPUWramWriteEvent
	PPUOamReadEvent
	PPUOamWriteEvent
	WramWriteEvent
	WramReadEvent
	IoReadEvent
	IoWriteEvent
	HramReadEvent
	HramWriteEvent
	NoEvent
	// Add more events as needed
)

type ExchangeType int

const (
	Request ExchangeType = iota
	Response

	// Add more events as needed
)

type Event struct {
	Operation OperationType
	Exchange  ExchangeType
}

type EventChannel[T MemoryArchitecture, U MemoryArchitecture] interface {
	Event() Event
	Data() T
	Address() U
}

type MemoryArchitecture interface {
	~uint16 | ~byte
}

type ReadEvent[T MemoryArchitecture, U MemoryArchitecture] struct {
	EventType   Event
	AddressType T
	DataType    U
}

type WriteEvent[T MemoryArchitecture, U MemoryArchitecture] struct {
	eventType   Event
	AddressType T
	DataType    U
}

func (e ReadEvent[T, U]) Event() Event {
	return e.EventType
}

func (e ReadEvent[T, U]) Data() interface{} {
	return e.DataType
}

func (e ReadEvent[T, U]) Address() interface{} {
	return e.Address
}

func (e WriteEvent[T, U]) Event() Event {
	return e.eventType
}

func (e WriteEvent[T, U]) Data() interface{} {
	return e.DataType
}

func (e WriteEvent[T, U]) Address() interface{} {
	return e.Address
}

/*
func (data Write8BitData) ProcessData(eventType Event) {
	event := WriteEvent[Write8BitData]{
		eventType: eventType,
		eventData: data,
	}

	PbManager.Publish(event)
}

func (data Write16BitData) ProcessData(eventType Event) {
	event := WriteEvent[Write16BitData]{
		eventType: eventType,
		eventData: data,
	}
	PbManager.Publish(event)
}

func (data Read8BitData) ProcessData(eventType Event) Read8BitData {
	eventData := <-PbManager.Subscribe(eventType)
	return eventData.(ReadEvent[Read8BitData]).Data().(Read8BitData)

}

func (data Read16BitData) ProcessData(eventType Event) Read16BitData {
	eventData := <-PbManager.Subscribe(eventType)
	return eventData.(ReadEvent[Read8BitData]).Data().(Read16BitData)
}
*/
