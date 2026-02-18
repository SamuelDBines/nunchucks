# Go Samples

Render a single file:

```bash
cd go
go run ./cmd/nunchucks render -views ../samples/go/views -template index.njk -data '{"user":{"name":"sam","isAdmin":true},"items":["alpha","beta"]}'
```

Precompile all templates:

```bash
cd go
go run ./cmd/nunchucks precompile -views ../samples/go/views -out ../samples/go/out -data '{"user":{"name":"sam","isAdmin":true},"items":["alpha","beta"]}'
```
