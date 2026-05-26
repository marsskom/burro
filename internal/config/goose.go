package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type GooseConfig struct {
	Driver       string
	MigrationDir string
}

func NewGooseConfig() (*GooseConfig, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("cannot load env: %w", err)
	}

	return &GooseConfig{
		Driver:       os.Getenv("GOOSE_DRIVER"),
		MigrationDir: os.Getenv("GOOSE_MIGRATION_DIR"),
	}, nil
}
