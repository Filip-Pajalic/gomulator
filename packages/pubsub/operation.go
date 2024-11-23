package pubsub

type OperationType int

const (
	MemoryReadEvent OperationType = iota
	MemoryWriteEvent
	PPUVramReadEvent
	PPUWramWriteEvent
	PPUOamReadEvent
	PPUOamWriteEvent
	WramReadEvent
	WramWriteEvent
	IoReadEvent
	IoWriteEvent
	HramReadEvent
	HramWriteEvent
	InterruptRequestEvent
	MemoryWramReadEvent
	MemoryWramWriteEvent
	MemoryHramReadEvent
	MemoryHramWriteEvent
	NoEvent
	// Add more events as needed
)

// ExchangeType represents whether the event is a request or a response
type ExchangeType int

const (
	Request ExchangeType = iota
	Response
)
