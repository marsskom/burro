package proxy

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// errReader simulates a body that errors mid-read.
type errReader struct {
	data []byte
	pos  int
	err  error
}

func (r *errReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, r.err
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func makeResp(body string, contentType string, status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{contentType}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestWriteHTTPStream_Successful(t *testing.T) {
	body := "hello world"
	resp := makeResp(body, "text/plain", http.StatusOK)

	w := httptest.NewRecorder()
	got, err := writeHTTPStream(w, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Code != http.StatusOK {
		t.Errorf("status: want %d got %d", http.StatusOK, w.Code)
	}
	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("content-type: want text/plain got %s", w.Header().Get("Content-Type"))
	}
	if w.Body.String() != body {
		t.Errorf("client body: want %q got %q", body, w.Body.String())
	}
	if string(got) != body {
		t.Errorf("captured body: want %q got %q", body, string(got))
	}
}

func TestWriteHTTPStream_EmptyBody(t *testing.T) {
	resp := makeResp("", "text/plain", http.StatusOK)

	w := httptest.NewRecorder()
	got, err := writeHTTPStream(w, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty capture, got %d bytes", len(got))
	}
	if w.Body.Len() != 0 {
		t.Errorf("expected empty client body")
	}
}

func TestWriteHTTPStream_StatusCodePreserved(t *testing.T) {
	resp := makeResp("not found", "text/plain", http.StatusNotFound)

	w := httptest.NewRecorder()
	_, err := writeHTTPStream(w, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Code != http.StatusNotFound {
		t.Errorf("status: want %d got %d", http.StatusNotFound, w.Code)
	}
}

func TestWriteHTTPStream_CaptureLimitedStream(t *testing.T) {
	// 2MB body — client should get all, buffer only first 1MB.
	big := bytes.Repeat([]byte("x"), 2<<20)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/octet-stream"}},
		Body:       io.NopCloser(bytes.NewReader(big)),
	}

	w := httptest.NewRecorder()
	got, err := writeHTTPStream(w, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1<<20 {
		t.Errorf("capture: want %d bytes got %d", 1<<20, len(got))
	}
	if w.Body.Len() != 2<<20 {
		t.Errorf("client body: want %d bytes got %d", 2<<20, w.Body.Len())
	}
}

func TestWriteHTTPStream_BodyReadError(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Body: io.NopCloser(&errReader{
			data: []byte("partial"),
			err:  errors.New("read error"),
		}),
	}

	w := httptest.NewRecorder()
	_, err := writeHTTPStream(w, resp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteHTTPStream_Exactly1MB(t *testing.T) {
	// Exactly 1MB — should be fully captured.
	exact := bytes.Repeat([]byte("y"), 1<<20)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/octet-stream"}},
		Body:       io.NopCloser(bytes.NewReader(exact)),
	}

	w := httptest.NewRecorder()
	got, err := writeHTTPStream(w, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1<<20 {
		t.Errorf("capture: want %d bytes got %d", 1<<20, len(got))
	}
}

func TestWriteHTTPStream_ContentTypePreserved(t *testing.T) {
	resp := makeResp("{}", "application/json", http.StatusOK)

	w := httptest.NewRecorder()
	_, err := writeHTTPStream(w, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type: want application/json got %s", ct)
	}
}

// mockConn is a net.Conn that writes to a buffer and can simulate write errors.
type mockConn struct {
	buf        bytes.Buffer
	writeErr   error
	writeCount int // fails after this many writes if writeErr is set
	writes     int
}

func (m *mockConn) Write(p []byte) (int, error) {
	m.writes++
	if m.writeErr != nil && m.writes >= m.writeCount {
		return 0, m.writeErr
	}
	return m.buf.Write(p)
}

func (m *mockConn) Read(p []byte) (int, error)         { return 0, io.EOF }
func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (m *mockConn) RemoteAddr() net.Addr               { return &net.TCPAddr{} }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func parseTunnelResponse(t *testing.T, raw []byte) *http.Response {
	t.Helper()
	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(raw)), nil)
	if err != nil {
		t.Fatalf("failed to parse tunnel response: %v", err)
	}
	return resp
}

func TestWriteTunnelResponse_Successful(t *testing.T) {
	body := "hello tunnel"
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	conn := &mockConn{}
	got, err := writeTunnelResponse(conn, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed := parseTunnelResponse(t, conn.buf.Bytes())
	if parsed.StatusCode != http.StatusOK {
		t.Errorf("status: want 200 got %d", parsed.StatusCode)
	}
	clientBody, _ := io.ReadAll(parsed.Body)
	if string(clientBody) != body {
		t.Errorf("client body: want %q got %q", body, string(clientBody))
	}
	if string(got) != body {
		t.Errorf("captured: want %q got %q", body, string(got))
	}
}

func TestWriteTunnelResponse_StatusLineFormat(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusNotFound,
		Status:     "404 Not Found",
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader("")),
	}

	conn := &mockConn{}
	_, err := writeTunnelResponse(conn, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw := conn.buf.String()
	if !strings.HasPrefix(raw, "HTTP/1.1 404 Not Found\r\n") {
		t.Errorf("unexpected status line in: %q", raw[:min(len(raw), 40)])
	}
}

func TestWriteTunnelResponse_CaptureLimitedTo1MB(t *testing.T) {
	big := bytes.Repeat([]byte("x"), 2<<20)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Body:       io.NopCloser(bytes.NewReader(big)),
	}

	conn := &mockConn{}
	got, err := writeTunnelResponse(conn, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1<<20 {
		t.Errorf("capture: want %d bytes got %d", 1<<20, len(got))
	}
	// Client must receive full 2MB (plus headers).
	parsed := parseTunnelResponse(t, conn.buf.Bytes())
	clientBody, _ := io.ReadAll(parsed.Body)
	if len(clientBody) != 2<<20 {
		t.Errorf("client body: want %d bytes got %d", 2<<20, len(clientBody))
	}
}

func TestWriteTunnelResponse_WriteError(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader("body")),
	}

	// Fails on the first write (status line).
	conn := &mockConn{writeErr: errors.New("write error"), writeCount: 1}
	_, err := writeTunnelResponse(conn, resp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteTunnelResponse_HeadersForwarded(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: http.Header{
			"Content-Type":    []string{"application/json"},
			"X-Custom-Header": []string{"myvalue"},
		},
		Body: io.NopCloser(strings.NewReader("{}")),
	}

	conn := &mockConn{}
	_, err := writeTunnelResponse(conn, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed := parseTunnelResponse(t, conn.buf.Bytes())
	if ct := parsed.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: want application/json got %s", ct)
	}
	if xc := parsed.Header.Get("X-Custom-Header"); xc != "myvalue" {
		t.Errorf("X-Custom-Header: want myvalue got %s", xc)
	}
}
