package policy

import (
	"fmt"

	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/plugin"
	"gitlab.com/marsskom/burro/internal/pluginapi"
	"gitlab.com/marsskom/burro/internal/response"
)

func init() {
	plugin.Register("policy", func() plugin.Plugin {
		return New()
	})
}

type PolicyConfig struct {
	Priority  int    `yaml:"priority"`
	Whitelist string `yaml:"whitelist"`
	Blacklist string `yaml:"blacklist"`
}

type PolicyPlugin struct {
	priority  int
	whitelist []string
	blacklist []string

	rt pluginapi.Runtime
}

func New() *PolicyPlugin {
	return &PolicyPlugin{}
}

func (p *PolicyPlugin) Priority() int {
	return p.priority
}

func (p *PolicyPlugin) Name() string {
	return "policy"
}

func (p *PolicyPlugin) Init(rt pluginapi.Runtime, cfg any) error {
	p.rt = rt

	p.rt.Log().Debug("Policy plugin is going to init with config", "config", cfg)

	var config PolicyConfig
	if err := plugin.DecodeYAML(cfg, &config); err != nil {
		return fmt.Errorf("Policy Plugin Init: cannot read plugin config: %w", err)
	}

	p.priority = config.Priority

	if config.Whitelist != "" {
		f, err := p.rt.Data().Read(config.Whitelist)
		if err != nil {
			return fmt.Errorf("Policy Plugin Init: cannot read whitelist file: %w", err)
		}

		whitelist, err := LoadDomains(f)
		if err != nil {
			f.Close()

			return fmt.Errorf("Policy Plugin Init: cannot load whitelist: %w", err)
		}
		f.Close()

		p.whitelist = whitelist
	}

	if config.Blacklist != "" {
		f, err := p.rt.Data().Read(config.Blacklist)
		if err != nil {
			return fmt.Errorf("Policy Plugin Init: cannot read blacklist file: %w", err)
		}

		blacklist, err := LoadDomains(f)
		if err != nil {
			f.Close()

			return fmt.Errorf("Policy Plugin Init: cannot load blacklist: %w", err)
		}
		f.Close()

		p.blacklist = blacklist
	}

	return nil
}

func (p *PolicyPlugin) OnRequest(ctx *model.RequestContext) error {
	if len(p.whitelist) > 0 && Match(ctx.Request.Host, p.whitelist) {
		p.rt.Log().Debug("Request host was found in whitelist", "host", ctx.Request.Host)

		return nil
	}

	if len(p.blacklist) > 0 && Match(ctx.Request.Host, p.blacklist) {
		p.rt.Log().Debug("Request host was found in blacklist", "host", ctx.Request.Host)

		ctx.Finish(response.Forbidden())
	}

	return nil
}
