package pluginwrapper

import (
	"encoding/base64"
	"errors"
	"testing"

	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/pluginapi"
)

// mockKV is a simple in-memory KeyValueStore
type mockKV struct {
	data   map[string][]byte
	getErr error
	setErr error
	delErr error
	lstErr error
}

func newMockKV() *mockKV {
	return &mockKV{data: make(map[string][]byte)}
}

func (m *mockKV) Get(key string) ([]byte, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	v, ok := m.data[key]
	if !ok {
		return nil, errors.New("key not found")
	}
	return v, nil
}

func (m *mockKV) Set(key string, value []byte) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.data[key] = value
	return nil
}

func (m *mockKV) Delete(key string) error {
	if m.delErr != nil {
		return m.delErr
	}
	delete(m.data, key)
	return nil
}

func (m *mockKV) List(prefix string) (map[string][]byte, error) {
	if m.lstErr != nil {
		return nil, m.lstErr
	}
	result := make(map[string][]byte)
	for k, v := range m.data {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			result[k] = v
		}
	}
	return result, nil
}

func newKVState(t *testing.T, kv pluginapi.KeyValueStore) *lua.LState {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })

	RegisterKeyValueStore(L, kv)

	return L
}

// get

func TestKV_Get_OK(t *testing.T) {
	kv := newMockKV()
	kv.data["foo"] = []byte("bar")
	L := newKVState(t, kv)

	if err := L.DoString(`val, err = kv.get("foo")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if v := L.GetGlobal("val").String(); v != "bar" {
		t.Errorf("val: want bar got %s", v)
	}
	if L.GetGlobal("err").Type() != lua.LTNil {
		t.Errorf("err: want nil got %v", L.GetGlobal("err"))
	}
}

func TestKV_Get_KeyNotFound(t *testing.T) {
	L := newKVState(t, newMockKV())

	if err := L.DoString(`val, err = kv.get("missing")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("val").Type() != lua.LTNil {
		t.Errorf("val: want nil got %v", L.GetGlobal("val"))
	}
	if L.GetGlobal("err").Type() != lua.LTString {
		t.Errorf("err: want string got %s", L.GetGlobal("err").Type())
	}
}

func TestKV_Get_StoreError(t *testing.T) {
	kv := newMockKV()
	kv.getErr = errors.New("store unavailable")
	L := newKVState(t, kv)

	if err := L.DoString(`val, err = kv.get("foo")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("val").Type() != lua.LTNil {
		t.Errorf("val: want nil")
	}
	if got := L.GetGlobal("err").String(); got != "store unavailable" {
		t.Errorf("err: want 'store unavailable' got %q", got)
	}
}

// get_base64

func TestKV_GetBase64_OK(t *testing.T) {
	kv := newMockKV()
	kv.data["img"] = []byte("binary\x00data")
	L := newKVState(t, kv)

	if err := L.DoString(`val, err = kv.get_base64("img")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	want := base64.StdEncoding.EncodeToString([]byte("binary\x00data"))
	if got := L.GetGlobal("val").String(); got != want {
		t.Errorf("val: want %q got %q", want, got)
	}
	if L.GetGlobal("err").Type() != lua.LTNil {
		t.Errorf("err: want nil got %v", L.GetGlobal("err"))
	}
}

func TestKV_GetBase64_Error(t *testing.T) {
	kv := newMockKV()
	kv.getErr = errors.New("get failed")
	L := newKVState(t, kv)

	if err := L.DoString(`val, err = kv.get_base64("key")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("val").Type() != lua.LTNil {
		t.Errorf("val: want nil")
	}
	if L.GetGlobal("err").String() != "get failed" {
		t.Errorf("err: want 'get failed' got %q", L.GetGlobal("err").String())
	}
}

// set

func TestKV_Set_OK(t *testing.T) {
	kv := newMockKV()
	L := newKVState(t, kv)

	if err := L.DoString(`ok, err = kv.set("name", "alice")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("ok").String() != "true" {
		t.Errorf("ok: want true got %v", L.GetGlobal("ok"))
	}
	if L.GetGlobal("err").Type() != lua.LTNil {
		t.Errorf("err: want nil got %v", L.GetGlobal("err"))
	}
	if string(kv.data["name"]) != "alice" {
		t.Errorf("stored value: want alice got %s", kv.data["name"])
	}
}

func TestKV_Set_Error(t *testing.T) {
	kv := newMockKV()
	kv.setErr = errors.New("write failed")
	L := newKVState(t, kv)

	if err := L.DoString(`ok, err = kv.set("k", "v")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("ok").String() != "false" {
		t.Errorf("ok: want false got %v", L.GetGlobal("ok"))
	}
	if L.GetGlobal("err").String() != "write failed" {
		t.Errorf("err: want 'write failed' got %q", L.GetGlobal("err").String())
	}
}

// delete

func TestKV_Delete_OK(t *testing.T) {
	kv := newMockKV()
	kv.data["todel"] = []byte("x")
	L := newKVState(t, kv)

	if err := L.DoString(`ok, err = kv.delete("todel")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("ok").String() != "true" {
		t.Errorf("ok: want true got %v", L.GetGlobal("ok"))
	}
	if _, exists := kv.data["todel"]; exists {
		t.Error("key should have been deleted")
	}
}

func TestKV_Delete_Error(t *testing.T) {
	kv := newMockKV()
	kv.delErr = errors.New("delete failed")
	L := newKVState(t, kv)

	if err := L.DoString(`ok, err = kv.delete("k")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("ok").String() != "false" {
		t.Errorf("ok: want false")
	}
	if L.GetGlobal("err").String() != "delete failed" {
		t.Errorf("err: want 'delete failed' got %q", L.GetGlobal("err").String())
	}
}

// list

func TestKV_List_OK(t *testing.T) {
	kv := newMockKV()
	kv.data["user:1"] = []byte("alice")
	kv.data["user:2"] = []byte("bob")
	kv.data["other"] = []byte("skip")
	L := newKVState(t, kv)

	if err := L.DoString(`tbl, err = kv.list("user:")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("err").Type() != lua.LTNil {
		t.Errorf("err: want nil got %v", L.GetGlobal("err"))
	}
	tbl, ok := L.GetGlobal("tbl").(*lua.LTable)
	if !ok {
		t.Fatalf("tbl: want table got %T", L.GetGlobal("tbl"))
	}
	if v := tbl.RawGetString("user:1").String(); v != "alice" {
		t.Errorf("user:1: want alice got %s", v)
	}
	if v := tbl.RawGetString("user:2").String(); v != "bob" {
		t.Errorf("user:2: want bob got %s", v)
	}
	if v := tbl.RawGetString("other"); v.Type() != lua.LTNil {
		t.Errorf("other: should not be in result, got %v", v)
	}
}

func TestKV_List_Empty(t *testing.T) {
	L := newKVState(t, newMockKV())

	if err := L.DoString(`tbl, err = kv.list("prefix:")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	tbl, ok := L.GetGlobal("tbl").(*lua.LTable)
	if !ok {
		t.Fatalf("tbl: want table got %T", L.GetGlobal("tbl"))
	}
	count := 0
	tbl.ForEach(func(_, _ lua.LValue) { count++ })
	if count != 0 {
		t.Errorf("expected empty table, got %d entries", count)
	}
}

func TestKV_List_Error(t *testing.T) {
	kv := newMockKV()
	kv.lstErr = errors.New("list failed")
	L := newKVState(t, kv)

	if err := L.DoString(`tbl, err = kv.list("x")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("tbl").Type() != lua.LTNil {
		t.Errorf("tbl: want nil")
	}
	if L.GetGlobal("err").String() != "list failed" {
		t.Errorf("err: want 'list failed' got %q", L.GetGlobal("err").String())
	}
}

// set then get roundtrip

func TestKV_SetGet_Roundtrip(t *testing.T) {
	kv := newMockKV()
	L := newKVState(t, kv)

	if err := L.DoString(`
		kv.set("round", "trip")
		val, err = kv.get("round")
	`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if v := L.GetGlobal("val").String(); v != "trip" {
		t.Errorf("roundtrip: want trip got %s", v)
	}
}
