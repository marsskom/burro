package harexport

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"gitlab.com/marsskom/burro/internal/export"
	"gitlab.com/marsskom/burro/internal/model"
)

func TestHAR_OnRequest_CreatesEntry(t *testing.T) {
	p := New()

	ctx := newCtx()

	err := p.OnRequest(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(p.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(p.entries))
	}

	entry, ok := p.entries[ctx.ID]
	if !ok {
		t.Fatal("entry not found")
	}

	if entry.Request.Method != "GET" {
		t.Fatalf("wrong method: %s", entry.Request.Method)
	}

	if entry.Request.URL == "" {
		t.Fatal("missing URL")
	}
}

func TestHAR_OnResponse_EnrichesEntry(t *testing.T) {
	p := New()

	ctx := newCtx()

	_ = p.OnRequest(ctx)

	ctx.ResponseSnapshot = &model.ResponseSnapshot{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body:        []byte(`{"ok":true}`),
		TimeDNS:     10 * time.Millisecond,
		TimeConnect: 20 * time.Millisecond,
		TimeWait:    30 * time.Millisecond,
		TimeSSL:     40 * time.Millisecond,
	}

	err := p.OnResponse(ctx)
	if err != nil {
		t.Fatal(err)
	}

	entry := p.entries[ctx.ID]

	if entry.Response.Status != 200 {
		t.Fatalf("expected 200, got %d", entry.Response.Status)
	}

	if entry.Timings.DNS != 10 {
		t.Fatalf("expected DNS 10, got %d", entry.Timings.DNS)
	}
}

func TestHAR_Response_TextEncoding(t *testing.T) {
	p := New()

	ctx := newCtx()

	_ = p.OnRequest(ctx)

	ctx.ResponseSnapshot = &model.ResponseSnapshot{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: []byte(`{"hello":"world"}`),
	}

	_ = p.OnResponse(ctx)

	entry := p.entries[ctx.ID]

	if entry.Response.Content.Encoding != "" {
		t.Fatalf("expected plain text encoding, got %s", entry.Response.Content.Encoding)
	}

	if entry.Response.Content.Text == "" {
		t.Fatal("expected response body text")
	}
}

func TestHAR_Response_Base64Encoding(t *testing.T) {
	p := New()

	ctx := newCtx()

	_ = p.OnRequest(ctx)

	ctx.ResponseSnapshot = &model.ResponseSnapshot{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		Headers: map[string][]string{
			"Content-Type": {"application/octet-stream"},
		},
		Body: []byte{0x01, 0x02, 0x03},
	}

	_ = p.OnResponse(ctx)

	entry := p.entries[ctx.ID]

	if entry.Response.Content.Encoding != "base64" {
		t.Fatalf("expected base64 encoding, got %s", entry.Response.Content.Encoding)
	}
}

func TestHAR_Flush_WritesFile(t *testing.T) {
	p := New()
	p.outputFile = t.TempDir() + "/test.har"

	ctx := newCtx()

	_ = p.OnRequest(ctx)

	ctx.ResponseSnapshot = &model.ResponseSnapshot{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		Headers:    map[string][]string{},
		Body:       []byte(`ok`),
	}

	_ = p.OnResponse(ctx)

	err := p.Flush(&export.FileNameVars{
		Session: "sess1",
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(p.outputFile)
	if err != nil {
		t.Fatal(err)
	}

	var har HAR
	if err := json.Unmarshal(data, &har); err != nil {
		t.Fatal(err)
	}

	if len(har.Log.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(har.Log.Entries))
	}
}

func TestHAR_Flush_Empty(t *testing.T) {
	p := New()
	p.outputFile = t.TempDir() + "/empty.har"

	err := p.Flush(&export.FileNameVars{
		Session: "x",
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestHAR_Flush_MissingSession(t *testing.T) {
	p := New()
	p.outputFile = "/tmp/%session%.har"

	p.entries["x"] = &HAREntry{}

	err := p.Flush(&export.FileNameVars{
		Session: "",
	})

	if err == nil {
		t.Fatal("expected error for missing session")
	}
}

func newCtx() *model.RequestContext {
	req, _ := http.NewRequest("GET", "http://example.com/test", nil)

	return &model.RequestContext{
		ID:        "ctx-1",
		StartTime: time.Now(),
		RequestSnapshot: &model.RequestSnapshot{
			Method: "GET",
			URL:    "http://example.com/test",
			Proto:  "HTTP/1.1",
			Headers: map[string][]string{
				"Content-Type": {"application/json"},
			},
			QueryParams: map[string][]string{
				"q": {"test"},
			},
			Cookies: []*model.CookieSnapshot{},
			Body:    []byte(`{"ok":true}`),
		},
		Request: req,
	}
}
