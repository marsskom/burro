package actions

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
	"testing"
)

type fakeFile struct {
	data []byte
	err  error
}

type fakeDS struct {
	files map[string]fakeFile
}

func (f *fakeDS) Exists(name string) bool {
	_, ok := f.files[name]

	return ok
}

func (f *fakeDS) Read(name string) (io.ReadCloser, error) {
	file, ok := f.files[name]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	if file.err != nil {
		return nil, file.err
	}
	return io.NopCloser(bytes.NewReader(file.data)), nil
}

func (f *fakeDS) List(path string, ext []string) ([]string, error) {
	return slices.Collect(maps.Keys(f.files)), nil
}

func yamlFile(actions string) []byte {
	return []byte("actions:\n" + actions)
}

func TestLoadActionRules_SuccessAndSort(t *testing.T) {
	ds := &fakeDS{
		files: map[string]fakeFile{
			"a.yml": {
				data: yamlFile(`
  - id: low
    priority: 10
    match:
      method: GET
    action:
      - op: allow
`),
			},
			"b.yml": {
				data: yamlFile(`
  - id: high
    priority: 100
    match:
      method: POST
    action:
      - op: deny
`),
			},
		},
	}

	rules, err := LoadActionRules(ds, []string{"a.yml", "b.yml"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rules) != 2 {
		t.Fatalf("expected 2 rules got %d", len(rules))
	}

	// MUST be sorted DESC priority
	if rules[0].ID != "high" {
		t.Fatalf("expected first rule to be 'high', got %s", rules[0].ID)
	}

	if rules[1].ID != "low" {
		t.Fatalf("expected second rule to be 'low', got %s", rules[1].ID)
	}
}

func TestLoadActionRules_ReadError(t *testing.T) {
	ds := &fakeDS{
		files: map[string]fakeFile{
			"a.yml": {err: fmt.Errorf("disk error")},
		},
	}

	_, err := LoadActionRules(ds, []string{"a.yml"})
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "cannot open file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadActionRules_InvalidYaml(t *testing.T) {
	ds := &fakeDS{
		files: map[string]fakeFile{
			"a.yml": {
				data: []byte("actions:\n  - id: bad:\n"), // broken yaml
			},
		},
	}

	_, err := LoadActionRules(ds, []string{"a.yml"})
	if err == nil {
		t.Fatal("expected yaml error")
	}

	if !strings.Contains(err.Error(), "unmarshall") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadActionRules_EmptyInput(t *testing.T) {
	ds := &fakeDS{
		files: map[string]fakeFile{},
	}

	rules, err := LoadActionRules(ds, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rules) != 0 {
		t.Fatalf("expected empty rules")
	}
}
