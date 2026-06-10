package proxy

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

func writeResponse(w http.ResponseWriter, resp *http.Response) {
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func writeHTTPStream(w http.ResponseWriter, resp *http.Response) ([]byte, error) {
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)

	flusher, _ := w.(http.Flusher)

	var buf bytes.Buffer
	tmp := make([]byte, 32<<10) // 32KB
	capturedLimit := 1 << 20    // 1MB
	captured := 0

	for {
		n, err := resp.Body.Read(tmp)
		if n > 0 {
			if captured < capturedLimit {
				take := min(n, capturedLimit-captured)
				buf.Write(tmp[:take])
				captured += take
			}

			if _, werr := w.Write(tmp[:n]); werr != nil {
				return nil, werr
			}

			if flusher != nil {
				flusher.Flush()
			}
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func isHTTPStreamingResponse(resp *http.Response) bool {
	if resp == nil {
		return false
	}

	if strings.Contains(resp.Header.Get("Transfer-Encoding"), "chunked") {
		return true
	}

	if strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		return true
	}

	// HTTP/2 often has no chunked encoding but still streams.
	if resp.ProtoMajor == 2 {
		return true
	}

	// Fallback.
	if resp.ContentLength == -1 {
		return true
	}

	return false
}

func writeTunnelResponse(clientConn net.Conn, resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	// Writes status line + headers manually.
	statusLine := fmt.Sprintf("HTTP/%d.%d %s\r\n", resp.ProtoMajor, resp.ProtoMinor, resp.Status)
	if _, err := fmt.Fprint(clientConn, statusLine); err != nil {
		return nil, err
	}
	if err := resp.Header.Write(clientConn); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprint(clientConn, "\r\n"); err != nil {
		return nil, err
	}

	// Caps buffering or only capture up to N bytes.
	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)
	limitedTee := io.LimitReader(tee, 1<<20) // capture first 1MB into buf
	io.Copy(clientConn, limitedTee)
	// After limit is hit, drains the rest to client only.
	io.Copy(clientConn, resp.Body)

	// Signals EOF to client.
	if tc, ok := clientConn.(*tls.Conn); ok {
		tc.CloseWrite()
	}

	return buf.Bytes(), nil
}
