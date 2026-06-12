package pluginwrapper

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/pluginapi"
)

type mockDataStore struct {
	data    map[string][]byte
	readErr error
	listErr error
}

func newMockDataStore() *mockDataStore {
	return &mockDataStore{data: make(map[string][]byte)}
}

func (m *mockDataStore) Exists(name string) bool {
	_, ok := m.data[name]

	return ok
}

func (m *mockDataStore) Read(name string) (io.ReadCloser, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}

	v, ok := m.data[name]
	if !ok {
		return nil, errors.New("not found")
	}

	return io.NopCloser(bytes.NewReader(v)), nil
}

func (m *mockDataStore) List(path string, exts []string) ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}

	var result []string
	for k := range m.data {
		// Filters by path prefix.
		if !strings.HasPrefix(k, path) {
			continue
		}

		// Filters by extension if provided.
		if len(exts) > 0 {
			matched := false
			for _, ext := range exts {
				if strings.HasSuffix(k, "."+ext) {
					matched = true
					break
				}
			}

			if !matched {
				continue
			}
		}

		result = append(result, k)
	}

	return result, nil
}

func newDataState(t *testing.T, store pluginapi.DataStore) *lua.LState {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })

	RegisterDataStore(L, store)

	return L
}

// Exists

func TestDataStore_Exists_True(t *testing.T) {
	store := newMockDataStore()
	store.data["config.json"] = []byte("{}")
	L := newDataState(t, store)

	if err := L.DoString(`result = data.exists("config.json")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("result").String() != "true" {
		t.Errorf("result: want true got %v", L.GetGlobal("result"))
	}
}

func TestDataStore_Exists_False(t *testing.T) {
	L := newDataState(t, newMockDataStore())

	if err := L.DoString(`result = data.exists("nope.txt")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("result").String() != "false" {
		t.Errorf("result: want false got %v", L.GetGlobal("result"))
	}
}

// Read

func TestDataStore_Read_OK(t *testing.T) {
	store := newMockDataStore()
	store.data["rules.txt"] = []byte("allow all")
	L := newDataState(t, store)

	if err := L.DoString(`content, err = data.read("rules.txt")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("err").Type() != lua.LTNil {
		t.Errorf("err: want nil got %v", L.GetGlobal("err"))
	}
	if L.GetGlobal("content").String() != "allow all" {
		t.Errorf("content: want 'allow all' got %q", L.GetGlobal("content").String())
	}
}

func TestDataStore_Read_NotFound(t *testing.T) {
	L := newDataState(t, newMockDataStore())

	if err := L.DoString(`content, err = data.read("missing.txt")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("content").Type() != lua.LTNil {
		t.Errorf("content: want nil got %v", L.GetGlobal("content"))
	}
	if L.GetGlobal("err").Type() != lua.LTString {
		t.Errorf("err: want string got %s", L.GetGlobal("err").Type())
	}
}

func TestDataStore_Read_StoreError(t *testing.T) {
	store := newMockDataStore()
	store.readErr = errors.New("io error")
	L := newDataState(t, store)

	if err := L.DoString(`content, err = data.read("f.txt")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("content").Type() != lua.LTNil {
		t.Errorf("content: want nil")
	}
	if L.GetGlobal("err").String() != "io error" {
		t.Errorf("err: want 'io error' got %q", L.GetGlobal("err").String())
	}
}

// List

func TestDataStore_List_NoFilter(t *testing.T) {
	store := newMockDataStore()
	store.data["dir/a.txt"] = []byte("a")
	store.data["dir/b.txt"] = []byte("b")
	store.data["other/c.txt"] = []byte("c")
	L := newDataState(t, store)

	if err := L.DoString(`files, err = data.list("dir/", nil)`); err != nil {
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
		t.Errorf("files: want 2 got %d", count)
	}
}

func TestDataStore_List_WithExtFilter(t *testing.T) {
	store := newMockDataStore()
	store.data["dir/a.txt"] = []byte("a")
	store.data["dir/b.json"] = []byte("{}")
	store.data["dir/c.txt"] = []byte("c")
	L := newDataState(t, store)

	if err := L.DoString(`files, err = data.list("dir/", { "txt" })`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	tbl, ok := L.GetGlobal("files").(*lua.LTable)
	if !ok {
		t.Fatalf("files: want table got %T", L.GetGlobal("files"))
	}
	count := 0
	tbl.ForEach(func(_, _ lua.LValue) { count++ })
	if count != 2 {
		t.Errorf("files: want 2 txt files got %d", count)
	}
}

func TestDataStore_List_MultipleExtFilters(t *testing.T) {
	store := newMockDataStore()
	store.data["dir/a.txt"] = []byte("a")
	store.data["dir/b.json"] = []byte("{}")
	store.data["dir/c.lua"] = []byte("--")
	store.data["dir/d.bin"] = []byte("x")
	L := newDataState(t, store)

	if err := L.DoString(`files, err = data.list("dir/", { "txt", "json" })`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	tbl, ok := L.GetGlobal("files").(*lua.LTable)
	if !ok {
		t.Fatalf("files: want table got %T", L.GetGlobal("files"))
	}
	count := 0
	tbl.ForEach(func(_, _ lua.LValue) { count++ })
	if count != 2 {
		t.Errorf("files: want 2 got %d", count)
	}
}

func TestDataStore_List_Empty(t *testing.T) {
	L := newDataState(t, newMockDataStore())

	if err := L.DoString(`files, err = data.list("dir/", nil)`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	tbl, ok := L.GetGlobal("files").(*lua.LTable)
	if !ok {
		t.Fatalf("files: want table got %T", L.GetGlobal("files"))
	}
	count := 0
	tbl.ForEach(func(_, _ lua.LValue) { count++ })
	if count != 0 {
		t.Errorf("files: want empty got %d", count)
	}
}

func TestDataStore_List_Error(t *testing.T) {
	store := newMockDataStore()
	store.listErr = errors.New("list failed")
	L := newDataState(t, store)

	if err := L.DoString(`files, err = data.list("dir/", nil)`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("files").Type() != lua.LTNil {
		t.Errorf("files: want nil")
	}
	if L.GetGlobal("err").String() != "list failed" {
		t.Errorf("err: want 'list failed' got %q", L.GetGlobal("err").String())
	}
}

// List returns 1-indexed lua table.

func TestDataStore_List_LuaIterable(t *testing.T) {
	store := newMockDataStore()
	store.data["p/x.txt"] = []byte("x")
	L := newDataState(t, store)

	// ipairs works only on 1-indexed tables.
	if err := L.DoString(`
		files, err = data.list("p/", nil)
		count = 0
		for i, v in ipairs(files) do
			count = count + 1
		end
	`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("count").String() != "1" {
		t.Errorf("count: want 1 got %v", L.GetGlobal("count"))
	}
}

// Exists then read sequence.

func TestDataStore_ExistsRead_Sequence(t *testing.T) {
	store := newMockDataStore()
	store.data["seq.txt"] = []byte("sequential")
	L := newDataState(t, store)

	if err := L.DoString(`
		if data.exists("seq.txt") then
			content, err = data.read("seq.txt")
		end
	`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("content").String() != "sequential" {
		t.Errorf("content: want 'sequential' got %q", L.GetGlobal("content").String())
	}
}
