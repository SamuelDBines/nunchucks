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
go run ./cmd/nunchucks render -views ../samples/go/views -template is-in-demo.njk -data '{"role":"admin","nums":[1,2,3],"cfg":{"a":1}}'
go run ./cmd/nunchucks render -views ../samples/go/views -template filter-parity-demo.njk -data '{"nums":[1,2,3,4],"data":{"b":2,"a":1,"C":3},"users":[{"name":"Zed","active":true,"role":"Dev","age":30},{"name":"Amy","active":false,"role":"ops","age":22},{"name":"Bob","active":true,"role":"dev","age":25}]}'
go run ./cmd/nunchucks render -views ../samples/go/views -template functions-demo.njk -data '{"cfg":{"b":2,"a":1}}'
```
