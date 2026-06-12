package pkg

import (
	"encoding/json"

	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/helper"
)

func RegisterJSON(L *lua.LState) error {
	tbl := L.NewTable()

	L.SetField(tbl, "encode", L.NewFunction(func(L *lua.LState) int {
		v := L.CheckAny(1)
		goVal := helper.LuaValueToGo(v)

		b, err := json.Marshal(goVal)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LString(b))
		L.Push(lua.LNil)
		return 2
	}))

	L.SetField(tbl, "decode", L.NewFunction(func(L *lua.LState) int {
		s := L.CheckString(1)
		var v any

		err := json.Unmarshal([]byte(s), &v)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(helper.GoValueToLua(L, v))
		L.Push(lua.LNil)
		return 2
	}))

	L.SetGlobal("json", tbl)

	return nil
}
