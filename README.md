# Nunchucks

Nunchucks is a Nunjucks/Jinja-inspired templating project with:

- TypeScript package (`@samuelbines/nunchucks`)
- Go runtime + CLI
- WASM target from Go
- Docs site in `docs/` (GitHub Pages)
- Playground editor demo with live preview

Repository: `https://github.com/SamuelDBines/nunchucks`

## Current Focus

This project is actively evolving toward strong Nunjucks-style parity plus a multi-language workflow.

Implemented in Go today includes core tags (`if`, `for`, `set`, `extends`, `block`, `include`, `macro`, `import`, `call`, `raw`, `verbatim`, `filter`), expression engine improvements, and a broad set of built-in filters.

## Installation

### TypeScript package

```bash
npm install @samuelbines/nunchucks
```

### Go usage

```bash
cd go
go test ./...
```

## Go CLI

From `go/`:

```bash
go run ./cmd/nunchucks help

go run ./cmd/nunchucks render \
  -views ./views \
  -template index.njk \
  -data '{"title":"Hello"}'

go run ./cmd/nunchucks precompile \
  -views ./views \
  -out ./public \
  -data '{"title":"Hello"}'
```

## Go Examples

From `go/`:

```bash
go run ./examples/basic
go run ./examples/advanced
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

Key pages:

- `docs/index.html` (main docs)
- `docs/templating.html` (templating reference)
- `docs/playground.html` (mini editor + preview)

## Tests

### TypeScript

```bash
npm test
```

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

BSD-2-Clause
