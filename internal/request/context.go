package request

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"slices"
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

type RequestContext struct {
	ID        string
	StartTime time.Time
	State     atomic.Int32

	Request  *http.Request
	Response *http.Response

	Context context.Context
	Cancel  context.CancelFunc

	Metadata map[string]any

	IsFinished bool
}

func New(r *http.Request) *RequestContext {
	base := r.Context()

	ctx, cancel := context.WithTimeout(base, 30*time.Second)

	return &RequestContext{
		ID:        uuid.NewString(),
		StartTime: time.Now(),
		Request:   r,
		Context:   ctx,
		Cancel:    cancel,
	}
}

func NewFromParent(parent *RequestContext, r *http.Request) *RequestContext {
	ctx, cancel := context.WithCancel(parent.Context)

	var mtdata map[string]any
	if parent.Metadata != nil {
		maps.Copy(mtdata, parent.Metadata)
	}

	return &RequestContext{
		ID:        uuid.NewString(),
		StartTime: time.Now(),
		Request:   r,
		Context:   ctx,
		Cancel:    cancel,
		Metadata:  mtdata,
	}
}

func (c *RequestContext) GetState() RequestState {
	return RequestState(c.State.Load())
}

func (c *RequestContext) Transition(next RequestState) error {
	current := RequestState(c.State.Load())

	if !slices.Contains(transitions[current], next) {
		return fmt.Errorf("invalid request state transition: %s -> %s", current, next)
	}

	c.State.Store(int32(next))

	return nil
}

func (c *RequestContext) Finish(resp *http.Response) {
	c.Response = resp
	c.IsFinished = true
}
