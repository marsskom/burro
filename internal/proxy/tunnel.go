package proxy

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"unicode/utf8"

	"gitlab.com/marsskom/burro/internal/logger"
)

type PeekConn struct {
	net.Conn
	r io.Reader
}

func dataToString(p []byte) string {
	if utf8.Valid(p) {
		return string(p)
	} else {
		return hex.EncodeToString(p)
	}
}

func (c *PeekConn) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)

	logger.Trace(
		"connection read",
		"bytes", n,
		"err", err,
		"data", dataToString(p[:n]),
	)

	return n, err
}

func (c *PeekConn) Write(p []byte) (int, error) {
	logger.Trace(
		"connection write",
		"bytes", len(p),
		"data", dataToString(p),
	)

	n, err := c.Conn.Write(p)

	logger.Trace(
		"write result",
		"written", n,
		"err", err,
	)

	return n, err
}

type TunnelProtocol string

const (
	ProtoUnknown TunnelProtocol = "unknown"
	ProtoTLS     TunnelProtocol = "TLS"
	ProtoHTTP    TunnelProtocol = "HTTP"
)

func detectTunnelProtocol(conn net.Conn) (TunnelProtocol, *PeekConn, error) {
	buf := make([]byte, 8)

	n, err := conn.Read(buf)
	if err != nil {
		return ProtoUnknown, nil, fmt.Errorf("cannot read from connection: %w", err)
	}

	peekConn := &PeekConn{
		Conn: conn,
		r: io.MultiReader(
			bytes.NewReader(buf[:n]),
			conn,
		),
	}

	// TLS ClientHello.
	if len(buf) > 3 &&
		buf[0] == 0x16 &&
		buf[1] == 0x03 {
		return ProtoTLS, peekConn, nil
	}

	// HTTP methods.
	methods := [][]byte{
		[]byte("GET"),
		[]byte("POST"),
		[]byte("PUT"),
		[]byte("HEAD"),
		[]byte("PATCH"),
		[]byte("DELETE"),
		[]byte("OPTIONS"),
		[]byte("TRACE"),
	}

	for _, m := range methods {
		if len(buf) >= len(m) &&
			bytes.Equal(buf[:len(m)], m) {
			return ProtoHTTP, peekConn, nil
		}
	}

	return ProtoUnknown, peekConn, nil
}
