package plugin

import "gitlab.com/marsskom/burro/internal/request"

type Plugin interface {
	Name() string
	Init(cfg any) error
}

type ConnectHook interface {
	OnConnect(ctx *request.RequestContext) error
}

type RequestHook interface {
	OnRequest(ctx *request.RequestContext) error
}

type ResponseHook interface {
	OnResponse(ctx *request.RequestContext) error
}

type ErrorHook interface {
	OnError(ctx *request.RequestContext, err error) error
}

type CloseHook interface {
	OnClose(ctx *request.RequestContext) error
}
