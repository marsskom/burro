package proxy

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"gitlab.com/marsskom/burro/internal/cert"
	"gitlab.com/marsskom/burro/internal/model"
)

func (px *Proxy) handleHTTPS(
	clientConn net.Conn,
	ctx *model.RequestContext,
) error {
	ctx.SetPrototol(model.HTTPS)

	err := ctx.Transition(model.StatePrepared)
	if err != nil {
		return fmt.Errorf("HTTPS: cannot transit context to prepared state: %w", err)
	}

	reqSnapshot, err := model.MakeRequestSnapshot(ctx.Request)
	if err != nil {
		return fmt.Errorf("HTTPS: error on request snapshot creation: %w", err)
	}

	ctx.SetRequestSnapshot(reqSnapshot)

	host := ctx.Request.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	cacheKey := cert.NewCertCacheHostKey(host)
	cachedCert, ok := px.clientCertCache.Get(cacheKey)
	var fakeCert *tls.Certificate
	if !ok {
		fakeCert, err = cert.GenerateHostCertificate(host, px.caCert, px.caKey)
		if err != nil {
			return fmt.Errorf("HTTPS: cannot generate host certificate: %w", err)
		}

		err = px.clientCertCache.Set(cacheKey, &cert.CertCacheItem{
			Certificate: fakeCert,
		})
		if err != nil {
			return fmt.Errorf("HTTPS: cannot add certificate into cache: %w", err)
		}
	} else {
		fakeCert = cachedCert.Certificate
	}

	tlsConn := tls.Server(clientConn, &tls.Config{
		Certificates:       []tls.Certificate{*fakeCert},
		InsecureSkipVerify: px.tls.Insecure,
	})
	defer tlsConn.Close()

	ioDeadline, ok := ctx.Context.Deadline()
	if !ok {
		ioDeadline = time.Now().Add(30 * time.Second)
	}

	tlsConn.SetReadDeadline(ioDeadline)

	handshakeDone := make(chan error, 1)
	go func() {
		handshakeDone <- tlsConn.Handshake()
	}()

	select {
	case err := <-handshakeDone:
		if err != nil {
			return fmt.Errorf("HTTPS: error during handshake: %w", err)
		}
	case <-ctx.Context.Done():
		err := ctx.Transition(model.StateCanceled)
		if err != nil {
			return fmt.Errorf("HTTPS: cannot transit context to canceled state: %w", err)
		}

		return ctx.Context.Err()
	}

	clientConn.SetDeadline(time.Time{})

	select {
	case <-ctx.Context.Done():
		err := ctx.Transition(model.StateCanceled)
		if err != nil {
			return fmt.Errorf("HTTPS: cannot transit context to canceled state: %w", err)
		}

		return ctx.Context.Err()
	default:
	}

	reader := bufio.NewReader(tlsConn)
	writer := bufio.NewWriter(tlsConn)

	for {
		// Set deadline for each request separately.
		tlsConn.SetReadDeadline(time.Now().Add(60 * time.Second))
		tlsConn.SetWriteDeadline(time.Now().Add(60 * time.Second))

		err := px.proceedTunnelRequest(
			tlsConn,
			"https",
			ctx,
			reader,
			writer,
		)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}

			return fmt.Errorf("HTTPS: %w", err)
		}

		tlsConn.SetWriteDeadline(time.Time{})
	}

	return nil
}
