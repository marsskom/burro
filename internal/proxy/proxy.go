package proxy

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"
	"time"

	"gitlab.com/marsskom/burro/internal/cert"
	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/plugin"
)

type Proxy struct {
	plugins *plugin.Manager
	session *model.Session

	client       *http.Client
	traceTimings *model.Timings

	caCert *x509.Certificate
	caKey  *rsa.PrivateKey

	clientCertCache *cert.CertCache
}

func NewProxy(
	pm *plugin.Manager,
	session *model.Session,
	caCert *x509.Certificate,
	caKey *rsa.PrivateKey,
) *Proxy {
	return &Proxy{
		plugins: pm,
		session: session,

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
		traceTimings: &model.Timings{},

		caCert: caCert,
		caKey:  caKey,

		clientCertCache: cert.NewCertCache(),
	}
}

func (px *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := model.NewCtx(px.session, r.WithContext(
		httptrace.WithClientTrace(r.Context(), px.traceTimings.AttachTrace()),
	))
	px.session.AddRequest(ctx)

	defer func() {
		err := px.plugins.EmitClose(ctx)
		if err != nil {
			px.plugins.EmitError(ctx, fmt.Errorf("ServeHTTP: error on EmitClose: %w", err))
		}

		ctx.Cancel()
	}()

	ctx.Transition(model.StateReceived)

	err := px.plugins.EmitConnect(ctx)
	if err != nil {
		ctx.Transition(model.StateFailed)
		px.plugins.EmitError(ctx, fmt.Errorf("ServeHTTP: error on EmitConnect: %w", err))

		return
	}

	if r.Method == http.MethodConnect {
		err = px.handleHTTPS(w, ctx)
	} else {
		err = px.handleHTTP(w, ctx)
	}

	if err != nil {
		ctx.Transition(model.StateFailed)
		px.plugins.EmitError(ctx, fmt.Errorf("ServeHTTP: error on handle request: %w", err))
	}
}
