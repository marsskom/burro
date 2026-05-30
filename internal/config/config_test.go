package config

import (
	"errors"
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

func TestResolvePath(t *testing.T) {
	t.Run("explicit wins", func(t *testing.T) {
		got, err := ResolvePath("/tmp/config.yml")
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

		got, err := ResolvePath("")
		if err != nil {
			t.Fatal(err)
		}

		if got != "/env/config.yml" {
			t.Fatalf("expected env path, got %s", got)
		}
	})

	t.Run("fallback to ~/.burro/config.yml", func(t *testing.T) {
		dir := t.TempDir()

		// Fakes home directory func.
		old := userHomeDir
		defer func() { userHomeDir = old }()

		userHomeDir = func() (string, error) {
			return dir, nil
		}

		expected := filepath.Join(dir, ".burro", "config.yml")
		if err := os.MkdirAll(filepath.Dir(expected), 0755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(expected, []byte("core:\n  log_level: debug"), 0644); err != nil {
			t.Fatal(err)
		}

		got, err := ResolvePath("")
		if err != nil {
			t.Fatal(err)
		}

		if got != expected {
			t.Fatalf("expected %s, got %s", expected, got)
		}
	})

	t.Run("fallback to ./config.yml", func(t *testing.T) {
		dir := t.TempDir()

		old := userHomeDir
		defer func() { userHomeDir = old }()

		userHomeDir = func() (string, error) {
			return dir, nil
		}

		// Ensures ~/.burro does NOT exist.
		homeCfg := filepath.Join(dir, ".burro", "config.yml")
		_ = os.RemoveAll(filepath.Dir(homeCfg))

		// Creates local config.
		local := filepath.Join(dir, "config.yml")
		if err := os.WriteFile(local, []byte("core:\n  log_level: debug"), 0644); err != nil {
			t.Fatal(err)
		}

		// Changes working dir.
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)

		if err := os.Chdir(dir); err != nil {
			t.Fatal(err)
		}

		got, err := ResolvePath("")
		if err != nil {
			t.Fatal(err)
		}

		if got != "./config.yml" {
			t.Fatalf("expected ./config.yml, got %s", got)
		}
	})

	t.Run("home directory doesn't exist", func(t *testing.T) {
		dir := t.TempDir()

		old := userHomeDir
		defer func() { userHomeDir = old }()

		userHomeDir = func() (string, error) {
			return "", errors.New("doesn't exist")
		}

		expected := filepath.Join(dir, ".burro", "config.yml")
		if err := os.MkdirAll(filepath.Dir(expected), 0755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(expected, []byte("core:\n  log_level: debug"), 0644); err != nil {
			t.Fatal(err)
		}

		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)

		_ = os.Chdir(dir)

		_, err := ResolvePath("")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("not found", func(t *testing.T) {
		dir := t.TempDir()

		old := userHomeDir
		defer func() { userHomeDir = old }()

		userHomeDir = func() (string, error) {
			return dir, nil
		}

		// Ensures nothing exists.
		_ = os.RemoveAll(filepath.Join(dir, ".burro"))

		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)

		_ = os.Chdir(dir)

		_, err := ResolvePath("")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
