package mapper

import (
	"database/sql"
	"time"

	"gitlab.com/marsskom/burro/internal/database"
	"gitlab.com/marsskom/burro/internal/model"
)

func FromStoredWorkspace(storedWorkspace database.Workspace) (*model.Workspace, error) {
	return &model.Workspace{
		ID:        storedWorkspace.ID,
		Name:      model.WorkspaceName(storedWorkspace.Name),
		CreatedAt: time.UnixMicro(storedWorkspace.CreatedAt.Int64),
		UpdatedAt: time.UnixMicro(storedWorkspace.UpdatedAt.Int64),
	}, nil
}

func ToStoredWorkspace(workspace *model.Workspace) (database.Workspace, error) {
	return database.Workspace{
		ID:   workspace.ID,
		Name: string(workspace.Name),
		CreatedAt: sql.NullInt64{
			Valid: true,
			Int64: workspace.CreatedAt.UnixMilli(),
		},
		UpdatedAt: sql.NullInt64{
			Valid: true,
			Int64: workspace.UpdatedAt.UnixMilli(),
		},
	}, nil
}
