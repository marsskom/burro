package model

import (
	"errors"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
)

type WorkspaceName string

var workspaceNameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func (wn WorkspaceName) Validate() error {
	if wn == "" {
		return errors.New("workspace name cannot be empty")
	}

	if !workspaceNameRe.MatchString(string(wn)) {
		return errors.New("invalid workspace name")
	}

	return nil
}

type Workspace struct {
	ID   string
	Name WorkspaceName

	CreatedAt time.Time
	UpdatedAt time.Time

	mu       sync.RWMutex
	Sessions []*Session

	IsNewWorkspace bool
}

func NewWorkspace() *Workspace {
	return &Workspace{
		ID:             uuid.NewString(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		IsNewWorkspace: true,
	}
}

func (w *Workspace) AddSession(session *Session) {
	if session == nil {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.Sessions = append(w.Sessions, session)
	w.UpdatedAt = time.Now()
}
