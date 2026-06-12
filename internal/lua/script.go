package lua

import (
	"fmt"
	"strings"

	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/logger"
	"gitlab.com/marsskom/burro/internal/lua/mapper"
	"gitlab.com/marsskom/burro/internal/lua/patch"
	"gitlab.com/marsskom/burro/internal/model"
)

func (w *PluginWrapper) runScript(
	script string,
	filename string,
	coll *mapper.LuaMapperCollection,
) (*model.Patch, error) {
	if _, ok := w.lStates[script]; !ok {
		return nil, fmt.Errorf("lua plugin wrapper: there is no state for the script '%s'", script)
	}

	calls, ok := w.lCalls[script]
	if !ok {
		return nil, fmt.Errorf("lua plugin wrapper: there is no calls for the script '%s'", script)
	}

	fn, ok := calls[filename]
	if !ok {
		return nil, fmt.Errorf("lua plugin wrapper: call '%s' not registered for a script '%s': %w", filename, script, ErrLuaScriptCallNotFound)
	}

	w.runtime.Log().Debug("lua plugin is going to be executed", "script", script, "filename", filename)

	L := w.lStates[script]
	L.SetTop(0)

	err := coll.InitMappers(L)
	if err != nil {
		return nil, fmt.Errorf("lua plugin '%s/%s' mappers errors: %w", script, filename, err)
	}
	defer coll.CloseMappers(L)

	// Adds ctx patch.
	p := model.NewPatch()

	patch.RegisterCtxPatch(L, p.Ctx, p.Req, p.Resp)
	defer func() {
		L.SetGlobal("mut", lua.LNil)
	}()

	// Executes shared.
	if sharedFn, ok := calls["_shared"]; ok {
		if err := L.CallByParam(lua.P{
			Fn:      sharedFn,
			NRet:    0,
			Protect: true,
		}); err != nil {
			return nil, fmt.Errorf("lua plugin '%s/%s' _shared execution error: %w", script, filename, err)
		}
	}

	// Traces lua globals.
	if w.runtime.Log().Enabled(logger.LevelTrace) {
		seen := make(map[*lua.LTable]bool)
		var b strings.Builder

		L.G.Global.ForEach(func(k, v lua.LValue) {
			fmt.Fprintf(&b, "[GLOBAL] %s => ", k.String())
			dumpLuaValue(L, v, 0, &b, seen)
		})
		w.runtime.Log().Trace(b.String())
	}

	// Executes script.
	if err := L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    0,
		Protect: true,
	}); err != nil {
		return nil, fmt.Errorf("lua plugin '%s/%s' execution error: %w : %w", script, filename, ErrLuaScriptExecute, err)
	}

	return p, nil
}

func dumpLuaValue(L *lua.LState, v lua.LValue, indent int, b *strings.Builder, seen map[*lua.LTable]bool) {
	prefix := strings.Repeat("  ", indent)

	switch val := v.(type) {

	case *lua.LTable:
		if seen[val] {
			fmt.Fprintf(b, "%s<cycle>\n", prefix)
			return
		}
		seen[val] = true

		fmt.Fprintf(b, "%s(table)\n", prefix)

		val.ForEach(func(k, v lua.LValue) {
			fmt.Fprintf(b, "%s  [%s] => ", prefix, k.String())
			dumpLuaValue(L, v, indent+1, b, seen)
		})

	case *lua.LFunction:
		fmt.Fprintf(b, "%s(function)\n", prefix)

	case lua.LString:
		fmt.Fprintf(b, "%s%s (string)\n", prefix, string(val))

	case lua.LNumber:
		fmt.Fprintf(b, "%s%v (number)\n", prefix, float64(val))

	case lua.LBool:
		fmt.Fprintf(b, "%s%v (bool)\n", prefix, bool(val))

	case *lua.LNilType:
		fmt.Fprintf(b, "%snil\n", prefix)

	default:
		fmt.Fprintf(b, "%s%s (%s)\n", prefix, val.String(), val.Type().String())
	}
}
