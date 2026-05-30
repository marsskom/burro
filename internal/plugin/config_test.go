package plugin

import (
	"testing"
)

type testCfg struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

func TestDecodeYAML_Success(t *testing.T) {
	input := map[string]any{
		"name": "burro",
		"port": 8080,
	}

	var out testCfg

	err := DecodeYAML(input, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Name != "burro" {
		t.Fatalf("expected name=burro, got %s", out.Name)
	}

	if out.Port != 8080 {
		t.Fatalf("expected port=8080, got %d", out.Port)
	}
}

func TestDecodeYAML_EmptyInput(t *testing.T) {
	input := map[string]any{}

	var out testCfg

	err := DecodeYAML(input, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Name != "" || out.Port != 0 {
		t.Fatalf("expected zero values, got %+v", out)
	}
}

func TestDecodeYAML_InvalidUnmarshalTarget(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("the TestDecodeYAML_InvalidUnmarshalTarget code did not panic")
		}
	}()

	input := map[string]any{
		"name": "burro",
	}

	var out testCfg

	// Passes not a pointer.
	err := DecodeYAML(input, out)
	if err == nil {
		t.Fatal("expected error for non-pointer output, got nil")
	}
}

func TestDecodeYAML_NonSerializableInput(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("the TestDecodeYAML_NonSerializableInput code did not panic")
		}
	}()
	// Channels cannot be marshaled by YAML.
	input := map[string]any{
		"bad": make(chan int),
	}

	var out map[string]any

	err := DecodeYAML(input, &out)
	if err == nil {
		t.Fatal("expected marshal error, got nil")
	}
}
