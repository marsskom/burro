package pluginapi

import "log/slog"

type Logger interface {
	Enabled(level slog.Level) bool
	Trace(msg string, args ...any)
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Audit(msg string, args ...any)
}
