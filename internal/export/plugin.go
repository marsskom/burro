package export

import "gitlab.com/marsskom/burro/internal/model"

type FileNameVars struct {
	Session string
}

type Exporter interface {
	OnConnect(ctx *model.RequestContext) error
	OnRequest(ctx *model.RequestContext) error
	OnResponse(ctx *model.RequestContext) error
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

func (ep *ExportPlugin) OnRequest(ctx *model.RequestContext) error {
	return ep.exporter.OnRequest(ctx)
}

func (ep *ExportPlugin) OnResponse(ctx *model.RequestContext) error {
	return ep.exporter.OnResponse(ctx)
}

func (ep *ExportPlugin) OnError(ctx *model.RequestContext, err error) error {
	return ep.exporter.OnError(ctx, err)
}

func (ep *ExportPlugin) OnClose(ctx *model.RequestContext) error {
	return ep.exporter.OnClose(ctx)
}
