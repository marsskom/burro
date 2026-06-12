package mapper

import lua "github.com/yuin/gopher-lua"

func ErrInfoLua(L *lua.LState, err error) (*lua.LTable, error) {
	tbl := L.NewTable()
	L.SetField(tbl, "msg", lua.LString(err.Error()))

	return tbl, nil
}
