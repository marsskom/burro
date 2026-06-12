package proxy

import (
	"context"
	"errors"
	"net"
	"net/http"
)

func mapProxyError(err error) (int, string) {
	if err == nil {
		return 200, ""
	}

	// Context timeout.
	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusGatewayTimeout, "upstream timeout"
	}

	// Net errors.
	if netErr, ok := errors.AsType[*net.OpError](err); ok && netErr.Op == "dial" {
		return http.StatusBadGateway, "upstream unreachable"
	}

	if _, ok := errors.AsType[*net.DNSError](err); ok {
		return http.StatusBadGateway, "dns resolution failed"
	}

	return http.StatusInternalServerError, "proxy internal error"
}

func writeError(w http.ResponseWriter, err error) {
	code, msg := mapProxyError(err)

	http.Error(w, msg, code)
}
