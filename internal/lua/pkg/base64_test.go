package pkg

import (
	"encoding/base64"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newBase64State(t *testing.T) *lua.LState {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })
	if err := RegisterBase64(L); err != nil {
		t.Fatalf("RegisterBase64: %v", err)
	}
	return L
}

// encode

func TestBase64_Encode_OK(t *testing.T) {
	L := newBase64State(t)
	if err := L.DoString(`result = base64.encode("hello world")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	want := base64.StdEncoding.EncodeToString([]byte("hello world"))
	if got := L.GetGlobal("result").String(); got != want {
		t.Errorf("encode: want %q got %q", want, got)
	}
}

func TestBase64_Encode_EmptyString(t *testing.T) {
	L := newBase64State(t)
	if err := L.DoString(`result = base64.encode("")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if got := L.GetGlobal("result").String(); got != "" {
		t.Errorf("encode empty: want empty string got %q", got)
	}
}

func TestBase64_Encode_BinaryDataInLuaFormat(t *testing.T) {
	L := newBase64State(t)
	if err := L.DoString(`result = base64.encode("\0\1\2")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	want := base64.StdEncoding.EncodeToString([]byte{0x00, 0x01, 0x02})
	if got := L.GetGlobal("result").String(); got != want {
		t.Errorf("encode binary: want %q got %q", want, got)
	}
}

func TestBase64_Encode_BinaryDataFromGoToLuaFormat(t *testing.T) {
	L := newBase64State(t)
	input := string([]byte{0x00, 0x01, 0x02})
	L.SetGlobal("input", lua.LString(input))
	if err := L.DoString(`result = base64.encode(input)`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	want := base64.StdEncoding.EncodeToString([]byte{0x00, 0x01, 0x02})
	if got := L.GetGlobal("result").String(); got != want {
		t.Errorf("encode binary: want %q got %q", want, got)
	}
}

func TestBase64_Encode_BinaryDataError(t *testing.T) {
	L := newBase64State(t)
	if err := L.DoString("result = base64.encode(\"\\x00\\x01\\x02\")"); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	want := base64.StdEncoding.EncodeToString([]byte{0x00, 0x01, 0x02})
	if got := L.GetGlobal("result").String(); got == want {
		t.Errorf("encode binary lua shouldn't encode \\x escape sequence correctly")
	}
}

func TestBase64_Encode_ReturnsOneValue(t *testing.T) {
	L := newBase64State(t)
	if err := L.DoString(`a, b = base64.encode("test")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("b").Type() != lua.LTNil {
		t.Errorf("encode: want single return value, second is %v", L.GetGlobal("b"))
	}
}

// decode

func TestBase64_Decode_OK(t *testing.T) {
	L := newBase64State(t)
	encoded := base64.StdEncoding.EncodeToString([]byte("hello world"))
	if err := L.DoString(`result, err = base64.decode("` + encoded + `")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("err").Type() != lua.LTNil {
		t.Errorf("err: want nil got %v", L.GetGlobal("err"))
	}
	if got := L.GetGlobal("result").String(); got != "hello world" {
		t.Errorf("decode: want 'hello world' got %q", got)
	}
}

func TestBase64_Decode_EmptyString(t *testing.T) {
	L := newBase64State(t)
	if err := L.DoString(`result, err = base64.decode("")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("err").Type() != lua.LTNil {
		t.Errorf("err: want nil got %v", L.GetGlobal("err"))
	}
	if got := L.GetGlobal("result").String(); got != "" {
		t.Errorf("decode empty: want empty got %q", got)
	}
}

func TestBase64_Decode_InvalidInput(t *testing.T) {
	L := newBase64State(t)
	if err := L.DoString(`result, err = base64.decode("not!valid@base64#")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("result").Type() != lua.LTNil {
		t.Errorf("result: want nil on error got %v", L.GetGlobal("result"))
	}
	if L.GetGlobal("err").Type() != lua.LTString {
		t.Errorf("err: want string got %s", L.GetGlobal("err").Type())
	}
}

func TestBase64_Decode_InvalidPadding(t *testing.T) {
	L := newBase64State(t)
	// Valid base64 chars but wrong padding.
	if err := L.DoString(`result, err = base64.decode("aGVsbG8")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("result").Type() != lua.LTNil {
		t.Errorf("result: want nil on padding error")
	}
	if L.GetGlobal("err").Type() != lua.LTString {
		t.Errorf("err: want string on padding error")
	}
}

// encode -> decode roundtrip

func TestBase64_Roundtrip(t *testing.T) {
	L := newBase64State(t)
	if err := L.DoString(`
		encoded = base64.encode("roundtrip value")
		result, err = base64.decode(encoded)
	`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if L.GetGlobal("err").Type() != lua.LTNil {
		t.Errorf("err: want nil got %v", L.GetGlobal("err"))
	}
	if got := L.GetGlobal("result").String(); got != "roundtrip value" {
		t.Errorf("roundtrip: want 'roundtrip value' got %q", got)
	}
}

func TestBase64_Roundtrip_Unicode(t *testing.T) {
	L := newBase64State(t)
	if err := L.DoString(`
		encoded = base64.encode("héllo wörld")
		result, err = base64.decode(encoded)
	`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	if got := L.GetGlobal("result").String(); got != "héllo wörld" {
		t.Errorf("roundtrip unicode: want 'héllo wörld' got %q", got)
	}
}
