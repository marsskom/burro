package model

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type RequestState int

const (
	StateReceived RequestState = iota
	StatePrepared
	StateForwarding
	StateResponding
	StateFinished
	StateCanceled
	StateFailed
)

var transitions = map[RequestState][]RequestState{
	StateReceived:   {StatePrepared, StateCanceled, StateFailed},
	StatePrepared:   {StateForwarding, StateCanceled, StateFailed},
	StateForwarding: {StateResponding, StateCanceled, StateFailed},
	StateResponding: {StateFinished},
}

// TODO: make request/response body data separate fields for transparency, move `is_new` from models.
type RequestContext struct {
	ID        string
	StartTime time.Time
	State     atomic.Int32

	CreatedAt time.Time
	UpdatedAt time.Time

	Session *Session

	Request  *http.Request
	Response *http.Response

	RequestBody  []byte
	ResponseBody []byte

	Context context.Context
	Cancel  context.CancelFunc

	Metadata map[string]any

	IsFinished bool

	mu           sync.RWMutex
	IsNewRequest bool
}

func NewCtx(session *Session, r *http.Request) *RequestContext {
	base := r.Context()

	ctx, cancel := context.WithTimeout(base, 30*time.Second)

	return &RequestContext{
		ID:           uuid.NewString(),
		StartTime:    time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Session:      session,
		Request:      r,
		Context:      ctx,
		Cancel:       cancel,
		IsNewRequest: true,
	}
}

func NewCtxFromParent(parent *RequestContext, r *http.Request) *RequestContext {
	parent.mu.RLock()
	defer parent.mu.RUnlock()

	ctx, cancel := context.WithCancel(parent.Context)

	var mtdata map[string]any
	if parent.Metadata != nil {
		maps.Copy(mtdata, parent.Metadata)
	}

	return &RequestContext{
		ID:           uuid.NewString(),
		StartTime:    time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Session:      parent.Session,
		Request:      r,
		Context:      ctx,
		Cancel:       cancel,
		Metadata:     mtdata,
		IsNewRequest: true,
	}
}

func (c *RequestContext) GetState() RequestState {
	return RequestState(c.State.Load())
}

func (c *RequestContext) Transition(next RequestState) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	current := RequestState(c.State.Load())

	if !slices.Contains(transitions[current], next) {
		return fmt.Errorf("invalid request state transition: %v -> %v", current, next)
	}

	c.State.Store(int32(next))
	c.UpdatedAt = time.Now()

	return nil
}

func (c *RequestContext) Finish(resp *http.Response) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Response = resp
	c.IsFinished = true
	c.UpdatedAt = time.Now()
}
