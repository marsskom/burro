package model

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTransition_Valid(t *testing.T) {
	c := &RequestContext{}

	err := c.Transition(StatePrepared)
	if err != nil {
		t.Fatal(err)
	}

	if c.GetState() != StatePrepared {
		t.Fatal("state not updated")
	}
}

func TestTransition_InvalidFromStart(t *testing.T) {
	c := &RequestContext{}

	err := c.Transition(StateReceived)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTransition_Invalid(t *testing.T) {
	c := &RequestContext{}
	c.State.Store(int32(StateFinished))

	err := c.Transition(StateReceived)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExtractCookies(t *testing.T) {
	r := httptest.NewRequest("GET", "http://example.com", nil)

	r.AddCookie(&http.Cookie{
		Name:  "test",
		Value: "123",
		Path:  "/",
	})

	cookies, err := ExtractCookies(r)
	if err != nil {
		t.Fatal(err)
	}

	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	if cookies[0].Name != "test" {
		t.Fatal("wrong cookie name")
	}
}

func TestMakeRequestSnapshot(t *testing.T) {
	body := bytes.NewBufferString("hello world")

	r := httptest.NewRequest("POST", "http://example.com/path?q=1", body)
	r.Header.Set("X-Test", "1")

	snap, err := MakeRequestSnapshot(r)
	if err != nil {
		t.Fatal(err)
	}

	if snap.Method != "POST" {
		t.Fatal("wrong method")
	}

	if snap.Path != "/path" {
		t.Fatal("wrong path")
	}

	if string(snap.Body) != "hello world" {
		t.Fatal("body not preserved")
	}

	// Body must be reusable.
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	if buf.String() != "hello world" {
		t.Fatal("body not restored")
	}
}

func TestMakeResponseSnapshot(t *testing.T) {
	start := time.Now()

	timings := &Timings{
		DNSStart:     start,
		DNSEnd:       start.Add(10 * time.Millisecond),
		ConnectStart: start.Add(10 * time.Millisecond),
		ConnectEnd:   start.Add(25 * time.Millisecond),
		TLSStart:     start.Add(25 * time.Millisecond),
		TLSEnd:       start.Add(40 * time.Millisecond),
		WroteRequest: start.Add(40 * time.Millisecond),
		FirstByte:    start.Add(70 * time.Millisecond),
	}

	res := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		Header:     http.Header{},
		Body:       ioNopCloser("ok"),
	}

	snap, err := MakeResponseSnapshot(res, timings)
	if err != nil {
		t.Fatal(err)
	}

	if snap.TimeDNS != 10*time.Millisecond {
		t.Fatalf("expected 10ms DNS, got %v", snap.TimeDNS)
	}

	if snap.TimeConnect != 15*time.Millisecond {
		t.Fatalf("expected 15ms connect, got %v", snap.TimeConnect)
	}

	if snap.TimeWait != 30*time.Millisecond {
		t.Fatalf("expected 30ms wait, got %v", snap.TimeWait)
	}
}

func ioNopCloser(s string) *nopCloser {
	return &nopCloser{bytes.NewBufferString(s)}
}

type nopCloser struct {
	*bytes.Buffer
}

func (n *nopCloser) Close() error { return nil }

func TestBuildAbsoluteURL(t *testing.T) {
	r := httptest.NewRequest("GET", "/path", nil)
	r.Host = "example.com"

	url := BuildAbsoluteURL(r)

	if url == "" {
		t.Fatal("empty url")
	}
}
