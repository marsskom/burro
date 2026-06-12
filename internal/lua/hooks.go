package lua

import (
	"errors"
	"fmt"

	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/export"
	"gitlab.com/marsskom/burro/internal/lua/mapper"
	"gitlab.com/marsskom/burro/internal/lua/patch"
	"gitlab.com/marsskom/burro/internal/model"
)

func (w *PluginWrapper) onHook(
	script string,
	hook string,
	coll *mapper.LuaMapperCollection,
	ctx *model.RequestContext,
) error {
	filename, ok := eventScriptMap[hook]
	if !ok {
		w.runtime.Log().Warn("non-existed hook", "hook", hook, "dir", w.dir)

		return nil
	}

	p, err := w.runScript(script, filename, coll)
	if err != nil {
		if errors.Is(err, ErrLuaScriptCallNotFound) {
			w.runtime.Log().Warn("lua call for hook not found", "script", script, "hook", hook, "err", err)

			return nil
		}

		return err
	}

	if ctx == nil {
		return nil
	}

	if err := patch.ApplyCtxPatch(ctx, p.Ctx); err != nil {
		return fmt.Errorf("cannot apply ctx patch: %w", err)
	}

	if ctx.Request == nil {
		return nil
	}

	if err := patch.ApplyRequestPatch(ctx.Request, p.Req); err != nil {
		return fmt.Errorf("cannot apply request patch: %w", err)
	}

	// Lua plugins get req and resp data from snapshots so updates them.
	if p.Req != nil {
		s, err := model.MakeRequestSnapshot(ctx.Request)
		if err != nil {
			return fmt.Errorf("cannot make request snapshot: %w", err)
		}

		ctx.SetRequestSnapshot(s)
	}

	if ctx.Response == nil {
		return nil
	}

	if err := patch.ApplyResponsePatch(ctx.Response, p.Resp); err != nil {
		return fmt.Errorf("cannot apply response patch: %w", err)
	}

	// Lua plugins get req and resp data from snapshots so updates them.
	if p.Resp != nil {
		s, err := model.MakeResponseSnapshot(ctx.Response, nil)
		if err != nil {
			return fmt.Errorf("cannot make response snapshot: %w", err)
		}

		ctx.SetResponse(ctx.Response, s)
	}

	return nil
}

func (w *PluginWrapper) OnConnect(script string, ctx *model.RequestContext) error {
	mc := mapper.NewLuaMapperCollection()
	mc.Add([]*mapper.LuaMapper{
		mapper.NewLuaMapper("ctx", func(L *lua.LState) (*lua.LTable, error) {
			return mapper.CtxInfoLua(L, ctx)
		}),
	})

	return w.onHook(script, "connect", mc, ctx)
}

func (w *PluginWrapper) OnBeforeRequestSend(script string, ctx *model.RequestContext) error {
	mc := mapper.NewLuaMapperCollection()
	mc.Add([]*mapper.LuaMapper{
		mapper.NewLuaMapper("ctx", func(L *lua.LState) (*lua.LTable, error) {
			return mapper.CtxInfoLua(L, ctx)
		}),
	})

	return w.onHook(script, "before_request", mc, ctx)
}

func (w *PluginWrapper) OnAfterRequestSend(script string, ctx *model.RequestContext) error {
	mc := mapper.NewLuaMapperCollection()
	mc.Add([]*mapper.LuaMapper{
		mapper.NewLuaMapper("ctx", func(L *lua.LState) (*lua.LTable, error) {
			return mapper.CtxInfoLua(L, ctx)
		}),
	})

	return w.onHook(script, "after_request", mc, ctx)
}

func (w *PluginWrapper) OnBeforeResponseSend(script string, ctx *model.RequestContext) error {
	mc := mapper.NewLuaMapperCollection()
	mc.Add([]*mapper.LuaMapper{
		mapper.NewLuaMapper("ctx", func(L *lua.LState) (*lua.LTable, error) {
			return mapper.CtxInfoLua(L, ctx)
		}),
	})

	return w.onHook(script, "before_response", mc, ctx)
}

func (w *PluginWrapper) OnAfterResponseSend(script string, ctx *model.RequestContext) error {
	mc := mapper.NewLuaMapperCollection()
	mc.Add([]*mapper.LuaMapper{
		mapper.NewLuaMapper("ctx", func(L *lua.LState) (*lua.LTable, error) {
			return mapper.CtxInfoLua(L, ctx)
		}),
	})

	return w.onHook(script, "after_response", mc, ctx)
}

func (w *PluginWrapper) OnError(script string, ctx *model.RequestContext, err error) error {
	mc := mapper.NewLuaMapperCollection()
	mc.Add([]*mapper.LuaMapper{
		mapper.NewLuaMapper("ctx", func(L *lua.LState) (*lua.LTable, error) {
			return mapper.CtxInfoLua(L, ctx)
		}),
		mapper.NewLuaMapper("err", func(L *lua.LState) (*lua.LTable, error) {
			return mapper.ErrInfoLua(L, err)
		}),
	})

	return w.onHook(script, "error", mc, ctx)
}

func (w *PluginWrapper) OnClose(script string, ctx *model.RequestContext) error {
	mc := mapper.NewLuaMapperCollection()
	mc.Add([]*mapper.LuaMapper{
		mapper.NewLuaMapper("ctx", func(L *lua.LState) (*lua.LTable, error) {
			return mapper.CtxInfoLua(L, ctx)
		}),
	})

	return w.onHook(script, "close", mc, ctx)
}

func (w *PluginWrapper) Flush(script string, opts *export.FileNameVars) error {
	mc := mapper.NewLuaMapperCollection()
	mc.Add([]*mapper.LuaMapper{
		mapper.NewLuaMapper("opts", func(L *lua.LState) (*lua.LTable, error) {
			return mapper.OptsIntoLua(L, opts)
		}),
	})

	return w.onHook(script, "flush", mc, nil)
}

// WS.

func (w *PluginWrapper) OnWSOpen(script string, ctx *model.RequestContext) error {
	mc := mapper.NewLuaMapperCollection()
	mc.Add([]*mapper.LuaMapper{
		mapper.NewLuaMapper("ctx", func(L *lua.LState) (*lua.LTable, error) {
			return mapper.CtxInfoLua(L, ctx)
		}),
	})

	return w.onHook(script, "ws_open", mc, ctx)
}

func (w *PluginWrapper) OnWSMessage(script string, ctx *model.RequestContext, msg *model.WSMessage) error {
	mc := mapper.NewLuaMapperCollection()
	mc.Add([]*mapper.LuaMapper{
		mapper.NewLuaMapper("ctx", func(L *lua.LState) (*lua.LTable, error) {
			return mapper.CtxInfoLua(L, ctx)
		}),
		mapper.NewLuaMapper("ws", func(L *lua.LState) (*lua.LTable, error) {
			return mapper.WSMsgToLua(L, msg)
		}),
	})

	return w.onHook(script, "ws_msg", mc, nil)
}

func (w *PluginWrapper) OnWSClose(script string, ctx *model.RequestContext) error {
	mc := mapper.NewLuaMapperCollection()
	mc.Add([]*mapper.LuaMapper{
		mapper.NewLuaMapper("ctx", func(L *lua.LState) (*lua.LTable, error) {
			return mapper.CtxInfoLua(L, ctx)
		}),
	})

	return w.onHook(script, "ws_close", mc, ctx)
}
