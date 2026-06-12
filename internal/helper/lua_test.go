package helper

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newL(t *testing.T) *lua.LState {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })
	return L
}

// LuaValueToGo

func TestLuaValueToGo_Nil(t *testing.T) {
	if got := LuaValueToGo(lua.LNil); got != nil {
		t.Errorf("nil: want nil got %v", got)
	}
}

func TestLuaValueToGo_BoolTrue(t *testing.T) {
	got := LuaValueToGo(lua.LTrue)
	if got != true {
		t.Errorf("bool true: want true got %v", got)
	}
}

func TestLuaValueToGo_BoolFalse(t *testing.T) {
	got := LuaValueToGo(lua.LFalse)
	if got != false {
		t.Errorf("bool false: want false got %v", got)
	}
}

func TestLuaValueToGo_Number(t *testing.T) {
	got := LuaValueToGo(lua.LNumber(42.5))
	if got != float64(42.5) {
		t.Errorf("number: want 42.5 got %v", got)
	}
}

func TestLuaValueToGo_NumberZero(t *testing.T) {
	got := LuaValueToGo(lua.LNumber(0))
	if got != float64(0) {
		t.Errorf("number zero: want 0.0 got %v", got)
	}
}

func TestLuaValueToGo_String(t *testing.T) {
	got := LuaValueToGo(lua.LString("hello"))
	if got != "hello" {
		t.Errorf("string: want hello got %v", got)
	}
}

func TestLuaValueToGo_StringEmpty(t *testing.T) {
	got := LuaValueToGo(lua.LString(""))
	if got != "" {
		t.Errorf("empty string: want '' got %v", got)
	}
}

func TestLuaValueToGo_Table(t *testing.T) {
	L := newL(t)
	tbl := L.NewTable()
	tbl.RawSetString("key", lua.LString("val"))
	tbl.RawSetString("num", lua.LNumber(7))

	got, ok := LuaValueToGo(tbl).(map[string]any)
	if !ok {
		t.Fatalf("table: want map[string]any got %T", LuaValueToGo(tbl))
	}
	if got["key"] != "val" {
		t.Errorf("table[key]: want val got %v", got["key"])
	}
	if got["num"] != float64(7) {
		t.Errorf("table[num]: want 7.0 got %v", got["num"])
	}
}

func TestLuaValueToGo_Default_Function(t *testing.T) {
	L := newL(t)
	fn := L.NewFunction(func(L *lua.LState) int { return 0 })
	got := LuaValueToGo(fn)
	if got == nil {
		t.Errorf("function: want non-nil string fallback got nil")
	}
	if _, ok := got.(string); !ok {
		t.Errorf("function: want string fallback got %T", got)
	}
}

// LuaTableToMap

func TestLuaTableToMap_Empty(t *testing.T) {
	L := newL(t)
	got, ok := LuaTableToMap(L.NewTable()).([]any)
	if !ok {
		t.Errorf("empty table: want slice got %v", got)
	}
	if len(got) != 0 {
		t.Errorf("empty table: want empty map got %v", got)
	}
}

func TestLuaTableToMap_StringKeys(t *testing.T) {
	L := newL(t)
	tbl := L.NewTable()
	tbl.RawSetString("a", lua.LString("1"))
	tbl.RawSetString("b", lua.LNumber(2))
	tbl.RawSetString("c", lua.LBool(true))

	got, ok := LuaTableToMap(tbl).(map[string]any)
	if !ok {
		t.Errorf("table: want map got %v", got)
	}
	if got["a"] != "1" {
		t.Errorf("a: want '1' got %v", got["a"])
	}
	if got["b"] != float64(2) {
		t.Errorf("b: want 2.0 got %v", got["b"])
	}
	if got["c"] != true {
		t.Errorf("c: want true got %v", got["c"])
	}
}

func TestLuaTableToMap_IntKeys(t *testing.T) {
	L := newL(t)
	tbl := L.NewTable()
	tbl.RawSetInt(1, lua.LString("first"))
	tbl.RawSetInt(2, lua.LString("second"))

	got, ok := LuaTableToMap(tbl).([]any)
	if !ok {
		t.Errorf("table: want slice got %v", got)
	}
	if got[0] != "first" {
		t.Errorf("key 0: want 'first' got %v", got[0])
	}
	if got[1] != "second" {
		t.Errorf("key 1: want 'second' got %v", got[1])
	}
}

func TestLuaTableToMap_IntKeysNonOrder(t *testing.T) {
	L := newL(t)
	tbl := L.NewTable()
	tbl.RawSetInt(1, lua.LString("first"))
	tbl.RawSetInt(2, lua.LString("second"))
	tbl.RawSetInt(4, lua.LString("third"))

	got, ok := LuaTableToMap(tbl).(map[string]any)
	if !ok {
		t.Errorf("table: want map got %v", got)
	}
	if got["1"] != "first" {
		t.Errorf("key 1: want 'first' got %v", got["1"])
	}
	if got["2"] != "second" {
		t.Errorf("key 2: want 'second' got %v", got["2"])
	}
	if got["4"] != "third" {
		t.Errorf("key 4: want 'second' got %v", got["4"])
	}
}

func TestLuaTableToMap_NestedTable(t *testing.T) {
	L := newL(t)
	inner := L.NewTable()
	inner.RawSetString("x", lua.LNumber(99))
	outer := L.NewTable()
	outer.RawSetString("inner", inner)

	got, ok := LuaTableToMap(outer).(map[string]any)
	if !ok {
		t.Errorf("table: want map got %v", got)
	}
	nested, ok := got["inner"].(map[string]any)
	if !ok {
		t.Fatalf("inner: want map[string]any got %T", got["inner"])
	}
	if nested["x"] != float64(99) {
		t.Errorf("inner[x]: want 99.0 got %v", nested["x"])
	}
}

func TestLuaTableToMap_NestedTableInSlice(t *testing.T) {
	L := newL(t)
	inner := L.NewTable()
	inner.RawSetString("x", lua.LNumber(99))
	outer := L.NewTable()
	outer.RawSetInt(1, inner)
	outer.RawSetInt(2, inner)

	got, ok := LuaTableToMap(outer).([]any)
	if !ok {
		t.Errorf("table: want slice got %v", got)
	}
	nested, ok := got[0].(map[string]any)
	if !ok {
		t.Fatalf("inner: want map[string]any got %T", got[0])
	}
	if nested["x"] != float64(99) {
		t.Errorf("inner[x]: want 99.0 got %v", nested["x"])
	}

	nested, ok = got[1].(map[string]any)
	if !ok {
		t.Fatalf("inner: want map[string]any got %T", got[1])
	}
	if nested["x"] != float64(99) {
		t.Errorf("inner[x]: want 99.0 got %v", nested["x"])
	}
}

func TestLuaTableToMap_NestedSliceInTable(t *testing.T) {
	L := newL(t)
	inner := L.NewTable()
	inner.RawSetInt(1, lua.LNumber(99))
	inner.RawSetInt(2, lua.LNumber(99.99))
	outer := L.NewTable()
	outer.RawSetString("inner", inner)

	got, ok := LuaTableToMap(outer).(map[string]any)
	if !ok {
		t.Errorf("table: want map got %v", got)
	}
	nested, ok := got["inner"].([]any)
	if !ok {
		t.Fatalf("inner: want []any got %T", got["inner"])
	}
	if nested[0] != float64(99) {
		t.Errorf("inner[0]: want 99.0 got %v", nested[0])
	}
	if nested[1] != float64(99.99) {
		t.Errorf("inner[1]: want 99.99 got %v", nested[1])
	}
}

// GoValueToLua

func TestGoValueToLua_String(t *testing.T) {
	L := newL(t)
	got := GoValueToLua(L, "hello")
	if got != lua.LString("hello") {
		t.Errorf("string: want LString(hello) got %v", got)
	}
}

func TestGoValueToLua_Float64(t *testing.T) {
	L := newL(t)
	got := GoValueToLua(L, float64(3.14))
	if got != lua.LNumber(3.14) {
		t.Errorf("float64: want LNumber(3.14) got %v", got)
	}
}

func TestGoValueToLua_Int64(t *testing.T) {
	L := newL(t)
	got := GoValueToLua(L, int64(100))
	if got != lua.LNumber(100) {
		t.Errorf("int64: want LNumber(100) got %v", got)
	}
}

func TestGoValueToLua_BoolTrue(t *testing.T) {
	L := newL(t)
	if GoValueToLua(L, true) != lua.LTrue {
		t.Errorf("bool true: want LTrue")
	}
}

func TestGoValueToLua_BoolFalse(t *testing.T) {
	L := newL(t)
	if GoValueToLua(L, false) != lua.LFalse {
		t.Errorf("bool false: want LFalse")
	}
}

func TestGoValueToLua_Nil_Default(t *testing.T) {
	L := newL(t)
	if GoValueToLua(L, nil) != lua.LNil {
		t.Errorf("nil: want LNil")
	}
}

func TestGoValueToLua_UnknownType_ReturnsString(t *testing.T) {
	L := newL(t)
	type custom struct{ x int }
	if got := GoValueToLua(L, custom{x: 1}); got != lua.LString("{1}") {
		t.Errorf("unknown type: want LNil got %v", got)
	}
}

func TestGoValueToLua_Map(t *testing.T) {
	L := newL(t)
	m := map[string]any{
		"name": "alice",
		"age":  float64(30),
	}
	got := GoValueToLua(L, m)
	tbl, ok := got.(*lua.LTable)
	if !ok {
		t.Fatalf("map: want *lua.LTable got %T", got)
	}
	if tbl.RawGetString("name") != lua.LString("alice") {
		t.Errorf("name: want alice got %v", tbl.RawGetString("name"))
	}
	if tbl.RawGetString("age") != lua.LNumber(30) {
		t.Errorf("age: want 30 got %v", tbl.RawGetString("age"))
	}
}

func TestGoValueToLua_Slice(t *testing.T) {
	L := newL(t)
	s := []any{"a", "b", "c"}
	got := GoValueToLua(L, s)
	tbl, ok := got.(*lua.LTable)
	if !ok {
		t.Fatalf("slice: want *lua.LTable got %T", got)
	}
	if tbl.RawGetInt(1) != lua.LString("a") {
		t.Errorf("slice[1]: want a got %v", tbl.RawGetInt(1))
	}
	if tbl.RawGetInt(2) != lua.LString("b") {
		t.Errorf("slice[2]: want b got %v", tbl.RawGetInt(2))
	}
	if tbl.RawGetInt(3) != lua.LString("c") {
		t.Errorf("slice[3]: want c got %v", tbl.RawGetInt(3))
	}
}

func TestGoValueToLua_Slice_Empty(t *testing.T) {
	L := newL(t)
	got := GoValueToLua(L, []any{})
	tbl, ok := got.(*lua.LTable)
	if !ok {
		t.Fatalf("empty slice: want *lua.LTable got %T", got)
	}
	count := 0
	tbl.ForEach(func(_, _ lua.LValue) { count++ })
	if count != 0 {
		t.Errorf("empty slice: want empty table got %d entries", count)
	}
}

func TestGoValueToLua_NestedMap(t *testing.T) {
	L := newL(t)
	m := map[string]any{
		"inner": map[string]any{
			"x": float64(42),
		},
	}
	got := GoValueToLua(L, m)
	outer := got.(*lua.LTable)
	inner, ok := outer.RawGetString("inner").(*lua.LTable)
	if !ok {
		t.Fatalf("inner: want *lua.LTable got %T", outer.RawGetString("inner"))
	}
	if inner.RawGetString("x") != lua.LNumber(42) {
		t.Errorf("inner[x]: want 42 got %v", inner.RawGetString("x"))
	}
}

func TestGoValueToLua_SliceOfMaps(t *testing.T) {
	L := newL(t)
	s := []any{
		map[string]any{"id": float64(1)},
		map[string]any{"id": float64(2)},
	}
	got := GoValueToLua(L, s)
	tbl := got.(*lua.LTable)
	first, ok := tbl.RawGetInt(1).(*lua.LTable)
	if !ok {
		t.Fatalf("slice[1]: want *lua.LTable got %T", tbl.RawGetInt(1))
	}
	if first.RawGetString("id") != lua.LNumber(1) {
		t.Errorf("slice[1][id]: want 1 got %v", first.RawGetString("id"))
	}
}

// Roundtrip.

func TestRoundtrip_GoToLuaToGo(t *testing.T) {
	L := newL(t)
	original := map[string]any{
		"name":   "bob",
		"score":  float64(99),
		"active": true,
	}

	lval := GoValueToLua(L, original)
	tbl, ok := lval.(*lua.LTable)
	if !ok {
		t.Fatalf("roundtrip: want *lua.LTable")
	}
	got := LuaTableToMap(tbl).(map[string]any)

	if got["name"] != "bob" {
		t.Errorf("name: want bob got %v", got["name"])
	}
	if got["score"] != float64(99) {
		t.Errorf("score: want 99.0 got %v", got["score"])
	}
	if got["active"] != true {
		t.Errorf("active: want true got %v", got["active"])
	}
}
