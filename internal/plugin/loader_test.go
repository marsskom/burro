package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/marsskom/burro/internal/config"
)

func TestResolvePluginConfig_NotFound(t *testing.T) {
	cfg := &config.CorePluginsConfig{
		Dir:    t.TempDir(),
		Config: "config.yml",
	}

	got, err := resolvePluginConfig(cfg, "policy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != nil {
		t.Fatalf("expected config nil, got %+v", got)
	}
}

func TestResolvePluginConfig_Loaded(t *testing.T) {
	dir := t.TempDir()

	expected := filepath.Join(dir, "policy", "config.yml")
	if err := os.MkdirAll(filepath.Dir(expected), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(expected, []byte("priority: 10\nenabled: true\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.CorePluginsConfig{
		Dir:    dir,
		Config: "config.yml",
	}

	got, err := resolvePluginConfig(cfg, "policy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("expect ok as true, got %v", ok)
	}

	if m["priority"] != 10 {
		t.Fatalf("expect priority equals to 10, got %v", m["priority"])
	}

	if m["enabled"] != true {
		t.Fatalf("expect enabled - true, got %v", m["enabled"])
	}
}

func TestResolvePluginConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()

	expected := filepath.Join(dir, "policy", "config.yml")
	if err := os.MkdirAll(filepath.Dir(expected), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(expected, []byte("{invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.CorePluginsConfig{
		Dir:    dir,
		Config: "config.yml",
	}

	_, err := resolvePluginConfig(cfg, "policy")
	if err == nil {
		t.Fatalf("expect error, got nil")
	}

	if !strings.Contains(err.Error(), "cannot unmarshall plugin config file") {
		t.Fatalf("expect unmarshall error, got %v", err)
	}
}
