package pubsub

import (
	"sync"
)

type Manager struct {
	subscribers map[EventType][]chan Event
	mu          sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		subscribers: make(map[EventType][]chan Event),
	}
}

func (m *Manager) Subscribe(eventType EventType) chan Event {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan Event, 1)
	m.subscribers[eventType] = append(m.subscribers[eventType], ch)
	return ch
}

func (m *Manager) Publish(event Event) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, ch := range m.subscribers[event.Type] {
		ch <- event
	}
}
