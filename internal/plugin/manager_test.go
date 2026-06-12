package plugin

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.com/marsskom/burro/internal/broker"
	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/pluginapi"
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

func (m *mockBurroPlugin) Shutdown() error { return nil }

func (m *mockBurroPlugin) Init(rt pluginapi.Runtime, cfg any) error { return nil }

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

func (m *mockBurroPlugin) OnBeforeRequestSend(*model.RequestContext) error {
	calls["before_request_send"] = append(calls["before_request_send"], m.name)

	m.called["before_request_send"] = true
	if m.fail["before_request_send"] {
		return errors.New("request error")
	}

	return nil
}

func (m *mockBurroPlugin) OnAfterRequestSend(*model.RequestContext) error {
	calls["after_request_send"] = append(calls["before_request_send"], m.name)

	m.called["after_request_send"] = true
	if m.fail["after_request_send"] {
		return errors.New("request error")
	}

	return nil
}

func (m *mockBurroPlugin) OnBeforeResponseSend(*model.RequestContext) error {
	calls["before_response_send"] = append(calls["before_response_send"], m.name)

	m.called["before_response_send"] = true
	if m.fail["before_response_send"] {
		return errors.New("response error")
	}

	return nil
}

func (m *mockBurroPlugin) OnAfterResponseSend(*model.RequestContext) error {
	calls["after_response_send"] = append(calls["after_response_send"], m.name)

	m.called["after_response_send"] = true
	if m.fail["after_response_send"] {
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

// Mocks only with OnConnect hook.

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

func (m *mockBurroOnConnectPlugin) Shutdown() error { return nil }

func (m *mockBurroOnConnectPlugin) Init(rt pluginapi.Runtime, cfg any) error { return nil }

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

func TestManager_EmitBeforeRequestSend(t *testing.T) {
	calls = make(map[string][]string)

	m := NewManager(broker.NewHub())

	p1 := newMock("p1")
	p2 := newMock("p2")

	m.Register(p1)
	m.Register(p2)

	m.Sort()

	r := httptest.NewRequest("GET", "http://example.com", nil)

	err := m.EmitBeforeRequestSend(model.NewCtx(model.NewSession(), &model.Timings{}, r))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !p1.called["before_request_send"] {
		t.Fatal("p1 should be called")
	}

	if !p2.called["before_request_send"] {
		t.Fatal("p2 should be called")
	}
}

func TestManager_EmitAfterRequestSend(t *testing.T) {
	calls = make(map[string][]string)

	m := NewManager(broker.NewHub())

	p1 := newMock("p1")
	p2 := newMock("p2")

	m.Register(p1)
	m.Register(p2)

	m.Sort()

	r := httptest.NewRequest("GET", "http://example.com", nil)

	err := m.EmitAfterRequestSend(model.NewCtx(model.NewSession(), &model.Timings{}, r))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !p1.called["after_request_send"] {
		t.Fatal("p1 should be called")
	}

	if !p2.called["after_request_send"] {
		t.Fatal("p2 should be called")
	}
}

func TestManager_DisabledPluginSkipped(t *testing.T) {
	calls = make(map[string][]string)

	m := NewManager(broker.NewHub())

	enabled := true
	disabled := false

	p1 := newMock("p1")
	p1.enabled = &enabled

	p2 := newMock("p2")
	p2.enabled = &disabled

	m.Register(p1)
	m.Register(p2)

	m.Sort()

	_ = m.EmitBeforeRequestSend(&model.RequestContext{})
	_ = m.EmitAfterRequestSend(&model.RequestContext{})

	if !p1.called["before_request_send"] {
		t.Fatal("enabled plugin should be called in before request")
	}

	if !p1.called["after_request_send"] {
		t.Fatal("enabled plugin should be called in after request")
	}

	if p2.called["before_request_send"] {
		t.Fatal("disabled plugin should NOT be called in before request")
	}

	if p2.called["after_request_send"] {
		t.Fatal("disabled plugin should NOT be called in after request")
	}
}

func TestManager_PriorityOrder(t *testing.T) {
	calls = make(map[string][]string)

	m := NewManager(broker.NewHub())

	p1 := newMock("low")
	p1.priority = 200

	p2 := newMock("high")
	p2.priority = 10

	m.Register(p1)
	m.Register(p2)

	m.Sort()

	_ = m.EmitBeforeRequestSend(&model.RequestContext{})

	for _, c := range calls {
		if len(c) != 2 {
			t.Fatalf("expected 2 calls, got %d", len(c))
		}

		if c[0] != "high" {
			t.Fatalf("expected high priority first, got %v", c)
		}
	}
}

func TestManager_DoNotStopOnError(t *testing.T) {
	calls = make(map[string][]string)

	m := NewManager(broker.NewHub())

	p1 := newMock("p1")
	p2 := newMock("p2")

	p1.fail["before_request_send"] = true
	p2.fail["before_request_send"] = true

	m.Register(p1)
	m.Register(p2)

	m.Sort()

	err := m.EmitBeforeRequestSend(&model.RequestContext{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !p1.called["before_request_send"] {
		t.Fatal("p1 should be called")
	}

	if !p2.called["before_request_send"] {
		t.Fatal("p2 should be called")
	}
}

func TestManager_HookFiltering(t *testing.T) {
	calls = make(map[string][]string)

	m := NewManager(broker.NewHub())

	p := newMockOnConnect("only-connect")

	m.Register(p)

	_ = m.EmitConnect(&model.RequestContext{})
	_ = m.EmitBeforeRequestSend(&model.RequestContext{})

	if !p.called["connect"] {
		t.Fatal("connect should be called")
	}

	if p.called["before_request_send"] {
		t.Fatal("request should NOT be called")
	}
}

func TestManager_EmitRequest_WithHub(t *testing.T) {
	calls = make(map[string][]string)

	hub := broker.NewHub()

	sub := hub.Subscribe([]broker.TransportType{}, []broker.EventType{})
	defer hub.Unsubscribe(sub)

	m := NewManager(hub)

	r := httptest.NewRequest("GET", "http://example.com", nil)

	err := m.EmitBeforeRequestSend(model.NewCtx(model.NewSession(), &model.Timings{}, r))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case event := <-sub.Ch:
		if event.Type != broker.EventBeforeRequestSend {
			t.Fatalf("unexpected event type: %v", event.Type)
		}

	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for hub event")
	}
}
