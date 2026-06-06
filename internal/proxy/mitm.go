package proxy

import (
	"fmt"
	"net/http"
	"time"

	"gitlab.com/marsskom/burro/internal/logger"
	"gitlab.com/marsskom/burro/internal/model"
)

func (px *Proxy) handleTunnel(w http.ResponseWriter, ctx *model.RequestContext) error {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return fmt.Errorf("tunnel: no hijack support")
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		return fmt.Errorf("tunnel: cannot hijack connection: %w", err)
	}
	defer clientConn.Close()

	handshakeDeadline := time.Now().Add(10 * time.Second)

	clientConn.SetDeadline(handshakeDeadline)

	_, err = clientConn.Write([]byte(
		"HTTP/1.1 200 Connection Established\r\n\r\n",
	))
	if err != nil {
		return fmt.Errorf("tunnel: cannot send 200 message to connection: %w", err)
	}

	proto, conn, err := detectTunnelProtocol(clientConn)
	if err != nil {
		return fmt.Errorf("tunnel: cannot detect protocol: %w", err)
	}

	logger.Debug("guessed connection protocol", "protocol", proto)

	clientConn.SetDeadline(time.Time{})

	switch proto {
	case ProtoTLS:
		return px.handleHTTPS(conn, ctx)
	case ProtoHTTP:
		return px.handleRawHTTPOverConn(conn, ctx)
	default:
		return fmt.Errorf("tunnel: protocol '%s' not implemented", proto)
	}
}
