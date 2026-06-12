package pkg

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newJSONState(t *testing.T) *lua.LState {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })
	if err := RegisterJSON(L); err != nil {
		t.Fatalf("RegisterJSON: %v", err)
	}
	return L
}

// encode

func TestJSON_Encode_String(t *testing.T) {
	L := newJSONState(t)
	if err := L.DoString(`result, err = json.encode("hello")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("err").Type() != lua.LTNil {
		t.Errorf("err: want nil got %v", L.GetGlobal("err"))
	}
	if got := L.GetGlobal("result").String(); got != `"hello"` {
		t.Errorf("encode string: want %q got %q", `"hello"`, got)
	}
}

func TestJSON_Encode_Number(t *testing.T) {
	L := newJSONState(t)
	if err := L.DoString(`result, err = json.encode(42)`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if got := L.GetGlobal("result").String(); got != "42" {
		t.Errorf("encode number: want 42 got %q", got)
	}
}

func TestJSON_Encode_BoolTrue(t *testing.T) {
	L := newJSONState(t)
	if err := L.DoString(`result, err = json.encode(true)`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if got := L.GetGlobal("result").String(); got != "true" {
		t.Errorf("encode bool: want true got %q", got)
	}
}

func TestJSON_Encode_Nil(t *testing.T) {
	L := newJSONState(t)
	if err := L.DoString(`result, err = json.encode(nil)`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if got := L.GetGlobal("result").String(); got != "null" {
		t.Errorf("encode nil: want null got %q", got)
	}
}

func TestJSON_Encode_Table(t *testing.T) {
	L := newJSONState(t)
	if err := L.DoString(`result, err = json.encode({ name = "alice", age = 30 })`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("err").Type() != lua.LTNil {
		t.Errorf("err: want nil got %v", L.GetGlobal("err"))
	}
	// Decodes result back to verify fields — map order is not guaranteed.
	if err := L.DoString(`decoded, _ = json.decode(result)`); err != nil {
		t.Fatalf("lua error on decode: %v", err)
	}
	tbl, ok := L.GetGlobal("decoded").(*lua.LTable)
	if !ok {
		t.Fatalf("decoded: want table got %T", L.GetGlobal("decoded"))
	}
	if tbl.RawGetString("name").String() != "alice" {
		t.Errorf("name: want alice got %v", tbl.RawGetString("name"))
	}
	if tbl.RawGetString("age") != lua.LNumber(30) {
		t.Errorf("age: want 30 got %v", tbl.RawGetString("age"))
	}
}

func TestJSON_Encode_NestedTable(t *testing.T) {
	L := newJSONState(t)
	if err := L.DoString(`
		result, err = json.encode({ outer = { inner = "val" } })
	`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("err").Type() != lua.LTNil {
		t.Errorf("err: want nil got %v", L.GetGlobal("err"))
	}
	// Verifies roundtrip.
	if err := L.DoString(`decoded, _ = json.decode(result)`); err != nil {
		t.Fatalf("lua decode error: %v", err)
	}
	tbl := L.GetGlobal("decoded").(*lua.LTable)
	outer := tbl.RawGetString("outer").(*lua.LTable)
	if outer.RawGetString("inner").String() != "val" {
		t.Errorf("outer.inner: want val got %v", outer.RawGetString("inner"))
	}
}

// decode

func TestJSON_Decode_String(t *testing.T) {
	L := newJSONState(t)
	if err := L.DoString(`result, err = json.decode('"hello"')`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("err").Type() != lua.LTNil {
		t.Errorf("err: want nil got %v", L.GetGlobal("err"))
	}
	if got := L.GetGlobal("result").String(); got != "hello" {
		t.Errorf("decode string: want hello got %q", got)
	}
}

func TestJSON_Decode_Number(t *testing.T) {
	L := newJSONState(t)
	if err := L.DoString(`result, err = json.decode("3.14")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if got := L.GetGlobal("result"); got != lua.LNumber(3.14) {
		t.Errorf("decode number: want 3.14 got %v", got)
	}
}

func TestJSON_Decode_Bool(t *testing.T) {
	L := newJSONState(t)
	if err := L.DoString(`result, err = json.decode("false")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if got := L.GetGlobal("result"); got != lua.LFalse {
		t.Errorf("decode bool: want false got %v", got)
	}
}

func TestJSON_Decode_Null(t *testing.T) {
	L := newJSONState(t)
	if err := L.DoString(`result, err = json.decode("null")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("result").Type() != lua.LTNil {
		t.Errorf("decode null: want LNil got %v", L.GetGlobal("result"))
	}
}

func TestJSON_Decode_Object(t *testing.T) {
	L := newJSONState(t)
	if err := L.DoString(`result, err = json.decode('{"name":"bob","score":99}')`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("err").Type() != lua.LTNil {
		t.Errorf("err: want nil got %v", L.GetGlobal("err"))
	}
	tbl, ok := L.GetGlobal("result").(*lua.LTable)
	if !ok {
		t.Fatalf("result: want table got %T", L.GetGlobal("result"))
	}
	if tbl.RawGetString("name").String() != "bob" {
		t.Errorf("name: want bob got %v", tbl.RawGetString("name"))
	}
	if tbl.RawGetString("score") != lua.LNumber(99) {
		t.Errorf("score: want 99 got %v", tbl.RawGetString("score"))
	}
}

func TestJSON_Decode_Array(t *testing.T) {
	L := newJSONState(t)
	if err := L.DoString(`result, err = json.decode('["a","b","c"]')`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	tbl, ok := L.GetGlobal("result").(*lua.LTable)
	if !ok {
		t.Fatalf("result: want table got %T", L.GetGlobal("result"))
	}
	if tbl.RawGetInt(1).String() != "a" {
		t.Errorf("result[1]: want a got %v", tbl.RawGetInt(1))
	}
	if tbl.RawGetInt(2).String() != "b" {
		t.Errorf("result[2]: want b got %v", tbl.RawGetInt(2))
	}
	if tbl.RawGetInt(3).String() != "c" {
		t.Errorf("result[3]: want c got %v", tbl.RawGetInt(3))
	}
}

func TestJSON_Decode_NestedObject(t *testing.T) {
	L := newJSONState(t)
	if err := L.DoString(`result, err = json.decode('{"outer":{"inner":42}}')`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	tbl := L.GetGlobal("result").(*lua.LTable)
	outer := tbl.RawGetString("outer").(*lua.LTable)
	if outer.RawGetString("inner") != lua.LNumber(42) {
		t.Errorf("outer.inner: want 42 got %v", outer.RawGetString("inner"))
	}
}

func TestJSON_Decode_InvalidJSON(t *testing.T) {
	L := newJSONState(t)
	if err := L.DoString(`result, err = json.decode("not json {")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("result").Type() != lua.LTNil {
		t.Errorf("result: want nil on error")
	}
	if L.GetGlobal("err").Type() != lua.LTString {
		t.Errorf("err: want string on invalid json")
	}
}

func TestJSON_Decode_EmptyString(t *testing.T) {
	L := newJSONState(t)
	if err := L.DoString(`result, err = json.decode("")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("result").Type() != lua.LTNil {
		t.Errorf("result: want nil on empty input")
	}
	if L.GetGlobal("err").Type() != lua.LTString {
		t.Errorf("err: want string on empty input")
	}
}

// Roundtrip.

func TestJSON_Roundtrip_TableToStringToTable(t *testing.T) {
	L := newJSONState(t)
	if err := L.DoString(`
		original = { city = "london", pop = 9000000 }
		encoded, _ = json.encode(original)
		decoded, err = json.decode(encoded)
	`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("err").Type() != lua.LTNil {
		t.Errorf("err: want nil got %v", L.GetGlobal("err"))
	}
	tbl, ok := L.GetGlobal("decoded").(*lua.LTable)
	if !ok {
		t.Fatalf("decoded: want table got %T", L.GetGlobal("decoded"))
	}
	if tbl.RawGetString("city").String() != "london" {
		t.Errorf("city: want london got %v", tbl.RawGetString("city"))
	}
	if tbl.RawGetString("pop") != lua.LNumber(9000000) {
		t.Errorf("pop: want 9000000 got %v", tbl.RawGetString("pop"))
	}
}
