package plugin

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gitlab.com/marsskom/burro/internal/config"
	"gopkg.in/yaml.v3"
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

		sepCfg, err := resolvePluginConfig(&cfg.Core.Plugins, name)
		if err != nil {
			return err
		}

		if sepCfg != nil {
			pluginCfg = sepCfg
		}

		if err := p.Init(pluginCfg); err != nil {
			return fmt.Errorf("cannot init plugin: %w", err)
		}

		pm.Register(p)
	}

	return nil
}

func resolvePluginConfig(cfg *config.CorePluginsConfig, name string) (any, error) {
	path := filepath.Join(cfg.Dir, name, cfg.Config)
	slog.Debug("Try to find separate plugin config file", "path", path)

	if _, err := os.Stat(path); err == nil {
		slog.Info("Separate plugin config has been found and is going to be used", "path", path, "name", name)

		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("cannot read plugin config file: %w", err)
		}

		var pCfg any
		if err := yaml.Unmarshal(data, &pCfg); err != nil {
			return nil, fmt.Errorf("cannot unmarshall plugin config file: %w", err)
		}

		return pCfg, nil
	}

	slog.Debug("Separate config for plugin wasn't found", "name", name)

	return nil, nil
}
