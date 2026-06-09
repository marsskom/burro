package domain

import (
	"os"
	"testing"
)

func TestLoadDomains(t *testing.T) {
	tmp, err := os.CreateTemp("", "domains")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	content := "example.com\ngoogle.com\n"
	if _, err := tmp.WriteString(content); err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()

	tmp.Seek(0, 0)

	domains, err := LoadDomains(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(domains) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(domains))
	}

	if domains[0] != "example.com" {
		t.Fatalf("unexpected first domain: %s", domains[0])
	}
}

func TestLoadDomains_NoEmptyLines(t *testing.T) {
	tmp, err := os.CreateTemp("", "domains")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	_ = os.WriteFile(tmp.Name(), []byte("example.com\n"), 0644)

	domains, err := LoadDomains(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if len(domains) != 1 {
		t.Fatalf("expected 1 domain, got %d", len(domains))
	}

	if domains[0] != "example.com" {
		t.Fatalf("unexpected domain: %s", domains[0])
	}
}
