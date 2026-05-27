package persistence

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"

	"gitlab.com/marsskom/burro/internal/config"
	_ "modernc.org/sqlite"
)

func MapToText(m any) (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("cannot convert map to json: %w", err)
	}

	return string(b), nil
}

func TextToMap[T any](s string) (T, error) {
	var m T
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		return m, fmt.Errorf("cannot unmarshall string json: %w", err)
	}

	return m, nil
}

func NullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}

	return ""
}

func IntToBool(i int) bool {
	if i > 0 {
		return true
	}

	return false
}

func getPathToDB(name string, path string) string {
	return filepath.Join(path, fmt.Sprintf("%s.sqlite3", string(name)))
}

func LoadDatabase(cfg *config.GooseConfig, name string, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", getPathToDB(name, path))
	if err != nil {
		return nil, fmt.Errorf("cannot load database: %w", err)
	}

	err = runMigrations(cfg, db)
	if err != nil {
		return nil, fmt.Errorf("cannot run migration on a new db: %w", err)
	}

	return db, nil
}

func CreateDatabase(cfg *config.GooseConfig, name string, path string) (*sql.DB, error) {
	db, err := LoadDatabase(cfg, name, path)
	if err != nil {
		return nil, fmt.Errorf("cannot create db: %w", err)
	}

	return db, nil
}
