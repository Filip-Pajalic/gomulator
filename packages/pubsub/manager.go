package pubsub

import (
	"sync"
)

var PbManager *Manager
var once sync.Once

type EventChannelBase struct {
	Event   Event
	Data    interface{}
	Address interface{}
}

type Manager struct {
	subscribers map[Event][]chan EventChannelBase
	mu          sync.RWMutex
}

func init() {
	PbManager = &Manager{
		subscribers: make(map[Event][]chan EventChannelBase),
	}
}

func (m *Manager) Subscribe(eventType Event) chan EventChannelBase {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan EventChannelBase, 1)
	m.subscribers[eventType] = append(m.subscribers[eventType], ch)
	return ch
}

func (m *Manager) Publish(event Event, data interface{}) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, ch := range m.subscribers[event] {
		// Non-blocking send
		select {
		case ch <- EventChannelBase{Event: event, Data: data}:
		default:
			// Handle the case where the channel is full
			// This can be logging, dropping the message, etc.
		}
	}
}
