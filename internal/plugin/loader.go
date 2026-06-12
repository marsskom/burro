package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"gitlab.com/marsskom/burro/internal/cli"
	"gitlab.com/marsskom/burro/internal/config"
	"gitlab.com/marsskom/burro/internal/logger"
	"gitlab.com/marsskom/burro/internal/pluginapi"
	rt "gitlab.com/marsskom/burro/internal/runtime"
	"gopkg.in/yaml.v3"
)

const (
	EventLoadPluginProgress string = "plugin_load_progress"
)

func LoadPlugins(cliIO cli.IO, paths *config.Paths, cfg *config.Config, pm *Manager) error {
	renderer := ProgressRenderer{
		cliIO: cliIO,
	}
	progress := GlobalProgress{
		Total:   0,
		Plugins: make(map[string]Progress),
	}

	eventBus := rt.NewEventBus()
	coreLogger := rt.NewCoreLogger()

	for name, pluginCfg := range cfg.Plugins {
		logger.Debug("try to init plugin", "plugin", name)

		factory, ok := registry[name]
		if !ok {
			logger.Warn("plugin is not in registry", "plugin", name)

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

		if !isPluginEnabled(pluginCfg) {
			continue
		}

		progress.Total = progress.Total + 1
		progress.Current = progress.Current + 1
		progress.Plugins[name] = Progress{
			Name:    name,
			Current: 1,
			Total:   1,
		}

		off := eventBus.On(EventLoadPluginProgress, func(event pluginapi.Event) {
			if event.From != name {
				return
			}

			m, ok := event.Data.(map[string]int)
			if !ok {
				return
			}

			p := progress.Plugins[name]

			cur, ok := m["current"]
			if ok {
				p.Current = cur
			}

			total, ok := m["total"]
			if ok {
				p.Total = total
			}

			progress.Plugins[name] = p

			renderer.Render(progress)
		})

		if err := p.Init(NewRuntime(
			rt.NewFileArtifactStore(filepath.Join(paths.Home, "artifacts", name)),
			rt.NewPluginDataStore(filepath.Join(paths.Home, cfg.Core.Plugins.Dir, name)),
			rt.NewKeyValue(),
			eventBus,
			coreLogger,
		), pluginCfg); err != nil {
			off()

			return fmt.Errorf("cannot init plugin: %w", err)
		}

		pm.Register(p)

		off()

		renderer.Render(progress)
	}

	pm.Sort()

	renderer.Render(progress)

	return nil
}

func resolvePluginConfig(home string, cfg config.CorePluginsConfig, name string) (any, error) {
	path := filepath.Join(home, cfg.Dir, name, cfg.Config)
	logger.Debug("try to find separate plugin config file", "path", path)

	if _, err := os.Stat(path); err == nil {
		logger.Info("separate plugin config has been found and is going to be used", "plugin", name, "path", path)

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

	logger.Debug("separate config for plugin wasn't found", "plugin", name, "path", path)

	return nil, nil
}
