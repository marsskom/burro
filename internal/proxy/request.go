package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"gitlab.com/marsskom/burro/internal/model"
)

func (px *Proxy) handleHTTP(w http.ResponseWriter, ctx *model.RequestContext) error {
	ctx.Transition(model.StatePrepared)

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return fmt.Errorf("HandleHTTP: error on read request body: %w", err)
	}

	ctx.Request.Body.Close()

	ctx.RequestBody = body

	// Restore request body for next request.
	ctx.Request.Body = io.NopCloser(bytes.NewReader(body))

	ctx.Transition(model.StateForwarding)

	err = px.handleRequest(ctx, ctx.Request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return fmt.Errorf("HandleHTTP: error on handle request: %w", err)
	}

	ctx.Transition(model.StateResponding)

	writeResponse(w, ctx.Response)

	ctx.Transition(model.StateFinished)

	return nil
}

func (px *Proxy) handleRequest(ctx *model.RequestContext, r *http.Request) error {
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("HandleRequest: error on read response body: %w", err)
	}

	resp.Body.Close()

	ctx.Response = resp
	ctx.ResponseBody = body

	// Restore response body for TLS connection.
	resp.Body = io.NopCloser(bytes.NewReader(body))

	err = px.plugins.EmitResponse(ctx)
	if err != nil {
		return fmt.Errorf("HandleRequest: error on EmitResponse: %w", err)
	}

	return nil
}
