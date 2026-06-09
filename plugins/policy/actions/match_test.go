package actions

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func req(method, host, path, ip string) *http.Request {
	r := httptest.NewRequest(method, "http://"+host+path, nil)
	r.Host = host
	r.RemoteAddr = ip + ":12345"
	return r
}

func TestActionMatch(t *testing.T) {

	tests := []struct {
		name  string
		match Match
		req   *http.Request
		want  bool
	}{
		{
			name:  "method match",
			match: Match{Method: "GET"},
			req:   req("GET", "example.com", "/a", "1.1.1.1"),
			want:  true,
		},
		{
			name:  "method mismatch",
			match: Match{Method: "POST"},
			req:   req("GET", "example.com", "/a", "1.1.1.1"),
			want:  false,
		},

		{
			name:  "domain match",
			match: Match{Domain: "example.com"},
			req:   req("GET", "example.com", "/a", "1.1.1.1"),
			want:  true,
		},
		{
			name:  "domain mismatch",
			match: Match{Domain: "other.com"},
			req:   req("GET", "example.com", "/a", "1.1.1.1"),
			want:  false,
		},

		{
			name:  "path exact match",
			match: Match{Path: "/a"},
			req:   req("GET", "example.com", "/a", "1.1.1.1"),
			want:  true,
		},

		{
			name:  "path prefix /*",
			match: Match{Path: "/admin/*"},
			req:   req("GET", "example.com", "/admin/user", "1.1.1.1"),
			want:  true,
		},

		{
			name:  "path prefix /* matches exact path",
			match: Match{Path: "/admin/*"},
			req:   req("GET", "example.com", "/admin", "1.1.1.1"),
			want:  true,
		},

		{
			name:  "path prefix fail",
			match: Match{Path: "/admin/*"},
			req:   req("GET", "example.com", "/admine/user", "1.1.1.1"),
			want:  false,
		},

		{
			name:  "path wildcard *",
			match: Match{Path: "/admin*"},
			req:   req("GET", "example.com", "/admin123", "1.1.1.1"),
			want:  true,
		},

		{
			name: "headers match",
			match: Match{
				Headers: map[string]string{
					"X-Test": "1",
				},
			},
			req: func() *http.Request {
				r := req("GET", "example.com", "/a", "1.1.1.1")
				r.Header.Set("X-Test", "1")
				return r
			}(),
			want: true,
		},

		{
			name: "headers mismatch",
			match: Match{
				Headers: map[string]string{
					"X-Test": "2",
				},
			},
			req: func() *http.Request {
				r := req("GET", "example.com", "/a", "1.1.1.1")
				r.Header.Set("X-Test", "1")
				return r
			}(),
			want: false,
		},

		{
			name:  "ip match",
			match: Match{IP: "1.1.1.1"},
			req:   req("GET", "example.com", "/a", "1.1.1.1"),
			want:  true,
		},

		{
			name:  "ip mismatch",
			match: Match{IP: "2.2.2.2"},
			req:   req("GET", "example.com", "/a", "1.1.1.1"),
			want:  false,
		},

		{
			name: "all success",
			match: Match{
				All: []Match{
					{Method: "GET"},
					{Domain: "example.com"},
				},
			},
			req:  req("GET", "example.com", "/a", "1.1.1.1"),
			want: true,
		},

		{
			name: "any success",
			match: Match{
				Any: []Match{
					{Method: "POST"},
					{Method: "GET"},
				},
			},
			req:  req("GET", "example.com", "/a", "1.1.1.1"),
			want: true,
		},

		{
			name: "not match",
			match: Match{
				Not: &Match{Method: "POST"},
			},
			req:  req("GET", "example.com", "/a", "1.1.1.1"),
			want: true,
		},

		{
			name:  "empty match",
			match: Match{},
			req:   req("GET", "example.com", "/a", "1.1.1.1"),
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := ActionMatch(tt.match, tt.req)

			if got != tt.want {
				t.Fatalf("expected %v got %v", tt.want, got)
			}
		})
	}
}
