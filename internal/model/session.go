package model

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID          string
	Name        string
	Description string

	CreatedAt time.Time
	UpdatedAt time.Time

	mu       sync.RWMutex
	Requests []*RequestContext

	Metadata map[string]any
}

func NewSession() *Session {
	return &Session{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (s *Session) SetName(name string) {
	s.Name = name
}

func (s *Session) SetDescription(description string) {
	s.Description = description
}

func (s *Session) AddRequest(ctx *RequestContext) {
	if ctx == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ctx.Session = s

	s.Requests = append(s.Requests, ctx)
	s.UpdatedAt = time.Now()
}
