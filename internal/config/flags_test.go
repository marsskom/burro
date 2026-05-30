package config

import "testing"

func TestMergeProxy(t *testing.T) {
	tests := []struct {
		name  string
		cfg   ProxyConfig
		flags ProxyFlags
		want  ProxyConfig
	}{
		{
			name:  "override port",
			cfg:   ProxyConfig{Port: 8080, Host: "localhost"},
			flags: ProxyFlags{Port: 9090},
			want:  ProxyConfig{Port: 9090, Host: "localhost"},
		},
		{
			name:  "no override when zero",
			cfg:   ProxyConfig{Port: 8080, Host: "localhost"},
			flags: ProxyFlags{Port: 0},
			want:  ProxyConfig{Port: 8080, Host: "localhost"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeProxy(tt.cfg, tt.flags)
			if got != tt.want {
				t.Fatalf("expected %+v, got %+v", tt.want, got)
			}
		})
	}
}
