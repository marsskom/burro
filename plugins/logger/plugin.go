package logger

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.com/marsskom/burro/internal/model"
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
	logger *slog.Logger
}

func New() *LoggerPlugin {
	return &LoggerPlugin{
		logger: slog.Default(),
	}
}

func (p *LoggerPlugin) Name() string {
	return "logger"
}

func (p *LoggerPlugin) Init(cfg any) error {
	return nil
}

func (p *LoggerPlugin) OnConnect(ctx *model.RequestContext) error {
	p.log(slog.LevelDebug, "Trying to connect", ctx)

	return nil
}

func (p *LoggerPlugin) OnRequest(ctx *model.RequestContext) error {
	p.log(slog.LevelInfo, "Request received", ctx)

	return nil
}

func (p *LoggerPlugin) OnResponse(ctx *model.RequestContext) error {
	p.log(slog.LevelInfo, "Response received", ctx)

	return nil
}

func (p *LoggerPlugin) OnError(ctx *model.RequestContext, err error) error {
	p.log(slog.LevelError, fmt.Sprintf("Error occurred: %v", err), ctx)

	return nil
}

func (p *LoggerPlugin) OnClose(ctx *model.RequestContext) error {
	p.log(slog.LevelDebug, "Connection closed", ctx)

	return nil
}

func (p *LoggerPlugin) log(level slog.Level, msg string, ctx *model.RequestContext) {
	args := []any{
		"ID", ctx.ID,
		"StartTime", ctx.StartTime,
		"State", ctx.State.Load(),
		"Metadata", ctx.Metadata,
	}

	if ctx.Request != nil {
		args = append(args,
			"Request Method", ctx.Request.Method,
			"Request URL", ctx.Request.URL.String(),
			"Request Proto", ctx.Request.Proto,
		)
	}

	if ctx.Response != nil {
		args = append(args,
			"Response Status", ctx.Response.Status,
			"Response StatusCode", ctx.Response.StatusCode,
			"Response Proto", ctx.Response.Proto,
		)
	}

	p.logger.Log(context.Background(), level, msg, args...)
}
