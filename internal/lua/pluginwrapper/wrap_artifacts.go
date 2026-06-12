package pluginwrapper

import (
	"io"
	"strings"

	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/pluginapi"
)

func RegisterArtifactsStore(L *lua.LState, store pluginapi.ArtifactStore) {
	tbl := L.NewTable()

	L.SetField(tbl, "write", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		content := L.CheckString(2)

		_, err := store.Write(name, strings.NewReader(content))
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		L.Push(lua.LNil)
		return 2
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

	L.SetField(tbl, "exists", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LBool(store.Exists(L.CheckString(1))))
		return 1
	}))

	L.SetField(tbl, "delete", L.NewFunction(func(L *lua.LState) int {
		err := store.Delete(L.CheckString(1))
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		L.Push(lua.LNil)
		return 2
	}))

	L.SetField(tbl, "rename", L.NewFunction(func(L *lua.LState) int {
		err := store.Rename(L.CheckString(1), L.CheckString(2))
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
		names, err := store.List()
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		result := L.NewTable()
		for i, name := range names {
			result.RawSetInt(i+1, lua.LString(name)) // 1-indexed like lua arrays
		}

		L.Push(result)
		L.Push(lua.LNil)
		return 2
	}))

	L.SetGlobal("artifacts", tbl)
}
