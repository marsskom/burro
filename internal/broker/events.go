package broker

import (
	"log/slog"
	"sync"
)

type EventType string

const (
	EventConnect  EventType = "connect"
	EventRequest  EventType = "request"
	EventResponse EventType = "response"
	EventError    EventType = "error"
	EventClose    EventType = "close"
)

type Event struct {
	Type EventType

	ID        string
	SessionID string

	Proto         string
	Scheme        string
	Host          string
	Method        string
	URL           string
	Path          string
	ContentLength int
	RemoteAddr    string

	StartTime  int64
	State      int
	IsFinished bool

	QueryParams string // JSON
	Headers     string // JSON
	Cookies     string // JSON

	RequestBody []byte

	ResponseStatus        string
	ResponseStatusCode    int
	ResponseProto         string
	ResponseHeaders       string // JSON
	ResponseContentLength int
	ResponseBody          []byte

	CreatedAt int64
	UpdatedAt int64

	Metadata string // JSON
}

type Hub struct {
	subs map[chan Event]struct{}

	mu sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		subs: make(map[chan Event]struct{}),
	}
}

func (h *Hub) Subscribe() chan Event {
	ch := make(chan Event, 100)

	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()

	slog.Debug("new subscriber is connected")

	return ch
}

func (h *Hub) Unsubscribe(ch chan Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.subs[ch]; ok {
		delete(h.subs, ch)
		close(ch)

		slog.Debug("subscriber has disconnected")
	}
}

func (h *Hub) Publish(e Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	slog.Debug("hub publish an event", "type", e.Type, "id", e.ID)

	for ch := range h.subs {
		select {
		case ch <- e:
		default:
		}
	}
}
