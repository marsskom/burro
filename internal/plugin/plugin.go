package plugin

import "gitlab.com/marsskom/burro/internal/events"

type Plugin interface {
	Name() string
}

type ConnectHook interface {
	OnConnect(ctx *events.Context) error
}

type RequestHook interface {
	OnRequest(ctx *events.Context) error
}

type ResponseHook interface {
	OnResponse(ctx *events.Context) error
}

type ErrorHook interface {
	OnError(ctx *events.Context) error
}

type CloseHook interface {
	OnClose(ctx *events.Context) error
}
