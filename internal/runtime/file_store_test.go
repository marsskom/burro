package runtime

import (
	"io"
	"strings"
	"testing"
)

func TestFileArtifactStore_CreateAndExists(t *testing.T) {
	dir := t.TempDir()
	s := NewFileArtifactStore(dir)

	w, err := s.Create("a/b.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer w.Close()

	if !s.Exists("a/b.txt") {
		t.Fatal("expected file to exist")
	}
}

func TestFileArtifactStore_WriteRead(t *testing.T) {
	dir := t.TempDir()
	s := NewFileArtifactStore(dir)

	_, err := s.Write("file.txt", strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r, err := s.Read("file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer r.Close()

	b, _ := io.ReadAll(r)

	if string(b) != "hello" {
		t.Fatalf("expected hello, got %s", b)
	}
}

func TestFileArtifactStore_Rename(t *testing.T) {
	dir := t.TempDir()
	s := NewFileArtifactStore(dir)

	_, _ = s.Write("a.txt", strings.NewReader("x"))

	err := s.Rename("a.txt", "b.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Exists("a.txt") {
		t.Fatal("old file should not exist")
	}

	if !s.Exists("b.txt") {
		t.Fatal("new file should exist")
	}
}

func TestFileArtifactStore_Rename_Errors(t *testing.T) {
	dir := t.TempDir()
	s := NewFileArtifactStore(dir)

	// old does not exist
	err := s.Rename("missing", "new")
	if err == nil {
		t.Fatal("expected error for missing old path")
	}

	// create file
	_, _ = s.Write("a.txt", strings.NewReader("x"))
	_, _ = s.Write("b.txt", strings.NewReader("y"))

	// new already exists
	err = s.Rename("a.txt", "b.txt")
	if err == nil {
		t.Fatal("expected error for existing new path")
	}
}

func TestFileArtifactStore_Delete(t *testing.T) {
	dir := t.TempDir()
	s := NewFileArtifactStore(dir)

	_, _ = s.Write("a.txt", strings.NewReader("x"))

	err := s.Delete("a.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Exists("a.txt") {
		t.Fatal("file should be deleted")
	}
}

func TestFileArtifactStore_List(t *testing.T) {
	dir := t.TempDir()
	s := NewFileArtifactStore(dir)

	_, _ = s.Write("a.txt", strings.NewReader("1"))
	_, _ = s.Write("dir/b.txt", strings.NewReader("2"))

	files, err := s.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) < 2 {
		t.Fatalf("expected at least 2 files, got %d", len(files))
	}
}

func TestFileArtifactStore_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	s := NewFileArtifactStore(dir)

	_, err := s.Write("../../evil.txt", strings.NewReader("x"))
	if err != nil {
		t.Fatal("expected write to be safe, not error")
	}

	if !s.Exists("evil.txt") {
		t.Fatal("expected cleaned path to be stored safely")
	}
}
