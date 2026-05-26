package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"gitlab.com/marsskom/burro/internal/database"
	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/persistence/mapper"
)

type WorkspaceRepository struct {
	q *database.Queries
}

func NewWorkspaceRepo(q *database.Queries) *WorkspaceRepository {
	return &WorkspaceRepository{
		q: q,
	}
}

func (r *WorkspaceRepository) LoadWorkspace(ctx context.Context, name string) (*model.Workspace, error) {
	storedWorkspace, err := r.q.GetWorkspaceByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("cannot find workspace '%s' in database: %w", name, err)
	}

	workspace, err := mapper.FromStoredWorkspace(storedWorkspace)
	if err != nil {
		return nil, fmt.Errorf("cannot convert workspace data from db: %w", err)
	}

	return workspace, nil
}

func (r *WorkspaceRepository) SaveWorkspace(ctx context.Context, db *sql.DB, workspace *model.Workspace) error {
	storedWorkspace, err := mapper.ToStoredWorkspace(workspace)
	if err != nil {
		return fmt.Errorf("cannot convert workspace date for db: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	q := database.New(tx)

	// Workspace.
	if workspace.IsNewWorkspace {
		_, err := q.CreateWorkspace(ctx, database.CreateWorkspaceParams{
			ID:        storedWorkspace.ID,
			Name:      storedWorkspace.Name,
			CreatedAt: storedWorkspace.CreatedAt,
			UpdatedAt: storedWorkspace.UpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("error on workspace insert: %w", err)
		}
	} else {
		err = q.UpdateWorkspace(ctx, database.UpdateWorkspaceParams{
			ID:        storedWorkspace.ID,
			UpdatedAt: storedWorkspace.UpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("error on workspace update: %w", err)
		}
	}

	// Sessions.
	if len(workspace.Sessions) == 0 {
		slog.Debug("Workspace doesn't have the sessions", "workspace", workspace.Name)

		return tx.Commit()
	}

	sessRepo := NewSessionRepo(q)
	reqRepo := NewRequestRepo(q)

	for _, s := range workspace.Sessions {
		err = sessRepo.SaveSession(ctx, s)
		if err != nil {
			return fmt.Errorf("cannot save session: %w", err)
		}

		// Requests.
		if len(s.Requests) == 0 {
			slog.Debug("Session doesn't have requests", "session", s.ID)

			continue
		}

		for _, req := range s.Requests {
			err := reqRepo.SaveRequest(ctx, req)
			if err != nil {
				return fmt.Errorf("cannot save request: %w", err)
			}
		}
	}

	return tx.Commit()
}
