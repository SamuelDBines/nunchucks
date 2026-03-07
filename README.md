# Nunchucks

Nunchucks is a templating project with Nunchucks-first syntax and Jinja-style concepts:

- JavaScript/TypeScript package backed by Go WASM (`@samuelbines/nunchucks`)
- Go runtime + CLI
- WASM target from Go
- Docs site in `docs/` (GitHub Pages)
- Playground editor demo with live preview

Repository: `https://github.com/SamuelDBines/nunchucks`

## Current Focus

This project is actively evolving toward strong syntax parity plus a multi-language workflow.

Implemented in Go today includes core tags (`if`, `for`, `set`, `extends`, `block`, `include`, `macro`, `import`, `call`, `raw`, `verbatim`, `filter`), expression engine improvements, and a broad set of built-in filters.

## Installation

### JS/TS package (Go WASM runtime)

```bash
npm install @samuelbines/nunchucks
```

```ts
import { loadNunchucksWasm } from "@samuelbines/nunchucks";

const nc = await loadNunchucksWasm({
  wasmURL: "/node_modules/@samuelbines/nunchucks/go/wasm/nunchucks.wasm",
  goURL: "/node_modules/@samuelbines/nunchucks/go/wasm/wasm_exec.js",
});

const html = nc.renderFromMap({
  template: "index.njk",
  files: { "index.njk": "Hello {{ name }}" },
  context: { name: "world" },
});
```

### CLI

Install the native CLI from GitHub Releases:

macOS and Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/SamuelDBines/nunchucks/main/scripts/install-cli.sh | sh
nunchucks version
nunchucks help
```

Windows PowerShell:

```powershell
irm https://raw.githubusercontent.com/SamuelDBines/nunchucks/main/scripts/install-cli.ps1 | iex
nunchucks.exe version
nunchucks.exe help
```

Go users can also install from source:

```bash
go install github.com/SamuelDBines/nunjucks/go/cmd/nunchucks@latest
```

Release binaries are published at `https://github.com/SamuelDBines/nunchucks/releases`.
The install scripts pick a writable bin directory and print PATH instructions when needed.

### Go usage

```bash
cd go
go test ./...
```

## Go CLI

From `go/` or from an installed release binary:

```bash
go run ./cmd/nunchucks help
go run ./cmd/nunchucks help render
go run ./cmd/nunchucks version

go run ./cmd/nunchucks render \
  -views ./views \
  -template index.njk \
  -data '{"title":"Hello"}'

go run ./cmd/nunchucks precompile \
  -views ./views \
  -out ./public \
  -data '{"title":"Hello"}'

go run ./cmd/nunchucks precompile \
  -views ./views \
  -out ./public \
  -out-format html

go run ./cmd/nunchucks precompile \
  -views ./views \
  -out ./public \
  -watch
```

## Go Examples

From `go/`:

```bash
go run ./examples/basic
go run ./examples/advanced
go run ./examples/client
```

Also see sample templates under `samples/go`.

## WASM Build (Go)

From `go/`:

```bash
GOOS=js GOARCH=wasm go build -o wasm/nunchucks.wasm ./cmd/nunchucks-wasm
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" wasm/wasm_exec.js
```

## Stateless Playground Render Server (Go)

A stateless memory-render API is available under `go/server`.

From `go/`:

```bash
go run ./server
```

Default endpoint:

- `POST /api/playground/render`
- `GET /healthz`

See `go/server/README.md` for request/response format.

## Docs

Docs source is in `docs/` and is intended for GitHub Pages.

Architecture docs:

- `SSR.md` (server-side rendering behavior)
- `CLIENT_RENDERING.md` (client-side behavior and roadmap)

Key pages:

- `docs/index.html` (main docs)
- `docs/templating.html` (templating reference)
- `docs/playground.html` (mini editor + preview)

## Local App Builder (Live Reload + Playground)

A local builder is available in `app/` to edit template files, add files, and preview output in one page with file watching.

Run:

```bash
pnpm app
```

Open:

- `http://localhost:5177`

Notes:

- `app/views` is the file workspace.
- Saving from the UI writes to `app/views`.
- File changes in `app/views` trigger live refresh.
- Rendering uses the Go playground server API (started automatically by `app/server.mjs`).

## Tests

### Go

```bash
cd go
go test ./...
```

## Status

This repo is under active development. Some advanced parity edge cases and async behavior are still being iterated.

## Contributing

Open an issue or PR on GitHub.

## License

- App/repo: BSD-2-Clause (`LICENSE`)
- VS Code extension: BSD-2-Clause (`vscode-nunchucks/LICENSE`)
