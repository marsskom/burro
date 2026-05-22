package proxy

import (
	"fmt"
	"net/http"

	"gitlab.com/marsskom/burro/internal/request"
)

func (px *Proxy) handleHTTP(w http.ResponseWriter, ctx *request.RequestContext) error {
	ctx.Transition(request.StatePrepared)
	ctx.Transition(request.StateForwarding)

	err := px.handleRequest(ctx, ctx.Request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return fmt.Errorf("HandleHTTP: error on handle request: %w", err)
	}

	ctx.Transition(request.StateResponding)

	writeResponse(w, ctx.Response)

	ctx.Transition(request.StateFinished)

	return nil
}

func (px *Proxy) handleRequest(ctx *request.RequestContext, r *http.Request) error {
	err := px.plugins.EmitRequest(ctx)
	if err != nil {
		return fmt.Errorf("HandleRequest: error on EmitRequest: %w", err)
	}

	if ctx.IsFinished {
		return nil
	}

	req, err := http.NewRequest(
		r.Method,
		r.URL.String(),
		r.Body,
	)
	if err != nil {
		return fmt.Errorf("HandleRequest: error on new request creation: %w", err)
	}

	req = req.WithContext(ctx.Context)
	req.Header = r.Header.Clone()

	resp, err := px.client.Do(req)
	if err != nil {
		return fmt.Errorf("HandleRequest: error on proceed request: %w", err)
	}

	ctx.Response = resp

	err = px.plugins.EmitResponse(ctx)
	if err != nil {
		return fmt.Errorf("HandleRequest: error on EmitResponse: %w", err)
	}

	return nil
}
