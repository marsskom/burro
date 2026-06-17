package main

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// resetInitFlags ensures initFlags package-level state doesn't leak between tests.
func resetInitFlags(t *testing.T) {
	t.Cleanup(func() {
		initFlags.DryRun = false
		initFlags.Force = false
	})
	initFlags.DryRun = false
	initFlags.Force = false
}

func TestCmdInitRejectsEmptyPath(t *testing.T) {
	resetInitFlags(t)

	err := cmdInit("", "runtime")
	if err == nil {
		t.Fatal("expected error for empty path, got nil")
	}
}

func TestCmdInitRejectsEmptyName(t *testing.T) {
	resetInitFlags(t)

	tmp := t.TempDir()

	err := cmdInit(tmp, "")
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestCmdInitRejectsNonExistentPath(t *testing.T) {
	resetInitFlags(t)

	missing := filepath.Join(t.TempDir(), "does-not-exist")

	err := cmdInit(missing, "runtime")
	if err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}
}

func TestCmdInitRejectsFileAsPath(t *testing.T) {
	resetInitFlags(t)

	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "not-a-dir")

	if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
		t.Fatalf("failed to set up test file: %v", err)
	}

	err := cmdInit(filePath, "runtime")
	if err == nil {
		t.Fatal("expected error when path is a file, got nil")
	}
}

func TestCmdInitCreatesExpectedDirectories(t *testing.T) {
	resetInitFlags(t)

	tmp := t.TempDir()
	name := "runtime"

	if err := cmdInit(tmp, name); err != nil {
		t.Fatalf("cmdInit returned error: %v", err)
	}

	appDir := filepath.Join(tmp, name)

	expectedDirs := []string{
		appDir,
		filepath.Join(appDir, "artifacts"),
		filepath.Join(appDir, "certs"),
		filepath.Join(appDir, "db"),
		filepath.Join(appDir, "plugins"),
		filepath.Join(appDir, "plugins", "policy"),
		filepath.Join(appDir, "plugins", "policy", "data"),
		filepath.Join(appDir, "plugins", "policy", "actions"),
		filepath.Join(appDir, "plugins", "luaplugin"),
		filepath.Join(appDir, "plugins", "luaplugin", "scripts"),
		filepath.Join(appDir, "plugins", "luaplugin", "scripts", "metric"),
	}

	for _, dir := range expectedDirs {
		fi, err := os.Stat(dir)
		if err != nil {
			t.Errorf("expected directory '%s' to exist: %v", dir, err)
			continue
		}
		if !fi.IsDir() {
			t.Errorf("expected '%s' to be a directory", dir)
		}
	}
}

func TestCmdInitCreatesExpectedFiles(t *testing.T) {
	resetInitFlags(t)

	tmp := t.TempDir()
	name := "runtime"

	if err := cmdInit(tmp, name); err != nil {
		t.Fatalf("cmdInit returned error: %v", err)
	}

	appDir := filepath.Join(tmp, name)

	expectedFiles := []string{
		filepath.Join(appDir, "config.yml"),
		filepath.Join(appDir, "plugins", "policy", "config.yml"),
		filepath.Join(appDir, "plugins", "policy", "data", "whitelist.txt"),
		filepath.Join(appDir, "plugins", "policy", "data", "blacklist.txt"),
	}

	for _, file := range expectedFiles {
		fi, err := os.Stat(file)
		if err != nil {
			t.Errorf("expected file '%s' to exist: %v", file, err)
			continue
		}
		if fi.IsDir() {
			t.Errorf("expected '%s' to be a file, not a directory", file)
		}
	}
}

func TestCmdInitDryRunCreatesNothing(t *testing.T) {
	resetInitFlags(t)
	initFlags.DryRun = true

	tmp := t.TempDir()
	name := "runtime"

	if err := cmdInit(tmp, name); err != nil {
		t.Fatalf("cmdInit returned error: %v", err)
	}

	appDir := filepath.Join(tmp, name)

	if _, err := os.Stat(appDir); !os.IsNotExist(err) {
		t.Errorf("expected '%s' not to exist in dry-run mode, stat err: %v", appDir, err)
	}
}

func TestCmdInitWhitelistAndBlacklistAreEmpty(t *testing.T) {
	resetInitFlags(t)

	tmp := t.TempDir()
	name := "runtime"

	if err := cmdInit(tmp, name); err != nil {
		t.Fatalf("cmdInit returned error: %v", err)
	}

	appDir := filepath.Join(tmp, name)

	for _, file := range []string{
		filepath.Join(appDir, "plugins", "policy", "data", "whitelist.txt"),
		filepath.Join(appDir, "plugins", "policy", "data", "blacklist.txt"),
	} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read '%s': %v", file, err)
		}
		if len(data) != 0 {
			t.Errorf("expected '%s' to be empty, got %d bytes", file, len(data))
		}
	}
}

func TestCmdInitDoesNotOverwriteExistingFilesByDefault(t *testing.T) {
	resetInitFlags(t)

	tmp := t.TempDir()
	name := "runtime"
	appDir := filepath.Join(tmp, name)
	configPath := filepath.Join(appDir, "config.yml")

	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("failed to set up app dir: %v", err)
	}

	sentinel := []byte("sentinel: true\n")
	if err := os.WriteFile(configPath, sentinel, 0644); err != nil {
		t.Fatalf("failed to seed existing config: %v", err)
	}

	if err := cmdInit(tmp, name); err != nil {
		t.Fatalf("cmdInit returned error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config after cmdInit: %v", err)
	}

	if string(data) != string(sentinel) {
		t.Errorf("expected existing config to be left untouched, got: %s", string(data))
	}
}

func TestCmdInitForceOverwritesExistingFiles(t *testing.T) {
	resetInitFlags(t)
	initFlags.Force = true

	tmp := t.TempDir()
	name := "runtime"
	appDir := filepath.Join(tmp, name)
	configPath := filepath.Join(appDir, "config.yml")

	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("failed to set up app dir: %v", err)
	}

	sentinel := []byte("sentinel: true\n")
	if err := os.WriteFile(configPath, sentinel, 0644); err != nil {
		t.Fatalf("failed to seed existing config: %v", err)
	}

	if err := cmdInit(tmp, name); err != nil {
		t.Fatalf("cmdInit returned error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config after cmdInit: %v", err)
	}

	if string(data) == string(sentinel) {
		t.Error("expected existing config to be overwritten when --force is set")
	}
}

func TestCmdInitAppConfigIsValidYAML(t *testing.T) {
	resetInitFlags(t)

	tmp := t.TempDir()
	name := "runtime"

	if err := cmdInit(tmp, name); err != nil {
		t.Fatalf("cmdInit returned error: %v", err)
	}

	configPath := filepath.Join(tmp, name, "config.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read app config: %v", err)
	}

	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("app config is not valid YAML: %v", err)
	}

	if _, ok := parsed["plugins"]; !ok {
		t.Error("expected 'plugins' key in app config")
	}
	if _, ok := parsed["proxy"]; !ok {
		t.Error("expected 'proxy' key in app config")
	}
}

func TestCmdInitPolicyConfigIsValidYAML(t *testing.T) {
	resetInitFlags(t)

	tmp := t.TempDir()
	name := "runtime"

	if err := cmdInit(tmp, name); err != nil {
		t.Fatalf("cmdInit returned error: %v", err)
	}

	configPath := filepath.Join(tmp, name, "plugins", "policy", "config.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read policy config: %v", err)
	}

	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("policy config is not valid YAML: %v", err)
	}

	expectedKeys := []string{"enabled", "priority", "whitelist", "blacklist", "action_dir"}
	for _, key := range expectedKeys {
		if _, ok := parsed[key]; !ok {
			t.Errorf("expected '%s' key in policy config", key)
		}
	}

	if whitelist, ok := parsed["whitelist"].(string); !ok || whitelist != "./data/whitelist.txt" {
		t.Errorf("unexpected whitelist value: %v", parsed["whitelist"])
	}
	if blacklist, ok := parsed["blacklist"].(string); !ok || blacklist != "./data/blacklist.txt" {
		t.Errorf("unexpected blacklist value: %v", parsed["blacklist"])
	}
}

func TestCmdInitDefaultNameIsRuntime(t *testing.T) {
	resetInitFlags(t)

	tmp := t.TempDir()

	if err := cmdInit(tmp, "runtime"); err != nil {
		t.Fatalf("cmdInit returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "runtime")); err != nil {
		t.Errorf("expected default 'runtime' directory to exist: %v", err)
	}
}

func TestCmdInitAcceptsCustomName(t *testing.T) {
	resetInitFlags(t)

	tmp := t.TempDir()
	name := "myapp"

	if err := cmdInit(tmp, name); err != nil {
		t.Fatalf("cmdInit returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, name)); err != nil {
		t.Errorf("expected custom '%s' directory to exist: %v", name, err)
	}
}

func TestCmdInitAcceptsRelativePath(t *testing.T) {
	resetInitFlags(t)

	tmp := t.TempDir()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir into temp dir: %v", err)
	}

	if err := cmdInit(".", "runtime"); err != nil {
		t.Fatalf("cmdInit returned error for relative path: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "runtime")); err != nil {
		t.Errorf("expected 'runtime' directory to exist under resolved absolute path: %v", err)
	}
}

func TestWriteSkipsExistingFileWithoutForce(t *testing.T) {
	resetInitFlags(t)

	tmp := t.TempDir()
	target := filepath.Join(tmp, "existing.txt")

	if err := os.WriteFile(target, []byte("original"), 0644); err != nil {
		t.Fatalf("failed to seed file: %v", err)
	}

	called := false
	err := write(target, func() error {
		called = true
		return os.WriteFile(target, []byte("overwritten"), 0644)
	})
	if err != nil {
		t.Fatalf("write returned error: %v", err)
	}

	if called {
		t.Error("expected fn not to be called when file exists and Force is false")
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read target file: %v", err)
	}
	if string(data) != "original" {
		t.Errorf("expected file contents to remain 'original', got '%s'", string(data))
	}
}

func TestWriteCallsFnWhenFileDoesNotExist(t *testing.T) {
	resetInitFlags(t)

	tmp := t.TempDir()
	target := filepath.Join(tmp, "new.txt")

	called := false
	err := write(target, func() error {
		called = true
		return os.WriteFile(target, []byte("created"), 0644)
	})
	if err != nil {
		t.Fatalf("write returned error: %v", err)
	}

	if !called {
		t.Error("expected fn to be called when file does not exist")
	}
}

func TestWriteCallsFnWhenForceIsSetEvenIfFileExists(t *testing.T) {
	resetInitFlags(t)
	initFlags.Force = true

	tmp := t.TempDir()
	target := filepath.Join(tmp, "existing.txt")

	if err := os.WriteFile(target, []byte("original"), 0644); err != nil {
		t.Fatalf("failed to seed file: %v", err)
	}

	called := false
	err := write(target, func() error {
		called = true
		return os.WriteFile(target, []byte("overwritten"), 0644)
	})
	if err != nil {
		t.Fatalf("write returned error: %v", err)
	}

	if !called {
		t.Error("expected fn to be called when Force is true, even if file exists")
	}
}

func TestWriteDryRunNeverCallsFn(t *testing.T) {
	resetInitFlags(t)
	initFlags.DryRun = true

	tmp := t.TempDir()
	target := filepath.Join(tmp, "new.txt")

	called := false
	err := write(target, func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("write returned error: %v", err)
	}

	if called {
		t.Error("expected fn not to be called in dry-run mode")
	}

	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("expected target file not to exist in dry-run mode, stat err: %v", err)
	}
}

func TestWritePropagatesFnError(t *testing.T) {
	resetInitFlags(t)

	tmp := t.TempDir()
	target := filepath.Join(tmp, "new.txt")

	wantErr := os.ErrPermission
	err := write(target, func() error {
		return wantErr
	})

	if err == nil {
		t.Fatal("expected error to be propagated, got nil")
	}
}

func TestWriteAppConfigProducesParsableYAML(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "config.yml")

	if err := writeAppConfig(target); err != nil {
		t.Fatalf("writeAppConfig returned error: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read generated app config: %v", err)
	}

	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("generated app config is not valid YAML: %v", err)
	}
}

func TestWritePolicyConfigProducesParsableYAML(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "config.yml")

	if err := writePolicyConfig(target); err != nil {
		t.Fatalf("writePolicyConfig returned error: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read generated policy config: %v", err)
	}

	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("generated policy config is not valid YAML: %v", err)
	}
}
