package logger

import (
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/marsskom/burro/internal/model"
)

func TestLogger_OnRequest(t *testing.T) {
	buf := captureLogs(t)

	p := New()

	req, _ := http.NewRequest("GET", "http://example.com/test", nil)

	ctx := &model.RequestContext{
		ID:      "req-1",
		Request: req,
	}

	err := p.OnRequest(ctx)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()

	if !strings.Contains(out, "Request received") {
		t.Fatalf("expected log output, got: %s", out)
	}

	if !strings.Contains(out, "req-1") {
		t.Fatalf("expected request ID in log")
	}
}

func TestLogger_OnError(t *testing.T) {
	buf := captureLogs(t)

	p := New()

	ctx := &model.RequestContext{
		ID: "err-1",
	}

	_ = p.OnError(ctx, http.ErrAbortHandler)

	out := buf.String()

	if !strings.Contains(out, "Error occurred") {
		t.Fatal("missing error log message")
	}

	if !strings.Contains(out, "err-1") {
		t.Fatal("missing context ID")
	}
}

func TestLogger_NilFieldsSafety(t *testing.T) {
	p := New()

	ctx := &model.RequestContext{
		ID: "nil-test",
	}

	// Must not panic.
	_ = p.OnConnect(ctx)
	_ = p.OnRequest(ctx)
	_ = p.OnResponse(ctx)
	_ = p.OnError(ctx, nil)
	_ = p.OnClose(ctx)
}

func TestLogger_ResponseContextIncluded(t *testing.T) {
	buf := captureLogs(t)

	p := New()

	ctx := &model.RequestContext{
		ID: "resp-1",
		Response: &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Proto:      "HTTP/1.1",
		},
	}

	_ = p.OnResponse(ctx)

	out := buf.String()

	if !strings.Contains(out, "Response received") {
		t.Fatal("missing response log")
	}

	if !strings.Contains(out, "200") {
		t.Fatal("missing status code in log output")
	}
}

func TestLogger_MetadataSafety(t *testing.T) {
	buf := captureLogs(t)

	p := New()

	ctx := &model.RequestContext{
		ID: "meta-1",
		Metadata: map[string]any{
			"user": "test",
		},
	}

	_ = p.OnRequest(ctx)

	out := buf.String()

	if !strings.Contains(out, "meta-1") {
		t.Fatal("missing ID in log")
	}

	if !strings.Contains(out, "user") {
		t.Fatal("missing metadata in log")
	}
}

func captureLogs(t *testing.T) *strings.Builder {
	t.Helper()

	var buf strings.Builder

	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	slog.SetDefault(slog.New(handler))

	return &buf
}
