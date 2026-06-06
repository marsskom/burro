package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
)

const LevelTrace = slog.Level(-8)

func SetDefault(verbosity int, level string) {
	var slogLevel slog.Level
	if verbosity > 0 {
		slogLevel = verbosityToLevel(verbosity)
	} else {
		slogLevel = parseLevel(level)
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
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

func Trace(msg string, args ...any) {
	if !slog.Default().Enabled(context.Background(), LevelTrace) {
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
