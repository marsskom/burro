package policy

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/testutils"
)

func TestLoadDomains(t *testing.T) {
	tmp, err := os.CreateTemp("", "domains")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	content := "example.com\ngoogle.com\n"
	if _, err := tmp.WriteString(content); err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()

	tmp.Seek(0, 0)

	domains, err := LoadDomains(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(domains) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(domains))
	}

	if domains[0] != "example.com" {
		t.Fatalf("unexpected first domain: %s", domains[0])
	}
}

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

func TestPolicy_WhitelistAllows(t *testing.T) {
	p := New()
	p.Init(testutils.NewForPlugin(""), map[string]any{})

	p.whitelist = []string{"example.com"}

	ctx := &model.RequestContext{
		Request: &http.Request{
			Host: "api.example.com",
		},
		Timings: &model.Timings{},
	}

	err := p.OnRequest(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if ctx.IsFinished {
		t.Fatal("request should NOT be finished for whitelist match")
	}
}

func TestPolicy_BlacklistBlocks(t *testing.T) {
	p := New()
	p.Init(testutils.NewForPlugin(""), map[string]any{})

	p.blacklist = []string{"example.com"}

	ctx := &model.RequestContext{
		Request: &http.Request{
			Host: "api.example.com",
		},
		Timings: &model.Timings{},
	}

	err := p.OnRequest(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !ctx.IsFinished {
		t.Fatal("request should be finished when blocked")
	}
}

func TestPolicy_WhitelistPriority(t *testing.T) {
	p := New()
	p.Init(testutils.NewForPlugin(""), map[string]any{})

	p.whitelist = []string{"example.com"}
	p.blacklist = []string{"example.com"}

	ctx := &model.RequestContext{
		Request: &http.Request{
			Host: "api.example.com",
		},
		Timings: &model.Timings{},
	}

	_ = p.OnRequest(ctx)

	if ctx.IsFinished {
		t.Fatal("whitelist should bypass blacklist")
	}
}

func TestLoadDomains_NoEmptyLines(t *testing.T) {
	tmp, err := os.CreateTemp("", "domains")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	_ = os.WriteFile(tmp.Name(), []byte("example.com\n"), 0644)

	domains, err := LoadDomains(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(domains) != 1 {
		t.Fatalf("expected 1 domain, got %d", len(domains))
	}

	if domains[0] != "example.com" {
		t.Fatalf("unexpected domain: %s", domains[0])
	}
}

func TestPolicy_Init(t *testing.T) {
	tmpW, _ := os.CreateTemp("", "whitelist")
	defer os.Remove(tmpW.Name())
	tmpW.WriteString("example.com\n")
	tmpW.Close()

	cfg := PolicyConfig{
		Priority:  10,
		Whitelist: filepath.Base(tmpW.Name()),
	}

	p := New()
	err := p.Init(testutils.NewForPlugin(filepath.Dir(tmpW.Name())), cfg)
	if err != nil {
		t.Fatal(err)
	}

	if p.priority != 10 {
		t.Fatalf("expected priority 10, got %d", p.priority)
	}

	if len(p.whitelist) != 1 {
		t.Fatalf("expected whitelist loaded")
	}
}
