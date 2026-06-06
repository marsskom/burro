package proxy

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"syscall"

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

	reqSnapshot, err := model.MakeRequestSnapshot(newCtx.Request)
	if err != nil {
		return fmt.Errorf("error on request snapshot creation: %w", err)
	}

	newCtx.SetRequestSnapshot(reqSnapshot)

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

	err = newCtx.Response.Write(writer)
	if err != nil && !errors.Is(err, syscall.EPIPE) {
		return fmt.Errorf("cannot write response: %w", err)
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("flush error: %w", err)
	}

	if newCtx.Response != nil && newCtx.Response.Body != nil {
		newCtx.Response.Body.Close()
	}

	err = newCtx.Transition(model.StateFinished)
	if err != nil {
		return fmt.Errorf("cannot transit new context to finished state: %w", err)
	}

	return nil
}

func (px *Proxy) proceedRawRequest(ctx *model.RequestContext, r *http.Request) error {
	err := px.plugins.EmitRequest(ctx)
	if err != nil {
		return fmt.Errorf("proceedRawRequest: error on EmitRequest: %w", err)
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
		return fmt.Errorf("proceedRawRequest: error on new request creation: %w", err)
	}

	req = req.WithContext(ctx.Context)
	req.Header = r.Header.Clone()

	res, err := px.client.Do(req)
	if err != nil {
		return fmt.Errorf("proceedRawRequest: error on proceed request: %w", err)
	}

	resSnapshot, err := model.MakeResponseSnapshot(res, px.traceTimings)
	if err != nil {
		return fmt.Errorf("proceedRawRequest: error on response snapshot creation: %w", err)
	}

	ctx.Finish(res, resSnapshot)

	err = px.plugins.EmitResponse(ctx)
	if err != nil {
		return fmt.Errorf("proccedRawRequest: error on EmitResponse: %w", err)
	}

	return nil
}

func (px *Proxy) proceedRequest(
	ctx *model.RequestContext,
	r *http.Request,
) error {
	err := px.plugins.EmitRequest(ctx)
	if err != nil {
		return fmt.Errorf("proccedRequest: error on EmitRequest: %w", err)
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
		return fmt.Errorf("proccedRequest: error on new request creation: %w", err)
	}

	req = req.WithContext(ctx.Context)
	req.Header = r.Header.Clone()

	res, err := px.client.Do(req)
	if err != nil {
		return fmt.Errorf("proccedRequest: error on proceed request: %w", err)
	}

	resSnapshot, err := model.MakeResponseSnapshot(res, px.traceTimings)
	if err != nil {
		return fmt.Errorf("proccedRequest: error on response snapshot creation: %w", err)
	}

	ctx.Finish(res, resSnapshot)

	err = px.plugins.EmitResponse(ctx)
	if err != nil {
		return fmt.Errorf("proccedRequest: error on EmitResponse: %w", err)
	}

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
