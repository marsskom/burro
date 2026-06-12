package patch

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"gitlab.com/marsskom/burro/internal/model"
)

func makeResponse(t *testing.T, statusCode int) *http.Response {
	t.Helper()
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader([]byte{})),
	}
}

func intPtr(i int) *int { return &i }

// nil patch

func TestApplyResponsePatch_NilPatch(t *testing.T) {
	resp := makeResponse(t, http.StatusOK)
	if err := ApplyResponsePatch(resp, nil); err != nil {
		t.Fatalf("nil patch: unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: want 200 got %d", resp.StatusCode)
	}
}

// status code

func TestApplyResponsePatch_StatusCode(t *testing.T) {
	resp := makeResponse(t, http.StatusOK)
	p := &model.ResponsePatch{StatusCode: intPtr(http.StatusNotFound)}
	if err := ApplyResponsePatch(resp, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode: want 404 got %d", resp.StatusCode)
	}
	if resp.Status != http.StatusText(http.StatusNotFound) {
		t.Errorf("Status: want %q got %q", http.StatusText(http.StatusNotFound), resp.Status)
	}
}

func TestApplyResponsePatch_StatusCode_StatusTextUpdated(t *testing.T) {
	resp := makeResponse(t, http.StatusOK)
	p := &model.ResponsePatch{StatusCode: intPtr(http.StatusInternalServerError)}
	if err := ApplyResponsePatch(resp, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := http.StatusText(http.StatusInternalServerError)
	if resp.Status != want {
		t.Errorf("Status text: want %q got %q", want, resp.Status)
	}
}

// body

func TestApplyResponsePatch_Body(t *testing.T) {
	resp := makeResponse(t, http.StatusOK)
	body := []byte(`{"patched":true}`)
	p := &model.ResponsePatch{Body: &body}
	if err := ApplyResponsePatch(resp, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(got) != `{"patched":true}` {
		t.Errorf("body: want '{\"patched\":true}' got %q", string(got))
	}
}

func TestApplyResponsePatch_Body_ContentLength(t *testing.T) {
	resp := makeResponse(t, http.StatusOK)
	body := []byte("hello")
	p := &model.ResponsePatch{Body: &body}
	if err := ApplyResponsePatch(resp, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ContentLength != 5 {
		t.Errorf("ContentLength: want 5 got %d", resp.ContentLength)
	}
}

func TestApplyResponsePatch_Body_Empty(t *testing.T) {
	resp := makeResponse(t, http.StatusOK)
	body := []byte{}
	p := &model.ResponsePatch{Body: &body}
	if err := ApplyResponsePatch(resp, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ContentLength != 0 {
		t.Errorf("ContentLength: want 0 got %d", resp.ContentLength)
	}
	got, _ := io.ReadAll(resp.Body)
	if len(got) != 0 {
		t.Errorf("body: want empty got %q", string(got))
	}
}

// headers

func TestApplyResponsePatch_Headers_Set(t *testing.T) {
	resp := makeResponse(t, http.StatusOK)
	p := &model.ResponsePatch{
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
	}
	if err := ApplyResponsePatch(resp, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := resp.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type: want application/json got %s", got)
	}
}

func TestApplyResponsePatch_Headers_MultipleValues(t *testing.T) {
	resp := makeResponse(t, http.StatusOK)
	p := &model.ResponsePatch{
		Headers: map[string][]string{
			"X-Multi": {"a", "b", "c"},
		},
	}
	if err := ApplyResponsePatch(resp, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	vals := resp.Header["X-Multi"]
	if len(vals) != 3 {
		t.Errorf("X-Multi: want 3 values got %d: %v", len(vals), vals)
	}
}

func TestApplyResponsePatch_Headers_Delete(t *testing.T) {
	resp := makeResponse(t, http.StatusOK)
	resp.Header.Set("X-Remove", "exists")
	p := &model.ResponsePatch{
		Headers: map[string][]string{
			"X-Remove": nil,
		},
	}
	if err := ApplyResponsePatch(resp, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := resp.Header.Get("X-Remove"); got != "" {
		t.Errorf("X-Remove: want deleted got %s", got)
	}
}

func TestApplyResponsePatch_Headers_SetAndDelete(t *testing.T) {
	resp := makeResponse(t, http.StatusOK)
	resp.Header.Set("X-Old", "old")
	p := &model.ResponsePatch{
		Headers: map[string][]string{
			"X-New": {"new"},
			"X-Old": nil,
		},
	}
	if err := ApplyResponsePatch(resp, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := resp.Header.Get("X-New"); got != "new" {
		t.Errorf("X-New: want new got %s", got)
	}
	if got := resp.Header.Get("X-Old"); got != "" {
		t.Errorf("X-Old: want deleted got %s", got)
	}
}

// empty patch

func TestApplyResponsePatch_EmptyPatch_NoChanges(t *testing.T) {
	resp := makeResponse(t, http.StatusCreated)
	resp.Header.Set("X-Keep", "value")
	p := &model.ResponsePatch{}
	if err := ApplyResponsePatch(resp, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status: want 201 got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("X-Keep"); got != "value" {
		t.Errorf("X-Keep: want value got %s", got)
	}
}

// combined

func TestApplyResponsePatch_Combined(t *testing.T) {
	resp := makeResponse(t, http.StatusOK)
	resp.Header.Set("X-Remove", "bye")
	body := []byte("patched body")
	p := &model.ResponsePatch{
		StatusCode: intPtr(http.StatusAccepted),
		Body:       &body,
		Headers: map[string][]string{
			"X-Add":    {"added"},
			"X-Remove": nil,
		},
	}
	if err := ApplyResponsePatch(resp, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("StatusCode: want 202 got %d", resp.StatusCode)
	}
	if resp.Status != http.StatusText(http.StatusAccepted) {
		t.Errorf("Status: want %q got %q", http.StatusText(http.StatusAccepted), resp.Status)
	}
	got, _ := io.ReadAll(resp.Body)
	if string(got) != "patched body" {
		t.Errorf("body: want 'patched body' got %q", string(got))
	}
	if resp.ContentLength != int64(len(body)) {
		t.Errorf("ContentLength: want %d got %d", len(body), resp.ContentLength)
	}
	if resp.Header.Get("X-Add") != "added" {
		t.Errorf("X-Add: want added got %s", resp.Header.Get("X-Add"))
	}
	if resp.Header.Get("X-Remove") != "" {
		t.Errorf("X-Remove: want deleted")
	}
}
