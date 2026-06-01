package plugin

import (
	"errors"
	"fmt"
	"log/slog"
	"sort"

	"gitlab.com/marsskom/burro/internal/broker"
	"gitlab.com/marsskom/burro/internal/export"
	"gitlab.com/marsskom/burro/internal/model"
)

type Manager struct {
	hub     *broker.Hub
	plugins []pluginMeta
}

type pluginMeta struct {
	plugin   Plugin
	enabled  bool
	priority int
}

func NewManager(hub *broker.Hub) *Manager {
	return &Manager{
		hub: hub,
	}
}

func (m *Manager) Register(p Plugin) {
	slog.Debug("register plugin", "plugin", p.Name())

	plugin := pluginMeta{
		plugin:   p,
		enabled:  getEnabled(p),
		priority: getPriority(p),
	}

	slog.Debug("priority", "priority", plugin.priority)

	m.plugins = append(m.plugins, plugin)

	m.sort()
}

func (m *Manager) sort() {
	sort.Slice(m.plugins, func(i, j int) bool {
		return m.plugins[i].priority < m.plugins[j].priority
	})
}

func (m *Manager) EmitExportPluginsFlush(opts *export.FileNameVars) error {
	var errs []error

	for _, p := range m.plugins {
		slog.Debug("EmitExportPluginsFlush: try plugin", "name", p.plugin.Name())
		if !p.enabled {
			slog.Debug("EmitExportPluginsFlush: plugin is disabled and ignores", "name", p.plugin.Name())

			continue
		}

		if e, ok := p.plugin.(export.Exporter); ok {
			if err := e.Flush(opts); err != nil {
				errs = append(errs, fmt.Errorf("Plugin Manager: error on EmitExportPluginsFlush: %w", err))
			}
		}
	}

	return errors.Join(errs...)
}

func (m *Manager) EmitConnect(ctx *model.RequestContext) error {
	var errs []error

	for _, p := range m.plugins {
		slog.Debug("EmitConnect: try plugin", "name", p.plugin.Name())
		if !p.enabled {
			slog.Debug("EmitConnect: plugin is disabled and ignores", "name", p.plugin.Name())

			continue
		}

		if h, ok := p.plugin.(ConnectHook); ok {
			if err := h.OnConnect(ctx); err != nil {
				errs = append(errs, fmt.Errorf("Plugin Manager: error on Connect hook: %w", err))
			}
		}
	}

	event, err := broker.ToBrokerEvent(broker.EventConnect, ctx)
	if err != nil {
		slog.Error("EmitConnect: cannot convert context to broker event", "err", err)
		errs = append(errs, err)
	} else {
		m.hub.Publish(event)
	}

	return errors.Join(errs...)
}

func (m *Manager) EmitRequest(ctx *model.RequestContext) error {
	var errs []error

	for _, p := range m.plugins {
		slog.Debug("EmitRequest: try plugin", "name", p.plugin.Name())
		if !p.enabled {
			slog.Debug("EmitRequest: plugin is disabled and ignores", "name", p.plugin.Name())

			continue
		}

		if h, ok := p.plugin.(RequestHook); ok {
			if err := h.OnRequest(ctx); err != nil {
				errs = append(errs, fmt.Errorf("Plugin Manager: error on Request hook: %w", err))
			}
		}
	}

	event, err := broker.ToBrokerEvent(broker.EventRequest, ctx)
	if err != nil {
		slog.Error("EmitRequest: cannot convert context to broker event", "err", err)
		errs = append(errs, err)
	} else {
		m.hub.Publish(event)
	}

	return errors.Join(errs...)
}

func (m *Manager) EmitResponse(ctx *model.RequestContext) error {
	var errs []error

	for _, p := range m.plugins {
		slog.Debug("EmitResponse: try plugin", "name", p.plugin.Name())
		if !p.enabled {
			slog.Debug("EmitResponse: plugin is disabled and ignores", "name", p.plugin.Name())

			continue
		}

		if h, ok := p.plugin.(ResponseHook); ok {
			if err := h.OnResponse(ctx); err != nil {
				errs = append(errs, fmt.Errorf("Plugin Manager: error on Response hook: %w", err))
			}
		}
	}

	event, err := broker.ToBrokerEvent(broker.EventResponse, ctx)
	if err != nil {
		slog.Error("EmitResponse: cannot convert context to broker event", "err", err)
		errs = append(errs, err)
	} else {
		m.hub.Publish(event)
	}

	return errors.Join(errs...)
}

func (m *Manager) EmitError(ctx *model.RequestContext, err error) error {
	var errs []error

	for _, p := range m.plugins {
		slog.Debug("EmitError: try plugin", "name", p.plugin.Name())
		if !p.enabled {
			slog.Debug("EmitError: plugin is disabled and ignores", "name", p.plugin.Name())

			continue
		}

		if h, ok := p.plugin.(ErrorHook); ok {
			if err := h.OnError(ctx, err); err != nil {
				errs = append(errs, fmt.Errorf("Plugin Manager: error on Error hook: %w", err))
			}
		}
	}

	event, err := broker.ToBrokerEvent(broker.EventError, ctx)
	if err != nil {
		slog.Error("EmitError: cannot convert context to broker event", "err", err)
		errs = append(errs, err)
	} else {
		m.hub.Publish(event)
	}

	return errors.Join(errs...)
}

func (m *Manager) EmitClose(ctx *model.RequestContext) error {
	var errs []error

	for _, p := range m.plugins {
		slog.Debug("EmitClose: try plugin", "name", p.plugin.Name())
		if !p.enabled {
			slog.Debug("EmitClose: plugin is disabled and ignores", "name", p.plugin.Name())

			continue
		}

		if h, ok := p.plugin.(CloseHook); ok {
			if err := h.OnClose(ctx); err != nil {
				errs = append(errs, fmt.Errorf("Plugin Manager: error on Close hook: %w", err))
			}
		}
	}

	event, err := broker.ToBrokerEvent(broker.EventClose, ctx)
	if err != nil {
		slog.Error("EmitClose: cannot convert context to broker event", "err", err)
		errs = append(errs, err)
	} else {
		m.hub.Publish(event)
	}

	return errors.Join(errs...)
}
