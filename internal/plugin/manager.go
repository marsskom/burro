package plugin

import (
	"log/slog"
	"sort"

	"gitlab.com/marsskom/burro/internal/request"
)

type Manager struct {
	plugins []pluginMeta
}

type pluginMeta struct {
	plugin   Plugin
	priority int
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Register(p Plugin) {
	slog.Debug("Register plugin", "plugin", p.Name())

	plugin := pluginMeta{
		plugin:   p,
		priority: getPriority(p),
	}

	m.plugins = append(m.plugins, plugin)

	m.sort()
}

func (m *Manager) sort() {
	sort.Slice(m.plugins, func(i, j int) bool {
		return m.plugins[i].priority < m.plugins[j].priority
	})
}

func (m *Manager) EmitConnect(ctx *request.RequestContext) error {
	for _, p := range m.plugins {
		slog.Debug("EmitConnect: try plugin", "name", p.plugin.Name())

		if h, ok := p.plugin.(ConnectHook); ok {
			err := h.OnConnect(ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) EmitRequest(ctx *request.RequestContext) error {
	for _, p := range m.plugins {
		slog.Debug("EmitRequest: try plugin", "name", p.plugin.Name())

		if h, ok := p.plugin.(RequestHook); ok {
			err := h.OnRequest(ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) EmitResponse(ctx *request.RequestContext) error {
	for _, p := range m.plugins {
		slog.Debug("EmitResponse: try plugin", "name", p.plugin.Name())

		if h, ok := p.plugin.(ResponseHook); ok {
			err := h.OnResponse(ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) EmitError(ctx *request.RequestContext, err error) error {
	for _, p := range m.plugins {
		slog.Debug("EmitError: try plugin", "name", p.plugin.Name())

		if h, ok := p.plugin.(ErrorHook); ok {
			err := h.OnError(ctx, err)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) EmitClose(ctx *request.RequestContext) error {
	for _, p := range m.plugins {
		slog.Debug("EmitClose: try plugin", "name", p.plugin.Name())

		if h, ok := p.plugin.(CloseHook); ok {
			err := h.OnClose(ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
