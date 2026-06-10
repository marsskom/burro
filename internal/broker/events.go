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
	EventConnect            EventType = "connect"
	EventBeforeRequestSend  EventType = "before_request_send"
	EventAfterRequestSend   EventType = "after_request_send"
	EventBeforeResponseSend EventType = "before_response_send"
	EventAfterResponseSend  EventType = "after_response_send"
	EventError              EventType = "error"
	EventClose              EventType = "close"

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

type Subscription struct {
	Ch         chan Event
	Transports map[TransportType]struct{}
	EventTypes map[EventType]struct{}
}

func (s *Subscription) Match(e Event) bool {
	if len(s.Transports) > 0 {
		if _, ok := s.Transports[e.TransportType]; !ok {
			return false
		}
	}

	if len(s.EventTypes) > 0 {
		if _, ok := s.EventTypes[e.Type]; !ok {
			return false
		}
	}

	return true
}

type Hub struct {
	subs map[*Subscription]struct{}

	mu sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		subs: make(map[*Subscription]struct{}),
	}
}

func (h *Hub) Subscribe(
	transports []TransportType,
	eventTypes []EventType,
) *Subscription {
	sub := &Subscription{
		Ch:         make(chan Event, 100),
		Transports: make(map[TransportType]struct{}),
		EventTypes: make(map[EventType]struct{}),
	}

	for _, t := range transports {
		sub.Transports[t] = struct{}{}
	}
	for _, et := range eventTypes {
		sub.EventTypes[et] = struct{}{}
	}

	h.mu.Lock()
	h.subs[sub] = struct{}{}
	h.mu.Unlock()

	logger.Debug("new subscriber is connected")

	return sub
}

func (h *Hub) Unsubscribe(sub *Subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.subs[sub]; ok {
		delete(h.subs, sub)
		close(sub.Ch)

		logger.Debug("subscriber has disconnected")
	}
}

func (h *Hub) Publish(e Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	logger.Debug("hub publish an event", "transportType", e.TransportType, "type", e.Type, "id", e.ID)

	for sub := range h.subs {
		if !sub.Match(e) {
			continue
		}

		select {
		case sub.Ch <- e:
		default:
		}
	}
}
