package plugin

import "gitlab.com/marsskom/burro/internal/model"

type Plugin interface {
	Name() string
	Init(cfg any) error
}

type ConnectHook interface {
	OnConnect(ctx *model.RequestContext) error
}

type RequestHook interface {
	OnRequest(ctx *model.RequestContext) error
}

type ResponseHook interface {
	OnResponse(ctx *model.RequestContext) error
}

type ErrorHook interface {
	OnError(ctx *model.RequestContext, err error) error
}

type CloseHook interface {
	OnClose(ctx *model.RequestContext) error
}
