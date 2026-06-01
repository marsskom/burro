package plugin

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gitlab.com/marsskom/burro/internal/config"
	rt "gitlab.com/marsskom/burro/internal/runtime"
	"gopkg.in/yaml.v3"
)

func LoadPlugins(paths *config.Paths, cfg *config.Config, pm *Manager) error {
	for name, pluginCfg := range cfg.Plugins {
		slog.Debug("try to init plugin", "plugin", name)

		factory, ok := registry[name]
		if !ok {
			slog.Warn("plugin is not in registry", "plugin", name)

			continue
		}

		p := factory()

		sepCfg, err := resolvePluginConfig(paths.Home, cfg.Core.Plugins, name)
		if err != nil {
			return err
		}

		if sepCfg != nil {
			pluginCfg = sepCfg
		}

		if err := p.Init(NewRuntime(
			rt.NewFileArtifactStore(filepath.Join(paths.Home, "artifacts", name)),
			rt.NewPluginDataStore(filepath.Join(paths.Home, cfg.Core.Plugins.Dir, name)),
			rt.NewKeyValue(),
			rt.NewEventBus(),
			rt.NewCoreLogger(),
		), pluginCfg); err != nil {
			return fmt.Errorf("cannot init plugin: %w", err)
		}

		pm.Register(p)
	}

	return nil
}

func resolvePluginConfig(home string, cfg config.CorePluginsConfig, name string) (any, error) {
	path := filepath.Join(home, cfg.Dir, name, cfg.Config)
	slog.Debug("try to find separate plugin config file", "path", path)

	if _, err := os.Stat(path); err == nil {
		slog.Info("separate plugin config has been found and is going to be used", "plugin", name, "path", path)

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

	slog.Debug("separate config for plugin wasn't found", "plugin", name, "path", path)

	return nil, nil
}
