package proxy

import (
	"net/http"

	"gitlab.com/marsskom/burro/internal/events"
)

func (px *Proxy) handleRequest(r *http.Request) (*http.Response, error) {
	ctx := &events.Context{
		Request:   r,
		Response:  nil,
		Metadata:  map[string]any{},
		IsHandled: false,
	}

	err := px.plugins.EmitRequest(ctx)
	if err != nil {
		px.plugins.EmitError(ctx)

		return nil, err
	}

	if ctx.IsHandled {
		return ctx.Response, nil
	}

	req, err := http.NewRequest(
		ctx.Request.Method,
		ctx.Request.URL.String(),
		ctx.Request.Body,
	)
	if err != nil {
		return nil, err
	}

	req.Header = ctx.Request.Header.Clone()

	resp, err := px.client.Do(req)
	if err != nil {
		return nil, err
	}

	ctx.Response = resp

	err = px.plugins.EmitResponse(ctx)
	if err != nil {
		return nil, err
	}

	return ctx.Response, nil
}
