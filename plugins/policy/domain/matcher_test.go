package domain

import "testing"

func TestMatch(t *testing.T) {
	tests := []struct {
		host     string
		domains  []string
		expected bool
	}{
		{"example.com", []string{"example.com"}, true},
		{"api.example.com", []string{"example.com"}, true},
		{"example.net", []string{"example.com"}, false},
		{"", []string{"example.com"}, false},
		{"example.com", []string{}, false},
	}

	for _, tt := range tests {
		got := Match(tt.host, tt.domains)
		if got != tt.expected {
			t.Fatalf("host=%s expected=%v got=%v", tt.host, tt.expected, got)
		}
	}
}
