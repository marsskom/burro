package pkg

import (
	"fmt"
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func newTimeState(t *testing.T) *lua.LState {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })
	if err := RegisterTime(L); err != nil {
		t.Fatalf("RegisterTime: %v", err)
	}
	return L
}

// time.unix

func TestTime_Unix_ReturnsNumber(t *testing.T) {
	L := newTimeState(t)
	before := time.Now().Unix()
	if err := L.DoString(`result = time.unix()`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	after := time.Now().Unix()

	v, ok := L.GetGlobal("result").(lua.LNumber)
	if !ok {
		t.Fatalf("result: want LNumber got %T", L.GetGlobal("result"))
	}
	got := int64(v)
	if got < before || got > after {
		t.Errorf("unix: want value between %d and %d got %d", before, after, got)
	}
}

// time.rfc3339

func TestTime_RFC3339_ParseableFormat(t *testing.T) {
	L := newTimeState(t)
	before := time.Now().Truncate(time.Second)
	if err := L.DoString(`result = time.rfc3339()`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	after := time.Now().Truncate(time.Second)

	got := L.GetGlobal("result").String()
	parsed, err := time.Parse(time.RFC3339, got)
	if err != nil {
		t.Fatalf("rfc3339: could not parse %q: %v", got, err)
	}
	parsed = parsed.UTC()
	before = before.UTC()
	after = after.UTC()
	if parsed.Before(before) || parsed.After(after) {
		t.Errorf("rfc3339: parsed time %v not between %v and %v", parsed, before, after)
	}
}

// time.date

func TestTime_Date_DefaultLayout(t *testing.T) {
	L := newTimeState(t)
	before := time.Now()
	if err := L.DoString(`result = time.date()`); err != nil {
		t.Fatalf("lua error: %v", err)
	}
	after := time.Now()

	got := L.GetGlobal("result").String()
	parsed, err := time.ParseInLocation("2006_01_02_15_04_05", got, time.Local)
	if err != nil {
		t.Fatalf("date default: could not parse %q: %v", got, err)
	}
	if parsed.Before(before.Truncate(time.Second)) || parsed.After(after.Add(time.Second)) {
		t.Errorf("date default: parsed %v not in expected range", parsed)
	}
}

func TestTime_Date_CustomLayout(t *testing.T) {
	L := newTimeState(t)
	if err := L.DoString(`result = time.date("%Y-%m-%d")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}

	got := L.GetGlobal("result").String()
	want := time.Now().Format("2006-01-02")
	if got != want {
		t.Errorf("date custom: want %q got %q", want, got)
	}
}

func TestTime_Date_TimeOnly(t *testing.T) {
	L := newTimeState(t)
	if err := L.DoString(`result = time.date("%H:%M:%S")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}

	got := L.GetGlobal("result").String()
	_, err := time.Parse("15:04:05", got)
	if err != nil {
		t.Fatalf("date time-only: could not parse %q: %v", got, err)
	}
}

func TestTime_Date_ShortYear(t *testing.T) {
	L := newTimeState(t)
	if err := L.DoString(`result = time.date("%y")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}

	got := L.GetGlobal("result").String()
	want := time.Now().Format("06")
	if got != want {
		t.Errorf("short year: want %q got %q", want, got)
	}
}

func TestTime_Date_ShortcutF(t *testing.T) {
	L := newTimeState(t)
	if err := L.DoString(`result = time.date("%F")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}

	got := L.GetGlobal("result").String()
	want := time.Now().Format("2006-01-02")
	if got != want {
		t.Errorf("%%F: want %q got %q", want, got)
	}
}

func TestTime_Date_ShortcutT(t *testing.T) {
	L := newTimeState(t)
	if err := L.DoString(`result = time.date("%T")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}

	got := L.GetGlobal("result").String()
	_, err := time.Parse("15:04:05", got)
	if err != nil {
		t.Fatalf("%%T: could not parse %q: %v", got, err)
	}
}

func TestTime_Date_AllTokensCombined(t *testing.T) {
	L := newTimeState(t)
	if err := L.DoString(`result = time.date("%Y/%m/%d %H:%M:%S")`); err != nil {
		t.Fatalf("lua error: %v", err)
	}

	got := L.GetGlobal("result").String()
	_, err := time.ParseInLocation("2006/01/02 15:04:05", got, time.Local)
	if err != nil {
		t.Fatalf("combined: could not parse %q: %v", got, err)
	}
}

// Lua layout to golang.

func TestLuaLayoutToGo_EmptyReturnsDefault(t *testing.T) {
	got := luaLayoutToGo("")
	want := "2006_01_02_15_04_05"
	if got != want {
		t.Errorf("empty: want %q got %q", want, got)
	}
}

func TestLuaLayoutToGo_Replacements(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"%Y", "2006"},
		{"%y", "06"},
		{"%m", "01"},
		{"%d", "02"},
		{"%H", "15"},
		{"%M", "04"},
		{"%S", "05"},
		{"%F", "2006-01-02"},
		{"%T", "15:04:05"},
		{"%Y-%m-%d", "2006-01-02"},
		{"%H:%M:%S", "15:04:05"},
		{"no_tokens", "no_tokens"},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("input=%s", tc.input), func(t *testing.T) {
			got := luaLayoutToGo(tc.input)
			if got != tc.want {
				t.Errorf("luaLayoutToGo(%q): want %q got %q", tc.input, tc.want, got)
			}
		})
	}
}
