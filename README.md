<p align="center">
  <img src="logo.webp" alt="Burro Logo" width="250"/>

  <br/>

<strong>Burro</strong>

</p>

---

# Burro

Burro (pronounced the same as word "burrow" /ˈbʌr.əʊ/) is a modular security and traffic inspection tool that allows extending its behavior through a plugin-based architecture.

---

## Features

- HTTP proxy with interception capabilities
- Plugin-based architecture (Go plugins + YAML configuration)
- Per-plugin configuration isolation
- Traffic export (e.g. HAR-like logs)
- Certificate-based HTTPS interception

---

> Please, consider use `-h` in CLI for help with Burro since the toll is still under developemnt and the documentation may be outdated.

---

## Quick Command Reference

| Command                      | Description                                    | Example                                |
| ---------------------------- | ---------------------------------------------- | -------------------------------------- |
| `burro proxy`                | Start HTTP proxy using runtime config          | `burro proxy`                          |
| `burro proxy -z <addr>`      | Run proxy in zero-config mode                  | `burro proxy -z localhost:8080`        |
| `burro proxy -w <name>`      | Run proxy with workspace persistence           | `burro proxy -w test-session`          |
| `burro serve <addr> <dir>`   | Start static file server                       | `burro serve localhost:8888 ./runtime` |
| `burro cert init`            | Generate CA certificate for TLS interception   | `burro cert init`                      |
| `burro cert generate [host]` | Generate host certificate (default: localhost) | `burro cert generate localhost`        |
| `burro proxy -v`             | Increase log verbosity (info/debug)            | `burro proxy -v`                       |
| `burro proxy -vv`            | Enable debug verbosity                         | `burro proxy -vv`                      |
| `burro proxy -vvv`           | Maximum verbosity (with trace)                 | `burro proxy -vvv`                     |
| `make build`                 | Build Burro binary                             | `make build`                           |
| `make browser`               | Launch Chromium configured for testing         | `make browser`                         |
| `make ca-install`            | Install CA into system trust store (macOS)     | `make ca-install`                      |
| `make ca-remove`             | Remove CA from system trust store              | `make ca-remove`                       |
| `make ca-find`               | Check if CA is installed in system             | `make ca-find`                         |

---

## Configuration

Main Burro directory is `runtime`.

It is shipped with binary and contains predefined structure for configs, plugins' data and artifacts.

You may override runtime directory via environment variable:

```text
BURRO_WORKDIR=runtime
```

Or using cli flag `-d runtime`.

Burro uses a main configuration file located at:

```text
%workdir%/config.yml
```

---

## Core configuration

```yml
core:
  log_level: error
  plugins:
    dir: "plugins"
    config: "config.yml"

proxy:
  listen: localhost:8080

grpc:
  enabled: true
  debug: false
  listen: localhost:7777

tls:
  enabled: false
  cert:
  key:

plugins:
  logger:

  policy: # separate config

  harexport:
    file: "%session%-%datetime%.har"
    override: true
```

---

## Zero configuration mode

You may run burro as standalone binary without runtime directory in zero configuratio mode (`-z` flag).

> In zero configuration mode Burro only runs logger to log the requests.

For this mode you must specify listen address since there are no defualt values: `burro proxy -z localhost:8080`.

Other flags is optional. For example, gRPC will be disabled if its listen address is not explicity set (`-g`).

For using CA certificates for proxy to connect with HTTPS sites you may specify both with `--ca-cert` and `--ca-key`.

Same for TLS certificates. If you have them for the host where you run Burro, you may specify both with `--tls-cert` and `--tls-key` to use HTTPS connection to the proxy: `https://localhost:8443/`.

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
runtime/plugins/policy/config.yml
```

Example:

```yml
priority: 10
whitelist: ./data/whitelist.txt
blacklist: ./data/blacklist.txt
```

This allows separation of:

- global orchestration config
- plugin-specific logic config

Please, pay attention that all paths in plugin section works under `%workdir%` directory.

Each plugin has an access to `%workdir%/artifacts/%plugin-name%` directory in a role of the files storage.

And `%workdir%/plugins/%plugin-name%` as configuration's data storage from where it can only read files. For example, whilelist of domains in Policy plugin.

---

Plugins directory name `plugins` and config name `config.yml` may be changed in global configuration:

```yml
plugins:
  dir: "plugins"
  config: "config.yml"
```

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

## Core Plugins

- Logger - just logs events
- [Policy](./plugins/policy/README.md)
- HAR Export - exports requests adn responses into HAR file

---

## License

Burro core is licensed under the GNU General Public License v3.0 (GPLv3). This applies to all source code included in the main repository, including bundled plugins located in the `plugins/` directory, which are considered part of the Burro codebase and distributed under GPLv3.

External plugins that are developed and distributed independently of this repository are not considered part of Burro itself if they interact only through the documented Plugin API (including hook events and IPC-based communication such as stdin/stdout, HTTP, or similar mechanisms). Such external plugins are independent works and may be licensed under any terms chosen by their authors, including permissive licenses (e.g. MIT, Apache-2.0) or proprietary licenses.

More details here:

- [PLUGIN_EXCEPTION.md](./PLUGIN_EXCEPTION.md)
- [PLUGINS](./PLUGINS.md)

---

## Running Burro

### Build

```shell
make build
```

Binary will be available in project root as `burro`.

### Run proxy

`burro proxy` runs proxy with default `runtime` directory as workdir. Load configuration from this directory for itself and plugins.

If you want to save working session use `-w %workspace-name%`. After you finished with Burro with `CTRL+C` it saves all the requests into `%workdir%/db/%workspace-name%.sqlite3`.

And you may reuse workspace in future to add more sessions to it.

### Logs

By default, configuration defines logger level in `config.yml`, zero configuration mode defines `info` level.

If you want to increase verbosity you may use `-v`, `-vv`, `-vvv`.

### Artifacts

By default, Burro on exit (`CTRL+C`) saves workspace into DB if workspace name ws set.

Moreover, some plugins also may create some artifacts - `%workdir%/artifacts/`.

For instance, HAR export plugin creates HAR report file under `%workdir%/artifacts/harexport/` directory, by default.

Basically even you didn't provide workspace name, HAR plugin writes artifacts according to its configuration settings: `enabled: true`.

### Browser usage

Of course, you may use raw `curl` just to test Burro.

However, `make browser` command provides to you Chromium browser, ready to go.

The only requirement here is - Chromium must be installed in your system.

---

## TLS interception

No one modern web portal works without HTTPS, means you Burro need a CA certificate installed in your system as allowed and trusted.

To generate CA certificates use:

```shell
burro cert init
```

By default it writes them to `%workdir%/certs/ca.{pem|key}` but you may specify another path with CLI flgas: `--dst-cert` and `--dst-key`.

To operate CA in MacOS you may consider following commands in the `Makefile`:

- `make ca-install`
- `make ca-remove`
- `make ca-find` - shows if certificate was found in the OS

For other OS, please, read the respective documentation.

### Host Certificate

If you need TLS certificates for `localhost`, as an example, and generated CA pair for Burro already added as trusted you may generate certificates for `hostname`:

```shell
burro cert generate [host]
```

By default, `host` is `localhost`.

To manage this command more precisely you may utilise CLI flags:

- `--src-cert` - path to CA certificate
- `--src-key` - path to CA key
- `--dst-cert` - where to save host certificate
- `--dst-key` - where to save host key

---

## File server

Burro also supports simple file server out of the box:

```shell
burro serve localhost:8888 ./runtime
```

Nothing special at all.

Additionally you may set-up HTTPS file server by specifying host certificates in CLI flags: `--cert` (`-c`) and `--key` (`-k`).

---

## Docker

`Makefile` provides additional docker commands as well.

If you just want to try Burro in an isolated environment.

---

## Notes on architecture

- Burro is not a passive proxy only — it is a plugin execution runtime for HTTP traffic
- Plugins define behavior, not the core
- Core is minimal and stable, plugins evolve independently

### Plugin Development

All plugins that use native hooks are located under `plugins` directory.

You may create a directory for your own plugin, also place a separate `config.yml` file into the directory.

But not forget to add plugin name (directory name) into global `%workdir%/config.yml` as well:

```yml
plugins:
  your-plugin-name:
```

### Plugin Requirements

The `Plugin` is basically an interface (`./internal/plugin/plugin.go`):

```golang
type Plugin interface {
 ...
```

For better understanding, please, look at any plugin under `plugins/` directory in details.

### Common Hooks

**HTTP**:

- `OnConnect` triggers on established connection
- `OnBefore...` hook doesn't guarantee a snapshot (req/resp) has been created
- `OnAfter...` hook contains a snapshot (req/resp)
- That means, for instance, the policy plugin works on `OnBeforeRequestSend` to change the request, or forbid connection
- And, for example, HAR export plugin waits for `OnAfterRequestSend` to collect a request's snapshot with a full data, after all plugins modify it and it was done
- Furthemore, `OnBeforeResponseSend` hook may be useful in a case of HTTP stream since it provides response before a stream has started
- `OnAfterResponseSend` contains a response with snapshot; in a case of HTTP stream it got only 1MB of the stream; regular HTTP(S) conenction contains a whole response
- `OnError` triggers if somewhere error was appeared with sending a data to the hooks
- `OnClose` - on close connection

**WebSocket**:

- `OnWSOpen` triggers on open WS connection
- `OnWSMessage` hook catches all WS messages in real time
- `OnWSClose` - on close connection

---

## gRPC API

Burro exposes an optional gRPC interface that can be enabled in the core configuration (or set in CLI flag):

```yml
grpc:
  enabled: true
  debug: false
  listen: localhost:7777
```

### Purpose

The gRPC layer is intended as a lightweight control and observation channel over Burro’s runtime. It allows external clients to:

- verify that Burro is running and responsive
- subscribe to live HTTP traffic events
- build external observers (e.g. UI dashboards, analyzers, security tooling)
- integrate Burro into larger distributed systems

It does not replace the proxy itself — it mirrors internal events in a structured, streaming-friendly way.

### Core Concepts

**1. Health Check (Ping)**

A simple request-response method used to confirm that the gRPC service is alive and reachable.

Typical use cases:

- service discovery checks
- monitoring systems (e.g. Prometheus-style health probes)
- debugging connectivity issues

It returns a minimal confirmation message and has no side effects.

**2. Event Streaming (Subscribe)**

The main feature of the gRPC API is a continuous stream of events representing HTTP traffic processed by Burro.

When a client subscribes, it receives a live feed of request/response lifecycle events.

### What is an Event?

Each event represents a single HTTP transaction observed by the proxy. It contains both request and response metadata, making it possible to reconstruct the full lifecycle of a request.

Or Event contains WebSocket message.

An event typically includes:

- Identity and context
  - unique request ID
  - session ID (useful for grouping traffic sessions)
- HTTP/HTTPS:
  - Request metadata
    - protocol (HTTP/HTTPS)
    - scheme, host, method, URL, path
    - remote address
    - headers, cookies, and query parameters (encoded as JSON)
    - request body (raw bytes)
  - Lifecycle information
    - start timestamp
    - current state (e.g. in-progress, completed)
    - whether the request is finished
  - Response metadata
    - HTTP status code and status text
    - response headers (JSON)
    - response body (raw bytes)
    - response size and protocol
  - Timing and tracking
    - creation and update timestamps
  - Extensions
    - metadata field for plugin-generated or internal auxiliary data
- WebSocket:
  - Direction (client -> server, server -> client)
  - OP Code
  - Data in bytes
  - Data as text
  - Timestamp of emitting

### Design Notes

The stream is append-only and real-time: events are pushed as they happen.
Payloads are intentionally verbose to avoid multiple round trips when analyzing traffic.

JSON-encoded fields (headers, cookies, query params) are used for flexibility and compatibility with external systems.

Raw bodies are preserved to allow deep inspection, replay, or forensic analysis.

### Typical Usage Scenarios

This API is useful when building:

- a custom UI similar to a lightweight Burp Suite dashboard
- automated traffic analyzers or anomaly detectors
- integration bridges to SIEM systems or logging pipelines
- debugging tools that need live visibility into HTTP traffic
- plugin-like external services that do not run inside Burro core

### CLI flags

If you want to run Burro in regular mode without gRPC, add `--no-grpc` flag.

For debugging gRPC you may enable it in config:

```yml
grpc:
  debug: true
```

Or pass flag `--grpc-d` to enable debug mode.

Debug mode allows to you use something like `grpcurl`:

```shell
grpcurl -plaintext localhost:7777 burro.v1.BurroService/Subscribe
```

To catch raw events.

You may investigare API with:

```shell
grpcurl -plaintext localhost:7777 list
```

Moreover, you may filter events by transport type:

```shell
grpcurl \
  -plaintext \
  -d '{
    "transportType": ["TRANSPORT_HTTP"]
  }' \
  localhost:7777 \
  burro.v1.BurroService/Subscribe
```

Event type:

```shell
grpcurl \
  -plaintext \
  -d '{
    "eventTypes": ["EVENT_WS_MESSAGE"]
  }' \
  localhost:7777 \
  burro.v1.BurroService/Subscribe
```

Or both:

```shell
grpcurl \
  -plaintext \
  -d '{
    "transportType": ["TRANSPORT_HTTP"],
    "eventTypes": ["EVENT_REQUEST", "EVENT_RESPONSE"]
  }' \
  localhost:7777 \
  burro.Burro/Subscribe
```

---

## `.gitlab`

This directory contains GitLab issue, merge request, and description templates used across the project.
