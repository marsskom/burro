package persistence

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
	"gitlab.com/marsskom/burro/internal/config"
)

func runMigrations(cfg *config.GooseConfig, db *sql.DB) error {
	err := goose.SetDialect(cfg.Driver)
	if err != nil {
		return fmt.Errorf("goose: error on set dialect: %w", err)
	}

	err = goose.Up(db, cfg.MigrationDir)
	if err != nil {
		return fmt.Errorf("goose: run migration error: %w", err)
	}

	return nil
}
