package proxy

import (
	"crypto/rsa"
	"crypto/x509"
	"net"
	"net/http"
	"time"

	"gitlab.com/marsskom/burro/internal/plugin"
	"gitlab.com/marsskom/burro/internal/request"
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
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,

				TLSHandshakeTimeout:   5 * time.Second,
				ResponseHeaderTimeout: 15 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,

				MaxIdleConns:        200,
				MaxIdleConnsPerHost: 50,
				IdleConnTimeout:     90 * time.Second,
			},
		},

		caCert: caCert,
		caKey:  caKey,
	}
}

func (px *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := request.New(r)

	defer func() {
		err := px.plugins.EmitClose(ctx)
		if err != nil {
			px.plugins.EmitError(ctx, err)
		}

		ctx.Cancel()
	}()

	ctx.Transition(request.StateReceived)

	err := px.plugins.EmitConnect(ctx)
	if err != nil {
		ctx.Transition(request.StateFailed)
		px.plugins.EmitError(ctx, err)

		return
	}

	if r.Method == http.MethodConnect {
		err = px.handleHTTPS(w, ctx)
	} else {
		err = px.handleHTTP(w, ctx)
	}

	if err != nil {
		ctx.Transition(request.StateFailed)
		px.plugins.EmitError(ctx, err)
	}
}
