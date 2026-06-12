package mapper

import (
	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/model"
)

func CtxInfoLua(L *lua.LState, ctx *model.RequestContext) (*lua.LTable, error) {
	tbl := L.NewTable()
	if ctx.ID == "" {
		return tbl, nil
	}

	L.SetField(tbl, "id", lua.LString(ctx.ID))
	L.SetField(tbl, "session_id", lua.LString(ctx.Session.ID))
	L.SetField(tbl, "is_finished", lua.LBool(ctx.IsFinished))

	if ctx.RequestSnapshot != nil {
		req := L.NewTable()
		L.SetField(req, "proto", lua.LString(ctx.RequestSnapshot.Proto))
		L.SetField(req, "host", lua.LString(ctx.RequestSnapshot.Host))
		L.SetField(req, "scheme", lua.LString(ctx.RequestSnapshot.Scheme))
		L.SetField(req, "method", lua.LString(ctx.RequestSnapshot.Method))
		L.SetField(req, "path", lua.LString(ctx.RequestSnapshot.Path))
		L.SetField(req, "query", stringMapToLua(L, ctx.RequestSnapshot.QueryParams))
		L.SetField(req, "url", lua.LString(ctx.RequestSnapshot.URL))
		L.SetField(req, "remote_addr", lua.LString(ctx.RequestSnapshot.RemoteAddr))

		L.SetField(req, "headers", stringMapToLua(L, ctx.RequestSnapshot.Headers))

		if len(ctx.RequestSnapshot.Cookies) == 0 {
			L.SetField(req, "cookies", lua.LNil)
		} else {
			L.SetField(req, "cookies", cookiesToLua(L, ctx.RequestSnapshot.Cookies))
		}

		if len(ctx.RequestSnapshot.Body) == 0 {
			L.SetField(req, "body", lua.LNil)
		} else {
			L.SetField(req, "body", lua.LString(ctx.RequestSnapshot.Body))
		}

		L.SetField(req, "content_length", lua.LNumber(ctx.RequestSnapshot.ContentLength))

		L.SetField(tbl, "req", req)
	} else {
		L.SetField(tbl, "req", lua.LNil)
	}

	if ctx.ResponseSnapshot != nil {
		resp := L.NewTable()
		L.SetField(resp, "status", lua.LString(ctx.ResponseSnapshot.Status))
		L.SetField(resp, "status_code", lua.LNumber(ctx.ResponseSnapshot.StatusCode))
		L.SetField(resp, "proto", lua.LString(ctx.ResponseSnapshot.Proto))

		L.SetField(resp, "headers", stringMapToLua(L, ctx.ResponseSnapshot.Headers))

		if len(ctx.ResponseSnapshot.Body) == 0 {
			L.SetField(resp, "body", lua.LNil)
		} else {
			L.SetField(resp, "body", lua.LString(ctx.ResponseSnapshot.Body))
		}

		L.SetField(resp, "content_length", lua.LNumber(ctx.ResponseSnapshot.ContentLength))

		L.SetField(tbl, "resp", resp)
	} else {
		L.SetField(tbl, "resp", lua.LNil)
	}

	return tbl, nil
}

func stringMapToLua(L *lua.LState, m map[string][]string) *lua.LTable {
	tbl := L.NewTable()

	for k, v := range m {
		arr := L.NewTable()
		for i, val := range v {
			arr.RawSetInt(i+1, lua.LString(val))
		}
		tbl.RawSetString(k, arr)
	}

	return tbl
}

func cookiesToLua(L *lua.LState, cookies []*model.CookieSnapshot) *lua.LTable {
	tbl := L.NewTable()

	for i, c := range cookies {
		cTbl := L.NewTable()
		L.SetField(cTbl, "name", lua.LString(c.Name))
		L.SetField(cTbl, "value", lua.LString(c.Value))
		L.SetField(cTbl, "quoted", lua.LBool(c.Quoted))
		L.SetField(cTbl, "path", lua.LString(c.Path))
		L.SetField(cTbl, "domain", lua.LString(c.Domain))
		L.SetField(cTbl, "expires", lua.LNumber(c.Expires.UnixMilli()))
		L.SetField(cTbl, "max_age", lua.LNumber(c.MaxAge))
		L.SetField(cTbl, "secure", lua.LBool(c.Secure))
		L.SetField(cTbl, "http_only", lua.LBool(c.HTTPOnly))
		L.SetField(cTbl, "same_site", lua.LNumber(c.SameSite))
		L.SetField(cTbl, "partitioned", lua.LBool(c.Partitioned))

		tbl.RawSetInt(i+1, cTbl)
	}
	return tbl
}
