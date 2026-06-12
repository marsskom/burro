package patch

import (
	"io"
	"net/http"
	"net/url"
	"testing"

	"gitlab.com/marsskom/burro/internal/model"
)

func makeRequest(t *testing.T, method, rawURL string) *http.Request {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("makeRequest: parse url: %v", err)
	}
	return &http.Request{
		Method: method,
		URL:    u,
		Host:   u.Host,
		Header: make(http.Header),
	}
}

func strPtr(s string) *string { return &s }

// nil patch

func TestApplyRequestPatch_NilPatch(t *testing.T) {
	r := makeRequest(t, "GET", "http://example.com/path")
	if err := ApplyRequestPatch(r, nil); err != nil {
		t.Fatalf("nil patch: unexpected error: %v", err)
	}
	if r.Method != "GET" {
		t.Errorf("method unchanged: want GET got %s", r.Method)
	}
}

// host

func TestApplyRequestPatch_Host(t *testing.T) {
	r := makeRequest(t, "GET", "http://old.example.com/path")
	p := &model.RequestPatch{Host: strPtr("new.example.com")}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Host != "new.example.com" {
		t.Errorf("r.Host: want new.example.com got %s", r.Host)
	}
	if r.URL.Host != "new.example.com" {
		t.Errorf("r.URL.Host: want new.example.com got %s", r.URL.Host)
	}
}

// scheme

func TestApplyRequestPatch_Scheme(t *testing.T) {
	r := makeRequest(t, "GET", "http://example.com/path")
	p := &model.RequestPatch{Scheme: strPtr("https")}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.URL.Scheme != "https" {
		t.Errorf("scheme: want https got %s", r.URL.Scheme)
	}
}

// method

func TestApplyRequestPatch_Method(t *testing.T) {
	r := makeRequest(t, "GET", "http://example.com/")
	p := &model.RequestPatch{Method: strPtr("POST")}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Method != "POST" {
		t.Errorf("method: want POST got %s", r.Method)
	}
}

// path

func TestApplyRequestPatch_Path(t *testing.T) {
	r := makeRequest(t, "GET", "http://example.com/old")
	p := &model.RequestPatch{Path: strPtr("/new/path")}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.URL.Path != "/new/path" {
		t.Errorf("path: want /new/path got %s", r.URL.Path)
	}
}

// url

func TestApplyRequestPatch_URL(t *testing.T) {
	r := makeRequest(t, "GET", "http://old.example.com/old")
	p := &model.RequestPatch{URL: strPtr("https://new.example.com/new?q=1")}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.URL.Scheme != "https" {
		t.Errorf("scheme: want https got %s", r.URL.Scheme)
	}
	if r.URL.Host != "new.example.com" {
		t.Errorf("host: want new.example.com got %s", r.URL.Host)
	}
	if r.URL.Path != "/new" {
		t.Errorf("path: want /new got %s", r.URL.Path)
	}
	if r.URL.RawQuery != "q=1" {
		t.Errorf("query: want q=1 got %s", r.URL.RawQuery)
	}
	if r.Host != "new.example.com" {
		t.Errorf("r.Host: want new.example.com got %s", r.Host)
	}
}

func TestApplyRequestPatch_URL_Invalid(t *testing.T) {
	r := makeRequest(t, "GET", "http://example.com/")
	p := &model.RequestPatch{URL: strPtr("://bad url\x00")}
	err := ApplyRequestPatch(r, p)
	if err == nil {
		t.Fatal("expected error on invalid url, got nil")
	}
}

// body

func TestApplyRequestPatch_Body(t *testing.T) {
	r := makeRequest(t, "POST", "http://example.com/")
	body := []byte("new body content")
	p := &model.RequestPatch{Body: &body}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ContentLength != int64(len(body)) {
		t.Errorf("ContentLength: want %d got %d", len(body), r.ContentLength)
	}
	got, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(got) != "new body content" {
		t.Errorf("body: want 'new body content' got %q", string(got))
	}
}

func TestApplyRequestPatch_Body_GetBody(t *testing.T) {
	r := makeRequest(t, "POST", "http://example.com/")
	body := []byte("repeatable")
	p := &model.RequestPatch{Body: &body}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.GetBody == nil {
		t.Fatal("GetBody: want non-nil func")
	}
	rc, err := r.GetBody()
	if err != nil {
		t.Fatalf("GetBody(): %v", err)
	}
	got, _ := io.ReadAll(rc)
	if string(got) != "repeatable" {
		t.Errorf("GetBody: want 'repeatable' got %q", string(got))
	}
}

func TestApplyRequestPatch_Body_Empty(t *testing.T) {
	r := makeRequest(t, "POST", "http://example.com/")
	body := []byte{}
	p := &model.RequestPatch{Body: &body}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ContentLength != 0 {
		t.Errorf("ContentLength: want 0 got %d", r.ContentLength)
	}
}

// headers

func TestApplyRequestPatch_Headers_Set(t *testing.T) {
	r := makeRequest(t, "GET", "http://example.com/")
	p := &model.RequestPatch{
		Headers: map[string][]string{
			"X-Custom": {"value1"},
		},
	}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := r.Header.Get("X-Custom"); got != "value1" {
		t.Errorf("X-Custom: want value1 got %s", got)
	}
}

func TestApplyRequestPatch_Headers_MultipleValues(t *testing.T) {
	r := makeRequest(t, "GET", "http://example.com/")
	p := &model.RequestPatch{
		Headers: map[string][]string{
			"X-Multi": {"a", "b", "c"},
		},
	}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	vals := r.Header["X-Multi"]
	if len(vals) != 3 {
		t.Errorf("X-Multi: want 3 values got %d: %v", len(vals), vals)
	}
}

func TestApplyRequestPatch_Headers_Delete(t *testing.T) {
	r := makeRequest(t, "GET", "http://example.com/")
	r.Header.Set("X-Remove", "exists")
	p := &model.RequestPatch{
		Headers: map[string][]string{
			"X-Remove": nil,
		},
	}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Header.Get("X-Remove") != "" {
		t.Errorf("X-Remove: want deleted got %s", r.Header.Get("X-Remove"))
	}
}

func TestApplyRequestPatch_Headers_SetAndDelete(t *testing.T) {
	r := makeRequest(t, "GET", "http://example.com/")
	r.Header.Set("X-Old", "old")
	p := &model.RequestPatch{
		Headers: map[string][]string{
			"X-New": {"new"},
			"X-Old": nil,
		},
	}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Header.Get("X-New") != "new" {
		t.Errorf("X-New: want new got %s", r.Header.Get("X-New"))
	}
	if r.Header.Get("X-Old") != "" {
		t.Errorf("X-Old: want deleted got %s", r.Header.Get("X-Old"))
	}
}

// empty patch (all nil fields)

func TestApplyRequestPatch_EmptyPatch_NoChanges(t *testing.T) {
	r := makeRequest(t, "DELETE", "http://example.com/original")
	r.Header.Set("X-Existing", "keep")
	p := &model.RequestPatch{}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Method != "DELETE" {
		t.Errorf("method: want DELETE got %s", r.Method)
	}
	if r.URL.Path != "/original" {
		t.Errorf("path: want /original got %s", r.URL.Path)
	}
	if r.Header.Get("X-Existing") != "keep" {
		t.Errorf("X-Existing: want keep got %s", r.Header.Get("X-Existing"))
	}
}

// combined patch

func TestApplyRequestPatch_Combined(t *testing.T) {
	r := makeRequest(t, "GET", "http://old.example.com/old")
	r.Header.Set("X-Remove", "bye")
	body := []byte("patched body")
	p := &model.RequestPatch{
		Host:   strPtr("new.example.com"),
		Method: strPtr("POST"),
		Path:   strPtr("/new"),
		Body:   &body,
		Headers: map[string][]string{
			"X-Add":    {"added"},
			"X-Remove": nil,
		},
	}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Host != "new.example.com" {
		t.Errorf("host: want new.example.com got %s", r.Host)
	}
	if r.Method != "POST" {
		t.Errorf("method: want POST got %s", r.Method)
	}
	if r.URL.Path != "/new" {
		t.Errorf("path: want /new got %s", r.URL.Path)
	}
	got, _ := io.ReadAll(r.Body)
	if string(got) != "patched body" {
		t.Errorf("body: want 'patched body' got %q", string(got))
	}
	if r.Header.Get("X-Add") != "added" {
		t.Errorf("X-Add: want added got %s", r.Header.Get("X-Add"))
	}
	if r.Header.Get("X-Remove") != "" {
		t.Errorf("X-Remove: want deleted")
	}
}

// cookies

func TestApplyRequestPatch_Cookie_Set(t *testing.T) {
	r := makeRequest(t, "GET", "http://example.com/")
	p := &model.RequestPatch{
		Cookies: []model.CookiePatch{
			{Name: "session", Value: "abc123"},
		},
	}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, err := r.Cookie("session")
	if err != nil {
		t.Fatalf("cookie session not found: %v", err)
	}
	if c.Value != "abc123" {
		t.Errorf("session value: want abc123 got %s", c.Value)
	}
}

func TestApplyRequestPatch_Cookie_AllFields(t *testing.T) {
	r := makeRequest(t, "GET", "http://example.com/")
	p := &model.RequestPatch{
		Cookies: []model.CookiePatch{
			{
				Name:     "full",
				Value:    "val",
				Path:     "/api",
				Domain:   "example.com",
				Secure:   true,
				HTTPOnly: true,
				MaxAge:   3600,
			},
		},
	}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, err := r.Cookie("full")
	if err != nil {
		t.Fatalf("cookie full not found: %v", err)
	}
	if c.Value != "val" {
		t.Errorf("value: want val got %s", c.Value)
	}
}

func TestApplyRequestPatch_Cookie_Delete(t *testing.T) {
	r := makeRequest(t, "GET", "http://example.com/")
	r.AddCookie(&http.Cookie{Name: "todel", Value: "bye"})
	r.AddCookie(&http.Cookie{Name: "keep", Value: "yes"})

	p := &model.RequestPatch{
		Cookies: []model.CookiePatch{
			{Name: "todel", Delete: true},
		},
	}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := r.Cookie("todel"); err == nil {
		t.Errorf("todel: want deleted but still present")
	}
	if _, err := r.Cookie("keep"); err != nil {
		t.Errorf("keep: want present but got: %v", err)
	}
}

func TestApplyRequestPatch_Cookie_Update_Existing(t *testing.T) {
	r := makeRequest(t, "GET", "http://example.com/")
	r.AddCookie(&http.Cookie{Name: "token", Value: "old"})

	p := &model.RequestPatch{
		Cookies: []model.CookiePatch{
			{Name: "token", Value: "new"},
		},
	}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, err := r.Cookie("token")
	if err != nil {
		t.Fatalf("cookie token not found: %v", err)
	}
	if c.Value != "new" {
		t.Errorf("token value: want new got %s", c.Value)
	}
}

func TestApplyRequestPatch_Cookie_PreservesUnpatched(t *testing.T) {
	r := makeRequest(t, "GET", "http://example.com/")
	r.AddCookie(&http.Cookie{Name: "a", Value: "1"})
	r.AddCookie(&http.Cookie{Name: "b", Value: "2"})
	r.AddCookie(&http.Cookie{Name: "c", Value: "3"})

	p := &model.RequestPatch{
		Cookies: []model.CookiePatch{
			{Name: "b", Value: "patched"},
		},
	}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c, err := r.Cookie("a"); err != nil || c.Value != "1" {
		t.Errorf("a: want 1 got %v (err: %v)", c, err)
	}
	if c, err := r.Cookie("b"); err != nil || c.Value != "patched" {
		t.Errorf("b: want patched got %v (err: %v)", c, err)
	}
	if c, err := r.Cookie("c"); err != nil || c.Value != "3" {
		t.Errorf("c: want 3 got %v (err: %v)", c, err)
	}
}

func TestApplyRequestPatch_Cookie_NoPatch_PreservesAll(t *testing.T) {
	r := makeRequest(t, "GET", "http://example.com/")
	r.AddCookie(&http.Cookie{Name: "x", Value: "1"})
	r.AddCookie(&http.Cookie{Name: "y", Value: "2"})

	p := &model.RequestPatch{}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c, err := r.Cookie("x"); err != nil || c.Value != "1" {
		t.Errorf("x: want 1 got %v (err: %v)", c, err)
	}
	if c, err := r.Cookie("y"); err != nil || c.Value != "2" {
		t.Errorf("y: want 2 got %v (err: %v)", c, err)
	}
}

func TestApplyRequestPatch_Cookie_AddNew_And_DeleteOld(t *testing.T) {
	r := makeRequest(t, "GET", "http://example.com/")
	r.AddCookie(&http.Cookie{Name: "old", Value: "bye"})

	p := &model.RequestPatch{
		Cookies: []model.CookiePatch{
			{Name: "old", Delete: true},
			{Name: "new", Value: "hello"},
		},
	}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := r.Cookie("old"); err == nil {
		t.Errorf("old: want deleted but still present")
	}
	if c, err := r.Cookie("new"); err != nil || c.Value != "hello" {
		t.Errorf("new: want hello got %v (err: %v)", c, err)
	}
}

func TestApplyRequestPatch_Cookie_Multiple(t *testing.T) {
	r := makeRequest(t, "GET", "http://example.com/")

	p := &model.RequestPatch{
		Cookies: []model.CookiePatch{
			{Name: "first", Value: "1"},
			{Name: "second", Value: "2"},
			{Name: "third", Value: "3"},
		},
	}
	if err := ApplyRequestPatch(r, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for name, want := range map[string]string{
		"first": "1", "second": "2", "third": "3",
	} {
		c, err := r.Cookie(name)
		if err != nil {
			t.Errorf("%s: not found: %v", name, err)
			continue
		}
		if c.Value != want {
			t.Errorf("%s: want %s got %s", name, want, c.Value)
		}
	}
}
