package runtime

import "log/slog"

type CoreLogger struct {
	log *slog.Logger
}

func NewCoreLogger() *CoreLogger {
	return &CoreLogger{
		log: slog.Default(),
	}
}

func (l *CoreLogger) Debug(msg string, args ...any) {
	l.log.Debug(msg, args...)
}
func (l *CoreLogger) Info(msg string, args ...any) {
	l.log.Info(msg, args...)
}
func (l *CoreLogger) Warn(msg string, args ...any) {
	l.log.Warn(msg, args...)
}
func (l *CoreLogger) Error(msg string, args ...any) {
	l.log.Error(msg, args...)
}
