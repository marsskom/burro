package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gitlab.com/marsskom/burro/internal/config"
)

type Env struct {
	Project string `env:"PROJECT"`
	Debug   bool   `env:"DEBUG" envDefault:"false"`
}

type State struct {
	BaseDir string
	Env     *Env
	Config  *config.Config
}

func Load(baseDirectory string) (*State, error) {
	envConfig, err := loadEnv()
	if err != nil {
		return &State{}, fmt.Errorf("error load env: %v", err)
	}

	return &State{
		baseDirectory,
		envConfig,
		&config.Config{},
	}, nil
}

func loadEnv() (*Env, error) {
	err := godotenv.Load()
	if err != nil {
		return &Env{}, fmt.Errorf("error reading env file: %v", err)
	}

	return &Env{
		os.Getenv("PROJECT"),
		getEnvBoolValue("DEBUG", false),
	}, nil
}

func getEnvBoolValue(key string, def bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return def
	}

	b, err := strconv.ParseBool(val)
	if err != nil {
		return def
	}

	return b
}
