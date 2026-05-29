package repository

import (
	"context"
	"fmt"

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

func (r *WorkspaceRepository) SaveWorkspace(ctx context.Context, workspace *model.Workspace) error {
	storedWorkspace, err := mapper.ToStoredWorkspace(workspace)
	if err != nil {
		return fmt.Errorf("cannot comvert workspace data for db: %w", err)
	}

	err = r.q.UpsertWorkspace(ctx, database.UpsertWorkspaceParams{
		ID:        storedWorkspace.ID,
		Name:      storedWorkspace.Name,
		CreatedAt: storedWorkspace.CreatedAt,
		UpdatedAt: storedWorkspace.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("error on workspace save into db: %w", err)
	}

	return nil
}
