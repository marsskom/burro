package patch

import (
	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/model"
)

func RegisterCtxPatch(
	L *lua.LState,
	ctxPatch *model.CtxPatch,
	reqPatch *model.RequestPatch,
	respPatch *model.ResponsePatch,
) {
	ctxMut := L.NewTable()
	L.SetField(ctxMut, "set_finish", L.NewFunction(func(L *lua.LState) int {
		v := true
		ctxPatch.IsFinished = &v
		return 0
	}))

	// ctx.req setters
	reqMut := L.NewTable()

	L.SetField(reqMut, "set_host", L.NewFunction(func(L *lua.LState) int {
		v := L.CheckString(1)
		reqPatch.Host = &v
		return 0
	}))
	L.SetField(reqMut, "set_scheme", L.NewFunction(func(L *lua.LState) int {
		v := L.CheckString(1)
		reqPatch.Scheme = &v
		return 0
	}))
	L.SetField(reqMut, "set_method", L.NewFunction(func(L *lua.LState) int {
		v := L.CheckString(1)
		reqPatch.Method = &v
		return 0
	}))
	L.SetField(reqMut, "set_path", L.NewFunction(func(L *lua.LState) int {
		v := L.CheckString(1)
		reqPatch.Path = &v
		return 0
	}))
	L.SetField(reqMut, "set_url", L.NewFunction(func(L *lua.LState) int {
		v := L.CheckString(1)
		reqPatch.URL = &v
		return 0
	}))
	L.SetField(reqMut, "set_body", L.NewFunction(func(L *lua.LState) int {
		v := []byte(L.CheckString(1))
		reqPatch.Body = &v
		return 0
	}))
	L.SetField(reqMut, "set_header", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		val := L.CheckString(2)
		if reqPatch.Headers == nil {
			reqPatch.Headers = make(map[string][]string)
		}
		reqPatch.Headers[key] = []string{val}
		return 0
	}))
	L.SetField(reqMut, "add_header", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		val := L.CheckString(2)
		if reqPatch.Headers == nil {
			reqPatch.Headers = make(map[string][]string)
		}
		reqPatch.Headers[key] = append(reqPatch.Headers[key], val)
		return 0
	}))
	L.SetField(reqMut, "del_header", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		if reqPatch.Headers == nil {
			reqPatch.Headers = make(map[string][]string)
		}
		reqPatch.Headers[key] = nil // nil slice = delete
		return 0
	}))
	L.SetField(reqMut, "set_cookie", L.NewFunction(func(L *lua.LState) int {
		tbl := L.CheckTable(1)

		patch := model.CookiePatch{}

		tbl.ForEach(func(key, val lua.LValue) {
			k := key.String()

			switch k {
			case "name":
				patch.Name = val.String()
			case "value":
				patch.Value = val.String()
			case "path":
				patch.Path = val.String()
			case "domain":
				patch.Domain = val.String()
			case "secure":
				patch.Secure = lua.LVAsBool(val)
			case "http_only":
				patch.HTTPOnly = lua.LVAsBool(val)
			case "max_age":
				patch.MaxAge = int(lua.LVAsNumber(val))
			}
		})

		reqPatch.Cookies = append(reqPatch.Cookies, patch)
		return 0
	}))
	L.SetField(reqMut, "del_cookie", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)

		reqPatch.Cookies = append(reqPatch.Cookies, model.CookiePatch{
			Name:   name,
			Delete: true,
		})

		return 0
	}))

	// ctx.resp setters
	respMut := L.NewTable()

	L.SetField(respMut, "set_status", L.NewFunction(func(L *lua.LState) int {
		v := int(L.CheckNumber(1))
		respPatch.StatusCode = &v
		return 0
	}))
	L.SetField(respMut, "set_body", L.NewFunction(func(L *lua.LState) int {
		v := []byte(L.CheckString(1))
		respPatch.Body = &v
		return 0
	}))
	L.SetField(respMut, "set_header", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		val := L.CheckString(2)
		if respPatch.Headers == nil {
			respPatch.Headers = make(map[string][]string)
		}
		respPatch.Headers[key] = []string{val}
		return 0
	}))
	L.SetField(respMut, "add_header", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		val := L.CheckString(2)
		if respPatch.Headers == nil {
			respPatch.Headers = make(map[string][]string)
		}
		respPatch.Headers[key] = append(respPatch.Headers[key], val)
		return 0
	}))
	L.SetField(respMut, "del_header", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		if respPatch.Headers == nil {
			respPatch.Headers = make(map[string][]string)
		}
		respPatch.Headers[key] = nil // nil slice = delete
		return 0
	}))

	mut := L.NewTable()
	L.SetField(mut, "ctx", ctxMut)
	L.SetField(mut, "req", reqMut)
	L.SetField(mut, "resp", respMut)

	L.SetGlobal("mut", mut)
}
