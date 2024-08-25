package pubsub

// Define event types as needed
type EventOperation int

const (
	MemoryReadEvent EventOperation = iota
	MemoryWriteEvent
	MemoryDataBusEvent
	DMATransferEvent
	PPUWramReadEvent
	PPUWramWriteEvent
	PPUOamReadEvent
	PPUOamWriteEvent
	WramWriteEvent
	WramReadEvent
	IoReadEvent
	IoWriteEvent
	HramReadEvent
	HramWriteEvent
	// Add more events as needed
)

type EventAction int

const (
	Request EventAction = iota
	Response

	// Add more events as needed
)

type Event struct {
	Type         EventOperation
	ExchangeType EventAction
}

type EventChannel interface {
	Event() Event
	Data() interface{}
}

type WriteData interface {
	WriteData(eventType Event)
}

type ReadData[T any] interface {
	ReadData(eventType Event) T
}

// Not sure if this is the best solution atm, cause it returns itself
type ReadEvent[T ReadData[T]] struct {
	EventType Event
	EventData T
}

type WriteEvent[T WriteData] struct {
	eventType Event
	eventData T
}

type Read8BitData struct {
	Address uint16
	Data    uint8
}

type Read16BitData struct {
	Address uint16
	Data    uint16
}

type Write8BitData struct {
	Address uint16
	Data    uint8
}

type Write16BitData struct {
	Address uint16
	Data    uint16
}

func (e ReadEvent[T]) Event() Event {
	return e.EventType
}

func (e ReadEvent[T]) Data() interface{} {
	return e.EventData
}

func (e WriteEvent[T]) Event() Event {
	return e.eventType
}

func (e WriteEvent[T]) Data() interface{} {
	return e.eventData
}

func (data Write8BitData) WriteData(eventType Event) {
	event := WriteEvent[Write8BitData]{
		eventType: eventType,
		eventData: data,
	}

	GetPubSubManager().Publish(event)
}

func (data Write16BitData) WriteData(eventType Event) {
	event := WriteEvent[Write16BitData]{
		eventType: eventType,
		eventData: data,
	}
	GetPubSubManager().Publish(event)
}

func (data Read8BitData) ReadData(eventType Event) Read8BitData {
	eventData := <-GetPubSubManager().Subscribe(eventType)
	return eventData.(ReadEvent[Read8BitData]).Data().(Read8BitData)

}

func (data Read16BitData) ReadData(eventType Event) Read16BitData {
	eventData := <-GetPubSubManager().Subscribe(eventType)
	return eventData.(ReadEvent[Read8BitData]).Data().(Read16BitData)
}
