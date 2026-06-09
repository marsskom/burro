package proxy

import (
	"bufio"
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

	err = px.proceedRawRequest(ctx, ctx.Request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return fmt.Errorf("HTTP: error on handle request: %w", err)
	}

	err = ctx.Transition(model.StateResponding)
	if err != nil {
		return fmt.Errorf("HTTP: canot transit context to responding state: %w", err)
	}

	writeResponse(w, ctx.Response)

	err = ctx.Transition(model.StateFinished)
	if err != nil {
		return fmt.Errorf("HTTP: canot transit context to finished state: %w", err)
	}

	return nil
}
