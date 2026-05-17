package logger

import (
	"log"

	"gitlab.com/marsskom/burro/internal/events"
)

type Plugin struct{}

func New() *Plugin {
	return &Plugin{}
}

func (p *Plugin) Name() string {
	return "logger"
}

func (p *Plugin) OnConnect(ctx *events.Context) error {
	print("Connected", ctx)

	return nil
}

func (p *Plugin) OnRequest(ctx *events.Context) error {
	print("Request received", ctx)

	return nil
}

func (p *Plugin) OnResponse(ctx *events.Context) error {
	print("Response received", ctx)

	return nil
}

func (p *Plugin) OnError(ctx *events.Context) error {
	print("Error occurred", ctx)

	return nil
}

func (p *Plugin) OnClose(ctx *events.Context) error {
	print("Connection closed", ctx)

	return nil
}

func print(msg string, ctx *events.Context) {
	log.Println(msg)

	if ctx.Request != nil {
		log.Printf(
			"Request data: [%s] %s %s\n",
			ctx.Request.Method,
			ctx.Request.URL.String(),
			ctx.Request.Proto,
		)
	}

	if ctx.Response != nil {
		log.Printf(
			"Response data: [%s] %d %s\n",
			ctx.Response.Status,
			ctx.Response.StatusCode,
			ctx.Response.Proto,
		)
	}
}
