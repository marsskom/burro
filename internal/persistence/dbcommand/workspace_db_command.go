package dbcommand

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"gitlab.com/marsskom/burro/internal/database"
	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/persistence/repository"
)

func UpsertWorkspaceCommand(ctx context.Context, db *sql.DB, workspace *model.Workspace) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	queries := database.New(tx)

	// Workspace.
	wRepo := repository.NewWorkspaceRepo(queries)
	err = wRepo.SaveWorkspace(ctx, workspace)
	if err != nil {
		return fmt.Errorf("error on workspace save into db: %w", err)
	}

	// Sessions.
	if len(workspace.Sessions) == 0 {
		slog.Info("workspace doesn't have the sessions", "workspace", workspace.Name)

		return tx.Commit()
	}

	sessRepo := repository.NewSessionRepo(queries)
	reqRepo := repository.NewRequestRepo(queries)

	for _, s := range workspace.Sessions {
		err = sessRepo.SaveSession(ctx, s)
		if err != nil {
			return fmt.Errorf("cannot save session: %w", err)
		}

		// Requests.
		if len(s.Requests) == 0 {
			slog.Debug("session doesn't have requests", "session", s.ID)

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
