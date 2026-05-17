package plugin

import "gitlab.com/marsskom/burro/internal/events"

type Manager struct {
	plugins []Plugin
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Register(p Plugin) {
	m.plugins = append(m.plugins, p)
}

func (m *Manager) EmitConnect(ctx *events.Context) error {
	for _, p := range m.plugins {
		if h, ok := p.(ConnectHook); ok {
			err := h.OnConnect(ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) EmitRequest(ctx *events.Context) error {
	for _, p := range m.plugins {
		if h, ok := p.(RequestHook); ok {
			err := h.OnRequest(ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) EmitResponse(ctx *events.Context) error {
	for _, p := range m.plugins {
		if h, ok := p.(ResponseHook); ok {
			err := h.OnResponse(ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) EmitError(ctx *events.Context) error {
	for _, p := range m.plugins {
		if h, ok := p.(ErrorHook); ok {
			err := h.OnError(ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) EmitClose(ctx *events.Context) error {
	for _, p := range m.plugins {
		if h, ok := p.(CloseHook); ok {
			err := h.OnClose(ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
