package persistence

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
	"gitlab.com/marsskom/burro/internal/migrations"
)

func runMigrations(db *sql.DB) error {
	goose.SetBaseFS(migrations.GooseFS)

	err := goose.SetDialect("sqlite3")
	if err != nil {
		return fmt.Errorf("goose: error on set dialect: %w", err)
	}

	err = goose.Up(db, "sql/schema")
	if err != nil {
		return fmt.Errorf("goose: run migration error: %w", err)
	}

	return nil
}
