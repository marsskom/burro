# Creating a Burro Plugin

Burro uses a plugin-based architecture. Every plugin must implement the base `plugin.Plugin` interface and may optionally implement one or more hook interfaces.

## Minimal Plugin

```go
package myplugin

import (
	"gitlab.com/marsskom/burro/internal/plugin"
	"gitlab.com/marsskom/burro/internal/pluginapi"
)

func init() {
	plugin.Register("myplugin", func() plugin.Plugin {
		return New()
	})
}

type Config struct {
	Enabled  *bool `yaml:"enabled"`
	Priority int   `yaml:"priority"`
}

type Plugin struct {
	enabled  *bool
	priority int

	rt pluginapi.Runtime
}

func New() *Plugin {
	return &Plugin{}
}

func (p *Plugin) Name() string {
	return "myplugin"
}

func (p *Plugin) Init(rt pluginapi.Runtime, cfg any) error {
	p.rt = rt

	var config Config
	if err := plugin.DecodeYAML(cfg, &config); err != nil {
		return err
	}

	p.enabled = config.Enabled
	p.priority = config.Priority

	return nil
}

func (p *Plugin) Shutdown() error {
	return nil
}

func (p *Plugin) Enabled() *bool {
	return p.enabled
}

func (p *Plugin) Priority() int {
	return p.priority
}
```

---

# Plugin Registration

Plugins register themselves during package initialization:

```go
func init() {
	plugin.Register("myplugin", func() plugin.Plugin {
		return New()
	})
}
```

But you need manually add them into registry plugin `plugins/registry/main.go`:

```go
package registry

import (
	_ "gitlab.com/marsskom/burro/plugins/harexport"
	_ "gitlab.com/marsskom/burro/plugins/logger"
	_ "gitlab.com/marsskom/burro/plugins/luaplugin"
	_ "gitlab.com/marsskom/burro/plugins/policy"
)
```

You may see `cmd/burro/proxy.go` uses it for plugin registration:

```go
	_ "gitlab.com/marsskom/burro/plugins/registry"
```

The registration name is used in the Burro configuration file.

---

# Runtime Access

The runtime object is provided during initialization:

```go
func (p *Plugin) Init(rt pluginapi.Runtime, cfg any) error {
	p.rt = rt
	return nil
}
```

Runtime provides access to services such as:

```go
p.rt.Log().Info("plugin started")
```

Artifact storage:

```go
p.rt.Artifacts().Create("output.txt")
```

Plugin's data storate to read an additional setting or configuration, under `%workdir%/plugins/%plugin-name%/` directory:

```go
p.rt.Data.Read("data/whitelist.txt")
```

Key-Value storage where plugin may set, get, delete, list values.

And Event-Bus to emit and subscribe for events.

Please, pay attention that Event-Bus is the one on a whole system that means plugins may connect between each other with the events.

---

# Configuration

Plugin configuration is decoded from YAML.

Example configuration structure:

```go
type Config struct {
	Enabled  *bool `yaml:"enabled"`
	Priority int   `yaml:"priority"`

	Output string `yaml:"output"`
}
```

Decode configuration inside `Init()`:

```go
var config Config

if err := plugin.DecodeYAML(cfg, &config); err != nil {
	return err
}
```

Example YAML:

```yaml
plugins:
  myplugin:
    enabled: true
    priority: 100
    output: output.txt
```

Focusing on `enabled` and `priority` you must know that it is required fields that have default values:

- enabled - `false`
- priority - 100

---

# Available Hooks

A plugin only needs to implement hooks it wants to use.

## Connection Hook

```go
type ConnectHook interface {
	OnConnect(ctx *model.RequestContext) error
}
```

Called when a client connection is established.

Example:

```go
func (p *Plugin) OnConnect(ctx *model.RequestContext) error {
	p.rt.Log().Info("new connection", "id", ctx.ID)
	return nil
}
```

---

## Request Hooks

```go
type RequestHook interface {
	OnBeforeRequestSend(ctx *model.RequestContext) error
	OnAfterRequestSend(ctx *model.RequestContext) error
}
```

### Before Request

Called before Burro sends the request upstream.

```go
func (p *Plugin) OnBeforeRequestSend(ctx *model.RequestContext) error {
	return nil
}
```

### After Request

Called after the request has been sent.

```go
func (p *Plugin) OnAfterRequestSend(ctx *model.RequestContext) error {
	return nil
}
```

---

## Response Hooks

```go
type ResponseHook interface {
	OnBeforeResponseSend(ctx *model.RequestContext) error
	OnAfterResponseSend(ctx *model.RequestContext) error
}
```

### Before Response

Called before the response is returned to the client.

```go
func (p *Plugin) OnBeforeResponseSend(ctx *model.RequestContext) error {
	return nil
}
```

### After Response

Called after the response has been sent.

```go
func (p *Plugin) OnAfterResponseSend(ctx *model.RequestContext) error {
	return nil
}
```

---

## Error Hook

```go
type ErrorHook interface {
	OnError(ctx *model.RequestContext, err error) error
}
```

Called when request processing fails.

```go
func (p *Plugin) OnError(ctx *model.RequestContext, err error) error {
	p.rt.Log().Error("request failed", "error", err)

	return nil
}
```

---

## Close Hook

```go
type CloseHook interface {
	OnClose(ctx *model.RequestContext) error
}
```

Called when request processing is completed and the context is about to be released.

```go
func (p *Plugin) OnClose(ctx *model.RequestContext) error {
	return nil
}
```

---

## WebSocket Hooks

```go
type WebSocketHook interface {
	OnWSOpen(ctx *model.RequestContext) error
	OnWSMessage(ctx *model.RequestContext, msg *model.WSMessage) error
	OnWSClose(ctx *model.RequestContext) error
}
```

Example:

```go
func (p *Plugin) OnWSMessage(
	ctx *model.RequestContext,
	msg *model.WSMessage,
) error {
	p.rt.Log().Debug("websocket message")

	return nil
}
```

---

# Export Plugins

Export plugins are special plugins that collect data during traffic processing and write results when Burro performs an export operation.

In addition to normal hooks, export plugins implement:

```go
type Exporter interface {
	Flush(opts *export.FileNameVars) error
}
```

Example:

```go
func (p *Plugin) Flush(opts *export.FileNameVars) error {
	file, err := p.rt.Artifacts().Create("export.json")
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.Write([]byte("{}"))

	return err
}
```

`Flush()` is typically responsible for:

- generating reports
- exporting traffic
- saving captured data
- writing artifacts

The `FileNameVars` structure currently provides:

```go
type FileNameVars struct {
	Session string
}
```

which can be used for filename templating:

```go
filename := strings.ReplaceAll(
	"%session%.json",
	"%session%",
	opts.Session,
)
```

---

# Request Context

Most hooks receive:

```go
ctx *model.RequestContext
```

which contains request and response snapshots.

Examples:

```go
ctx.RequestSnapshot.Method
ctx.RequestSnapshot.URL
ctx.RequestSnapshot.Headers
ctx.RequestSnapshot.Body
```

and

```go
ctx.ResponseSnapshot.StatusCode
ctx.ResponseSnapshot.Headers
ctx.ResponseSnapshot.Body
```

The context also provides timing information and a unique request ID:

```go
ctx.ID
ctx.StartTime
```

---

# Best Practices

- Keep hook handlers fast.
- Use mutexes when storing shared state.
- Avoid modifying requests or responses unless necessary.
- Clean up resources in `Shutdown()`.
- Use `Flush()` only for export plugins.
- Use artifact storage instead of writing directly to the filesystem.
- Log important events through `Runtime.Log()`.

---

# Typical Plugin Lifecycle

```text
Init()
    │
    ├── OnConnect()
    │
    ├── OnBeforeRequestSend()
    ├── OnAfterRequestSend()
    │
    ├── OnBeforeResponseSend()
    ├── OnAfterResponseSend()
    │
    ├── OnError()        (optional)
    │
    ├── OnClose()
    │
    ├── Flush()          (export plugins only)
    │
Shutdown()
```
