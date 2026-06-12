package pluginwrapper

import (
	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/helper"
	"gitlab.com/marsskom/burro/internal/pluginapi"
)

func RegisterEventBus(L *lua.LState, bus pluginapi.EventBus) {
	tbl := L.NewTable()

	L.SetField(tbl, "emit", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)

		data := helper.LuaValueToGo(L.Get(2))

		err := bus.Emit(pluginapi.Event{
			Name: name,
			Data: data,
			From: "lua",
		})

		if err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}

		L.Push(lua.LNil)
		return 1
	}))

	L.SetGlobal("bus", tbl)
}
