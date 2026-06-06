package broker

import (
	"sync"

	"gitlab.com/marsskom/burro/internal/logger"
)

type TransportType string

const (
	TransportHTTP TransportType = "http"
	TransportWS   TransportType = "ws"
)

type EventType string

const (
	EventConnect  EventType = "connect"
	EventRequest  EventType = "request"
	EventResponse EventType = "response"
	EventError    EventType = "error"
	EventClose    EventType = "close"

	EventWSConnect EventType = "ws_connect"
	EventWSMessage EventType = "ws_message"
	EventWSClose   EventType = "ws_close"
)

type HTTPEvent struct {
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

type WSEvent struct {
	Direction string
	OpCode    int
	Data      []byte
	Text      string
	Timestamp int64
}

type Event struct {
	TransportType TransportType
	Type          EventType

	ID        string
	SessionID string

	Timestamp int64

	HTTP *HTTPEvent
	WS   *WSEvent
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

	logger.Debug("new subscriber is connected")

	return ch
}

func (h *Hub) Unsubscribe(ch chan Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.subs[ch]; ok {
		delete(h.subs, ch)
		close(ch)

		logger.Debug("subscriber has disconnected")
	}
}

func (h *Hub) Publish(e Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	logger.Debug("hub publish an event", "type", e.Type, "id", e.ID)

	for ch := range h.subs {
		select {
		case ch <- e:
		default:
		}
	}
}
