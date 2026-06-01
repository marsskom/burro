package broker

import (
	"net/http"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"gitlab.com/marsskom/burro/internal/model"
)

func baseCtx() *model.RequestContext {
	return &model.RequestContext{
		ID: "req-1",
		Session: &model.Session{
			ID: "sess-1",
		},
		Request: &http.Request{
			Method: "GET",
			Host:   "example.com",
			Proto:  "HTTP/1.1",
			URL: &url.URL{
				Scheme:   "http",
				Host:     "example.com",
				Path:     "/test",
				RawQuery: "a=1",
			},
			RemoteAddr: "127.0.0.1",
		},
		StartTime:  time.UnixMilli(1000),
		CreatedAt:  time.UnixMilli(1000),
		UpdatedAt:  time.UnixMilli(2000),
		State:      atomic.Int32{},
		IsFinished: true,
	}
}

func TestToBrokerEvent_MissingID(t *testing.T) {
	ctx := baseCtx()
	ctx.ID = ""

	_, err := ToBrokerEvent(EventRequest, ctx)
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestToBrokerEvent_NoRequestSnapshot(t *testing.T) {
	ctx := baseCtx()
	ctx.RequestSnapshot = nil

	ev, err := ToBrokerEvent(EventRequest, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ev.ID != "req-1" {
		t.Fatalf("wrong ID: %s", ev.ID)
	}

	if ev.Host != "example.com" {
		t.Fatalf("expected host from Request: %s", ev.Host)
	}

	if ev.QueryParams != "" {
		t.Fatalf("expected empty query params in fallback")
	}
}

func TestToBrokerEvent_WithRequestSnapshot(t *testing.T) {
	ctx := baseCtx()

	ctx.RequestSnapshot = &model.RequestSnapshot{
		Proto:         "HTTP/2",
		Scheme:        "https",
		Host:          "api.test.com",
		Method:        "POST",
		URL:           "https://api.test.com/x",
		Path:          "/x",
		ContentLength: 10,
		RemoteAddr:    "10.0.0.1",
		Body:          []byte("hello"),

		Headers:     map[string][]string{"h": {"v"}},
		QueryParams: map[string][]string{"q": {"1"}},
		Cookies:     make([]*model.CookieSnapshot, 0),
	}

	ev, err := ToBrokerEvent(EventRequest, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ev.Method != "POST" {
		t.Fatalf("expected POST, got %s", ev.Method)
	}

	if ev.Scheme != "https" {
		t.Fatalf("expected https scheme")
	}

	if ev.ContentLength != 10 {
		t.Fatalf("wrong content length")
	}

	if len(ev.RequestBody) != 5 {
		t.Fatalf("expected body length 5")
	}
}

func TestToBrokerEvent_MetadataConversion(t *testing.T) {
	ctx := baseCtx()

	ctx.Metadata = map[string]any{
		"k": "v",
	}

	ev, err := ToBrokerEvent(EventRequest, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ev.Metadata == "" {
		t.Fatal("expected metadata string")
	}
}

func TestToBrokerEvent_WithResponseSnapshot(t *testing.T) {
	ctx := baseCtx()

	ctx.RequestSnapshot = &model.RequestSnapshot{
		Proto:         "HTTP/1.1",
		Scheme:        "http",
		Host:          "x",
		Method:        "GET",
		URL:           "/",
		Path:          "/",
		ContentLength: 1,
		Body:          []byte("a"),
		Headers:       map[string][]string{},
		QueryParams:   map[string][]string{},
		Cookies:       make([]*model.CookieSnapshot, 0),
	}

	ctx.ResponseSnapshot = &model.ResponseSnapshot{
		Status:        "OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ContentLength: 3,
		Body:          []byte("abc"),
		Headers:       map[string][]string{"r": {"1"}},
	}

	ev, err := ToBrokerEvent(EventRequest, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ev.ResponseStatusCode != 200 {
		t.Fatalf("expected 200, got %d", ev.ResponseStatusCode)
	}

	if len(ev.ResponseBody) != 3 {
		t.Fatal("wrong response body")
	}
}
