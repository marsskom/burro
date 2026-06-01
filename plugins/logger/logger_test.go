package logger

import (
	"net/http"
	"strings"
	"testing"

	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/testutils"
)

func TestLogger_OnRequest(t *testing.T) {
	runtime := testutils.NewForPlugin("")

	p := New()
	p.Init(runtime, map[string]any{})

	req, _ := http.NewRequest("GET", "http://example.com/test", nil)

	ctx := &model.RequestContext{
		ID:      "req-1",
		Request: req,
	}

	err := p.OnRequest(ctx)
	if err != nil {
		t.Fatal(err)
	}

	logger, _ := runtime.Log().(*testutils.MemoryLogger)
	if len(logger.Messages["info"]) != 1 {
		t.Fatalf("expected 1 info message, got: %d", len(logger.Messages["info"]))
	}

	message := logger.Messages["info"][0]
	if !strings.Contains(message, "request received") {
		t.Fatalf("expected log output, got: %s", message)
	}

	if !strings.Contains(message, "req-1") {
		t.Fatalf("expected request ID in log")
	}
}

func TestLogger_OnError(t *testing.T) {
	runtime := testutils.NewForPlugin("")

	p := New()
	p.Init(runtime, map[string]any{})

	ctx := &model.RequestContext{
		ID: "err-1",
	}

	_ = p.OnError(ctx, http.ErrAbortHandler)

	logger, _ := runtime.Log().(*testutils.MemoryLogger)
	if len(logger.Messages["error"]) != 1 {
		t.Fatalf("expected 1 error message, got: %d", len(logger.Messages["error"]))
	}

	message := logger.Messages["error"][0]

	if !strings.Contains(message, "error occurred") {
		t.Fatal("missing error log message")
	}

	if !strings.Contains(message, "err-1") {
		t.Fatal("missing context ID")
	}
}

func TestLogger_NilFieldsSafety(t *testing.T) {
	runtime := testutils.NewForPlugin("")

	p := New()
	p.Init(runtime, map[string]any{})

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
	runtime := testutils.NewForPlugin("")

	p := New()
	p.Init(runtime, map[string]any{})

	ctx := &model.RequestContext{
		ID: "resp-1",
		Response: &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Proto:      "HTTP/1.1",
		},
	}

	_ = p.OnResponse(ctx)

	logger, _ := runtime.Log().(*testutils.MemoryLogger)
	if len(logger.Messages["info"]) != 1 {
		t.Fatalf("expected 1 info message, got: %d", len(logger.Messages["info"]))
	}

	message := logger.Messages["info"][0]

	if !strings.Contains(message, "response received") {
		t.Fatal("missing response log")
	}

	if !strings.Contains(message, "200") {
		t.Fatal("missing status code in log output")
	}
}

func TestLogger_MetadataSafety(t *testing.T) {
	runtime := testutils.NewForPlugin("")

	p := New()
	p.Init(runtime, map[string]any{})

	ctx := &model.RequestContext{
		ID: "meta-1",
		Metadata: map[string]any{
			"user": "test",
		},
	}

	_ = p.OnRequest(ctx)

	logger, _ := runtime.Log().(*testutils.MemoryLogger)
	if len(logger.Messages["info"]) != 1 {
		t.Fatalf("expected 1 info message, got: %d", len(logger.Messages["info"]))
	}

	message := logger.Messages["info"][0]

	if !strings.Contains(message, "meta-1") {
		t.Fatal("missing ID in log")
	}

	if !strings.Contains(message, "user") {
		t.Fatal("missing metadata in log")
	}
}
