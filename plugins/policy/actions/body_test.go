package actions

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestRedactJSONField_RemovesField(t *testing.T) {
	in := []byte(`{"a":1,"secret":"hide","b":2}`)

	out := redactJSONField(in, "secret")

	if strings.Contains(string(out), "secret") {
		t.Fatalf("expected field to be removed, got %s", out)
	}
}

func TestRedactJSONField_FieldNotExists(t *testing.T) {
	in := []byte(`{"a":1,"b":2}`)

	out := redactJSONField(in, "secret")

	if string(out) != string(in) {
		t.Fatalf("expected unchanged JSON")
	}
}

func TestRedactJSONField_InvalidJSON(t *testing.T) {
	in := []byte(`{broken-json`)

	out := redactJSONField(in, "secret")

	if string(out) != string(in) {
		t.Fatalf("expected original data on invalid JSON")
	}
}

func TestRedactJSONField_NonObjectJSON(t *testing.T) {
	in := []byte(`"string value"`)

	out := redactJSONField(in, "secret")

	if string(out) != string(in) {
		t.Fatalf("expected unchanged for non-object JSON")
	}
}

func TestReadRequestBody_ReadAndRestore(t *testing.T) {
	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("hello world")),
	}

	b, err := readRequestBody(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(b) != "hello world" {
		t.Fatalf("unexpected body: %s", b)
	}

	// Ensures body is restored.
	b2, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read restored body error: %v", err)
	}

	if string(b2) != "hello world" {
		t.Fatalf("body not restored correctly")
	}
}

func TestReadRequestBody_NilBody(t *testing.T) {
	req := &http.Request{Body: nil}

	b, err := readRequestBody(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if b != nil {
		t.Fatalf("expected nil body")
	}
}

func TestReplaceRequestBody(t *testing.T) {
	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("old")),
	}

	replaceRequestBody(req, []byte("new-body"))

	b, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	if string(b) != "new-body" {
		t.Fatalf("expected new-body got %s", b)
	}

	if req.ContentLength != int64(len("new-body")) {
		t.Fatalf("content length not updated")
	}
}
