package actions

import "testing"

func TestDecode_SetHeaderArgs(t *testing.T) {
	in := map[string]any{
		"name":  "X-Test",
		"value": "123",
	}

	out, err := decode[SetHeaderArgs](in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Name != "X-Test" {
		t.Fatalf("expected X-Test got %s", out.Name)
	}

	if out.Value != "123" {
		t.Fatalf("expected 123 got %s", out.Value)
	}
}

func TestDecode_RemoveHeaderArgs(t *testing.T) {
	in := map[string]any{
		"names": []any{"A", "B"},
	}

	out, err := decode[RemoveHeaderArgs](in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out.Names) != 2 {
		t.Fatalf("expected 2 names got %d", len(out.Names))
	}
}

func TestDecode_RedactBodyArgs(t *testing.T) {
	in := map[string]any{
		"fields": []any{"token", "password"},
	}

	out, err := decode[RedactBodyArgs](in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out.Fields) != 2 {
		t.Fatalf("expected 2 fields got %d", len(out.Fields))
	}
}

func TestDecode_LogArgs(t *testing.T) {
	in := map[string]any{
		"level":   "info",
		"message": "hello",
	}

	out, err := decode[LogArgs](in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Level != "info" {
		t.Fatalf("expected info got %s", out.Level)
	}

	if out.Message != "hello" {
		t.Fatalf("expected hello got %s", out.Message)
	}
}

func TestDecode_StructInput(t *testing.T) {
	in := SetHeaderArgs{
		Name:  "X",
		Value: "Y",
	}

	out, err := decode[SetHeaderArgs](in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Name != "X" || out.Value != "Y" {
		t.Fatalf("unexpected decode result: %+v", out)
	}
}

func TestDecode_InvalidType(t *testing.T) {
	// Functions cannot be JSON marshaled.
	in := func() {}

	_, err := decode[SetHeaderArgs](in)
	if err == nil {
		t.Fatal("expected error")
	}
}
