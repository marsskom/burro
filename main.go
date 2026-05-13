package main

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"
)

func execDir() string {
	exe, err := os.Executable()
	if err != nil {
		log.Panicln(err)
	}

	return filepath.Dir(exe)
}

func main() {
	state, err := Load(execDir())
	if err != nil {
		log.Fatalln(err)
	}

	InitLogger(state)

	slog.Debug("Burro has been initialized")
}
