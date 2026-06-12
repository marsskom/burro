package luaplugin

import (
	"errors"
	"fmt"
	"sort"

	"gitlab.com/marsskom/burro/internal/export"
	internalLua "gitlab.com/marsskom/burro/internal/lua"
	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/plugin"
	"gitlab.com/marsskom/burro/internal/pluginapi"
)

func init() {
	plugin.Register("luaplugin", func() plugin.Plugin {
		return New()
	})
}

type LuaPluginConfig struct {
	Enabled   *bool                            `yaml:"enabled"`
	Priority  int                              `yaml:"priority"`
	ScriptDir string                           `yaml:"dir"`
	Scripts   map[string]LuaPluginScriptConfig `yaml:"scripts"`
}

type LuaPluginScriptConfig struct {
	Enabled  bool `yaml:"enabled"`
	Priority int  `yaml:"priority"`
}

type LuaPlugin struct {
	enabled  *bool
	priority int

	rt pluginapi.Runtime

	scripts    []LuaPluginScript
	scriptsDir string
	lWrapper   *internalLua.PluginWrapper
}

type LuaPluginScript struct {
	name     string
	enabled  bool
	priority int
}

func New() *LuaPlugin {
	return &LuaPlugin{}
}

func (p *LuaPlugin) Enabled() *bool {
	return p.enabled
}

func (p *LuaPlugin) Priority() int {
	return p.priority
}

func (p *LuaPlugin) Name() string {
	return "luaplugin"
}

func (p *LuaPlugin) Init(rt pluginapi.Runtime, cfg any) error {
	p.rt = rt

	p.rt.Log().Debug("lua plugin is going to init with config", "config", cfg)

	var config LuaPluginConfig
	if err := plugin.DecodeYAML(cfg, &config); err != nil {
		return fmt.Errorf("lua plugin cannot read its config: %w", err)
	}

	p.enabled = config.Enabled
	p.priority = config.Priority
	p.scriptsDir = config.ScriptDir

	// Lua wrapper.
	p.lWrapper = internalLua.NewPluginWrapper(rt, p.scriptsDir)

	// Emits plugin starts loading.
	eventData := map[string]int{
		"current": 1,
		"total":   len(config.Scripts) + 1, // lua plugin itself makes it plus 1
	}
	p.rt.Events().Emit(pluginapi.Event{
		Name: plugin.EventLoadPluginProgress,
		From: p.Name(),
		Data: eventData,
	})

	// Sets scripts.
	p.scripts = make([]LuaPluginScript, 0, len(config.Scripts))
	for k, v := range config.Scripts {
		if !v.Enabled {
			eventData["current"] = eventData["current"] + 1

			continue
		}

		p.scripts = append(p.scripts, LuaPluginScript{
			name:     k,
			enabled:  v.Enabled,
			priority: v.Priority,
		})

		err := p.lWrapper.InitState(k)
		if err != nil {
			return err
		}

		err = p.lWrapper.InitCalls(k)
		if err != nil {
			return err
		}

		eventData["current"] = eventData["current"] + 1

		p.rt.Events().Emit(pluginapi.Event{
			Name: plugin.EventLoadPluginProgress,
			From: p.Name(),
			Data: eventData,
		})
	}

	sort.Slice(p.scripts, func(i, j int) bool {
		return p.scripts[i].priority < p.scripts[j].priority
	})

	return nil
}

func (p *LuaPlugin) Shutdown() error {
	return p.lWrapper.Shutdown()
}

func (p *LuaPlugin) trigger(fn func(LuaPluginScript) error) error {
	var errs []error
	for _, s := range p.scripts {
		p.rt.Log().Debug("lua plugin tries script", "script", s.name)
		if !s.enabled {
			p.rt.Log().Debug("lua plugin ignores script because it is disabled", "script", s.name, "enabled", s.enabled)
		}

		err := fn(s)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (p *LuaPlugin) OnConnect(ctx *model.RequestContext) error {
	p.rt.Log().Debug("lua plugin connect hook is triggered")

	return p.trigger(func(s LuaPluginScript) error {
		return p.lWrapper.OnConnect(s.name, ctx)
	})
}

func (p *LuaPlugin) OnBeforeRequestSend(ctx *model.RequestContext) error {
	p.rt.Log().Debug("lua plugin before request hook is triggered")

	return p.trigger(func(s LuaPluginScript) error {
		return p.lWrapper.OnBeforeRequestSend(s.name, ctx)
	})
}

func (p *LuaPlugin) OnAfterRequestSend(ctx *model.RequestContext) error {
	p.rt.Log().Debug("lua plugin after request hook is triggered")

	return p.trigger(func(s LuaPluginScript) error {
		return p.lWrapper.OnAfterRequestSend(s.name, ctx)
	})
}

func (p *LuaPlugin) OnBeforeResponseSend(ctx *model.RequestContext) error {
	p.rt.Log().Debug("lua plugin before response hook is triggered")

	return p.trigger(func(s LuaPluginScript) error {
		return p.lWrapper.OnBeforeResponseSend(s.name, ctx)
	})
}

func (p *LuaPlugin) OnAfterResponseSend(ctx *model.RequestContext) error {
	p.rt.Log().Debug("lua plugin after response hook is triggered")

	return p.trigger(func(s LuaPluginScript) error {
		return p.lWrapper.OnAfterResponseSend(s.name, ctx)
	})
}

func (p *LuaPlugin) OnError(ctx *model.RequestContext, err error) error {
	p.rt.Log().Debug("lua plugin error hook is triggered")

	return p.trigger(func(s LuaPluginScript) error {
		return p.lWrapper.OnError(s.name, ctx, err)
	})
}

func (p *LuaPlugin) OnClose(ctx *model.RequestContext) error {
	p.rt.Log().Debug("lua plugin close hook is triggered")

	return p.trigger(func(s LuaPluginScript) error {
		return p.lWrapper.OnClose(s.name, ctx)
	})
}

func (p *LuaPlugin) Flush(opts *export.FileNameVars) error {
	p.rt.Log().Debug("lua plugin flush hook is triggered")

	return p.trigger(func(s LuaPluginScript) error {
		return p.lWrapper.Flush(s.name, opts)
	})
}

// WS.

func (p *LuaPlugin) OnWSOpen(ctx *model.RequestContext) error {
	p.rt.Log().Debug("lua plugin ws_open hook is triggered")

	return p.trigger(func(s LuaPluginScript) error {
		return p.lWrapper.OnWSOpen(s.name, ctx)
	})
}

func (p *LuaPlugin) OnWSMessage(ctx *model.RequestContext, msg *model.WSMessage) error {
	p.rt.Log().Debug("lua plugin ws_msg hook is triggered")

	return p.trigger(func(s LuaPluginScript) error {
		return p.lWrapper.OnWSMessage(s.name, ctx, msg)
	})
}

func (p *LuaPlugin) OnWSClose(ctx *model.RequestContext) error {
	p.rt.Log().Debug("lua plugin ws_close hook is triggered")

	return p.trigger(func(s LuaPluginScript) error {
		return p.lWrapper.OnWSClose(s.name, ctx)
	})
}
