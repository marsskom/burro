package proxy

import (
	"fmt"
	"net/http"

	"gitlab.com/marsskom/burro/internal/model"
)

func (px *Proxy) handleHTTP(w http.ResponseWriter, ctx *model.RequestContext) error {
	ctx.Transition(model.StatePrepared)

	reqSnapshot, err := model.MakeRequestSnapshot(ctx.Request)
	if err != nil {
		return fmt.Errorf("HandleHTTP: error on request snapshot creation: %w", err)
	}

	ctx.RequestSnapshot = reqSnapshot

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

	res, err := px.client.Do(req)
	if err != nil {
		return fmt.Errorf("HandleRequest: error on proceed request: %w", err)
	}

	resSnapshot, err := model.MakeResponseSnapshot(res, px.traceTimings)
	if err != nil {
		return fmt.Errorf("HandleRequest: error on response snapshot creation: %w", err)
	}

	ctx.Response = res
	ctx.ResponseSnapshot = resSnapshot

	err = px.plugins.EmitResponse(ctx)
	if err != nil {
		return fmt.Errorf("HandleRequest: error on EmitResponse: %w", err)
	}

	return nil
}
