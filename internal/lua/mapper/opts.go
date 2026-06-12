package mapper

import (
	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/export"
)

func OptsIntoLua(L *lua.LState, opts *export.FileNameVars) (*lua.LTable, error) {
	tbl := L.NewTable()
	L.SetField(tbl, "session", lua.LString(opts.Session))

	return tbl, nil
}
