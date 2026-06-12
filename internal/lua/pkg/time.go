package pkg

import (
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func RegisterTime(L *lua.LState) error {
	tbl := L.NewTable()

	L.SetField(tbl, "unix", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(time.Now().Unix()))

		return 1
	}))

	L.SetField(tbl, "rfc3339", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(
			time.Now().Format(time.RFC3339),
		))

		return 1
	}))

	L.SetField(tbl, "date", L.NewFunction(func(L *lua.LState) int {
		layout := luaLayoutToGo(
			L.OptString(1, "%Y_%m_%d_%H_%M_%S"),
		)

		L.Push(lua.LString(
			time.Now().Format(layout),
		))

		return 1
	}))

	L.SetGlobal("time", tbl)

	return nil
}

var luaDateReplacer = strings.NewReplacer(
	"%Y", "2006",
	"%y", "06",

	"%m", "01",
	"%d", "02",

	"%H", "15",
	"%M", "04",
	"%S", "05",

	"%F", "2006-01-02",
	"%T", "15:04:05",
)

func luaLayoutToGo(layout string) string {
	if layout == "" {
		return "2006_01_02_15_04_05"
	}

	return luaDateReplacer.Replace(layout)
}
