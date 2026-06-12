# Lua Plugin (Burro)

Lua scripting plugin for Burro that allows runtime request/response manipulation, event handling, and persistent integration with internal system services.

---

## Overview

The Lua plugin executes user-defined scripts at specific lifecycle events of a request/response pipeline and WebSocket sessions.

Each script runs in an isolated Lua environment with access to:

- Request/response context
- Ability mutate some fields of the request/response objects
- Key-value storage
- Event bus (only emit)
- Artifacts storage
- Time utilities
- JSON/Base64 utilities
- Logging system

---

## Configuration

Example `config.yml`:

```yml
luaplugin:
  enabled: true
  priority: 90
  dir: scripts
  scripts:
    metric:
      enabled: true
      priority: 100
```

### Fields

- `enabled` — enable/disable plugin (all lua scripts respectively)
- `priority` — execution priority relative to other Go plugins
- `dir` — root directory for Lua scripts (relative to `%workdir%/plugins/luaplugin/` - regular data store directory of the plugin)
- `scripts` — per-script configuration:
  - `enabled` — enable script
  - `priority` — execution order among all lua scripts

## Script location

`%workdir%/plugins/luaplugin/scripts/%script_name%`

Lua metric script may be used as an example.

Each script contains files that corresponds to an hook name:

```text
flush.lua
connect.lua
before_request.lua
after_request.lua
before_response.lua
after_response.lua
error.lua
close.lua

ws_open.lua
ws_msg.lua
ws_close.lua
```

Creates only the files with hook names you want to use.

Specific `_shared.lua` file may be used for reusable functions or part of the logic among other lua scripts.

---

## Execution Model

Each script is triggered by a specific hook:

| Event             | Description                           |
| ----------------- | ------------------------------------- |
| `connect`         | New connection established            |
| `before_request`  | Before request is processed           |
| `after_request`   | After request is processed            |
| `before_response` | Before response is sent               |
| `after_response`  | After response is sent                |
| `error`           | Error handling hook                   |
| `close`           | Connection closed                     |
| `ws_open`         | WebSocket opened                      |
| `ws_msg`          | WebSocket message received            |
| `ws_close`        | WebSocket closed                      |
| `flush`           | Flush is triggered for export plugins |

---

## Global Lua API

For newest changes you may follow the file `%workdir%/plugins/luaplugin/scripts/_annotations.lua` for more information.

### Logging

```lua
log.trace(msg, attrs?)
log.debug(msg, attrs?)
log.info(msg, attrs?)
log.warn(msg, attrs?)
log.error(msg, attrs?)
log.audit(msg, attrs?)
```

### Context

```lua
ctx.id
ctx.session_id
ctx.is_finished
ctx.req
ctx.resp
```

### Request

```lua
ctx.req.proto
ctx.req.host
ctx.req.scheme
ctx.req.method
ctx.req.path
ctx.req.query
ctx.req.url
ctx.req.remote_addr
ctx.req.headers
ctx.req.cookies
ctx.req.body
ctx.req.cookies
ctx.req.content_length
```

### Response

```lua
ctx.resp.status
ctx.resp.status_code
ctx.resp.proto
ctx.resp.headers
ctx.resp.body
ctx.resp.content_length
```

---

## Mutation API

### Context control

```lua
mut.ctx.set_finish()
```

### Request mutation

```lua
mut.req.set_host(host)
mut.req.set_scheme(scheme)
mut.req.set_method(method)
mut.req.set_path(path)
mut.req.set_url(url)
mut.req.set_body(body)

mut.req.set_header(key, value)
mut.req.add_header(key, value)
mut.req.del_header(key)

mut.req.set_cookie(cookie)
mut.req.del_cookie(name)
```

### Response mutation

```lua
mut.resp.set_status(status)
mut.resp.set_body(body)

mut.resp.set_header(key, value)
mut.resp.add_header(key, value)
mut.resp.del_header(key)
```

---

## Key-Value Storage

```lua
kv.get(key) -> value, error
kv.get_base64(key) -> value, error
kv.set(key, value) -> ok, error
kv.delete(key) -> ok, error
kv.list(prefix) -> table<string, string>, error
```

---

## Event Bus

```lua
bus.emit(name, data?) -> error
```

---

## Artifacts Storage

```lua
artifacts.write(name, content) -> ok, error
artifacts.read(name) -> content, error
artifacts.exists(name) -> ok
artifacts.delete(name) -> ok, error
artifacts.rename(oldpath, newpath) -> ok, error
artifacts.list() -> result, error
```

---

## Data Store

Read-only structured data access:

```lua
data_store.exists(name) -> ok
data_store.read(name) -> content, error
data_store.list(path, exts?) -> result, error
```

---

## Packages

### Time

```lua
time.unix()        -- Unix timestamp
time.rfc3339()     -- RFC3339 timestamp
time.date(fmt?)    -- formatted date
```

### Base64

```lua
base64.encode(data)
base64.decode(data) -> result, error
```

### JSON

```lua
json.encode(data) -> result, error
json.decode(data) -> result, error
```

---

## Specific Script Contexts

### error.lua

```lua
---@class Err
---@field msg string

err.msg
```

### flush.lua

```lua
---@class Opts
---@field session string

opts.session
```

---

## Example Script

```lua
log.info("request started", {
    path = ctx.req.path,
    method = ctx.req.method
})

if ctx.req.path == "/block" then
    mut.resp.set_status(403)
    mut.resp.set_body("Blocked by Lua plugin")
    mut.ctx.set_finish()
end
```

---

## Notes

- Scripts execute in isolated event contexts
- Execution order is controlled by priority
- Returning errors from scripts may trigger `error.lua`
- `mut.ctx.set_finish()` stops further processing pipeline
