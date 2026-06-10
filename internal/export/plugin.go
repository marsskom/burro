package export

import "gitlab.com/marsskom/burro/internal/model"

type FileNameVars struct {
	Session string
}

type Exporter interface {
	OnConnect(ctx *model.RequestContext) error
	OnBeforeRequestSend(ctx *model.RequestContext) error
	OnAfterRequestSend(ctx *model.RequestContext) error
	OnBeforeResponseSend(ctx *model.RequestContext) error
	OnAfterResponseSend(ctx *model.RequestContext) error
	OnError(ctx *model.RequestContext, err error) error
	OnClose(ctx *model.RequestContext) error
	Flush(opts *FileNameVars) error
}

type ExportPlugin struct {
	exporter Exporter
}

func (ep *ExportPlugin) OnConnect(ctx *model.RequestContext) error {
	return ep.exporter.OnConnect(ctx)
}

func (ep *ExportPlugin) OnBeforeRequestSend(ctx *model.RequestContext) error {
	return ep.exporter.OnBeforeRequestSend(ctx)
}

func (ep *ExportPlugin) OnAfterRequestSend(ctx *model.RequestContext) error {
	return ep.exporter.OnAfterRequestSend(ctx)
}

func (ep *ExportPlugin) OnBeforeResponseSend(ctx *model.RequestContext) error {
	return ep.exporter.OnBeforeResponseSend(ctx)
}

func (ep *ExportPlugin) OnAfterResponseSend(ctx *model.RequestContext) error {
	return ep.exporter.OnAfterResponseSend(ctx)
}

func (ep *ExportPlugin) OnError(ctx *model.RequestContext, err error) error {
	return ep.exporter.OnError(ctx, err)
}

func (ep *ExportPlugin) OnClose(ctx *model.RequestContext) error {
	return ep.exporter.OnClose(ctx)
}
