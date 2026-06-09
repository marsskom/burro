package policy

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/testutils"
)

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
