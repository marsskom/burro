package proxy

import (
	"crypto/rsa"
	"crypto/x509"
	"io"
	"net/http"

	"gitlab.com/marsskom/burro/internal/events"
	"gitlab.com/marsskom/burro/internal/plugin"
)

type Proxy struct {
	plugins *plugin.Manager
	client  *http.Client

	caCert *x509.Certificate
	caKey  *rsa.PrivateKey
}

func New(
	pm *plugin.Manager,
	caCert *x509.Certificate,
	caKey *rsa.PrivateKey,
) *Proxy {
	return &Proxy{
		plugins: pm,
		client:  &http.Client{},

		caCert: caCert,
		caKey:  caKey,
	}
}

func (px *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &events.Context{
		Request:   r,
		Response:  nil,
		LastError: nil,
		Metadata:  map[string]any{},
		IsHandled: false,
	}

	err := px.plugins.EmitConnect(ctx)
	if err != nil {
		ctx.LastError = err
		px.plugins.EmitError(ctx)
	}

	if r.Method == http.MethodConnect {
		err = px.handleHTTPS(w, r)
		if err != nil {
			ctx.LastError = err
			px.plugins.EmitError(ctx)
		}

		err := px.plugins.EmitClose(ctx)
		if err != nil {
			ctx.LastError = err
			px.plugins.EmitError(ctx)
		}

		return
	}

	response, err := px.handleRequest(ctx.Request)
	if err != nil {
		ctx.LastError = err
		px.plugins.EmitError(ctx)

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer response.Body.Close()

	writeResponse(w, response)

	err = px.plugins.EmitClose(ctx)
	if err != nil {
		ctx.LastError = err
		px.plugins.EmitError(ctx)
	}
}

func writeResponse(w http.ResponseWriter, resp *http.Response) {
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
