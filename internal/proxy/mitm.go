package proxy

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"syscall"
	"time"

	"gitlab.com/marsskom/burro/internal/cert"
	"gitlab.com/marsskom/burro/internal/model"
)

// TODO: optimize.
func (px *Proxy) handleHTTPS(w http.ResponseWriter, ctx *model.RequestContext) error {
	hijacker := w.(http.Hijacker)

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		return fmt.Errorf("handleHTTPS: cannot hijack connection: %w", err)
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
		return fmt.Errorf("handleHTTPS: cannot send 200 message to connection: %w", err)
	}

	err = ctx.Transition(model.StatePrepared)
	if err != nil {
		return fmt.Errorf("handleHTTPS: cannot transit context to prepared state: %w", err)
	}

	reqSnapshot, err := model.MakeRequestSnapshot(ctx.Request)
	if err != nil {
		return fmt.Errorf("HandleHTTP: error on request snapshot creation: %w", err)
	}

	ctx.RequestSnapshot = reqSnapshot

	host := ctx.Request.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	cacheKey := cert.CertCacheHostKey(host)
	cachedCert, ok := px.clientCertCache.Get(cacheKey)
	var fakeCert *tls.Certificate
	if !ok {
		fakeCert, err = cert.GenerateHostCertificate(host, px.caCert, px.caKey)
		if err != nil {
			return fmt.Errorf("handleHTTPS: cannot generate host certificate: %w", err)
		}

		err = px.clientCertCache.Set(cacheKey, &cert.CertCacheItem{
			Certificate: fakeCert,
		})
		if err != nil {
			return fmt.Errorf("handleHTTPS: cannot add certificate into cache: %w", err)
		}
	} else {
		fakeCert = cachedCert.Certificate
	}

	tlsConn := tls.Server(clientConn, &tls.Config{
		Certificates: []tls.Certificate{*fakeCert},
	})
	defer tlsConn.Close()

	tlsConn.SetReadDeadline(ioDeadline)

	handshakeDone := make(chan error, 1)
	go func() {
		handshakeDone <- tlsConn.Handshake()
	}()

	select {
	case err := <-handshakeDone:
		if err != nil {
			return fmt.Errorf("handleHTTPS: error during handshake: %w", err)
		}
	case <-ctx.Context.Done():
		err := ctx.Transition(model.StateCanceled)
		if err != nil {
			return fmt.Errorf("handleHTTPS: cannot transit context to canceled state: %w", err)
		}

		return ctx.Context.Err()
	}

	clientConn.SetDeadline(time.Time{})

	select {
	case <-ctx.Context.Done():
		err := ctx.Transition(model.StateCanceled)
		if err != nil {
			return fmt.Errorf("handleHTTPS: cannot transit context to canceled state: %w", err)
		}

		return ctx.Context.Err()
	default:
	}

	reader := bufio.NewReader(tlsConn)

	for {
		// Set deadline for each request separately.
		tlsConn.SetReadDeadline(time.Now().Add(60 * time.Second))

		req, err := http.ReadRequest(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}

			return fmt.Errorf("handleHTTPS: problem to read request: %w", err)
		}

		if req == nil {
			continue
		}

		if req.Close {
			break
		}

		newCtx := model.NewCtxFromParent(ctx, req)
		px.session.AddRequest(newCtx)

		err = newCtx.Transition(model.StatePrepared)
		if err != nil {
			return fmt.Errorf("handleHTTPS: cannot transit new context to prepared state: %w", err)
		}

		reqSnapshot, err := model.MakeRequestSnapshot(newCtx.Request)
		if err != nil {
			return fmt.Errorf("HandleHTTP: error on request snapshot creation: %w", err)
		}

		newCtx.RequestSnapshot = reqSnapshot

		req = req.WithContext(newCtx.Context)
		req.URL.Scheme = "https"
		req.URL.Host = newCtx.Request.Host
		req.Host = newCtx.Request.Host
		req.RequestURI = ""

		err = newCtx.Transition(model.StateForwarding)
		if err != nil {
			return fmt.Errorf("handleHTTPS: cannot transit new context to forwarding state: %w", err)
		}

		err = px.handleRequest(newCtx, req)
		if err != nil {
			return fmt.Errorf("handleHTTPS: error on handle request: %w", err)
		}

		newCtx.Transition(model.StateResponding)
		if err != nil {
			return fmt.Errorf("handleHTTPS: cannot transit new context to responding state: %w", err)
		}

		tlsConn.SetWriteDeadline(time.Now().Add(60 * time.Second))

		err = newCtx.Response.Write(tlsConn)
		if err != nil && !errors.Is(err, syscall.EPIPE) {
			return fmt.Errorf("handleHTTPS: cannot write response: %w", err)
		}

		tlsConn.SetWriteDeadline(time.Time{})

		if newCtx.Response != nil && newCtx.Response.Body != nil {
			newCtx.Response.Body.Close()
		}

		err = newCtx.Transition(model.StateFinished)
		if err != nil {
			return fmt.Errorf("handleHTTPS: cannot transit new context to finished state: %w", err)
		}
	}

	return nil
}
