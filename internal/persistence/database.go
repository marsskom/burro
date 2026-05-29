package persistence

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
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

type DBConnection struct {
	FileName    string
	Path        string
	DB          *sql.DB
	GooseConfig *config.GooseConfig
}

func NewConnection(cfg *config.GooseConfig, filename string, path string) *DBConnection {
	return &DBConnection{
		FileName:    filename,
		Path:        path,
		GooseConfig: cfg,
	}
}

func (c *DBConnection) getFullPath() string {
	return filepath.Join(c.Path, fmt.Sprintf("%s.sqlite3", c.FileName))
}

func (c *DBConnection) connect() error {
	db, err := sql.Open("sqlite", c.getFullPath())
	if err != nil {
		return fmt.Errorf("cannot open connection to database: %w", err)
	}

	c.DB = db
	err = runMigrations(c.GooseConfig, c.DB)
	if err != nil {
		c.DB.Close()

		return fmt.Errorf("cannot run migration on the db: %w", err)
	}

	return nil
}

func (c *DBConnection) Create() error {
	if _, err := os.Stat(c.getFullPath()); err == nil {
		return fmt.Errorf("db file already exist")
	}

	return c.connect()
}

func (c *DBConnection) Open() error {
	if _, err := os.Stat(c.getFullPath()); err != nil {
		return fmt.Errorf("db file doesn't exist: %w", err)
	}

	return c.connect()
}

func (c *DBConnection) Close() {
	if c.DB != nil {
		c.DB.Close()
	}
}

func (c *DBConnection) Remove() error {
	c.Close()

	err := os.Remove(c.getFullPath())
	if err != nil {
		return fmt.Errorf("cannot remove db file: %w", err)
	}

	return nil
}
