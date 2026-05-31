package runtime

import (
	"sync"

	"gitlab.com/marsskom/burro/internal/pluginapi"
)

type EventBus struct {
	subs map[string][]pluginapi.EventHandler

	mu sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subs: make(map[string][]pluginapi.EventHandler),
	}
}

func (b *EventBus) Emit(event pluginapi.Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, h := range b.subs[event.Name] {
		go h(event)
	}

	return nil
}

func (b *EventBus) On(name string, handler pluginapi.EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subs[name] = append(b.subs[name], handler)
}
