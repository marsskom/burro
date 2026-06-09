# Policy Plugin (Burro)

The Policy plugin provides rule-based request filtering and transformation for Burro.

It allows defining match conditions + actions pipeline applied to incoming HTTP requests before forwarding to upstream services.

---

## Configuration

Main plugin config (`config.yml` at `%workdir%/plugins/policy/`):

```yml
priority: 10

whitelist: ./data/whitelist.txt
blacklist: ./data/blacklist.txt

action_dir: actions
```

---

## Request Processing Flow

Policy execution order:

1. Whitelist check
2. Blacklist check
3. Action rules execution (sorted by priority DESC)

---

## Action File Structure

Each YAML file under `%workdir%/plugins/policy/actions/` must contain:

```yml
actions:
  - id: string # required, unique rule id
    priority: integer # 0..1000
    on_match: stop|continue # optional (default: continue)

    match: # required (see Match below)
      ...

    action: # required (list of actions)
      - op: string # required
        args: {} # optional object (free-form per op)
```

### Match

```yml
match:
  method: string
  domain: string
  path: string
  ip: string

  headers:
    Header-Name: "value"
```

With logic constructions:

```yml
match:
  all:
    - match: ...
    - match: ...
...
match:
  any:
    - match: ...
    - match: ...
...
match:
  not:
    match: ...
```

### Action

```yml
action:
  - op: deny | allow | set_header | remove_header | redact_body | log
    args: {}
```

Multiple files are supported and merged into a single rule set.

---

## Rule Execution Model

Rules are executed in order of: `priority DESC` **(higher first)**.

For each rule:

- Check match
- If matched, execute actions sequentially
- Apply on_match:
  - continue - proceed to next rule
  - stop - stop rule processing immediately

---

## Match System

Match supports **logical composition** + **leaf conditions**.

Leaf conditions

```yml
match:
  method: GET
  domain: example.com
  path: /admin/*
  ip: 127.0.0.1

  headers:
    Authorization: "secret"
```

### Logical operators

#### ALL (AND)

All conditions must match:

```yml
match:
  all:
    - domain: example.com
    - method: DELETE
```

#### ANY (OR)

At least one must match:

```yml
match:
  any:
    - method: GET
    - method: POST
```

#### NOT

Negates a condition:

```yml
match:
  not:
    method: DELETE
```

### Rules

You can use ONLY one of: `all`, `any`, `not`.

You cannot mix logical operators with leaf conditions.

---

## Actions

Actions are executed sequentially.

```yml
action:
  - op: set_header
    args: ...
```

### Supported operations

#### deny

Immediately blocks request.

```yml
- op: deny
```

#### allow

Immediately allows request (skips further processing).

```yml
- op: allow
```

#### set_header

Sets request header.

```yml
- op: set_header
  args:
    name: X-Debug
    value: "true"
```

#### remove_header

Removes headers.

```yml
- op: remove_header
  args:
    names:
      - Authorization
      - Cookie
```

#### redact_body

Removes fields from JSON body.

```yml
- op: redact_body
  args:
    fields:
      - password
      - token
```

#### log

Writes log entry.

```yml
- op: log
  args:
    level: audit | debug | info | warn | error | trace
    message: "custom message"
```

---

## Action Execution Rules

- Actions are executed **in order**
- Each action can:
  - modify request
  - return response
  - stop execution

### Example Rules

#### Example 1: add debug header

```yml
actions:
  - id: example_add_debug_header
    priority: 1

    match:
      domain: example.com

    action:
      - op: set_header
        args:
          name: X-Debug
          value: "true"

      - op: log
        args:
          level: audit
          message: "X-Debug was set to true for example.com"
```

#### Example 2: block DELETE requests

```yml
actions:
  - id: example_block_delete_requests
    priority: 0

    match:
      all:
        - domain: example.com
        - method: DELETE

    action:
      - op: log
        args:
          level: audit
          message: "DELETE request denied for example.com"

      - op: deny

    on_match: stop
```

---

## Execution Semantics (important)

| Concept        | Behavior                   |
| -------------- | -------------------------- |
| whitelist      | always allows early        |
| blacklist      | always blocks early        |
| priority       | higher runs first          |
| on_match: stop | stops rule processing      |
| deny           | returns forbidden response |
| allow          | short-circuits pipeline    |
| log            | side effect only           |
| redact_body    | modifies request payload   |
