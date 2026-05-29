package model

import (
	"crypto/tls"
	"net/http/httptrace"
	"time"
)

type Timings struct {
	DNSStart     time.Time
	DNSEnd       time.Time
	ConnectStart time.Time
	ConnectEnd   time.Time
	TLSStart     time.Time
	TLSEnd       time.Time
	GotConn      time.Time
	WroteRequest time.Time
	FirstByte    time.Time
}

func (t *Timings) AttachTrace() *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		DNSStart: func(httptrace.DNSStartInfo) {
			t.DNSStart = time.Now()
		},
		DNSDone: func(httptrace.DNSDoneInfo) {
			t.DNSEnd = time.Now()
		},

		ConnectStart: func(_, _ string) {
			t.ConnectStart = time.Now()
		},
		ConnectDone: func(_, _ string, _ error) {
			t.ConnectEnd = time.Now()
		},

		TLSHandshakeStart: func() {
			t.TLSStart = time.Now()
		},
		TLSHandshakeDone: func(tls.ConnectionState, error) {
			t.TLSEnd = time.Now()
		},

		WroteRequest: func(httptrace.WroteRequestInfo) {
			t.WroteRequest = time.Now()
		},

		GotFirstResponseByte: func() {
			t.FirstByte = time.Now()
		},
	}
}
