package runtime

import (
	"log/slog"

	"gitlab.com/marsskom/burro/internal/logger"
)

type CoreLogger struct{}

func NewCoreLogger() *CoreLogger {
	return &CoreLogger{}
}

func (l *CoreLogger) Enabled(level slog.Level) bool {
	return logger.Level() >= level
}

func (l *CoreLogger) Trace(msg string, args ...any) {
	logger.Trace(msg, args...)
}
func (l *CoreLogger) Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}
func (l *CoreLogger) Info(msg string, args ...any) {
	logger.Info(msg, args...)
}
func (l *CoreLogger) Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}
func (l *CoreLogger) Error(msg string, args ...any) {
	logger.Error(msg, args...)
}
func (l *CoreLogger) Audit(msg string, args ...any) {
	logger.Audit(msg, args...)
}
