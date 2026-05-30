package cli

import (
	"io"
	"testing"
)

func TestConfirm_DefaultYes_EmptyInput(t *testing.T) {
	cliIO := testIO("\n")

	result, err := Confirm(cliIO, "Proceed?", ChoiceYes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result {
		t.Fatal("expected true for default YES")
	}
}

func TestConfirm_DefaultNo_EmptyInput(t *testing.T) {
	cliIO := testIO("\n")

	result, err := Confirm(cliIO, "Proceed?", ChoiceNo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result {
		t.Fatal("expected false for default NO")
	}
}

func TestConfirm_ExplicitYes(t *testing.T) {
	cliIO := testIO("y\n")

	result, err := Confirm(cliIO, "Proceed?", ChoiceNo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result {
		t.Fatal("expected true for 'y'")
	}
}

func TestConfirm_YesWord(t *testing.T) {
	cliIO := testIO("yes\n")

	result, err := Confirm(cliIO, "Proceed?", ChoiceNo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result {
		t.Fatal("expected true for 'yes'")
	}
}

func TestConfirm_NoInput(t *testing.T) {
	cliIO := testIO("n\n")

	result, err := Confirm(cliIO, "Proceed?", ChoiceYes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result {
		t.Fatal("expected false for 'n'")
	}
}

func TestConfirm_EOF_Empty(t *testing.T) {
	cliIO := testIO("")

	result, err := Confirm(cliIO, "Proceed?", ChoiceYes)

	if err == nil {
		t.Fatal("expected io.EOF error")
	}

	if err != io.EOF {
		t.Fatalf("expected io.EOF, got %v", err)
	}

	if result {
		t.Fatal("expected false result on EOF")
	}
}
