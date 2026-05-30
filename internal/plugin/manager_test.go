package plugin

import (
	"errors"
	"testing"

	"gitlab.com/marsskom/burro/internal/model"
)

var calls map[string][]string

type mockBurroPlugin struct {
	name     string
	enabled  *bool
	priority int

	called map[string]bool
	fail   map[string]bool
}

func newMock(name string) *mockBurroPlugin {
	return &mockBurroPlugin{
		name:   name,
		called: map[string]bool{},
		fail:   map[string]bool{},
	}
}

func (m *mockBurroPlugin) Init(cfg any) error { return nil }

func (m *mockBurroPlugin) Name() string { return m.name }

func (m *mockBurroPlugin) Enabled() *bool { return m.enabled }

func (m *mockBurroPlugin) Priority() int { return m.priority }

func (m *mockBurroPlugin) OnConnect(*model.RequestContext) error {
	calls["connect"] = append(calls["connect"], m.name)

	m.called["connect"] = true
	if m.fail["connect"] {
		return errors.New("connect error")
	}

	return nil
}

func (m *mockBurroPlugin) OnRequest(*model.RequestContext) error {
	calls["request"] = append(calls["request"], m.name)

	m.called["request"] = true
	if m.fail["request"] {
		return errors.New("request error")
	}

	return nil
}

func (m *mockBurroPlugin) OnResponse(*model.RequestContext) error {
	calls["response"] = append(calls["response"], m.name)

	m.called["response"] = true
	if m.fail["response"] {
		return errors.New("response error")
	}

	return nil
}

func (m *mockBurroPlugin) OnError(*model.RequestContext, error) error {
	calls["error"] = append(calls["error"], m.name)

	m.called["error"] = true
	if m.fail["error"] {
		return errors.New("error hook failed")
	}

	return nil
}

func (m *mockBurroPlugin) OnClose(*model.RequestContext) error {
	calls["close"] = append(calls["close"], m.name)

	m.called["close"] = true
	if m.fail["close"] {
		return errors.New("close error")
	}

	return nil
}

// Mock only with OnConnect hook.

type mockBurroOnConnectPlugin struct {
	name     string
	enabled  *bool
	priority int

	called map[string]bool
	fail   map[string]bool
}

func newMockOnConnect(name string) *mockBurroOnConnectPlugin {
	return &mockBurroOnConnectPlugin{
		name:   name,
		called: map[string]bool{},
		fail:   map[string]bool{},
	}
}

func (m *mockBurroOnConnectPlugin) Init(cfg any) error { return nil }

func (m *mockBurroOnConnectPlugin) Name() string { return m.name }

func (m *mockBurroOnConnectPlugin) Enabled() *bool { return m.enabled }

func (m *mockBurroOnConnectPlugin) Priority() int { return m.priority }

func (m *mockBurroOnConnectPlugin) OnConnect(*model.RequestContext) error {
	calls["connect"] = append(calls["connect"], m.name)

	m.called["connect"] = true
	if m.fail["connect"] {
		return errors.New("connect error")
	}

	return nil
}

func TestManager_EmitRequest(t *testing.T) {
	calls = make(map[string][]string)

	m := NewManager()

	p1 := newMock("p1")
	p2 := newMock("p2")

	m.Register(p1)
	m.Register(p2)

	err := m.EmitRequest(&model.RequestContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !p1.called["request"] {
		t.Fatal("p1 should be called")
	}

	if !p2.called["request"] {
		t.Fatal("p2 should be called")
	}
}

func TestManager_DisabledPluginSkipped(t *testing.T) {
	calls = make(map[string][]string)

	m := NewManager()

	enabled := true
	disabled := false

	p1 := newMock("p1")
	p1.enabled = &enabled

	p2 := newMock("p2")
	p2.enabled = &disabled

	m.Register(p1)
	m.Register(p2)

	_ = m.EmitRequest(&model.RequestContext{})

	if !p1.called["request"] {
		t.Fatal("enabled plugin should be called")
	}

	if p2.called["request"] {
		t.Fatal("disabled plugin should NOT be called")
	}
}

func TestManager_PriorityOrder(t *testing.T) {
	calls = make(map[string][]string)

	m := NewManager()

	p1 := newMock("low")
	p1.priority = 200

	p2 := newMock("high")
	p2.priority = 10

	m.Register(p1)
	m.Register(p2)

	_ = m.EmitRequest(&model.RequestContext{})

	for _, c := range calls {
		if len(c) != 2 {
			t.Fatalf("expected 2 calls, got %d", len(c))
		}

		if c[0] != "high" {
			t.Fatalf("expected high priority first, got %v", c)
		}
	}
}

func TestManager_StopOnError(t *testing.T) {
	calls = make(map[string][]string)

	m := NewManager()

	p1 := newMock("p1")
	p2 := newMock("p2")

	p1.fail["request"] = true

	m.Register(p1)
	m.Register(p2)

	err := m.EmitRequest(&model.RequestContext{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !p1.called["request"] {
		t.Fatal("p1 should be called")
	}

	if p2.called["request"] {
		t.Fatal("p2 should NOT be called after error")
	}
}

func TestManager_HookFiltering(t *testing.T) {
	calls = make(map[string][]string)

	m := NewManager()

	p := newMockOnConnect("only-connect")

	m.Register(p)

	_ = m.EmitConnect(&model.RequestContext{})
	_ = m.EmitRequest(&model.RequestContext{})

	if !p.called["connect"] {
		t.Fatal("connect should be called")
	}

	if p.called["request"] {
		t.Fatal("request should NOT be called")
	}
}
