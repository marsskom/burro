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

func (r *SessionRepository) SaveSession(ctx context.Context, session *model.Session) error {
	storedSession, err := mapper.ToStoredSession(session)
	if err != nil {
		return fmt.Errorf("cannot comvert session data for db: %w", err)
	}

	err = r.q.UpsertSession(ctx, database.UpsertSessionParams{
		ID:          storedSession.ID,
		Name:        storedSession.Name,
		Description: storedSession.Description,
		Metadata:    storedSession.Metadata,
		CreatedAt:   storedSession.CreatedAt,
		UpdatedAt:   storedSession.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("error on session save into db: %w", err)
	}

	return nil
}
