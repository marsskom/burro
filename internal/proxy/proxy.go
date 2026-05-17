package proxy

import (
	"io"
	"net/http"

	"gitlab.com/marsskom/burro/internal/events"
	"gitlab.com/marsskom/burro/internal/plugin"
)

type Proxy struct {
	plugins *plugin.Manager
	client  *http.Client
}

func New(pm *plugin.Manager) *Proxy {
	return &Proxy{
		plugins: pm,
		client:  &http.Client{},
	}
}

func (px *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &events.Context{
		Request:     r,
		Response:    nil,
		Metadata:    map[string]any{},
		IsCancelled: false,
	}

	err := px.plugins.EmitConnect(ctx)
	if err != nil {
		px.plugins.EmitError(ctx)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest(
		r.Method,
		r.URL.String(),
		r.Body,
	)
	if err != nil {
		px.plugins.EmitError(ctx)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = px.plugins.EmitRequest(ctx)
	if err != nil {
		px.plugins.EmitError(ctx)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header = r.Header.Clone()

	resp, err := px.client.Do(req)
	if err != nil {
		px.plugins.EmitError(ctx)

		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	defer resp.Body.Close()

	ctx.Response = resp

	err = px.plugins.EmitResponse(ctx)
	if err != nil {
		px.plugins.EmitError(ctx)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)
}
