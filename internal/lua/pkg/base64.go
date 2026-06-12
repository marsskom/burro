package pkg

import (
	"encoding/base64"

	lua "github.com/yuin/gopher-lua"
)

func RegisterBase64(L *lua.LState) error {
	tbl := L.NewTable()

	L.SetField(tbl, "encode", L.NewFunction(func(L *lua.LState) int {
		s := L.CheckString(1)
		res := base64.StdEncoding.EncodeToString([]byte(s))

		L.Push(lua.LString(res))
		return 1
	}))

	L.SetField(tbl, "decode", L.NewFunction(func(L *lua.LState) int {
		s := L.CheckString(1)
		decode, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LString(decode))
		L.Push(lua.LNil)
		return 2
	}))

	L.SetGlobal("base64", tbl)

	return nil
}
