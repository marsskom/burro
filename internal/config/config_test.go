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
  port: 8080
  host: localhost
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

	if cfg.Proxy.Port != 8080 {
		t.Fatalf("expected 8080, got %d", cfg.Proxy.Port)
	}

	if cfg.Proxy.Host != "localhost" {
		t.Fatalf("expected localhost, got %s", cfg.Proxy.Host)
	}
}

func TestLoadWithFlags(t *testing.T) {
	content := `
core:
  log_level: info
proxy:
  port: 8080
  host: 127.0.0.1
`

	tmp := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadWithFlags(tmp, ProxyFlags{
		Port: 9999,
	})
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Proxy.Port != 9999 {
		t.Fatalf("expected overridden port 9999, got %d", cfg.Proxy.Port)
	}

	if cfg.Proxy.Host != "127.0.0.1" {
		t.Fatalf("host should stay unchanged, got %s", cfg.Proxy.Host)
	}
}

func TestResolveHome(t *testing.T) {
	t.Run("explicit wins", func(t *testing.T) {
		got := ResolveHome("/tmp")

		if got != "/tmp" {
			t.Fatalf("expected explicit path, got %s", got)
		}
	})

	t.Run("env wins over default", func(t *testing.T) {
		os.Setenv("BURRO_HOME", "/env")
		defer os.Unsetenv("BURRO_HOME")

		got := ResolveHome("")

		if got != "/env" {
			t.Fatalf("expected env path, got %s", got)
		}
	})

	t.Run("fallback to ./runtime", func(t *testing.T) {
		got := ResolveHome("")

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
