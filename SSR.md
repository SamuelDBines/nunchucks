# Nunchucks SSR

This document defines the current **server-side rendering (SSR)** behavior in Nunchucks.

## Scope

- Primary runtime: Go (`go/`)
- JS/TS package: thin wrapper over Go WASM runtime
- Default rendering mode: **SSR**

If you render a template without a client block, output is plain rendered text/HTML from the server/runtime.

## Runtime Entry Points

- Go API:
  - `Env.Render(name, ctx)`
  - `Env.RenderString(source, ctx)`
  - `Env.PrecompileDir(outDir, ctx)`
- Go CLI:
  - `nunchucks render`
  - `nunchucks precompile`
- WASM API:
  - `renderFromMap({ template, files, context })`
  - `renderString({ source, context })`

## Current SSR Features

### Core tags

- `if / elif / else / endif`
- `for / endfor`
- `set`
- `extends`
- `block / endblock`
- `include` (with/without context, ignore missing)
- `macro / endmacro`
- `import`
- `from ... import ...`
- `call / endcall`
- `filter / endfilter`
- `raw / endraw`
- `verbatim / endverbatim`

### Expressions

- Variable lookup and dot paths
- Math and logic expressions
- Function/macro calls (including named args)
- Filter pipelines
- `super()` in inherited blocks

### Built-ins and compatibility work

Go runtime includes a broad set of built-ins and parity-focused behavior (see tests in `go/` for exact cases):

- sorting/grouping and selection filters
- string/number/list utility filters
- import/macro context modes

## SSR Example

```njk
{% extends "base.njk" %}
{% block body %}
  {% set active = users | selectattr("active") %}
  <h2>{{ team.name | title }}</h2>
  <p>Active: {{ active | length }}</p>
  {% for user in active %}
    {% include "partials/user-card.njk" with context %}
  {% endfor %}
{% endblock %}
```

## Local Verification

```bash
cd go
go test ./...
go run ./examples/basic
go run ./examples/advanced
```

## SSR Contract

- Server/runtime output is deterministic for the same template + context.
- No automatic client reactivity is assumed in SSR output.
- Client behavior must be explicitly opted into (see `CLIENT_RENDERING.md`).

