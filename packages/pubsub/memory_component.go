package pubsub

// Memory represents a simple memory component
type Memory struct {
	data [0x10000]byte // 64KB of memory
}

// NewMemory creates a new memory component
func NewMemory() *Memory {
	return &Memory{}
}

// MemoryReadProcess reads a byte from memory
func (m *Memory) MemoryReadProcess(address uint16) byte {
	return m.data[address]
}

// MemoryWriteProcess writes a byte to memory
func (m *Memory) MemoryWriteProcess(address uint16, data byte) {
	m.data[address] = data
}

// StartMemoryComponent initializes the memory component and starts processing
func (m *Memory) StartMemoryComponent() {
	// Create a ReadWriteConfig for memory
	config := NewReadWriteConfig[uint16, byte](
		MemoryReadEvent,
		MemoryWriteEvent,
		m.MemoryReadProcess,
		m.MemoryWriteProcess,
	)

	// Start processing events for the memory component
	go ProcessChannelTransactions(config)
}
