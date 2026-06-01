package logger

import (
	"log/slog"
	"os"
	"strings"

	"gitlab.com/marsskom/burro/internal/config"
)

const LevelTrace = slog.Level(-8)

func SetDefault(verbosity int, cfg config.CoreConfig) {
	var level slog.Level
	if verbosity > 0 {
		level = verbosityToLevel(verbosity)
	} else {
		level = parseLevel(cfg.LogLevel)
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)

	slog.SetDefault(slog.New(handler))
}

func verbosityToLevel(v int) slog.Level {
	switch {
	case v >= 3:
		return LevelTrace
	case v == 2:
		return slog.LevelDebug
	case v == 1:
		return slog.LevelInfo
	default:
		return slog.LevelError
	}
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
