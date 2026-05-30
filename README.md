# Burro

Burro is a modular security and traffic inspection tool that allows extending its behavior through a plugin-based architecture.

---

## Features

- HTTP proxy with interception capabilities
- Plugin-based architecture (Go plugins + YAML configuration)
- Per-plugin configuration isolation
- Traffic export (e.g. HAR-like logs)
- Certificate-based HTTPS interception

---

## Configuration

Burro uses a main configuration file located at:

```text
config.yml
```

You can override its location via environment variable:

```text
BURRO_CONFIG=config.yml
```

---

## Core configuration

```yml
core:
  log_level: debug

  plugins:
    dir: "./plugins"
    config: "config.yml"

proxy:
  port: 8080
  host: 0.0.0.0

plugins:
  logger:

  policy:

  harexport:
    file: "./bin/%session%-%datetime%.har"
    override: true
```

---

## Plugin configuration model

Each plugin can have its own configuration.

### Global plugin config (from config.yml)

```yml
plugins:
  policy:
    enabled: true
    priority: 10
```

### Local plugin config

Plugins may also define their own configuration file:

```text
plugins/policy/config.yml
```

Example:

```yml
priority: 10
whitelist: ./data/policy/whitelist.txt
blacklist: ./data/policy/blacklist.txt
```

This allows separation of:

- global orchestration config
- plugin-specific logic config

---

## Plugin system

Burro uses a plugin-based architecture.

Plugins:

- are loaded from `plugins/` directory
- are configured via YAML
- communicate through the public plugin API (hooks/events/IPC)

### Plugin isolation rules

Plugins MUST NOT:

- access internal Burro state directly
- bypass plugin API
- depend on internal implementation details

Plugins MAY:

- react to hook events
- process requests/responses
- export data
- use their own configuration files

---

## License

Burro core is licensed under the GNU General Public License v3.0 (GPLv3). This applies to all source code included in the main repository, including bundled plugins located in the `plugins/` directory, which are considered part of the Burro codebase and distributed under GPLv3.

External plugins that are developed and distributed independently of this repository are not considered part of Burro itself if they interact only through the documented Plugin API (including hook events and IPC-based communication such as stdin/stdout, HTTP, or similar mechanisms). Such external plugins are independent works and may be licensed under any terms chosen by their authors, including permissive licenses (e.g. MIT, Apache-2.0) or proprietary licenses.

More details here:

- [PLUGIN_EXCEPTION.md](./PLUGIN_EXCEPTION.md)
- [PLUGINS](./PLUGINS.md)

--

## Running Burro

### Build

```shell
make build
```

Binary will be available in:

```text
./bin/burro-proxy
```

### Run proxy

You may just run proxy without building it:

- `make urn` - runs proxy without saving any artifacts (db, plugins' files)
- `make run` - runs new empty session, but at the end you may save it, even to existed one
- `make run ARGS="-w workspace-name"` - loads workspace from db and continue session under this workspace

Or use raw commands:

- `go run ./cmd/proxy` - (`burro`) - same as `make urn`
- `go run ./cmd/proxy -i` - (`burro -i`) - same as `make run`
- `go run ./cmd/proxy -i -w workspace-name` - (`burro -i -w workspace-name`) - same as `make run ARGS="-w workspace-name"` (flag `-i` is optional, depends do you want to save a session afterwards)

### Artifacts

By default, Burro on exit (`CTRL+C`) asks if you want to save session.

The session is a basically workspace in Burro and if you agreed to save it and provided a workspace's name it will create SQLite db file with this name under `./bin` directory.

Moreover, some plugins also may create some artifacts.

For instance, HAR export plugin creates HAR report file under `./bin` directory by default.

Even you chose do not save workspace since HAR export plugin must be configured in config file.

### Browser usage

Of course, you may use raw `curl` just to test Burro.

However, `make browser` command provides to you Chromium browser, ready to go.

The only requirement here is - Chromium must be installed in your system already.

---

## TLS interception

No one modern web portal works without HTTPS, means you Burro need a CA certificate installed in your system as allowed and trusted.

For MacOS you may consider to look at commands in the `Makefile`:

- `make certs` - (`certgen`) - generates certificates
- `make ca-install`
- `make ca-remove`
- `make ca-find` - shows if certificate was found in the OS

For other OS, please, read respective documentation.

---

## Docker

`Makefile` provides additional docker commands for safety.

If you just want to try Burro in isolated environment.

---

## Notes on architecture

- Burro is not a passive proxy only — it is a plugin execution runtime for HTTP traffic
- Plugins define behavior, not the core
- Core is minimal and stable, plugins evolve independently

---

### Plugin Development

All plugins that use native hooks are located under `plugins` directory.

You may create a directory for your own plugin, also place a separate `config.yml` file into the directory.

But not forget to add plugin name (directory name) into global `config.yml` as well:

```yml
plugins:
  your-plugin-name:
```

### Plugin Requirements

The `Plugin` is basically an interface (`./internal/plugin/plugin.go`):

```golang
type Plugin interface {
	Name() string
	Init(cfg any) error
}
```

Therefore here some requirements for each plugin for now.

Let's look at Policy plugin as an example:

```golang
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
		return fmt.Errorf("Policy Plugin Init: cannot read plugin config: %w", err)
	}

	p.priority = config.Priority

	if config.Whitelist != "" {
		whitelist, err := LoadDomains(config.Whitelist)
		if err != nil {
			return fmt.Errorf("Policy Plugin Init: cannot load whitelist: %w", err)
		}

		p.whitelist = whitelist
	}

	if config.Blacklist != "" {
		blacklist, err := LoadDomains(config.Blacklist)
		if err != nil {
			return fmt.Errorf("Policy Plugin Init: cannot load blacklist: %w", err)
		}

		p.blacklist = blacklist
	}

	return nil
}

func (p *PolicyPlugin) OnRequest(ctx *model.RequestContext) error {
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
```

- `init()` function that registers plugin under the same name as directory and its name in `config.yml`
- Plugin has its struct with fields, aldo may have some config struct if it depends on config fields
- `* Name()` function returns the same plugin name
- `* Init()` function sets plugin's settings according to its config
- Optional `* Priority()` and `* Enabled()` functions are common for all plugins, if not defined, default values are used (`./internal/plugin/priority.go`)
- And the hooks' functions only on those that the plugin expects, like `* OnRequest()` in this case

### Hooks

All hooks available in `./internal/plugin/plugin.go`:

```golang
type ConnectHook interface {
	OnConnect(ctx *model.RequestContext) error
}

type RequestHook interface {
	OnRequest(ctx *model.RequestContext) error
}

type ResponseHook interface {
	OnResponse(ctx *model.RequestContext) error
}

type ErrorHook interface {
	OnError(ctx *model.RequestContext, err error) error
}

type CloseHook interface {
	OnClose(ctx *model.RequestContext) error
}
```
