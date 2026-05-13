package main

import (
	"log/slog"
	"os"
)

func InitLogger(state *State) {
	logLevel := slog.LevelDebug
	if !state.Env.Debug {
		logLevel = slog.LevelInfo
	}

	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		}),
	)

	slog.SetDefault(logger)
}
