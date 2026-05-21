package proxy

import (
	"bufio"
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab.com/marsskom/burro/internal/cert"
)

func (px *Proxy) handleHTTPS(w http.ResponseWriter, r *http.Request) error {
	hijacker := w.(http.Hijacker)

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		return err
	}
	defer clientConn.Close()

	clientConn.SetDeadline(time.Now().Add(30 * time.Second))

	_, err = clientConn.Write([]byte(
		"HTTP/1.1 200 Connection Established\r\n\r\n",
	))
	if err != nil {
		return err
	}

	host := r.Host
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

	err = tlsConn.Handshake()
	if err != nil {
		return err
	}

	reader := bufio.NewReader(tlsConn)

	req, err := http.ReadRequest(reader)
	if err != nil {
		if err == io.EOF {
			return nil
		}

		return err
	}

	if req.Body != nil {
		io.Copy(io.Discard, req.Body)
		req.Body.Close()
	}

	req.URL.Scheme = "https"
	req.URL.Host = r.Host
	req.RequestURI = ""

	resp, err := px.handleRequest(req)
	if err != nil {
		return err
	}

	err = resp.Write(tlsConn)
	if err != nil {
		return err
	}

	resp.Body.Close()

	return nil
}
