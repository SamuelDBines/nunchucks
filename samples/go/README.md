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

Additional demos:

```bash
cd go
go run ./cmd/nunchucks render -views ../samples/go/views -template is-in-demo.njk -data '{"role":"admin","nums":[1,2,3]}'
go run ./cmd/nunchucks render -views ../samples/go/views -template filter-parity-demo.njk -data '{"users":[{"name":"Zed","active":true,"role":"dev"},{"name":"Amy","active":false,"role":"ops"},{"name":"Bob","active":true,"role":"dev"}]}'
```
