package patch

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/model"
)

func newPatchState(t *testing.T) (*lua.LState, *model.CtxPatch, *model.RequestPatch, *model.ResponsePatch) {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })
	ctxPatch := &model.CtxPatch{}
	reqPatch := &model.RequestPatch{}
	respPatch := &model.ResponsePatch{}
	RegisterCtxPatch(L, ctxPatch, reqPatch, respPatch)
	return L, ctxPatch, reqPatch, respPatch
}

// mut.ctx

func TestCtxPatch_SetFinish(t *testing.T) {
	L, ctx, _, _ := newPatchState(t)
	if err := L.DoString(`mut.ctx.set_finish()`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if ctx.IsFinished == nil {
		t.Fatal("IsFinished: want non-nil")
	}
	if !*ctx.IsFinished {
		t.Errorf("IsFinished: want true got false")
	}
}

func TestCtxPatch_SetFinish_ReturnsNothing(t *testing.T) {
	L, _, _, _ := newPatchState(t)
	if err := L.DoString(`local x = mut.ctx.set_finish(); assert(x == nil)`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
}

// mut.req

func TestReqPatch_SetHost(t *testing.T) {
	L, _, req, _ := newPatchState(t)
	if err := L.DoString(`mut.req.set_host("example.com")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if req.Host == nil || *req.Host != "example.com" {
		t.Errorf("Host: want example.com got %v", req.Host)
	}
}

func TestReqPatch_SetScheme(t *testing.T) {
	L, _, req, _ := newPatchState(t)
	if err := L.DoString(`mut.req.set_scheme("https")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if req.Scheme == nil || *req.Scheme != "https" {
		t.Errorf("Scheme: want https got %v", req.Scheme)
	}
}

func TestReqPatch_SetMethod(t *testing.T) {
	L, _, req, _ := newPatchState(t)
	if err := L.DoString(`mut.req.set_method("POST")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if req.Method == nil || *req.Method != "POST" {
		t.Errorf("Method: want POST got %v", req.Method)
	}
}

func TestReqPatch_SetPath(t *testing.T) {
	L, _, req, _ := newPatchState(t)
	if err := L.DoString(`mut.req.set_path("/new/path")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if req.Path == nil || *req.Path != "/new/path" {
		t.Errorf("Path: want /new/path got %v", req.Path)
	}
}

func TestReqPatch_SetURL(t *testing.T) {
	L, _, req, _ := newPatchState(t)
	if err := L.DoString(`mut.req.set_url("https://new.example.com/path")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if req.URL == nil || *req.URL != "https://new.example.com/path" {
		t.Errorf("URL: want https://new.example.com/path got %v", req.URL)
	}
}

func TestReqPatch_SetBody(t *testing.T) {
	L, _, req, _ := newPatchState(t)
	if err := L.DoString(`mut.req.set_body("request body")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if req.Body == nil {
		t.Fatal("Body: want non-nil")
	}
	if string(*req.Body) != "request body" {
		t.Errorf("Body: want 'request body' got %q", string(*req.Body))
	}
}

func TestReqPatch_SetHeader(t *testing.T) {
	L, _, req, _ := newPatchState(t)
	if err := L.DoString(`mut.req.set_header("X-Custom", "value")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if req.Headers == nil {
		t.Fatal("Headers: want non-nil map")
	}
	vals := req.Headers["X-Custom"]
	if len(vals) != 1 || vals[0] != "value" {
		t.Errorf("X-Custom: want [value] got %v", vals)
	}
}

func TestReqPatch_SetHeader_Overwrites(t *testing.T) {
	L, _, req, _ := newPatchState(t)
	if err := L.DoString(`
		mut.req.set_header("X-Custom", "first")
		mut.req.set_header("X-Custom", "second")
	`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	vals := req.Headers["X-Custom"]
	if len(vals) != 1 || vals[0] != "second" {
		t.Errorf("X-Custom: want [second] got %v", vals)
	}
}

func TestReqPatch_AddHeader(t *testing.T) {
	L, _, req, _ := newPatchState(t)
	if err := L.DoString(`
		mut.req.add_header("X-Multi", "one")
		mut.req.add_header("X-Multi", "two")
	`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	vals := req.Headers["X-Multi"]
	if len(vals) != 2 || vals[0] != "one" || vals[1] != "two" {
		t.Errorf("X-Multi: want [one two] got %v", vals)
	}
}

func TestReqPatch_DelHeader(t *testing.T) {
	L, _, req, _ := newPatchState(t)
	if err := L.DoString(`mut.req.del_header("X-Remove")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if req.Headers == nil {
		t.Fatal("Headers: want non-nil map after del")
	}
	vals, exists := req.Headers["X-Remove"]
	if !exists {
		t.Fatal("X-Remove: want key present with nil value")
	}
	if vals != nil {
		t.Errorf("X-Remove: want nil slice got %v", vals)
	}
}

func TestReqPatch_SetThenDelHeader(t *testing.T) {
	L, _, req, _ := newPatchState(t)
	if err := L.DoString(`
		mut.req.set_header("X-Temp", "value")
		mut.req.del_header("X-Temp")
	`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if req.Headers["X-Temp"] != nil {
		t.Errorf("X-Temp: want nil after del got %v", req.Headers["X-Temp"])
	}
}

func TestReqPatch_UnsetFieldsRemainNil(t *testing.T) {
	_, _, req, _ := newPatchState(t)
	if req.Host != nil {
		t.Errorf("Host: want nil")
	}
	if req.Method != nil {
		t.Errorf("Method: want nil")
	}
	if req.Body != nil {
		t.Errorf("Body: want nil")
	}
	if req.Headers != nil {
		t.Errorf("Headers: want nil")
	}
}

// mut.resp

func TestRespPatch_SetStatus(t *testing.T) {
	L, _, _, resp := newPatchState(t)
	if err := L.DoString(`mut.resp.set_status(404)`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if resp.StatusCode == nil || *resp.StatusCode != 404 {
		t.Errorf("StatusCode: want 404 got %v", resp.StatusCode)
	}
}

func TestRespPatch_SetBody(t *testing.T) {
	L, _, _, resp := newPatchState(t)
	if err := L.DoString(`mut.resp.set_body('{"ok":true}')`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if resp.Body == nil {
		t.Fatal("Body: want non-nil")
	}
	if string(*resp.Body) != `{"ok":true}` {
		t.Errorf("Body: want '{\"ok\":true}' got %q", string(*resp.Body))
	}
}

func TestRespPatch_SetHeader(t *testing.T) {
	L, _, _, resp := newPatchState(t)
	if err := L.DoString(`mut.resp.set_header("Content-Type", "application/json")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	vals := resp.Headers["Content-Type"]
	if len(vals) != 1 || vals[0] != "application/json" {
		t.Errorf("Content-Type: want [application/json] got %v", vals)
	}
}

func TestRespPatch_SetHeader_Overwrites(t *testing.T) {
	L, _, _, resp := newPatchState(t)
	if err := L.DoString(`
		mut.resp.set_header("X-H", "first")
		mut.resp.set_header("X-H", "second")
	`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	vals := resp.Headers["X-H"]
	if len(vals) != 1 || vals[0] != "second" {
		t.Errorf("X-H: want [second] got %v", vals)
	}
}

func TestRespPatch_AddHeader(t *testing.T) {
	L, _, _, resp := newPatchState(t)
	if err := L.DoString(`
		mut.resp.add_header("X-Multi", "a")
		mut.resp.add_header("X-Multi", "b")
	`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	vals := resp.Headers["X-Multi"]
	if len(vals) != 2 || vals[0] != "a" || vals[1] != "b" {
		t.Errorf("X-Multi: want [a b] got %v", vals)
	}
}

func TestRespPatch_DelHeader(t *testing.T) {
	L, _, _, resp := newPatchState(t)
	if err := L.DoString(`mut.resp.del_header("X-Remove")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	vals, exists := resp.Headers["X-Remove"]
	if !exists {
		t.Fatal("X-Remove: want key present with nil value")
	}
	if vals != nil {
		t.Errorf("X-Remove: want nil slice got %v", vals)
	}
}

func TestRespPatch_UnsetFieldsRemainNil(t *testing.T) {
	_, _, _, resp := newPatchState(t)
	if resp.StatusCode != nil {
		t.Errorf("StatusCode: want nil")
	}
	if resp.Body != nil {
		t.Errorf("Body: want nil")
	}
	if resp.Headers != nil {
		t.Errorf("Headers: want nil")
	}
}

// Multiple patches in one script.

func TestPatch_FullScript(t *testing.T) {
	L, ctx, req, resp := newPatchState(t)
	if err := L.DoString(`
		mut.ctx.set_finish()
		mut.req.set_host("proxy.example.com")
		mut.req.set_method("POST")
		mut.req.set_header("X-Forwarded-For", "1.2.3.4")
		mut.req.add_header("X-Tags", "a")
		mut.req.add_header("X-Tags", "b")
		mut.resp.set_status(200)
		mut.resp.set_body('{"patched":true}')
		mut.resp.set_header("Content-Type", "application/json")
	`); err != nil {
		t.Fatalf("lua error: %v", err)
	}

	if ctx.IsFinished == nil || !*ctx.IsFinished {
		t.Errorf("IsFinished: want true")
	}
	if req.Host == nil || *req.Host != "proxy.example.com" {
		t.Errorf("Host: want proxy.example.com got %v", req.Host)
	}
	if req.Method == nil || *req.Method != "POST" {
		t.Errorf("Method: want POST got %v", req.Method)
	}
	if req.Headers["X-Forwarded-For"][0] != "1.2.3.4" {
		t.Errorf("X-Forwarded-For: want 1.2.3.4 got %v", req.Headers["X-Forwarded-For"])
	}
	if len(req.Headers["X-Tags"]) != 2 {
		t.Errorf("X-Tags: want 2 values got %v", req.Headers["X-Tags"])
	}
	if resp.StatusCode == nil || *resp.StatusCode != 200 {
		t.Errorf("StatusCode: want 200 got %v", resp.StatusCode)
	}
	if string(*resp.Body) != `{"patched":true}` {
		t.Errorf("Body: want '{\"patched\":true}' got %q", string(*resp.Body))
	}
	if resp.Headers["Content-Type"][0] != "application/json" {
		t.Errorf("Content-Type: want application/json got %v", resp.Headers["Content-Type"])
	}
}
