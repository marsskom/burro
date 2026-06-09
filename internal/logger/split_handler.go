package logger

import (
	"context"
	"log/slog"
	"os"
)

type SplitHandler struct {
	stdout slog.Handler
	stderr slog.Handler
	level  slog.Leveler
}

func NewSplitHandler(opts *slog.HandlerOptions) *SplitHandler {
	h := &SplitHandler{
		stdout: slog.NewJSONHandler(os.Stdout, opts),
		stderr: slog.NewJSONHandler(os.Stderr, opts),
		level:  opts.Level,
	}

	if h.level == nil {
		h.level = slog.LevelInfo
	}

	return h
}

func (h *SplitHandler) Enabled(ctx context.Context, level slog.Level) bool {
	min := h.level.Level()

	if level == LevelAudit {
		return true
	}

	return level >= min
}

func (h *SplitHandler) Handle(ctx context.Context, r slog.Record) error {
	if !h.Enabled(ctx, r.Level) {
		return nil
	}

	if r.Level == LevelAudit {
		return h.stdout.Handle(ctx, r)
	}

	if r.Level >= slog.LevelError {
		return h.stderr.Handle(ctx, r)
	}

	return h.stdout.Handle(ctx, r)
}

func (h *SplitHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SplitHandler{
		stdout: h.stdout.WithAttrs(attrs),
		stderr: h.stderr.WithAttrs(attrs),
	}
}

func (h *SplitHandler) WithGroup(name string) slog.Handler {
	return &SplitHandler{
		stdout: h.stdout.WithGroup(name),
		stderr: h.stderr.WithGroup(name),
	}
}
