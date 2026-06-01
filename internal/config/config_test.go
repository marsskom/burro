package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	content := `
core:
  log_level: debug
proxy:
  listen: localhost:8080
plugins:
  test: true
`

	tmp := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Core.LogLevel != "debug" {
		t.Fatalf("expected debug, got %s", cfg.Core.LogLevel)
	}

	if cfg.Proxy.Listen != "localhost:8080" {
		t.Fatalf("expected localhost:8080, got %s", cfg.Proxy.Listen)
	}
}

func TestLoadWithFlags(t *testing.T) {
	content := `
core:
  log_level: info
proxy:
  listen: 127.0.0.1:8080
tls:
  enabled: false
  cert:
  key:
`

	t.Run("load valid flags", func(t *testing.T) {
		tmp := filepath.Join(t.TempDir(), "config.yml")
		if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadWithFlags(tmp, ProxyFlags{
			Listen:  "localhost:7676",
			TLSCert: "cert.pem",
			TLSKey:  "key.key",
		})
		if err != nil {
			t.Fatal(err)
		}

		if cfg.Proxy.Listen != "localhost:7676" {
			t.Fatalf("expected localhost:7676, got %s", cfg.Proxy.Listen)
		}

		if !cfg.TLS.Enabled {
			t.Fatalf("expected enabled TLS, got %v", cfg.TLS.Enabled)
		}

		if cfg.TLS.Cert != "cert.pem" {
			t.Fatalf("expected cert.pem, got %s", cfg.TLS.Cert)
		}

		if cfg.TLS.Key != "key.key" {
			t.Fatalf("expected key.key, got %s", cfg.TLS.Key)
		}
	})

	t.Run("load invalid listen string", func(t *testing.T) {
		tmp := filepath.Join(t.TempDir(), "config.yml")
		if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadWithFlags(tmp, ProxyFlags{
			Listen: "localhost-7676",
		})
		if err != nil {
			t.Fatal(err)
		}

		if cfg.Proxy.Listen != "127.0.0.1:8080" {
			t.Fatalf("expected 127.0.0.1:8080, got %s", cfg.Proxy.Listen)
		}
	})
}

func TestResolveWorkdir(t *testing.T) {
	t.Run("explicit wins", func(t *testing.T) {
		got := ResolveWorkdir("/tmp")

		if got != "/tmp" {
			t.Fatalf("expected explicit path, got %s", got)
		}
	})

	t.Run("env wins over default", func(t *testing.T) {
		os.Setenv("BURRO_WORKDIR", "/env")
		defer os.Unsetenv("BURRO_WORKDIR")

		got := ResolveWorkdir("")

		if got != "/env" {
			t.Fatalf("expected env path, got %s", got)
		}
	})

	t.Run("fallback to ./runtime", func(t *testing.T) {
		got := ResolveWorkdir("")

		if got != "./runtime" {
			t.Fatalf("expected ./runtime, got %s", got)
		}
	})
}

func TestGetConfigPath(t *testing.T) {
	t.Run("explicit wins", func(t *testing.T) {
		paths := &Paths{
			Home: "~",
		}

		got, err := paths.GetConfigPath("/tmp/config.yml")

		if err != nil {
			t.Fatal(err)
		}

		if got != "/tmp/config.yml" {
			t.Fatalf("expected explicit path, got %s", got)
		}
	})

	t.Run("env wins over default", func(t *testing.T) {
		os.Setenv("BURRO_CONFIG", "/env/config.yml")
		defer os.Unsetenv("BURRO_CONFIG")

		paths := &Paths{
			Home: "~",
		}

		got, err := paths.GetConfigPath("")

		if err != nil {
			t.Fatal(err)
		}

		if got != "/env/config.yml" {
			t.Fatalf("expected env path, got %s", got)
		}
	})

	t.Run("fallback to home directory / config.yml", func(t *testing.T) {
		dir := t.TempDir()

		paths := &Paths{
			Home: dir,
		}

		expected := filepath.Join(paths.Home, "config.yml")

		if err := os.WriteFile(expected, []byte("core:\n  log_level: debug"), 0644); err != nil {
			t.Fatal(err)
		}

		got, err := paths.GetConfigPath("")
		if err != nil {
			t.Fatal(err)
		}

		if got != expected {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	})

	t.Run("not found", func(t *testing.T) {
		dir := t.TempDir()

		paths := &Paths{
			Home: dir,
		}

		_, err := paths.GetConfigPath("")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
