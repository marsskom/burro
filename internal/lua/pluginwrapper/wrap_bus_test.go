package pluginwrapper

import (
	"errors"
	"testing"

	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/pluginapi"
)

// mockBus is a simple EventBus mock
type mockBus struct {
	emitted []pluginapi.Event
	emitErr error
}

func (m *mockBus) Emit(event pluginapi.Event) error {
	m.emitted = append(m.emitted, event)
	return m.emitErr
}

func (m *mockBus) On(name string, handler pluginapi.EventHandler) func() { return func() {} }

func (b *mockBus) Off(name string, id int) {}

func newState(t *testing.T, bus pluginapi.EventBus) *lua.LState {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })

	RegisterEventBus(L, bus)

	return L
}

func TestRegisterEventBus_EmitStringData(t *testing.T) {
	bus := &mockBus{}
	L := newState(t, bus)

	err := L.DoString(`bus.emit("user.created", "john")`)
	if err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}
	if len(bus.emitted) != 1 {
		t.Fatalf("expected 1 event, got %d", len(bus.emitted))
	}
	ev := bus.emitted[0]
	if ev.Name != "user.created" {
		t.Errorf("name: want user.created got %s", ev.Name)
	}
	if ev.Data != "john" {
		t.Errorf("data: want john got %v", ev.Data)
	}
	if ev.From != "lua" {
		t.Errorf("from: want lua got %s", ev.From)
	}
}

func TestRegisterEventBus_EmitNilData(t *testing.T) {
	bus := &mockBus{}
	L := newState(t, bus)

	if err := L.DoString(`bus.emit("ping", nil)`); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}
	if len(bus.emitted) != 1 {
		t.Fatalf("expected 1 event, got %d", len(bus.emitted))
	}
	if bus.emitted[0].Data != nil {
		t.Errorf("expected nil data, got %v", bus.emitted[0].Data)
	}
}

func TestRegisterEventBus_EmitNumberData(t *testing.T) {
	bus := &mockBus{}
	L := newState(t, bus)

	if err := L.DoString(`bus.emit("score", 42)`); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}
	if bus.emitted[0].Data != float64(42) {
		t.Errorf("data: want 42.0 got %v", bus.emitted[0].Data)
	}
}

func TestRegisterEventBus_EmitBoolData(t *testing.T) {
	bus := &mockBus{}
	L := newState(t, bus)

	if err := L.DoString(`bus.emit("flag", true)`); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}
	if bus.emitted[0].Data != true {
		t.Errorf("data: want true got %v", bus.emitted[0].Data)
	}
}

func TestRegisterEventBus_EmitTableData(t *testing.T) {
	bus := &mockBus{}
	L := newState(t, bus)

	if err := L.DoString(`bus.emit("payload", { key = "val", num = 1 })`); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}
	m, ok := bus.emitted[0].Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", bus.emitted[0].Data)
	}
	if m["key"] != "val" {
		t.Errorf("map[key]: want val got %v", m["key"])
	}
	if m["num"] != float64(1) {
		t.Errorf("map[num]: want 1.0 got %v", m["num"])
	}
}

func TestRegisterEventBus_EmitError_ReturnedToLua(t *testing.T) {
	bus := &mockBus{emitErr: errors.New("bus unavailable")}
	L := newState(t, bus)

	// capture return value in lua
	if err := L.DoString(`err = bus.emit("fail", "data")`); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}
	errVal := L.GetGlobal("err")
	if errVal.Type() != lua.LTString {
		t.Fatalf("expected string error, got %s", errVal.Type())
	}
	if errVal.String() != "bus unavailable" {
		t.Errorf("error msg: want 'bus unavailable' got %q", errVal.String())
	}
}

func TestRegisterEventBus_EmitSuccess_ReturnsNil(t *testing.T) {
	bus := &mockBus{}
	L := newState(t, bus)

	if err := L.DoString(`err = bus.emit("ok", "data")`); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}
	errVal := L.GetGlobal("err")
	if errVal.Type() != lua.LTNil {
		t.Errorf("expected nil on success, got %s: %v", errVal.Type(), errVal)
	}
}

func TestRegisterEventBus_MultipleEmits(t *testing.T) {
	bus := &mockBus{}
	L := newState(t, bus)

	if err := L.DoString(`
		bus.emit("first",  "a")
		bus.emit("second", "b")
		bus.emit("third",  "c")
	`); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}
	if len(bus.emitted) != 3 {
		t.Fatalf("expected 3 events, got %d", len(bus.emitted))
	}
	names := []string{"first", "second", "third"}
	for i, want := range names {
		if bus.emitted[i].Name != want {
			t.Errorf("event[%d].Name: want %s got %s", i, want, bus.emitted[i].Name)
		}
	}
}

func TestLuaValueToGo_NestedTable(t *testing.T) {
	bus := &mockBus{}
	L := newState(t, bus)

	if err := L.DoString(`bus.emit("nested", { inner = { x = 99 } })`); err != nil {
		t.Fatalf("unexpected lua error: %v", err)
	}
	m, ok := bus.emitted[0].Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map got %T", bus.emitted[0].Data)
	}
	inner, ok := m["inner"].(map[string]any)
	if !ok {
		t.Fatalf("expected inner map got %T", m["inner"])
	}
	if inner["x"] != float64(99) {
		t.Errorf("inner[x]: want 99.0 got %v", inner["x"])
	}
}
