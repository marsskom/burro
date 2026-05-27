package mapper

import (
	"database/sql"
	"fmt"
	"time"

	"gitlab.com/marsskom/burro/internal/database"
	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/persistence"
)

func FromStoredSession(storedSession database.Session) (*model.Session, error) {
	mtdata := map[string]any{}
	if storedSession.Metadata.Valid {
		m, err := persistence.TextToMap[map[string]any](storedSession.Metadata.String)
		if err != nil {
			return nil, fmt.Errorf("cannot convert metadata to map: %w", err)
		}
		mtdata = m
	}

	return &model.Session{
		ID:          storedSession.ID,
		Name:        persistence.NullString(storedSession.Name),
		Description: persistence.NullString(storedSession.Description),
		CreatedAt:   time.UnixMilli(storedSession.CreatedAt.Int64),
		UpdatedAt:   time.UnixMilli(storedSession.UpdatedAt.Int64),
		Metadata:    mtdata,
	}, nil
}

func ToStoredSession(session *model.Session) (database.Session, error) {
	mtdata := sql.NullString{
		Valid:  false,
		String: "",
	}
	if session.Metadata != nil {
		m, err := persistence.MapToText(session.Metadata)
		if err != nil {
			return database.Session{}, fmt.Errorf("cannot convert metadata to string: %w", err)
		}
		mtdata = sql.NullString{
			Valid:  true,
			String: m,
		}
	}

	name := sql.NullString{
		Valid:  false,
		String: "",
	}
	if session.Name != "" {
		name = sql.NullString{
			Valid:  true,
			String: session.Name,
		}
	}

	desc := sql.NullString{
		Valid:  false,
		String: "",
	}
	if session.Description != "" {
		desc = sql.NullString{
			Valid:  true,
			String: session.Description,
		}
	}

	return database.Session{
		ID:          session.ID,
		Name:        name,
		Description: desc,
		Metadata:    mtdata,
		CreatedAt: sql.NullInt64{
			Valid: true,
			Int64: session.CreatedAt.UnixMilli(),
		},
		UpdatedAt: sql.NullInt64{
			Valid: true,
			Int64: session.UpdatedAt.UnixMilli(),
		},
	}, nil
}
