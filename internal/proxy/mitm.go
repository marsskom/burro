package proxy

import (
	"bufio"
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab.com/marsskom/burro/internal/cert"
	"gitlab.com/marsskom/burro/internal/request"
)

func (px *Proxy) handleHTTPS(w http.ResponseWriter, ctx *request.RequestContext) error {
	hijacker := w.(http.Hijacker)

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		return err
	}
	defer clientConn.Close()

	handshakeDeadline := time.Now().Add(10 * time.Second)
	ioDeadline, ok := ctx.Context.Deadline()
	if !ok {
		ioDeadline = time.Now().Add(30 * time.Second)
	}

	clientConn.SetDeadline(handshakeDeadline)

	_, err = clientConn.Write([]byte(
		"HTTP/1.1 200 Connection Established\r\n\r\n",
	))
	if err != nil {
		return err
	}

	err = ctx.Transition(request.StatePrepared)
	if err != nil {
		return err
	}

	host := ctx.Request.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	fakeCert, err := cert.GenerateHostCertificate(host, px.caCert, px.caKey)
	if err != nil {
		return err
	}

	tlsConn := tls.Server(clientConn, &tls.Config{
		Certificates: []tls.Certificate{*fakeCert},
	})
	defer tlsConn.Close()

	handshakeDone := make(chan error, 1)
	go func() {
		handshakeDone <- tlsConn.Handshake()
	}()

	select {
	case err := <-handshakeDone:
		if err != nil {
			return err
		}
	case <-ctx.Context.Done():
		err := ctx.Transition(request.StateCanceled)
		if err != nil {
			return err
		}

		return ctx.Context.Err()
	}

	clientConn.SetDeadline(time.Time{})

	select {
	case <-ctx.Context.Done():
		err := ctx.Transition(request.StateCanceled)
		if err != nil {
			return err
		}

		return ctx.Context.Err()
	default:
	}

	reader := bufio.NewReader(tlsConn)

	for {
		tlsConn.SetReadDeadline(ioDeadline)

		req, err := http.ReadRequest(reader)
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return err
		}

		if req.Close {
			return nil
		}

		newCtx := request.NewFromParent(ctx, req)

		req = req.WithContext(newCtx.Context)
		req.URL.Scheme = "https"
		req.URL.Host = newCtx.Request.Host
		req.RequestURI = ""

		newCtx.Transition(request.StateForwarding)

		err = px.handleRequest(newCtx, req)
		if err != nil {
			return err
		}

		newCtx.Transition(request.StateResponding)

		tlsConn.SetWriteDeadline(ioDeadline)

		err = newCtx.Response.Write(tlsConn)
		if err != nil {
			return err
		}

		tlsConn.SetWriteDeadline(time.Time{})

		if newCtx.Response != nil && newCtx.Response.Body != nil {
			newCtx.Response.Body.Close()
		}

		newCtx.Transition(request.StateFinished)
	}
}
