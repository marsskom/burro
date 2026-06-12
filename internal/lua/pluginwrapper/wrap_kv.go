package pluginwrapper

import (
	"encoding/base64"

	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/pluginapi"
)

func RegisterKeyValueStore(L *lua.LState, kv pluginapi.KeyValueStore) {
	tbl := L.NewTable()

	L.SetField(tbl, "get", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		b, err := kv.Get(key)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LString(b))
		L.Push(lua.LNil)
		return 2
	}))

	L.SetField(tbl, "get_base64", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		b, err := kv.Get(key)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LString(base64.StdEncoding.EncodeToString(b)))
		L.Push(lua.LNil)
		return 2
	}))

	L.SetField(tbl, "set", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value := L.CheckString(2)

		err := kv.Set(key, []byte(value))
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		L.Push(lua.LNil)
		return 2
	}))

	L.SetField(tbl, "delete", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)

		err := kv.Delete(key)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		L.Push(lua.LNil)
		return 2
	}))

	L.SetField(tbl, "list", L.NewFunction(func(L *lua.LState) int {
		prefix := L.CheckString(1)

		m, err := kv.List(prefix)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		tbl := L.NewTable()

		for k, v := range m {
			tbl.RawSetString(k, lua.LString(v))
		}

		L.Push(tbl)
		L.Push(lua.LNil)
		return 2
	}))

	L.SetGlobal("kv", tbl)
}
