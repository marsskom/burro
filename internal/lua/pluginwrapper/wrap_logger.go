package pluginwrapper

import (
	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/pluginapi"
)

func RegisterLogger(L *lua.LState, log pluginapi.Logger) {
	tbl := L.NewTable()

	var collectArgs = func(t *lua.LTable) []any {
		args := []any{}

		if t != nil {
			t.ForEach(func(k, v lua.LValue) {
				args = append(args, k.String(), v.String())
			})
		}

		return args
	}

	L.SetField(tbl, "trace", L.NewFunction(func(L *lua.LState) int {
		msg := L.CheckString(1)
		args := collectArgs(L.OptTable(2, nil))

		log.Trace(msg, args...)

		return 0
	}))
	L.SetField(tbl, "debug", L.NewFunction(func(L *lua.LState) int {
		msg := L.CheckString(1)
		args := collectArgs(L.OptTable(2, nil))

		log.Debug(msg, args...)

		return 0
	}))
	L.SetField(tbl, "info", L.NewFunction(func(L *lua.LState) int {
		msg := L.CheckString(1)
		args := collectArgs(L.OptTable(2, nil))

		log.Info(msg, args...)

		return 0
	}))
	L.SetField(tbl, "warn", L.NewFunction(func(L *lua.LState) int {
		msg := L.CheckString(1)
		args := collectArgs(L.OptTable(2, nil))

		log.Warn(msg, args...)

		return 0
	}))
	L.SetField(tbl, "error", L.NewFunction(func(L *lua.LState) int {
		msg := L.CheckString(1)
		args := collectArgs(L.OptTable(2, nil))

		log.Error(msg, args...)

		return 0
	}))
	L.SetField(tbl, "audit", L.NewFunction(func(L *lua.LState) int {
		msg := L.CheckString(1)
		args := collectArgs(L.OptTable(2, nil))

		log.Audit(msg, args...)

		return 0
	}))

	L.SetGlobal("log", tbl)
}
