package pubsub

import (
	"sync"
)

var pbManager *Manager
var once sync.Once

type Manager struct {
	subscribers map[Event][]chan EventChannel
	mu          sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		subscribers: make(map[Event][]chan EventChannel),
	}
}

func GetPubSubManager() *Manager {
	once.Do(func() {
		pbManager = NewManager()
	})
	return pbManager
}

func (m *Manager) Subscribe(eventType Event) chan EventChannel {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan EventChannel, 1)
	m.subscribers[eventType] = append(m.subscribers[eventType], ch)
	return ch
}

func (m *Manager) Publish(event EventChannel) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, ch := range m.subscribers[event.Event()] {
		// Non-blocking send
		select {
		case ch <- event:
		default:
			// Handle the case where the channel is full
			// This can be logging, dropping the message, etc.
		}
	}
}
