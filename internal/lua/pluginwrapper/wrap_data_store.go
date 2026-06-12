package pluginwrapper

import (
	"io"

	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/pluginapi"
)

func RegisterDataStore(L *lua.LState, store pluginapi.DataStore) {
	tbl := L.NewTable()

	L.SetField(tbl, "exists", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LBool(store.Exists(L.CheckString(1))))
		return 1
	}))

	L.SetField(tbl, "read", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		rc, err := store.Read(name)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		defer rc.Close()

		b, err := io.ReadAll(rc)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LString(b))
		L.Push(lua.LNil)
		return 2
	}))

	L.SetField(tbl, "list", L.NewFunction(func(L *lua.LState) int {
		path := L.CheckString(1)
		var exts []string
		if extTbl := L.OptTable(2, nil); extTbl != nil {
			extTbl.ForEach(func(_, v lua.LValue) {
				if s, ok := v.(lua.LString); ok {
					exts = append(exts, string(s))
				}
			})
		}

		names, err := store.List(path, exts)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		result := L.NewTable()
		for i, name := range names {
			result.RawSetInt(i+1, lua.LString(name))
		}

		L.Push(result)
		L.Push(lua.LNil)
		return 2
	}))

	L.SetGlobal("data", tbl)
}
