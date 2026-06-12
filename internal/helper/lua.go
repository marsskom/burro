package helper

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

func LuaValueToGo(v lua.LValue) any {
	switch v.Type() {
	case lua.LTNil:
		return nil

	case lua.LTBool:
		return lua.LVAsBool(v)

	case lua.LTNumber:
		return float64(v.(lua.LNumber))

	case lua.LTString:
		return string(v.(lua.LString))

	case lua.LTTable:
		return LuaTableToMap(v.(*lua.LTable))

	default:
		return v.String()
	}
}

func LuaTableToMap(tbl *lua.LTable) any {
	if isArrayTable(tbl) {
		result := make([]any, tbl.Len())

		for i := 1; i <= tbl.Len(); i++ {
			result[i-1] = LuaValueToGo(tbl.RawGetInt(i))
		}

		return result
	}

	result := make(map[string]any)

	tbl.ForEach(func(key, value lua.LValue) {
		result[key.String()] = LuaValueToGo(value)
	})

	return result
}

func isArrayTable(tbl *lua.LTable) bool {
	count := 0

	tbl.ForEach(func(key, _ lua.LValue) {
		count++

		n, ok := key.(lua.LNumber)
		if !ok {
			count = -1
			return
		}

		i := int(n)

		if float64(i) != float64(n) || i < 1 {
			count = -1
		}
	})

	return count >= 0 && count == tbl.Len()
}

func GoValueToLua(L *lua.LState, v any) lua.LValue {
	switch x := v.(type) {
	case nil:
		return lua.LNil

	case map[string]any:
		t := L.NewTable()

		for k, val := range x {
			t.RawSetString(k, GoValueToLua(L, val))
		}

		return t

	case []any:
		t := L.NewTable()

		for i, val := range x {
			t.RawSetInt(i+1, GoValueToLua(L, val))
		}

		return t

	case string:
		return lua.LString(x)

	case bool:
		return lua.LBool(x)

	case float64:
		return lua.LNumber(x)

	case float32:
		return lua.LNumber(x)

	case int:
		return lua.LNumber(x)

	case int8:
		return lua.LNumber(x)

	case int16:
		return lua.LNumber(x)

	case int32:
		return lua.LNumber(x)

	case int64:
		return lua.LNumber(x)

	case uint:
		return lua.LNumber(x)

	case uint8:
		return lua.LNumber(x)

	case uint16:
		return lua.LNumber(x)

	case uint32:
		return lua.LNumber(x)

	case uint64:
		return lua.LNumber(x)

	default:
		return lua.LString(fmt.Sprint(v))
	}
}
