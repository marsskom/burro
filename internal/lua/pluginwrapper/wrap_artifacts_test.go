package pluginwrapper

import (
	"bytes"
	"errors"
	"io"
	"testing"

	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/pluginapi"
)

// mockArtifactStore is an in-memory ArtifactStore
type mockArtifactStore struct {
	data      map[string][]byte
	writeErr  error
	readErr   error
	deleteErr error
	renameErr error
	listErr   error
}

func newMockArtifactStore() *mockArtifactStore {
	return &mockArtifactStore{data: make(map[string][]byte)}
}

type nopWriteCloser struct{ buf *bytes.Buffer }

func (w *nopWriteCloser) Write(p []byte) (int, error) { return w.buf.Write(p) }
func (w *nopWriteCloser) Close() error                { return nil }

func (m *mockArtifactStore) Create(name string) (io.WriteCloser, error) {
	buf := &bytes.Buffer{}
	m.data[name] = buf.Bytes()
	return &nopWriteCloser{buf: buf}, nil
}

func (m *mockArtifactStore) Exists(name string) bool {
	_, ok := m.data[name]
	return ok
}

func (m *mockArtifactStore) Rename(oldpath, newpath string) error {
	if m.renameErr != nil {
		return m.renameErr
	}
	v, ok := m.data[oldpath]
	if !ok {
		return errors.New("not found")
	}
	m.data[newpath] = v
	delete(m.data, oldpath)
	return nil
}

func (m *mockArtifactStore) Write(name string, r io.Reader) (string, error) {
	if m.writeErr != nil {
		return "", m.writeErr
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	m.data[name] = b
	return name, nil
}

func (m *mockArtifactStore) Read(name string) (io.ReadCloser, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	v, ok := m.data[name]
	if !ok {
		return nil, errors.New("not found")
	}
	return io.NopCloser(bytes.NewReader(v)), nil
}

func (m *mockArtifactStore) Delete(name string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.data, name)
	return nil
}

func (m *mockArtifactStore) List() ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	names := make([]string, 0, len(m.data))
	for k := range m.data {
		names = append(names, k)
	}
	return names, nil
}

func newArtifactState(t *testing.T, store pluginapi.ArtifactStore) *lua.LState {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })

	RegisterArtifactsStore(L, store)

	return L
}

// Write

func TestArtifacts_Write_OK(t *testing.T) {
	store := newMockArtifactStore()
	L := newArtifactState(t, store)

	if err := L.DoString(`ok, err = artifacts.write("file.txt", "hello")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("ok").String() != "true" {
		t.Errorf("ok: want true got %v", L.GetGlobal("ok"))
	}
	if L.GetGlobal("err").Type() != lua.LTNil {
		t.Errorf("err: want nil got %v", L.GetGlobal("err"))
	}
	if string(store.data["file.txt"]) != "hello" {
		t.Errorf("stored: want hello got %s", store.data["file.txt"])
	}
}

func TestArtifacts_Write_Error(t *testing.T) {
	store := newMockArtifactStore()
	store.writeErr = errors.New("disk full")
	L := newArtifactState(t, store)

	if err := L.DoString(`ok, err = artifacts.write("file.txt", "hello")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("ok").String() != "false" {
		t.Errorf("ok: want false got %v", L.GetGlobal("ok"))
	}
	if L.GetGlobal("err").String() != "disk full" {
		t.Errorf("err: want 'disk full' got %q", L.GetGlobal("err").String())
	}
}

// Read

func TestArtifacts_Read_OK(t *testing.T) {
	store := newMockArtifactStore()
	store.data["hello.txt"] = []byte("world")
	L := newArtifactState(t, store)

	if err := L.DoString(`content, err = artifacts.read("hello.txt")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("err").Type() != lua.LTNil {
		t.Errorf("err: want nil got %v", L.GetGlobal("err"))
	}
	if L.GetGlobal("content").String() != "world" {
		t.Errorf("content: want world got %s", L.GetGlobal("content").String())
	}
}

func TestArtifacts_Read_NotFound(t *testing.T) {
	L := newArtifactState(t, newMockArtifactStore())

	if err := L.DoString(`content, err = artifacts.read("missing.txt")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("content").Type() != lua.LTNil {
		t.Errorf("content: want nil got %v", L.GetGlobal("content"))
	}
	if L.GetGlobal("err").Type() != lua.LTString {
		t.Errorf("err: want string got %s", L.GetGlobal("err").Type())
	}
}

func TestArtifacts_Read_Error(t *testing.T) {
	store := newMockArtifactStore()
	store.readErr = errors.New("read failed")
	L := newArtifactState(t, store)

	if err := L.DoString(`content, err = artifacts.read("f.txt")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("err").String() != "read failed" {
		t.Errorf("err: want 'read failed' got %q", L.GetGlobal("err").String())
	}
}

// Exists

func TestArtifacts_Exists_True(t *testing.T) {
	store := newMockArtifactStore()
	store.data["exists.txt"] = []byte("x")
	L := newArtifactState(t, store)

	if err := L.DoString(`result = artifacts.exists("exists.txt")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("result").String() != "true" {
		t.Errorf("result: want true got %v", L.GetGlobal("result"))
	}
}

func TestArtifacts_Exists_False(t *testing.T) {
	L := newArtifactState(t, newMockArtifactStore())

	if err := L.DoString(`result = artifacts.exists("nope.txt")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("result").String() != "false" {
		t.Errorf("result: want false got %v", L.GetGlobal("result"))
	}
}

// Delete

func TestArtifacts_Delete_OK(t *testing.T) {
	store := newMockArtifactStore()
	store.data["todel.txt"] = []byte("bye")
	L := newArtifactState(t, store)

	if err := L.DoString(`ok, err = artifacts.delete("todel.txt")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("ok").String() != "true" {
		t.Errorf("ok: want true got %v", L.GetGlobal("ok"))
	}
	if _, exists := store.data["todel.txt"]; exists {
		t.Error("file should have been deleted")
	}
}

func TestArtifacts_Delete_Error(t *testing.T) {
	store := newMockArtifactStore()
	store.deleteErr = errors.New("delete failed")
	L := newArtifactState(t, store)

	if err := L.DoString(`ok, err = artifacts.delete("f.txt")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("ok").String() != "false" {
		t.Errorf("ok: want false")
	}
	if L.GetGlobal("err").String() != "delete failed" {
		t.Errorf("err: want 'delete failed' got %q", L.GetGlobal("err").String())
	}
}

// Rename

func TestArtifacts_Rename_OK(t *testing.T) {
	store := newMockArtifactStore()
	store.data["old.txt"] = []byte("content")
	L := newArtifactState(t, store)

	if err := L.DoString(`ok, err = artifacts.rename("old.txt", "new.txt")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("ok").String() != "true" {
		t.Errorf("ok: want true got %v", L.GetGlobal("ok"))
	}
	if _, exists := store.data["old.txt"]; exists {
		t.Error("old.txt should not exist after rename")
	}
	if string(store.data["new.txt"]) != "content" {
		t.Errorf("new.txt: want content got %s", store.data["new.txt"])
	}
}

func TestArtifacts_Rename_Error(t *testing.T) {
	store := newMockArtifactStore()
	store.renameErr = errors.New("rename failed")
	L := newArtifactState(t, store)

	if err := L.DoString(`ok, err = artifacts.rename("a.txt", "b.txt")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("ok").String() != "false" {
		t.Errorf("ok: want false")
	}
	if L.GetGlobal("err").String() != "rename failed" {
		t.Errorf("err: want 'rename failed' got %q", L.GetGlobal("err").String())
	}
}

// List

func TestArtifacts_List_OK(t *testing.T) {
	store := newMockArtifactStore()
	store.data["a.txt"] = []byte("a")
	store.data["b.txt"] = []byte("b")
	L := newArtifactState(t, store)

	if err := L.DoString(`files, err = artifacts.list()`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("err").Type() != lua.LTNil {
		t.Errorf("err: want nil got %v", L.GetGlobal("err"))
	}
	tbl, ok := L.GetGlobal("files").(*lua.LTable)
	if !ok {
		t.Fatalf("files: want table got %T", L.GetGlobal("files"))
	}
	count := 0
	tbl.ForEach(func(_, _ lua.LValue) { count++ })
	if count != 2 {
		t.Errorf("files: want 2 entries got %d", count)
	}
}

func TestArtifacts_List_Empty(t *testing.T) {
	L := newArtifactState(t, newMockArtifactStore())

	if err := L.DoString(`files, err = artifacts.list()`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	tbl, ok := L.GetGlobal("files").(*lua.LTable)
	if !ok {
		t.Fatalf("files: want table got %T", L.GetGlobal("files"))
	}
	count := 0
	tbl.ForEach(func(_, _ lua.LValue) { count++ })
	if count != 0 {
		t.Errorf("files: want empty table got %d entries", count)
	}
}

func TestArtifacts_List_Error(t *testing.T) {
	store := newMockArtifactStore()
	store.listErr = errors.New("list failed")
	L := newArtifactState(t, store)

	if err := L.DoString(`files, err = artifacts.list()`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("files").Type() != lua.LTNil {
		t.Errorf("files: want nil")
	}
	if L.GetGlobal("err").String() != "list failed" {
		t.Errorf("err: want 'list failed' got %q", L.GetGlobal("err").String())
	}
}

// Write then read roundtrip.

func TestArtifacts_WriteRead_Roundtrip(t *testing.T) {
	store := newMockArtifactStore()
	L := newArtifactState(t, store)

	if err := L.DoString(`
		artifacts.write("round.txt", "trip content")
		content, err = artifacts.read("round.txt")
	`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("content").String() != "trip content" {
		t.Errorf("roundtrip: want 'trip content' got %q", L.GetGlobal("content").String())
	}
}

// Write then exists then delete.

func TestArtifacts_WriteExistsDelete_Sequence(t *testing.T) {
	store := newMockArtifactStore()
	L := newArtifactState(t, store)

	if err := L.DoString(`
		artifacts.write("seq.txt", "data")
		existed_before = artifacts.exists("seq.txt")
		artifacts.delete("seq.txt")
		existed_after = artifacts.exists("seq.txt")
	`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("existed_before").String() != "true" {
		t.Errorf("existed_before: want true")
	}
	if L.GetGlobal("existed_after").String() != "false" {
		t.Errorf("existed_after: want false")
	}
}
