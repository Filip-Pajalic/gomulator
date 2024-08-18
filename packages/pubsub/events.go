package pubsub

// Define event types as needed
type EventType int

const (
	MemoryReadEvent EventType = iota
	MemoryWriteEvent
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

// Event represents a generic event with a type and data
type Event struct {
	Type EventType
	Data interface{}
}
