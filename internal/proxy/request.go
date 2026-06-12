package proxy

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"gitlab.com/marsskom/burro/internal/model"
)

func (px *Proxy) proceedTunnelRequest(
	clientConn net.Conn,
	scheme string,
	ctx *model.RequestContext,
	reader *bufio.Reader,
	writer *bufio.Writer,
) error {
	req, err := http.ReadRequest(reader)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return io.EOF
		}
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return io.ErrUnexpectedEOF
		}

		return fmt.Errorf("problem to read request: %w", err)
	}

	if req == nil {
		return nil
	}

	// WebSocket.
	if isWebSocketRequest(req) {
		if ctx.Protocol == model.HTTPS {
			ctx.SetPrototol(model.WSS)
		} else {
			ctx.SetPrototol(model.WS)
		}

		return px.handleWebSocket(clientConn, ctx, req)
	}

	if req.Close {
		return io.EOF
	}

	newCtx := model.NewCtxFromParent(ctx, px.traceTimings, req)
	px.session.AddRequest(newCtx)

	err = newCtx.Transition(model.StatePrepared)
	if err != nil {
		return fmt.Errorf("cannot transit new context to prepared state: %w", err)
	}

	req = req.WithContext(newCtx.Context)
	req.URL.Scheme = scheme
	req.URL.Host = newCtx.Request.Host
	req.Host = newCtx.Request.Host
	req.RequestURI = ""

	err = newCtx.Transition(model.StateForwarding)
	if err != nil {
		return fmt.Errorf("cannot transit new context to forwarding state: %w", err)
	}

	err = px.proceedRequest(newCtx, req)
	if err != nil {
		return fmt.Errorf("error on handle request: %w", err)
	}

	err = newCtx.Transition(model.StateResponding)
	if err != nil {
		return fmt.Errorf("cannot transit new context to responding state: %w", err)
	}

	err = px.plugins.EmitBeforeResponseSend(newCtx)
	if err != nil {
		return fmt.Errorf("error on EmitBeforeResponseSend: %w", err)
	}

	body, err := writeTunnelResponse(clientConn, newCtx.Response)
	if err != nil {
		return fmt.Errorf("error on read response body: %w", err)
	}
	defer newCtx.Response.Body.Close()

	newCtx.Response.Body = io.NopCloser(bytes.NewReader(body))

	resSnapshot, err := model.MakeResponseSnapshot(newCtx.Response, px.traceTimings)
	if err != nil {
		return fmt.Errorf("error on response snapshot creation: %w", err)
	}

	newCtx.SetResponse(newCtx.Response, resSnapshot)
	newCtx.Finish()

	err = px.plugins.EmitAfterResponseSend(newCtx)
	if err != nil {
		return fmt.Errorf("error on EmitAfterResponseSend: %w", err)
	}

	err = newCtx.Transition(model.StateFinished)
	if err != nil {
		return fmt.Errorf("cannot transit new context to finished state: %w", err)
	}

	return nil
}

func (px *Proxy) proceedRequest(
	ctx *model.RequestContext,
	r *http.Request,
) error {
	// Creates snapshot before triggers `before_request` creating snapshot for plugins.
	reqSnapshot, err := model.MakeRequestSnapshot(ctx.Request)
	if err != nil {
		return fmt.Errorf("error on request snapshot creation: %w", err)
	}

	ctx.SetRequestSnapshot(reqSnapshot)

	// Before the request send.
	err = px.plugins.EmitBeforeRequestSend(ctx)
	if err != nil {
		return fmt.Errorf("error on emit before request send: %w", err)
	}

	// Creates snapshot on after `before_request` trigger to update snapshot.
	reqSnapshot, err = model.MakeRequestSnapshot(ctx.Request)
	if err != nil {
		return fmt.Errorf("error on request snapshot creation: %w", err)
	}

	ctx.SetRequestSnapshot(reqSnapshot)

	// Verifies context.
	if ctx.IsFinished {
		return nil
	}

	// Prepares and sends the request.
	req, err := http.NewRequest(
		r.Method,
		r.URL.String(),
		r.Body,
	)
	if err != nil {
		return fmt.Errorf("error on new request creation: %w", err)
	}

	req = req.WithContext(ctx.Context)
	req.Header = r.Header.Clone()

	res, err := px.client.Do(req)
	if err != nil {
		return fmt.Errorf("error on proceed request: %w", err)
	}

	// Makes final req snapshot after the request was sent.
	reqSnapshot, err = model.MakeRequestSnapshot(ctx.Request)
	if err != nil {
		return fmt.Errorf("error on request snapshot creation: %w", err)
	}

	ctx.SetRequestSnapshot(reqSnapshot)

	// Emits with he request's snapshot.
	err = px.plugins.EmitAfterRequestSend(ctx)
	if err != nil {
		return fmt.Errorf("error on emit after request was sent: %w", err)
	}

	// Sets response but not its snapshot.
	ctx.SetResponse(res, nil)

	return nil
}

func isWebSocketRequest(r *http.Request) bool {
	if !headerContains(r.Header.Get("Connection"), "upgrade") {
		return false
	}

	if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		return false
	}

	if r.Header.Get("Sec-WebSocket-Key") == "" {
		return false
	}

	return true
}

func headerContains(v string, target string) bool {
	for part := range strings.SplitSeq(v, ",") {
		if strings.EqualFold(strings.TrimSpace(part), target) {
			return true
		}
	}
	return false
}
