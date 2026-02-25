# Nunchucks Client Rendering

This document defines the **client-side rendering direction** for Nunchucks.

Goal: keep Nunchucks backend-first, but allow opt-in client behavior for teams that do not want full SPA frameworks.

## Design Goal

- SSR remains the default.
- Client behavior is explicit and localized.
- Template language stays Nunchucks-first.

## Current Client Features (Implemented)

## 1) `client` block

Wrap logic inside:

```njk
{% client %}
  // client-side code region
{% endclient %}
```

Behavior:

- Compiles to a browser module script:
  - `<script type="module" data-nunchucks-client>...</script>`
- Keeps SSR output intact outside the block.

## 2) `fetch` inside `client`

Current supported syntax:

```njk
{% fetch | '/api/users/' + userID | as user | %}
```

or text mode:

```njk
{% fetch | '/status.txt' | as statusText | text | %}
```

Compiled client JS shape:

```js
const __nc_fetch_res_0 = await fetch("/api/users/42");
const user = await __nc_fetch_res_0.json();
```

Notes:

- Endpoint expression is evaluated during SSR compile (using template context).
- Default mode is `json`.
- `text` mode is supported.

## What This Is Not Yet

- Not a full reactive DOM runtime yet.
- Not keyed diff/patch rendering for client-side loops yet.
- Not a component framework replacement yet.

## Planned Client Rendering Model

## Principle

Within `{% client %}...{% endclient %}`, Nunchucks constructs should be compilable to reactive client runtime instructions, while SSR keeps first paint.

## State Model (Proposed)

State should be emitted near the top of the page (head or first body segment), so SSR and client runtime start from the same model.

### Chosen Direction

- `state` is **local by default** (template/file scope)
- state sharing is **explicit** (export/use)
- simple events use **inline expressions**
- long event logic can use optional named actions

### Core syntax

```njk
{% client %}
  {% state app | { users: users, alerts: [], ui: { modalOpen: false } } | %}
{% endclient %}
```

### Local component state example

```njk
{% client %}
  {% state app | { pressed: false } | %}
{% endclient %}

<button onClick={{ app.pressed = true }}>
  Click me
</button>
```

### Explicit cross-file sharing (proposed)

Export in producer template:

```njk
{% client %}
  {% state app | { pressed: false } | export "ButtonState" %}
{% endclient %}
```

Use in consumer template:

```njk
{% client %}
  {% state use "ButtonState" as button %}
{% endclient %}
```

This avoids overloading SSR `{% set %}` semantics for client runtime state.

## Lifecycle Contract (Required)

Every reactive action should have predictable lifecycle states:

- `idle`
- `loading`
- `success`
- `error`

This applies to:

- fetch requests
- SSE connections
- form submissions

Suggested runtime shape:

```js
window.__nunchucks.state.app
window.__nunchucks.meta["users.fetch"].status // idle/loading/success/error
```

## Reactive Primitives (Proposed)

These should be built-in and lightweight (no framework dependency):

1. `state`
   - defines reactive model
   - local by default
   - optional `export/use` for explicit sharing
2. `fetch`
   - request + state write + lifecycle status
3. `sse`
   - EventSource stream + reconnect strategy + handler mapping
4. `form`
   - enhanced submit with validation and target patch
5. `alerts`
   - global queue for success/error/info
6. `modal`
   - open/close + focus trap + escape close
7. `sidebar`
   - open/close + overlay + responsive behavior
8. `action` (optional)
   - named handlers for longer event logic

## Built-in UI Blocks (Syntax Ideas)

### Alerts

```njk
{% client %}
  {% alerts app.alerts | %}
{% endclient %}
```

### Modal

```njk
{% client %}
  {% modal app.ui.modalOpen | id "user-modal" | %}
    <h2>User</h2>
  {% endmodal %}
{% endclient %}
```

### Sidebar

```njk
{% client %}
  {% sidebar app.ui.sidebarOpen | side "left" | %}
    ...
  {% endsidebar %}
{% endclient %}
```

## SSE Support (High Priority)

SSE should be first-class for backend-heavy apps.

### Syntax idea

```njk
{% client %}
  {% sse | "/events" | on "message" => "events.push($event)" | %}
{% endclient %}
```

### Runtime requirements

- auto reconnect with backoff
- explicit close on teardown/navigation
- event-type routing (`message`, `update`, `alert`, etc.)
- lifecycle flags:
  - `connecting`
  - `open`
  - `error`
  - `closed`

## Forms (High Priority)

Forms are a core backend workflow and should be first-class.

### Syntax idea

```njk
{% client %}
  {% form | "#create-user" | submit "/api/users" | target "#users-list" | %}
{% endclient %}
```

### Expected behavior

- progressive enhancement (normal POST works without JS)
- JS-enhanced submit when client runtime is active
- validation mapping into state/errors
- lifecycle states (`loading/success/error`)
- optional optimistic UI

## Event Handling Style

### Default for simple cases: inline expression

```njk
<button onClick={{ app.pressed = true }}>Press</button>
```

### Optional for long logic: named action block

```njk
{% client %}
  {% action buttonPress %}
    app.pressed = true
    app.alerts.push({ type: "success", message: "Pressed" })
  {% endaction %}
{% endclient %}

<button onClick={{ buttonPress }}>Press</button>
```

Recommendation: keep inline first; use `action` only when expressions become hard to read.

## Proposed v1 roadmap

1. Named local `state` model in client blocks
2. Inline event expressions (`onClick={{ ... }}`)
3. Fetch lifecycle + state writes
4. SSE primitive with reconnect and event routing
5. Form submit primitive with progressive enhancement
6. Explicit state `export/use` sharing
7. `alerts` built-in
8. `modal` + `sidebar` built-ins
9. Reactive `for` in client with keyed patching

## Example Target UX

Template:

```njk
<ul id="users-list">
{% client %}
  {% for user in users %}
    <li>{{ user.name }}</li>
  {% endfor %}
{% endclient %}
</ul>
```

Target behavior:

- SSR renders initial list.
- Client runtime hydrates list binding.
- Updating `users` updates only changed DOM rows.

## Local Testing Today

```bash
cd go
go test ./...
go run ./examples/client
```

`go/examples/client` currently demonstrates emitted client script + fetch transform.

## Separation Contract

- **SSR.md** = server/runtime features and guarantees.
- **CLIENT_RENDERING.md** = client-only feature set, semantics, and roadmap.

This separation keeps Nunchucks backend-centric while enabling progressive client capability.
