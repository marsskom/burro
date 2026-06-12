package logger

import (
	"context"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
)

const (
	LevelTrace = slog.Level(-8)
	LevelAudit = slog.Level(12)
)

var logger = slog.Default()
var loggerLevel = LevelAudit

func Default() *slog.Logger {
	return logger
}

func Level() slog.Level {
	return loggerLevel
}

func SetDefault(verbosity int, level string) {
	if verbosity > 0 {
		loggerLevel = verbosityToLevel(verbosity)
	} else {
		loggerLevel = parseLevel(level)
	}

	opts := &slog.HandlerOptions{
		Level: loggerLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				switch v := a.Value.Any().(type) {
				case slog.Level:
					return formatLevel(v)

				case string:
					return a

				case int:
					return formatLevel(slog.Level(v))
				}
			}

			return a
		},
	}

	logger = slog.New(NewSplitHandler(opts))

	slog.SetDefault(logger)
}

func formatLevel(l slog.Level) slog.Attr {
	switch l {
	case LevelTrace:
		return slog.String(slog.LevelKey, "TRACE")
	case LevelAudit:
		return slog.String(slog.LevelKey, "AUDIT")
	default:
		return slog.String(slog.LevelKey, l.String())
	}
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
	case "trace":
		return LevelTrace
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "audit":
		return LevelAudit
	default:
		return slog.LevelInfo
	}
}

func Trace(msg string, args ...any) {
	if loggerLevel != LevelTrace {
		return
	}

	pcs := make([]uintptr, 32)
	n := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:n])

	var b strings.Builder

	for {
		frame, more := frames.Next()

		b.WriteString(frame.Function)
		b.WriteByte('\n')
		b.WriteByte('\t')
		b.WriteString(frame.File)
		b.WriteByte(':')
		b.WriteString(strconv.Itoa(frame.Line))
		b.WriteByte('\n')

		if !more {
			break
		}
	}

	args = append(args, "stack", b.String())

	slog.Log(context.Background(), LevelTrace, msg, args...)
}

func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

func Audit(msg string, args ...any) {
	slog.Log(context.Background(), LevelAudit, msg, args...)
}
