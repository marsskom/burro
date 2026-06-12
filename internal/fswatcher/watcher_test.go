package fswatcher

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const testDelay = 1 * time.Second

func waitEvent(t *testing.T, w *FSWatcher, timeout time.Duration) (FSEvent, bool) {
	t.Helper()
	select {
	case e := <-w.Events:
		return e, true
	case <-time.After(timeout):
		return FSEvent{}, false
	}
}

func noEvent(t *testing.T, w *FSWatcher, wait time.Duration) {
	t.Helper()
	select {
	case e := <-w.Events:
		t.Errorf("unexpected event: %+v", e)
	case <-time.After(wait):
	}
}

func newTestWatcher(t *testing.T, root string, ext []string, recursive bool) *FSWatcher {
	t.Helper()
	w, err := NewFSWatcher(root, ext, recursive)
	if err != nil {
		t.Fatalf("NewFSWatcher: %v", err)
	}
	w.delay = 50 * time.Millisecond // speeds up debounce for tests
	t.Cleanup(func() { w.Stop() })
	if err := w.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return w
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}

// NewFSWatcher

func TestNewFSWatcher_InvalidRoot(t *testing.T) {
	_, err := NewFSWatcher("/nonexistent/path/that/does/not/exist", nil, false)
	if err == nil {
		t.Fatal("expected error for invalid root, got nil")
	}
}

func TestNewFSWatcher_ValidRoot(t *testing.T) {
	dir := t.TempDir()
	w, err := NewFSWatcher(dir, nil, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer w.Stop()
}

func TestNewFSWatcher_InvalidFileRoot(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "filename.lua"), "filename.lua")

	_, err := NewFSWatcher(filepath.Join(dir, "filename.lua"), nil, false)
	if err == nil {
		t.Fatal("expected error for invalid file root, got nil")
	}
}

// file create

func TestFSWatcher_FileCreate(t *testing.T) {
	dir := t.TempDir()
	w := newTestWatcher(t, dir, nil, false)

	writeFile(t, filepath.Join(dir, "hello.txt"), "hello")

	e, ok := waitEvent(t, w, testDelay)
	if !ok {
		t.Fatal("timeout: expected create event")
	}
	if e.Op != OpCreated && e.Op != OpUpdate {
		t.Errorf("op: want OpCreated or OpUpdate got %v", e.Op)
	}
	if filepath.Base(e.Path) != "hello.txt" {
		t.Errorf("path: want hello.txt got %s", filepath.Base(e.Path))
	}
}

// file write

func TestFSWatcher_FileWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	writeFile(t, path, "initial")

	w := newTestWatcher(t, dir, nil, false)

	writeFile(t, path, "updated")

	e, ok := waitEvent(t, w, testDelay)
	if !ok {
		t.Fatal("timeout: expected write event")
	}
	if e.Op != OpUpdate {
		t.Errorf("op: want OpUpdate got %v", e.Op)
	}
}

// file delete

func TestFSWatcher_FileDelete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "todel.txt")
	writeFile(t, path, "bye")

	w := newTestWatcher(t, dir, nil, false)

	if err := os.Remove(path); err != nil {
		t.Fatalf("remove: %v", err)
	}

	e, ok := waitEvent(t, w, testDelay)
	if !ok {
		t.Fatal("timeout: expected delete event")
	}
	if e.Op != OpDelete {
		t.Errorf("op: want OpDelete got %v", e.Op)
	}
}

// file rename

func TestFSWatcher_FileRename(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "old.txt")
	writeFile(t, src, "content")

	w := newTestWatcher(t, dir, nil, false)

	dst := filepath.Join(dir, "new.txt")
	if err := os.Rename(src, dst); err != nil {
		t.Fatalf("rename: %v", err)
	}

	e, ok := waitEvent(t, w, testDelay)
	if !ok {
		t.Fatal("timeout: expected rename/delete event")
	}
	if e.Op != OpDelete && e.Op != OpCreated {
		t.Errorf("op: want OpDelete or OpCreate got %v", e.Op)
	}
}

// extension filter

func TestFSWatcher_ExtFilter_Matching(t *testing.T) {
	dir := t.TempDir()
	w := newTestWatcher(t, dir, []string{".lua"}, false)

	writeFile(t, filepath.Join(dir, "script.lua"), "-- lua")

	e, ok := waitEvent(t, w, testDelay)
	if !ok {
		t.Fatal("timeout: expected event for .lua file")
	}
	if filepath.Ext(e.Path) != ".lua" {
		t.Errorf("ext: want .lua got %s", filepath.Ext(e.Path))
	}
}

func TestFSWatcher_ExtFilter_NotMatching(t *testing.T) {
	dir := t.TempDir()
	w := newTestWatcher(t, dir, []string{".lua"}, false)

	writeFile(t, filepath.Join(dir, "readme.txt"), "text")

	noEvent(t, w, testDelay)
}

func TestFSWatcher_ExtFilter_Multiple(t *testing.T) {
	dir := t.TempDir()
	w := newTestWatcher(t, dir, []string{".lua", ".json"}, false)

	writeFile(t, filepath.Join(dir, "data.json"), "{}")

	e, ok := waitEvent(t, w, testDelay)
	if !ok {
		t.Fatal("timeout: expected event for .json file")
	}
	if filepath.Ext(e.Path) != ".json" {
		t.Errorf("ext: want .json got %s", filepath.Ext(e.Path))
	}
}

func TestFSWatcher_ExtFilter_None(t *testing.T) {
	dir := t.TempDir()
	w := newTestWatcher(t, dir, nil, false) // no filter = all files

	writeFile(t, filepath.Join(dir, "any.xyz"), "data")

	_, ok := waitEvent(t, w, testDelay)
	if !ok {
		t.Fatal("timeout: expected event with no ext filter")
	}
}

// hidden files ignored

func TestFSWatcher_HiddenFile_Ignored(t *testing.T) {
	dir := t.TempDir()
	w := newTestWatcher(t, dir, nil, false)

	writeFile(t, filepath.Join(dir, ".hidden"), "secret")

	noEvent(t, w, testDelay)
}

func TestFSWatcher_HiddenFile_DotPrefix_Ignored(t *testing.T) {
	dir := t.TempDir()
	w := newTestWatcher(t, dir, nil, false)

	writeFile(t, filepath.Join(dir, ".gitignore"), "node_modules")

	noEvent(t, w, testDelay)
}

// debounce

func TestFSWatcher_Debounce_MultipleWritesOneEvent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	writeFile(t, path, "v0")

	w := newTestWatcher(t, dir, nil, false)

	// rapid writes
	for i := range 5 {
		writeFile(t, path, fmt.Sprintf("v%d", i+1))
	}

	// should receive only one event
	_, ok := waitEvent(t, w, testDelay)
	if !ok {
		t.Fatal("timeout: expected at least one event")
	}
	noEvent(t, w, testDelay)
}

// recursive

func TestFSWatcher_Recursive_SubdirFile(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "subdir")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	w := newTestWatcher(t, dir, nil, true)

	writeFile(t, filepath.Join(sub, "nested.txt"), "hello")

	_, ok := waitEvent(t, w, testDelay)
	if !ok {
		t.Fatal("timeout: expected event in subdirectory")
	}
}

func TestFSWatcher_NonRecursive_SubdirFile_NoEvent(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "subdir")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	w := newTestWatcher(t, dir, nil, false)

	writeFile(t, filepath.Join(sub, "nested.txt"), "hello")

	noEvent(t, w, testDelay)
}

// stop

func TestFSWatcher_Stop_NoEventAfterStop(t *testing.T) {
	dir := t.TempDir()
	w := newTestWatcher(t, dir, nil, false)

	w.Stop()
	time.Sleep(50 * time.Millisecond)

	writeFile(t, filepath.Join(dir, "after_stop.txt"), "data")

	noEvent(t, w, testDelay)
}

func TestFSWatcher_Stop_Idempotent(t *testing.T) {
	dir := t.TempDir()
	w, err := NewFSWatcher(dir, nil, false)
	if err != nil {
		t.Fatalf("NewFSWatcher: %v", err)
	}
	if err := w.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// Stopping twice should not panic.
	w.Stop()
	w.Stop()
}
