package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var (
	DBErrorFileAlreadyExists = errors.New("db: file already exists")
	DBErrorFileNotFound      = errors.New("db: file not found")
)

type DBConnection struct {
	FileName string
	Path     string
	DB       *sql.DB
}

func NewConnection(filename string, path string) *DBConnection {
	return &DBConnection{
		FileName: filename,
		Path:     path,
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
	err = runMigrations(c.DB)
	if err != nil {
		c.DB.Close()

		return fmt.Errorf("cannot run migration on the db: %w", err)
	}

	return nil
}

func (c *DBConnection) Create() error {
	if _, err := os.Stat(c.getFullPath()); err == nil {
		return DBErrorFileAlreadyExists
	}

	return c.connect()
}

func (c *DBConnection) Open() error {
	if _, err := os.Stat(c.getFullPath()); err != nil {
		return fmt.Errorf("%w: %s", DBErrorFileNotFound, err)
	}

	return c.connect()
}

func (c *DBConnection) OpenOrCreate() error {
	err := c.Open()
	if err == nil {
		return nil
	}

	if !errors.Is(err, DBErrorFileNotFound) {
		return err
	}

	if err := c.Create(); err != nil {
		return err
	}

	return nil
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
