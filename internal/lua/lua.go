package lua

import (
	lua "github.com/yuin/gopher-lua"
)

func newSandboxState() *lua.LState {
	L := lua.NewState(lua.Options{SkipOpenLibs: true})

	// Whitelist.
	for _, pair := range []struct {
		name string
		fn   lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage},
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.StringLibName, lua.OpenString},
		{lua.MathLibName, lua.OpenMath},
	} {
		L.Push(L.NewFunction(pair.fn))
		L.Push(lua.LString(pair.name))
		L.Call(1, 0)
	}

	return L
}
