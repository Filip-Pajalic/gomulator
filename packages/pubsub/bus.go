package pubsub

// Bus struct remains empty as per your current implementation
type Bus struct{}

var busInstance *Bus

// NewBus creates a new Bus instance
func NewBus() *Bus {
	return &Bus{}
}

// BusCtx returns the singleton Bus instance
func BusCtx() *Bus {
	once.Do(func() {
		busInstance = NewBus()
	})
	return busInstance
}

// BusRead reads a byte from the bus at the specified address
func (m *Bus) BusRead(address uint16) byte {
	readType := toReadEvent(address)

	// Create a channel to receive the response
	responseChan := make(chan interface{})

	// Publish the read request event with the address and response channel
	PbManager.Publish(Event{
		Operation: readType,
		Exchange:  Request,
	}, EventChannelBase{
		Address:      address,
		ResponseChan: responseChan,
	})

	// Wait for the response from the component handling the read
	data := <-responseChan
	return data.(byte)
}

// BusWrite writes a byte to the bus at the specified address
func (m *Bus) BusWrite(address uint16, data byte) {
	writeType := toWriteEvent(address)

	// Publish the write request event with the address and data
	PbManager.Publish(Event{
		Operation: writeType,
		Exchange:  Request,
	}, EventChannelBase{
		Address: address,
		Data:    data,
	})
}

// BusRead16 reads two bytes from the bus starting at the specified address
func (m *Bus) BusRead16(address uint16) uint16 {
	lo := uint16(m.BusRead(address))
	hi := uint16(m.BusRead(address + 1))
	return lo | (hi << 8)
}

// BusWrite16 writes two bytes to the bus starting at the specified address
func (m *Bus) BusWrite16(address uint16, data uint16) {
	m.BusWrite(address, byte(data&0xFF))
	m.BusWrite(address+1, byte((data>>8)&0xFF))
}
