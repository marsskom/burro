package runtime

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPluginDataStore_Read(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(
		filepath.Join(dir, "test.txt"),
		[]byte("hello"),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	ds := NewPluginDataStore(dir)

	r, err := ds.Read("test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer r.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(b) != "hello" {
		t.Fatalf("expected hello, got %s", string(b))
	}
}

func TestPluginDataStore_Exists(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(
		filepath.Join(dir, "test.txt"),
		[]byte("hello"),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	ds := NewPluginDataStore(dir)

	if !ds.Exists("test.txt") {
		t.Fatal("expected file to exist")
	}

	if ds.Exists("missing.txt") {
		t.Fatal("expected file not to exist")
	}
}

func TestPluginDataStore_List(t *testing.T) {
	dir := t.TempDir()

	_ = os.MkdirAll(filepath.Join(dir, "nested"), 0755)

	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "nested", "b.txt"), []byte("b"), 0644)

	ds := NewPluginDataStore(dir)

	files, err := ds.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
}

func TestPluginDataStore_UnsupportedOperations(t *testing.T) {
	ds := NewPluginDataStore(t.TempDir())

	tests := []struct {
		name string
		fn   func() error
	}{
		{
			name: "create",
			fn: func() error {
				_, err := ds.Create("a.txt")
				return err
			},
		},
		{
			name: "rename",
			fn: func() error {
				return ds.Rename("a", "b")
			},
		},
		{
			name: "write",
			fn: func() error {
				_, err := ds.Write("a.txt", strings.NewReader("test"))
				return err
			},
		},
		{
			name: "delete",
			fn: func() error {
				return ds.Delete("a.txt")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fn(); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestPluginDataStore_PathTraversal(t *testing.T) {
	ds := NewPluginDataStore(t.TempDir())

	if ds.Exists("../../etc/passwd") {
		t.Fatal("path traversal should not be allowed")
	}

	_, err := ds.Read("../../etc/passwd")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPluginDataStore_ReadMissingFile(t *testing.T) {
	ds := NewPluginDataStore(t.TempDir())

	_, err := ds.Read("missing.txt")
	if err == nil {
		t.Fatal("expected error")
	}
}
