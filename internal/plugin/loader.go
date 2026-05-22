package plugin

import (
	"fmt"
	"log/slog"

	"gitlab.com/marsskom/burro/internal/config"
)

func LoadPlugins(cfg *config.Config, pm *Manager) error {
	for name, pluginCfg := range cfg.Plugins {
		slog.Debug("Try to init plugin", "plugin", name)

		factory, ok := registry[name]
		if !ok {
			slog.Warn("Plugin is not in registry", "plugin", name)

			continue
		}

		p := factory()
		if err := p.Init(pluginCfg); err != nil {
			return fmt.Errorf("LoadPlugins: cannot init plugin: %w", err)
		}

		pm.Register(p)
	}

	return nil
}
