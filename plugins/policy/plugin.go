package policy

import (
	"errors"
	"fmt"

	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/plugin"
	"gitlab.com/marsskom/burro/internal/pluginapi"
	"gitlab.com/marsskom/burro/plugins/policy/actions"
	"gitlab.com/marsskom/burro/plugins/policy/domain"
	"gitlab.com/marsskom/burro/plugins/policy/response"
)

func init() {
	plugin.Register("policy", func() plugin.Plugin {
		return New()
	})
}

type PolicyConfig struct {
	Enabled   *bool  `yaml:"enabled"`
	Priority  int    `yaml:"priority"`
	Whitelist string `yaml:"whitelist"`
	Blacklist string `yaml:"blacklist"`
	ActionDir string `yaml:"action_dir"`
}

type PolicyPlugin struct {
	enabled     *bool
	priority    int
	whitelist   []string
	blacklist   []string
	actionRules []actions.ActionRule

	rt pluginapi.Runtime
}

func New() *PolicyPlugin {
	return &PolicyPlugin{}
}

func (p *PolicyPlugin) Enabled() *bool {
	return p.enabled
}

func (p *PolicyPlugin) Priority() int {
	return p.priority
}

func (p *PolicyPlugin) Name() string {
	return "policy"
}

func (p *PolicyPlugin) Init(rt pluginapi.Runtime, cfg any) error {
	p.rt = rt

	p.rt.Log().Debug("policy plugin is going to init with config", "config", cfg)

	var config PolicyConfig
	if err := plugin.DecodeYAML(cfg, &config); err != nil {
		return fmt.Errorf("policy: cannot read plugin config: %w", err)
	}

	p.enabled = config.Enabled
	p.priority = config.Priority

	if config.Whitelist != "" {
		f, err := p.rt.Data().Read(config.Whitelist)
		if err != nil {
			return fmt.Errorf("policy: cannot read whitelist file: %w", err)
		}

		whitelist, err := domain.LoadDomains(f)
		if err != nil {
			f.Close()

			return fmt.Errorf("policy: cannot load whitelist: %w", err)
		}
		f.Close()

		p.whitelist = whitelist
	}

	if config.Blacklist != "" {
		f, err := p.rt.Data().Read(config.Blacklist)
		if err != nil {
			return fmt.Errorf("policy: cannot read blacklist file: %w", err)
		}

		blacklist, err := domain.LoadDomains(f)
		if err != nil {
			f.Close()

			return fmt.Errorf("policy: cannot load blacklist: %w", err)
		}
		f.Close()

		p.blacklist = blacklist
	}

	if config.ActionDir != "" {
		fileList, err := p.rt.Data().List(config.ActionDir, []string{".yml", ".yaml"})
		if err != nil {
			return fmt.Errorf("policy: error on reading action dir '%s': %w", config.ActionDir, err)
		}

		actionRules, err := actions.LoadActionRules(p.rt.Data(), fileList)
		if err != nil {
			return fmt.Errorf("policy: load actions error: %w", err)
		}

		var errs []error
		for _, ar := range actionRules {
			err = ar.Validate()
			if err != nil {
				errs = append(errs, err)
			}
		}

		err = errors.Join(errs...)
		if err != nil {
			return fmt.Errorf("policy: error in action rule: %w", err)
		}

		p.actionRules = actionRules
	}

	return nil
}

func (p *PolicyPlugin) OnRequest(ctx *model.RequestContext) error {
	if len(p.whitelist) > 0 && domain.Match(ctx.Request.Host, p.whitelist) {
		p.rt.Log().Debug("request host was found in whitelist", "host", ctx.Request.Host)

		return nil
	}

	if len(p.blacklist) > 0 && domain.Match(ctx.Request.Host, p.blacklist) {
		p.rt.Log().Debug("request host was found in blacklist", "host", ctx.Request.Host)

		resp := response.Forbidden()
		snapshot, err := model.MakeResponseSnapshot(resp, ctx.Timings)
		if err != nil {
			return fmt.Errorf("cannot create response snapshot: %w", err)
		}

		ctx.Finish(resp, snapshot)

		return nil
	}

	resp := actions.Execute(p.rt.Log(), p.actionRules, ctx.Request)
	if resp != nil {
		snapshot, err := model.MakeResponseSnapshot(resp, ctx.Timings)
		if err != nil {
			return fmt.Errorf("cannot create response snapshot: %w", err)
		}

		ctx.Finish(resp, snapshot)
	}

	return nil
}
