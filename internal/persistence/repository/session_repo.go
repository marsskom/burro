package repository

import (
	"context"
	"fmt"

	"gitlab.com/marsskom/burro/internal/database"
	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/persistence/mapper"
)

type SessionRepository struct {
	q *database.Queries
}

func NewSessionRepo(q *database.Queries) *SessionRepository {
	return &SessionRepository{
		q: q,
	}
}

func (r *SessionRepository) LoadSession(ctx context.Context, id string) (*model.Session, error) {
	storedSession, err := r.q.GetSession(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("cannot find session '%s' in database: %w", id, err)
	}

	session, err := mapper.FromStoredSession(storedSession)
	if err != nil {
		return nil, fmt.Errorf("cannot convert session data from db: %w", err)
	}

	return session, nil
}

func (r *SessionRepository) SaveSession(ctx context.Context, session *model.Session) error {
	storedSession, err := mapper.ToStoredSession(session)
	if err != nil {
		return fmt.Errorf("cannot comvert session data for db: %w", err)
	}

	if session.IsNewSession {
		_, err := r.q.CreateSession(ctx, database.CreateSessionParams{
			ID:          storedSession.ID,
			Name:        storedSession.Name,
			Description: storedSession.Description,
			Metadata:    storedSession.Metadata,
			CreatedAt:   storedSession.CreatedAt,
			UpdatedAt:   storedSession.UpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("error on session insert: %w", err)
		}

		return nil
	}

	err = r.q.UpdateSession(ctx, database.UpdateSessionParams{
		ID:          storedSession.ID,
		Name:        storedSession.Name,
		Description: storedSession.Description,
		Metadata:    storedSession.Metadata,
		UpdatedAt:   storedSession.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("error on session update: %w", err)
	}

	return nil
}
