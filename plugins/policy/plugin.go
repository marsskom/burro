package policy

import (
	"log/slog"

	"gitlab.com/marsskom/burro/internal/plugin"
	"gitlab.com/marsskom/burro/internal/request"
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

func (p *PolicyPlugin) Init(cfg any) error {
	slog.Debug("Policy plugin is going to init with config", "config", cfg)

	var config PolicyConfig
	if err := plugin.DecodeYAML(cfg, &config); err != nil {
		return err
	}

	p.priority = config.Priority

	if config.Whitelist != "" {
		whitelist, err := LoadDomains(config.Whitelist)
		if err != nil {
			return err
		}

		p.whitelist = whitelist
	}

	if config.Blacklist != "" {
		blacklist, err := LoadDomains(config.Blacklist)
		if err != nil {
			return err
		}

		p.blacklist = blacklist
	}

	return nil
}

func (p *PolicyPlugin) OnRequest(ctx *request.RequestContext) error {
	if len(p.whitelist) > 0 && Match(ctx.Request.Host, p.whitelist) {
		slog.Debug("Request host was found in whitelist", "host", ctx.Request.Host)

		return nil
	}

	if len(p.blacklist) > 0 && Match(ctx.Request.Host, p.blacklist) {
		slog.Debug("Request host was found in blacklist", "host", ctx.Request.Host)

		ctx.Finish(response.Forbidden())
	}

	return nil
}
