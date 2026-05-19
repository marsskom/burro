package logger

import (
	"log/slog"
	"os"
	"strings"

	"gitlab.com/marsskom/burro/internal/config"
)

func SetDefault(cfg config.CoreConfig) {
	opts := &slog.HandlerOptions{
		Level: parseLevel(cfg.LogLevel),
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)

	slog.SetDefault(slog.New(handler))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
