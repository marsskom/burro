package actions

import (
	"net"
	"net/http"
	"strings"
)

func ActionMatch(m Match, req *http.Request) bool {
	switch {
	case len(m.All) > 0:
		for _, sub := range m.All {
			if !ActionMatch(sub, req) {
				return false
			}
		}
		return true

	case len(m.Any) > 0:
		for _, sub := range m.Any {
			if ActionMatch(sub, req) {
				return true
			}
		}
		return false

	case m.Not != nil:
		return !ActionMatch(*m.Not, req)
	}

	// Leaf evaluation.
	if m.Method != "" && req.Method != m.Method {
		return false
	}

	if m.Domain != "" && req.Host != m.Domain {
		return false
	}

	if m.Path != "" && !matchPattern(m.Path, req.URL.Path) {
		return false
	}

	if m.IP != "" {
		ip := extractIP(req.RemoteAddr)
		if ip != m.IP {
			return false
		}
	}

	for k, v := range m.Headers {
		if req.Header.Get(k) != v {
			return false
		}
	}

	return true
}

func matchPattern(pattern, path string) bool {
	// Case 1: /admin/*
	if prefix, ok := strings.CutSuffix(pattern, "/*"); ok {
		return path == prefix || strings.HasPrefix(path, prefix+"/")
	}

	// Case 2: /admin*
	if prefix, ok := strings.CutSuffix(pattern, "*"); ok {
		return strings.HasPrefix(path, prefix)
	}

	// Exact match.
	return pattern == path
}

func extractIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}
