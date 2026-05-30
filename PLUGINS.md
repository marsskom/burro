# Burro Plugin System

## Overview

Burro supports a plugin system that allows extending functionality via hooks and a defined plugin API.

Plugins in Burro can exist in two forms:

1. **Bundled plugins (shipped with Burro repository)**
2. **External plugins (installed or provided by users)**

The licensing of each type depends on how the plugin is distributed and how it integrates with Burro.

---

## 1. Bundled Plugins (Inside This Repository)

All plugins located in the `plugins/` directory **within the Burro repository** are considered part of the Burro source distribution.

### Licensing

- Bundled plugins are licensed under **GNU GPLv3**
- They are considered part of the Burro codebase
- Any modification or redistribution must comply with GPLv3 terms

### Rationale

Bundled plugins:

- Are distributed **together** with Burro
- May use internal hooks and internal Go packages
- Are tightly coupled with the core system

Therefore, they are treated as a single work with Burro and fall under GPLv3.

---

## 2. External Plugins (User-Provided or Third-Party)

External plugins are plugins that are:

- Installed separately from the Burro repository
- Not part of the _official_ `plugins/` directory
- Developed and distributed independently by third-party authors
- Communicate with Burro only through the public plugin API

### Communication Model

External plugins MUST interact with Burro only via the documented plugin interface:

- Hook system (event callbacks)
- Public plugin API
- IPC mechanisms (e.g., stdin/stdout, HTTP, gRPC, sockets)

External plugins MUST NOT:

- Import internal Burro packages
- Rely on undocumented internal behavior
- Be compiled into Burro binary

### Licensing

External plugins are considered **independent works**.

They may be licensed under **any license chosen by their author**, including but not limited to:

- MIT License
- Apache 2.0
- BSD licenses
- Proprietary licenses
- GPL-compatible licenses

Burro does not impose licensing restrictions on external plugins.

---

## 3. Plugin Boundary Definition

A plugin is considered **external and independent** if all conditions are met:

- It is executed or loaded as a separate unit
- It communicates only through the official plugin API
- It does not import or link against internal Burro code
- It can function independently of Burro core internals

If any of these conditions are violated, the plugin is considered **bundled/internal** and falls under GPLv3.

---

## 4. Summary of Licensing Model

| Plugin Type      | Location             | License     |
| ---------------- | -------------------- | ----------- |
| Bundled plugins  | `plugins/` in repo   | GPLv3       |
| External plugins | Installed separately | Any license |

---

## 5. Important Notes

- Directory structure alone does NOT define licensing.
- Licensing is determined by distribution method and coupling level.
- The intent of this model is to keep Burro core GPLv3 while enabling a flexible plugin ecosystem.

---

## 6. No Warranty

Plugins are provided "as is", without warranty of any kind.
Burro is not responsible for external plugins, their behavior, or their licensing compliance.
