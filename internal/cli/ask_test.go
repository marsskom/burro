package cli

import (
	"errors"
	"io"
	"testing"
)

func TestAsk_OK(t *testing.T) {
	cliIO := testIO("hello world\n")

	got, err := Ask(cliIO, "Enter:")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "hello world" {
		t.Fatalf("expected 'hello world', got %q", got)
	}
}

func TestAsk_EOF(t *testing.T) {
	cliIO := testIO("")

	_, err := Ask(cliIO, "Enter:")
	if err == nil {
		t.Fatal("expected EOF error")
	}

	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF, got %v", err)
	}
}

func TestAskWithValidator_Valid(t *testing.T) {
	cliIO := testIO("ok\n")

	validator := func(s string) error {
		if s != "ok" {
			t.Fatal("validator called with wrong input")
		}
		return nil
	}

	got, err := AskWithValidator(cliIO, "Enter:", validator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "ok" {
		t.Fatalf("expected ok, got %s", got)
	}
}

func TestAskWithValidator_Retry(t *testing.T) {
	cliIO := testIO("bad\nok\n")

	calls := 0

	validator := func(s string) error {
		calls++
		if s == "ok" {
			return nil
		}
		return errors.New("invalid")
	}

	got, err := AskWithValidator(cliIO, "Enter:", validator)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "ok" {
		t.Fatalf("expected ok, got %s", got)
	}

	if calls != 2 {
		t.Fatalf("expected 2 validator calls, got %d", calls)
	}
}

func TestAskWithValidator_EOF(t *testing.T) {
	cliIO := testIO("bad\n")

	validator := func(s string) error {
		return errors.New("always invalid")
	}

	_, err := AskWithValidator(cliIO, "Enter:", validator)
	if err == nil {
		t.Fatal("expected EOF error")
	}

	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected wrapped io.EOF, got %v", err)
	}
}
