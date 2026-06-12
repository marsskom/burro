package proxy

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"

	"gitlab.com/marsskom/burro/internal/model"
)

func (px *Proxy) handleRawHTTPOverConn(clientConn net.Conn, ctx *model.RequestContext) error {
	ctx.SetPrototol(model.HTTP)

	err := ctx.Transition(model.StatePrepared)
	if err != nil {
		return fmt.Errorf("HTTP Over Tunnel: cannot transit context to prepared state: %w", err)
	}

	reader := bufio.NewReader(clientConn)
	writer := bufio.NewWriter(clientConn)

	for {
		err := px.proceedTunnelRequest(
			clientConn,
			"http",
			ctx,
			reader,
			writer,
		)
		if err != nil {
			if err == io.EOF {
				break
			}

			return fmt.Errorf("HTTP: %w", err)
		}
	}

	return nil
}

func (px *Proxy) handleRawHTTP(w http.ResponseWriter, ctx *model.RequestContext) error {
	ctx.SetPrototol(model.HTTP)

	err := ctx.Transition(model.StatePrepared)
	if err != nil {
		return fmt.Errorf("HTTP: cannot transit context to prepared state: %w", err)
	}

	err = ctx.Transition(model.StateForwarding)
	if err != nil {
		return fmt.Errorf("HTTP: canot transit context to forwarding state: %w", err)
	}

	err = px.proceedRequest(ctx, ctx.Request)
	if err != nil {
		return fmt.Errorf("HTTP: error on handle request: %w", err)
	}

	err = ctx.Transition(model.StateResponding)
	if err != nil {
		return fmt.Errorf("HTTP: canot transit context to responding state: %w", err)
	}

	err = px.plugins.EmitBeforeResponseSend(ctx)
	if err != nil {
		return fmt.Errorf("HTTP: error on EmitBeforeResponseSend: %w", err)
	}

	if !isHTTPStreamingResponse(ctx.Response) {
		// Regular HTTP request - response.
		writeResponse(w, ctx.Response)
	} else {
		// Stream HTTP.
		body, err := writeHTTPStream(w, ctx.Response)
		if err != nil {
			return fmt.Errorf("HTTP: error on write HTTP stream: %w", err)
		}

		ctx.Response.Body = io.NopCloser(bytes.NewReader(body))
	}

	resSnapshot, err := model.MakeResponseSnapshot(ctx.Response, px.traceTimings)
	if err != nil {
		return fmt.Errorf("HTTP: error on response snapshot creation: %w", err)
	}

	ctx.SetResponse(ctx.Response, resSnapshot)
	ctx.Finish()

	err = px.plugins.EmitAfterResponseSend(ctx)
	if err != nil {
		return fmt.Errorf("HTTP: error on EmitAfterResponseSend: %w", err)
	}

	err = ctx.Transition(model.StateFinished)
	if err != nil {
		return fmt.Errorf("HTTP: canot transit context to finished state: %w", err)
	}

	return nil
}
