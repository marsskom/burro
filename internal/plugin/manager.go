package plugin

import (
	"errors"
	"fmt"
	"sort"

	"gitlab.com/marsskom/burro/internal/broker"
	"gitlab.com/marsskom/burro/internal/export"
	"gitlab.com/marsskom/burro/internal/logger"
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
	logger.Debug("register plugin", "plugin", p.Name())

	plugin := pluginMeta{
		plugin:   p,
		enabled:  getEnabled(p),
		priority: getPriority(p),
	}

	logger.Debug("enabled", "enabled", plugin.enabled)
	logger.Debug("priority", "priority", plugin.priority)

	if !plugin.enabled {
		return
	}

	m.plugins = append(m.plugins, plugin)
}

func (m *Manager) Sort() {
	sort.Slice(m.plugins, func(i, j int) bool {
		return m.plugins[i].priority < m.plugins[j].priority
	})
}

func (m *Manager) Close() error {
	var errs []error
	for _, p := range m.plugins {
		if err := p.plugin.Shutdown(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (m *Manager) EmitExportPluginsFlush(opts *export.FileNameVars) error {
	var errs []error

	for _, p := range m.plugins {
		logger.Debug("EmitExportPluginsFlush: try plugin", "name", p.plugin.Name())
		if !p.enabled {
			logger.Debug("EmitExportPluginsFlush: plugin is disabled and ignores", "name", p.plugin.Name())

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
		logger.Debug("EmitConnect: try plugin", "name", p.plugin.Name())
		if !p.enabled {
			logger.Debug("EmitConnect: plugin is disabled and ignores", "name", p.plugin.Name())

			continue
		}

		if h, ok := p.plugin.(ConnectHook); ok {
			if err := h.OnConnect(ctx); err != nil {
				errs = append(errs, fmt.Errorf("Plugin Manager: error on Connect hook: %w", err))
			}
		}
	}

	event, err := broker.ToHTTPBrokerEvent(broker.EventConnect, ctx)
	if err != nil {
		logger.Error("EmitConnect: cannot convert context to broker event", "err", err)
		errs = append(errs, err)
	} else {
		m.hub.Publish(event)
	}

	return errors.Join(errs...)
}

func (m *Manager) EmitBeforeRequestSend(ctx *model.RequestContext) error {
	var errs []error

	for _, p := range m.plugins {
		logger.Debug("EmitBeforeRequestSend: try plugin", "name", p.plugin.Name())
		if !p.enabled {
			logger.Debug("EmitBeforeRequestSend: plugin is disabled and ignores", "name", p.plugin.Name())

			continue
		}

		if h, ok := p.plugin.(RequestHook); ok {
			if err := h.OnBeforeRequestSend(ctx); err != nil {
				errs = append(errs, fmt.Errorf("Plugin Manager: error on Before Request hook: %w", err))
			}
		}
	}

	event, err := broker.ToHTTPBrokerEvent(broker.EventBeforeRequestSend, ctx)
	if err != nil {
		logger.Error("EmitBeforeRequestSend: cannot convert context to broker event", "err", err)
		errs = append(errs, err)
	} else {
		m.hub.Publish(event)
	}

	return errors.Join(errs...)
}

func (m *Manager) EmitAfterRequestSend(ctx *model.RequestContext) error {
	var errs []error

	for _, p := range m.plugins {
		logger.Debug("EmitAfterRequestSend: try plugin", "name", p.plugin.Name())
		if !p.enabled {
			logger.Debug("EmitAfterRequestSend: plugin is disabled and ignores", "name", p.plugin.Name())

			continue
		}

		if h, ok := p.plugin.(RequestHook); ok {
			if err := h.OnAfterRequestSend(ctx); err != nil {
				errs = append(errs, fmt.Errorf("Plugin Manager: error on After Request hook: %w", err))
			}
		}
	}

	event, err := broker.ToHTTPBrokerEvent(broker.EventAfterRequestSend, ctx)
	if err != nil {
		logger.Error("EmitAfterRequestSend: cannot convert context to broker event", "err", err)
		errs = append(errs, err)
	} else {
		m.hub.Publish(event)
	}

	return errors.Join(errs...)
}

func (m *Manager) EmitBeforeResponseSend(ctx *model.RequestContext) error {
	var errs []error

	for _, p := range m.plugins {
		logger.Debug("EmitBeforeResponseSend: try plugin", "name", p.plugin.Name())
		if !p.enabled {
			logger.Debug("EmitBeforeResponseSend: plugin is disabled and ignores", "name", p.plugin.Name())

			continue
		}

		if h, ok := p.plugin.(ResponseHook); ok {
			if err := h.OnBeforeResponseSend(ctx); err != nil {
				errs = append(errs, fmt.Errorf("Plugin Manager: error on Before Response hook: %w", err))
			}
		}
	}

	event, err := broker.ToHTTPBrokerEvent(broker.EventBeforeResponseSend, ctx)
	if err != nil {
		logger.Error("EmitBeforeResponseSend: cannot convert context to broker event", "err", err)
		errs = append(errs, err)
	} else {
		m.hub.Publish(event)
	}

	return errors.Join(errs...)
}

func (m *Manager) EmitAfterResponseSend(ctx *model.RequestContext) error {
	var errs []error

	for _, p := range m.plugins {
		logger.Debug("EmitAfterResponseSend: try plugin", "name", p.plugin.Name())
		if !p.enabled {
			logger.Debug("EmitAfterResponseSend: plugin is disabled and ignores", "name", p.plugin.Name())

			continue
		}

		if h, ok := p.plugin.(ResponseHook); ok {
			if err := h.OnAfterResponseSend(ctx); err != nil {
				errs = append(errs, fmt.Errorf("Plugin Manager: error on After Response hook: %w", err))
			}
		}
	}

	event, err := broker.ToHTTPBrokerEvent(broker.EventAfterResponseSend, ctx)
	if err != nil {
		logger.Error("EmitAfterResponseSend: cannot convert context to broker event", "err", err)
		errs = append(errs, err)
	} else {
		m.hub.Publish(event)
	}

	return errors.Join(errs...)
}

func (m *Manager) EmitError(ctx *model.RequestContext, err error) error {
	var errs []error

	for _, p := range m.plugins {
		logger.Debug("EmitError: try plugin", "name", p.plugin.Name())
		if !p.enabled {
			logger.Debug("EmitError: plugin is disabled and ignores", "name", p.plugin.Name())

			continue
		}

		if h, ok := p.plugin.(ErrorHook); ok {
			if err := h.OnError(ctx, err); err != nil {
				errs = append(errs, fmt.Errorf("Plugin Manager: error on Error hook: %w", err))
			}
		}
	}

	event, err := broker.ToHTTPBrokerEvent(broker.EventError, ctx)
	if err != nil {
		logger.Error("EmitError: cannot convert context to broker event", "err", err)
		errs = append(errs, err)
	} else {
		m.hub.Publish(event)
	}

	return errors.Join(errs...)
}

func (m *Manager) EmitClose(ctx *model.RequestContext) error {
	var errs []error

	for _, p := range m.plugins {
		logger.Debug("EmitClose: try plugin", "name", p.plugin.Name())
		if !p.enabled {
			logger.Debug("EmitClose: plugin is disabled and ignores", "name", p.plugin.Name())

			continue
		}

		if h, ok := p.plugin.(CloseHook); ok {
			if err := h.OnClose(ctx); err != nil {
				errs = append(errs, fmt.Errorf("Plugin Manager: error on Close hook: %w", err))
			}
		}
	}

	event, err := broker.ToHTTPBrokerEvent(broker.EventClose, ctx)
	if err != nil {
		logger.Error("EmitClose: cannot convert context to broker event", "err", err)
		errs = append(errs, err)
	} else {
		m.hub.Publish(event)
	}

	return errors.Join(errs...)
}

func (m *Manager) EmitWSOpen(ctx *model.RequestContext) error {
	var errs []error

	for _, p := range m.plugins {
		logger.Debug("EmitWSOpen: try plugin", "name", p.plugin.Name())
		if !p.enabled {
			logger.Debug("EmitWSOpen: plugin is disabled and ignores", "name", p.plugin.Name())

			continue
		}

		if h, ok := p.plugin.(WebSocketHook); ok {
			if err := h.OnWSOpen(ctx); err != nil {
				errs = append(errs, fmt.Errorf("Plugin Manager: error on WSOpen hook: %w", err))
			}
		}
	}

	event, err := broker.ToWSBrokerEvent(broker.EventWSConnect, ctx, nil)
	if err != nil {
		logger.Error("WSOpen: cannot convert context to broker event", "err", err)
		errs = append(errs, err)
	} else {
		m.hub.Publish(event)
	}

	return errors.Join(errs...)
}

func (m *Manager) EmitWSMessage(ctx *model.RequestContext, msg *model.WSMessage) error {
	var errs []error

	for _, p := range m.plugins {
		logger.Debug("EmitWSMessage: try plugin", "name", p.plugin.Name())
		if !p.enabled {
			logger.Debug("EmitWSMessage: plugin is disabled and ignores", "name", p.plugin.Name())

			continue
		}

		if h, ok := p.plugin.(WebSocketHook); ok {
			if err := h.OnWSMessage(ctx, msg); err != nil {
				errs = append(errs, fmt.Errorf("Plugin Manager: error on WSMessage hook: %w", err))
			}
		}
	}

	event, err := broker.ToWSBrokerEvent(broker.EventWSMessage, ctx, msg)
	if err != nil {
		logger.Error("EmitWSMessage: cannot convert context to broker event", "err", err)
		errs = append(errs, err)
	} else {
		m.hub.Publish(event)
	}

	return errors.Join(errs...)
}

func (m *Manager) EmitWSClose(ctx *model.RequestContext) error {
	var errs []error

	for _, p := range m.plugins {
		logger.Debug("EmitWSClose: try plugin", "name", p.plugin.Name())
		if !p.enabled {
			logger.Debug("EmitWSClose: plugin is disabled and ignores", "name", p.plugin.Name())

			continue
		}

		if h, ok := p.plugin.(WebSocketHook); ok {
			if err := h.OnWSClose(ctx); err != nil {
				errs = append(errs, fmt.Errorf("Plugin Manager: error on WSClose hook: %w", err))
			}
		}
	}

	event, err := broker.ToWSBrokerEvent(broker.EventWSClose, ctx, nil)
	if err != nil {
		logger.Error("EmitWSClose: cannot convert context to broker event", "err", err)
		errs = append(errs, err)
	} else {
		m.hub.Publish(event)
	}

	return errors.Join(errs...)
}
