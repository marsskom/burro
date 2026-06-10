package plugin

import (
	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/pluginapi"
)

type Plugin interface {
	Name() string
	Init(rt pluginapi.Runtime, cfg any) error
}

type ConnectHook interface {
	OnConnect(ctx *model.RequestContext) error
}

type RequestHook interface {
	OnBeforeRequestSend(ctx *model.RequestContext) error
	OnAfterRequestSend(ctx *model.RequestContext) error
}

type ResponseHook interface {
	OnBeforeResponseSend(ctx *model.RequestContext) error
	OnAfterResponseSend(ctx *model.RequestContext) error
}

type ErrorHook interface {
	OnError(ctx *model.RequestContext, err error) error
}

type CloseHook interface {
	OnClose(ctx *model.RequestContext) error
}

type WebSocketHook interface {
	OnWSOpen(ctx *model.RequestContext) error
	OnWSMessage(ctx *model.RequestContext, msg *model.WSMessage) error
	OnWSClose(ctx *model.RequestContext) error
}
