package runtime

import (
	"sync"

	"gitlab.com/marsskom/burro/internal/pluginapi"
)

type handlerEntity struct {
	id      int
	handler pluginapi.EventHandler
}

type EventBus struct {
	subs   map[string][]handlerEntity
	nextID int

	mu sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subs:   make(map[string][]handlerEntity),
		nextID: 1,
	}
}

func (b *EventBus) Emit(event pluginapi.Event) error {
	b.mu.RLock()
	// Copies handlers on lock avoiding race condition on `Off` during `Emit`.
	handlers := append([]handlerEntity(nil), b.subs[event.Name]...)
	b.mu.RUnlock()

	for _, h := range handlers {
		go h.handler(event)
	}

	return nil
}

func (b *EventBus) On(name string, handler pluginapi.EventHandler) func() {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := b.nextID
	b.nextID++

	b.subs[name] = append(b.subs[name], handlerEntity{
		id:      id,
		handler: handler,
	})

	return func() {
		b.Off(name, id)
	}
}

func (b *EventBus) Off(name string, id int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	list := b.subs[name]
	for i := 0; i < len(list); i++ {
		if list[i].id != id {
			continue
		}

		list[i] = list[len(list)-1]
		list = list[:len(list)-1]
		i--
	}

	if len(list) == 0 {
		delete(b.subs, name)
	} else {
		b.subs[name] = list
	}
}
