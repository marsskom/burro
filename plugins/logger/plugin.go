package logger

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.com/marsskom/burro/internal/events"
	"gitlab.com/marsskom/burro/internal/plugin"
)

func init() {
	plugin.Register("logger", func() plugin.Plugin {
		return New()
	})
}

type LoggerConfig struct {
}

type LoggerPlugin struct {
}

func New() *LoggerPlugin {
	return &LoggerPlugin{}
}

func (p *LoggerPlugin) Name() string {
	return "logger"
}

func (p *LoggerPlugin) Init(cfg any) error {
	return nil
}

func (p *LoggerPlugin) OnConnect(ctx *events.Context) error {
	print(slog.LevelDebug, "Connected", ctx)

	return nil
}

func (p *LoggerPlugin) OnRequest(ctx *events.Context) error {
	print(slog.LevelInfo, "Request received", ctx)

	return nil
}

func (p *LoggerPlugin) OnResponse(ctx *events.Context) error {
	print(slog.LevelInfo, "Response received", ctx)

	return nil
}

func (p *LoggerPlugin) OnError(ctx *events.Context) error {
	print(slog.LevelError, "Error occurred", ctx)

	return nil
}

func (p *LoggerPlugin) OnClose(ctx *events.Context) error {
	print(slog.LevelDebug, "Connection closed", ctx)

	return nil
}

func print(level slog.Level, msg string, ctx *events.Context) {
	slog.Log(context.Background(), level, msg)

	if ctx.Request != nil {
		slog.Log(context.Background(), level,
			fmt.Sprintf(
				"Request data: [%s] %s %s\n",
				ctx.Request.Method,
				ctx.Request.URL.String(),
				ctx.Request.Proto,
			),
		)
	}

	if ctx.Response != nil {
		slog.Log(context.Background(), level,
			fmt.Sprintf(
				"Response data: [%s] %d %s\n",
				ctx.Response.Status,
				ctx.Response.StatusCode,
				ctx.Response.Proto,
			),
		)
	}
}
