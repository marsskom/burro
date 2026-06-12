package mapper

import (
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/model"
)

func newL(t *testing.T) *lua.LState {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })
	return L
}

func getField(t *testing.T, tbl *lua.LTable, key string) lua.LValue {
	t.Helper()
	v := tbl.RawGetString(key)
	if v == lua.LNil {
		t.Errorf("field %q is nil", key)
	}
	return v
}

func getTable(t *testing.T, tbl *lua.LTable, key string) *lua.LTable {
	t.Helper()
	v := tbl.RawGetString(key)
	sub, ok := v.(*lua.LTable)
	if !ok {
		t.Fatalf("field %q: want table got %T", key, v)
	}
	return sub
}

func getString(t *testing.T, tbl *lua.LTable, key string) string {
	t.Helper()
	return getField(t, tbl, key).String()
}

func getNumber(t *testing.T, tbl *lua.LTable, key string) float64 {
	t.Helper()
	v := getField(t, tbl, key)
	n, ok := v.(lua.LNumber)
	if !ok {
		t.Fatalf("field %q: want number got %T", key, v)
	}
	return float64(n)
}

func getBool(t *testing.T, tbl *lua.LTable, key string) bool {
	t.Helper()
	v := getField(t, tbl, key)
	b, ok := v.(lua.LBool)
	if !ok {
		t.Fatalf("field %q: want bool got %T", key, v)
	}
	return bool(b)
}

// nil snapshots.

func TestCtxInfoLua_NilSnapshots(t *testing.T) {
	L := newL(t)
	ctx := &model.RequestContext{}

	tbl, err := CtxInfoLua(L, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tbl.RawGetString("id") != lua.LNil {
		t.Errorf("tbl: want nil on id got %v", tbl.RawGetString("id"))
	}
}

func TestCtxInfoLua_NilSnapshotsOnHTTP(t *testing.T) {
	L := newL(t)
	ctx := &model.RequestContext{
		ID: "id",
		Session: &model.Session{
			ID: "session_id",
		},
		IsFinished: false,
	}

	tbl, err := CtxInfoLua(L, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tbl.RawGetString("req") != lua.LNil {
		t.Errorf("req: want nil got %v", tbl.RawGetString("req"))
	}
	if tbl.RawGetString("resp") != lua.LNil {
		t.Errorf("resp: want nil got %v", tbl.RawGetString("resp"))
	}
}

// Request snapshot.

func TestCtxInfoLua_RequestSnapshot_BasicFields(t *testing.T) {
	L := newL(t)
	ctx := &model.RequestContext{
		ID: "id",
		Session: &model.Session{
			ID: "session_id",
		},
		IsFinished: false,
		RequestSnapshot: &model.RequestSnapshot{
			Proto:         "HTTP/1.1",
			Host:          "example.com",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/foo/bar",
			URL:           "https://example.com/foo/bar",
			RemoteAddr:    "127.0.0.1:9999",
			ContentLength: 0,
		},
	}

	tbl, err := CtxInfoLua(L, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := getTable(t, tbl, "req")
	if got := getString(t, req, "proto"); got != "HTTP/1.1" {
		t.Errorf("proto: want HTTP/1.1 got %s", got)
	}
	if got := getString(t, req, "host"); got != "example.com" {
		t.Errorf("host: want example.com got %s", got)
	}
	if got := getString(t, req, "scheme"); got != "https" {
		t.Errorf("scheme: want https got %s", got)
	}
	if got := getString(t, req, "method"); got != "GET" {
		t.Errorf("method: want GET got %s", got)
	}
	if got := getString(t, req, "path"); got != "/foo/bar" {
		t.Errorf("path: want /foo/bar got %s", got)
	}
	if got := getString(t, req, "url"); got != "https://example.com/foo/bar" {
		t.Errorf("url: want https://example.com/foo/bar got %s", got)
	}
	if got := getString(t, req, "remote_addr"); got != "127.0.0.1:9999" {
		t.Errorf("remote_addr: want 127.0.0.1:9999 got %s", got)
	}
}

func TestCtxInfoLua_RequestSnapshot_ContentLength(t *testing.T) {
	L := newL(t)
	ctx := &model.RequestContext{
		ID: "id",
		Session: &model.Session{
			ID: "session_id",
		},
		IsFinished: false,
		RequestSnapshot: &model.RequestSnapshot{
			Body:          []byte("hello"),
			ContentLength: 5,
		},
	}

	tbl, err := CtxInfoLua(L, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := getTable(t, tbl, "req")
	if got := getNumber(t, req, "content_length"); got != 5 {
		t.Errorf("content_length: want 5 got %v", got)
	}
}

func TestCtxInfoLua_RequestSnapshot_Body_Set(t *testing.T) {
	L := newL(t)
	ctx := &model.RequestContext{
		ID: "id",
		Session: &model.Session{
			ID: "session_id",
		},
		IsFinished: false,
		RequestSnapshot: &model.RequestSnapshot{
			Body: []byte("request body"),
		},
	}

	tbl, err := CtxInfoLua(L, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := getTable(t, tbl, "req")
	if got := getString(t, req, "body"); got != "request body" {
		t.Errorf("body: want 'request body' got %q", got)
	}
}

func TestCtxInfoLua_RequestSnapshot_Body_Empty(t *testing.T) {
	L := newL(t)
	ctx := &model.RequestContext{
		ID: "id",
		Session: &model.Session{
			ID: "session_id",
		},
		IsFinished: false,
		RequestSnapshot: &model.RequestSnapshot{
			Body: []byte{},
		},
	}

	tbl, err := CtxInfoLua(L, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := getTable(t, tbl, "req")
	if req.RawGetString("body") != lua.LNil {
		t.Errorf("body: want nil for empty body")
	}
}

func TestCtxInfoLua_RequestSnapshot_Headers(t *testing.T) {
	L := newL(t)
	ctx := &model.RequestContext{
		ID: "id",
		Session: &model.Session{
			ID: "session_id",
		},
		IsFinished: false,
		RequestSnapshot: &model.RequestSnapshot{
			Headers: map[string][]string{
				"Content-Type": {"application/json"},
				"X-Custom":     {"val1", "val2"},
			},
		},
	}

	tbl, err := CtxInfoLua(L, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := getTable(t, tbl, "req")
	headers := getTable(t, req, "headers")

	ct := headers.RawGetString("Content-Type").(*lua.LTable)
	if ct.RawGetInt(1).String() != "application/json" {
		t.Errorf("Content-Type[1]: want application/json got %s", ct.RawGetInt(1))
	}

	xc := headers.RawGetString("X-Custom").(*lua.LTable)
	if xc.RawGetInt(1).String() != "val1" {
		t.Errorf("X-Custom[1]: want val1 got %s", xc.RawGetInt(1))
	}
	if xc.RawGetInt(2).String() != "val2" {
		t.Errorf("X-Custom[2]: want val2 got %s", xc.RawGetInt(2))
	}
}

func TestCtxInfoLua_RequestSnapshot_QueryParams(t *testing.T) {
	L := newL(t)
	ctx := &model.RequestContext{
		ID: "id",
		Session: &model.Session{
			ID: "session_id",
		},
		IsFinished: false,
		RequestSnapshot: &model.RequestSnapshot{
			QueryParams: map[string][]string{
				"page": {"1"},
				"q":    {"foo", "bar"},
			},
		},
	}

	tbl, err := CtxInfoLua(L, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := getTable(t, tbl, "req")
	query := getTable(t, req, "query")

	page := query.RawGetString("page").(*lua.LTable)
	if page.RawGetInt(1).String() != "1" {
		t.Errorf("page[1]: want 1 got %s", page.RawGetInt(1))
	}

	q := query.RawGetString("q").(*lua.LTable)
	if q.RawGetInt(1).String() != "foo" {
		t.Errorf("q[1]: want foo got %s", q.RawGetInt(1))
	}
	if q.RawGetInt(2).String() != "bar" {
		t.Errorf("q[2]: want bar got %s", q.RawGetInt(2))
	}
}

func TestCtxInfoLua_RequestSnapshot_Cookies(t *testing.T) {
	expires := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	L := newL(t)
	ctx := &model.RequestContext{
		ID: "id",
		Session: &model.Session{
			ID: "session_id",
		},
		IsFinished: false,
		RequestSnapshot: &model.RequestSnapshot{
			Cookies: []*model.CookieSnapshot{
				{
					Name:        "session",
					Value:       "abc123",
					Quoted:      true,
					Path:        "/",
					Domain:      "example.com",
					Expires:     expires,
					MaxAge:      3600,
					Secure:      true,
					HTTPOnly:    true,
					SameSite:    2,
					Partitioned: false,
				},
			},
		},
	}

	tbl, err := CtxInfoLua(L, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := getTable(t, tbl, "req")
	cookies := getTable(t, req, "cookies")
	c := cookies.RawGetInt(1).(*lua.LTable)

	if getString(t, c, "name") != "session" {
		t.Errorf("name: want session got %s", getString(t, c, "name"))
	}
	if getString(t, c, "value") != "abc123" {
		t.Errorf("value: want abc123 got %s", getString(t, c, "value"))
	}
	if !getBool(t, c, "quoted") {
		t.Errorf("quoted: want true")
	}
	if getString(t, c, "path") != "/" {
		t.Errorf("path: want / got %s", getString(t, c, "path"))
	}
	if getString(t, c, "domain") != "example.com" {
		t.Errorf("domain: want example.com got %s", getString(t, c, "domain"))
	}
	if got := getNumber(t, c, "expires"); got != float64(expires.UnixMilli()) {
		t.Errorf("expires: want %v got %v", expires.UnixMilli(), got)
	}
	if got := getNumber(t, c, "max_age"); got != 3600 {
		t.Errorf("max_age: want 3600 got %v", got)
	}
	if !getBool(t, c, "secure") {
		t.Errorf("secure: want true")
	}
	if !getBool(t, c, "http_only") {
		t.Errorf("http_only: want true")
	}
	if got := getNumber(t, c, "same_site"); got != 2 {
		t.Errorf("same_site: want 2 got %v", got)
	}
	if getBool(t, c, "partitioned") {
		t.Errorf("partitioned: want false")
	}
}

func TestCtxInfoLua_RequestSnapshot_Cookies_Empty(t *testing.T) {
	L := newL(t)
	ctx := &model.RequestContext{
		ID: "id",
		Session: &model.Session{
			ID: "session_id",
		},
		IsFinished: false,
		RequestSnapshot: &model.RequestSnapshot{
			Cookies: []*model.CookieSnapshot{},
		},
	}

	tbl, err := CtxInfoLua(L, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := getTable(t, tbl, "req")
	if req.RawGetString("cookies") != lua.LNil {
		t.Errorf("cookies: want nil for empty slice")
	}
}

func TestCtxInfoLua_RequestSnapshot_MultipleCookies(t *testing.T) {
	L := newL(t)
	ctx := &model.RequestContext{
		ID: "id",
		Session: &model.Session{
			ID: "session_id",
		},
		IsFinished: false,
		RequestSnapshot: &model.RequestSnapshot{
			Cookies: []*model.CookieSnapshot{
				{Name: "a", Value: "1"},
				{Name: "b", Value: "2"},
				{Name: "c", Value: "3"},
			},
		},
	}

	tbl, err := CtxInfoLua(L, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := getTable(t, tbl, "req")
	cookies := getTable(t, req, "cookies")

	count := 0
	cookies.ForEach(func(_, _ lua.LValue) { count++ })
	if count != 3 {
		t.Errorf("cookies: want 3 got %d", count)
	}

	c2 := cookies.RawGetInt(2).(*lua.LTable)
	if getString(t, c2, "name") != "b" {
		t.Errorf("cookie[2].name: want b got %s", getString(t, c2, "name"))
	}
}

// Response snapshot.

func TestCtxInfoLua_ResponseSnapshot_BasicFields(t *testing.T) {
	L := newL(t)
	ctx := &model.RequestContext{
		ID: "id",
		Session: &model.Session{
			ID: "session_id",
		},
		IsFinished: false,
		ResponseSnapshot: &model.ResponseSnapshot{
			Status:        "200 OK",
			StatusCode:    200,
			Proto:         "HTTP/1.1",
			ContentLength: 0,
		},
	}

	tbl, err := CtxInfoLua(L, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp := getTable(t, tbl, "resp")
	if got := getString(t, resp, "status"); got != "200 OK" {
		t.Errorf("status: want '200 OK' got %s", got)
	}
	if got := getNumber(t, resp, "status_code"); got != 200 {
		t.Errorf("status_code: want 200 got %v", got)
	}
	if got := getString(t, resp, "proto"); got != "HTTP/1.1" {
		t.Errorf("proto: want HTTP/1.1 got %s", got)
	}
}

func TestCtxInfoLua_ResponseSnapshot_Body_Set(t *testing.T) {
	L := newL(t)
	ctx := &model.RequestContext{
		ID: "id",
		Session: &model.Session{
			ID: "session_id",
		},
		IsFinished: false,
		ResponseSnapshot: &model.ResponseSnapshot{
			Body: []byte(`{"ok":true}`),
		},
	}

	tbl, err := CtxInfoLua(L, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp := getTable(t, tbl, "resp")
	if got := getString(t, resp, "body"); got != `{"ok":true}` {
		t.Errorf("body: want '{\"ok\":true}' got %q", got)
	}
}

func TestCtxInfoLua_ResponseSnapshot_Body_Empty(t *testing.T) {
	L := newL(t)
	ctx := &model.RequestContext{
		ID: "id",
		Session: &model.Session{
			ID: "session_id",
		},
		IsFinished: false,
		ResponseSnapshot: &model.ResponseSnapshot{
			Body: []byte{},
		},
	}

	tbl, err := CtxInfoLua(L, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp := getTable(t, tbl, "resp")
	if resp.RawGetString("body") != lua.LNil {
		t.Errorf("body: want nil for empty body")
	}
}

func TestCtxInfoLua_ResponseSnapshot_Headers(t *testing.T) {
	L := newL(t)
	ctx := &model.RequestContext{
		ID: "id",
		Session: &model.Session{
			ID: "session_id",
		},
		IsFinished: false,
		ResponseSnapshot: &model.ResponseSnapshot{
			Headers: map[string][]string{
				"Content-Type": {"text/html"},
			},
		},
	}

	tbl, err := CtxInfoLua(L, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp := getTable(t, tbl, "resp")
	headers := getTable(t, resp, "headers")
	ct := headers.RawGetString("Content-Type").(*lua.LTable)
	if ct.RawGetInt(1).String() != "text/html" {
		t.Errorf("Content-Type[1]: want text/html got %s", ct.RawGetInt(1))
	}
}

func TestCtxInfoLua_ResponseSnapshot_ContentLength(t *testing.T) {
	L := newL(t)
	ctx := &model.RequestContext{
		ID: "id",
		Session: &model.Session{
			ID: "session_id",
		},
		IsFinished: false,
		ResponseSnapshot: &model.ResponseSnapshot{
			Body:          []byte("response"),
			ContentLength: 8,
		},
	}

	tbl, err := CtxInfoLua(L, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp := getTable(t, tbl, "resp")
	if got := getNumber(t, resp, "content_length"); got != 8 {
		t.Errorf("content_length: want 8 got %v", got)
	}
}

// Both snapshots set.

func TestCtxInfoLua_BothSnapshots(t *testing.T) {
	L := newL(t)
	ctx := &model.RequestContext{
		ID: "id",
		Session: &model.Session{
			ID: "session_id",
		},
		IsFinished: false,
		RequestSnapshot: &model.RequestSnapshot{
			Method: "POST",
			URL:    "https://example.com/api",
		},
		ResponseSnapshot: &model.ResponseSnapshot{
			StatusCode: 201,
			Status:     "201 Created",
		},
	}

	tbl, err := CtxInfoLua(L, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := getTable(t, tbl, "req")
	if getString(t, req, "method") != "POST" {
		t.Errorf("req.method: want POST")
	}

	resp := getTable(t, tbl, "resp")
	if getNumber(t, resp, "status_code") != 201 {
		t.Errorf("resp.status_code: want 201")
	}
}
