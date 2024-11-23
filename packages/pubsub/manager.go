// pubsub/manager.go
package pubsub

import (
	"log"
	"sync"
)

// PubSubManager manages subscribers and event dispatching
type PubSubManager struct {
	subscribers map[OperationType]chan EventChannelBase
	mu          sync.RWMutex
}

// Singleton instance
var PbManager *PubSubManager
var once sync.Once

// Initialize the PubSubManager singleton
func init() {
	once.Do(func() {
		PbManager = &PubSubManager{
			subscribers: make(map[OperationType]chan EventChannelBase),
		}
	})
}

// Subscribe allows a component to subscribe to an Event
// Returns a channel to receive EventChannelBase
func (m *PubSubManager) Subscribe(event Event) chan EventChannelBase {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch, exists := m.subscribers[event.Operation]
	if !exists {
		// Create a new buffered channel for this OperationType
		ch = make(chan EventChannelBase, 100) // Adjust buffer size as needed
		m.subscribers[event.Operation] = ch
		log.Printf("pubsub: Subscribed to event type %s", event.Operation)
	}
	return ch
}

// Publish sends data to all subscribers of an OperationType
func (m *PubSubManager) Publish(event Event, data EventChannelBase) {
	m.mu.RLock()
	ch, exists := m.subscribers[event.Operation]
	m.mu.RUnlock()

	if !exists {
		log.Printf("pubsub: No subscribers for event type %s", event.Operation)
		return
	}

	select {
	case ch <- data:
		// Successfully sent
	default:
		// Channel is full; handle overflow
		log.Printf("pubsub: Event queue for %s is full. Dropping event.", event.Operation)
	}
}

// CloseAll closes all subscriber channels (for graceful shutdown)
func (m *PubSubManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for eventType, ch := range m.subscribers {
		close(ch)
		log.Printf("pubsub: Closed channel for %s", eventType)
		delete(m.subscribers, eventType)
	}
}
