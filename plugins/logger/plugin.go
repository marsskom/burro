package logger

import (
	"fmt"
	"log/slog"
	"unicode/utf8"

	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/plugin"
	"gitlab.com/marsskom/burro/internal/pluginapi"
)

func init() {
	plugin.Register("logger", func() plugin.Plugin {
		return New()
	})
}

type LoggerConfig struct {
}

type LoggerPlugin struct {
	rt pluginapi.Runtime
}

func New() *LoggerPlugin {
	return &LoggerPlugin{}
}

func (p *LoggerPlugin) Name() string {
	return "logger"
}

func (p *LoggerPlugin) Init(rt pluginapi.Runtime, cfg any) error {
	p.rt = rt

	return nil
}

func (p *LoggerPlugin) OnConnect(ctx *model.RequestContext) error {
	p.log(slog.LevelDebug, "trying to connect", ctx)

	return nil
}

func (p *LoggerPlugin) OnBeforeRequestSend(ctx *model.RequestContext) error {
	p.log(slog.LevelDebug, "before request send", ctx)

	return nil
}

func (p *LoggerPlugin) OnAfterRequestSend(ctx *model.RequestContext) error {
	p.log(slog.LevelInfo, "after request was sent", ctx)

	return nil
}

func (p *LoggerPlugin) OnBeforeResponseSend(ctx *model.RequestContext) error {
	p.log(slog.LevelDebug, "before response send", ctx)

	return nil
}

func (p *LoggerPlugin) OnAfterResponseSend(ctx *model.RequestContext) error {
	p.log(slog.LevelInfo, "after response was sent", ctx)

	return nil
}

func (p *LoggerPlugin) OnError(ctx *model.RequestContext, err error) error {
	p.log(slog.LevelError, fmt.Sprintf("error occurred: %v", err), ctx)

	return nil
}

func (p *LoggerPlugin) OnClose(ctx *model.RequestContext) error {
	p.log(slog.LevelDebug, "connection closed", ctx)

	return nil
}

func (p *LoggerPlugin) OnWSOpen(ctx *model.RequestContext) error {
	p.log(slog.LevelDebug, "ws connection opened", ctx)

	return nil
}

func (p *LoggerPlugin) OnWSMessage(ctx *model.RequestContext, msg *model.WSMessage) error {
	dataPreview := formatWSData(msg)

	p.log(
		slog.LevelDebug,
		fmt.Sprintf(
			"ws message | dir=%s opcode=%d time=%d data=%s text=%s",
			msg.Direction,
			msg.OpCode,
			msg.Timestamp,
			dataPreview,
			msg.Text,
		),
		ctx,
	)

	return nil
}

func formatWSData(msg *model.WSMessage) string {
	if len(msg.Data) == 0 {
		return "<empty>"
	}

	// Tries to treat as UTF-8 text first.
	if utf8.Valid(msg.Data) {
		s := string(msg.Data)

		if len(s) > 200 {
			return s[:200] + "...(truncated)"
		}
		return s
	}

	// Fallback: hex preview for binary data.
	const max = 64
	if len(msg.Data) > max {
		return fmt.Sprintf("binary(hex)=%x...(truncated)", msg.Data[:max])
	}

	return fmt.Sprintf("binary(hex)=%x", msg.Data)
}

func (p *LoggerPlugin) OnWSClose(ctx *model.RequestContext) error {
	p.log(slog.LevelDebug, "ws connection closed", ctx)

	return nil
}

func (p *LoggerPlugin) log(level slog.Level, msg string, ctx *model.RequestContext) {
	args := []any{
		"Context ID", ctx.ID,
		"Context StartTime", ctx.StartTime,
		"Context State", ctx.State.Load(),
		"Context Metadata", ctx.Metadata,
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

	switch level {
	case slog.LevelDebug:
		p.rt.Log().Debug(msg, args...)
	case slog.LevelWarn:
		p.rt.Log().Warn(msg, args...)
	case slog.LevelError:
		p.rt.Log().Error(msg, args...)
	default:
		p.rt.Log().Info(msg, args...)
	}
}
