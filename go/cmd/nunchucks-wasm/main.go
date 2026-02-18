//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	nunchucks "github.com/SamuelDBines/nunjucks/go"
)

type renderMapRequest struct {
	Template string            `json:"template"`
	Files    map[string]string `json:"files"`
	Context  map[string]any    `json:"context"`
}

type renderStringRequest struct {
	Source  string         `json:"source"`
	Context map[string]any `json:"context"`
}

type wasmResponse struct {
	OK     bool   `json:"ok"`
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

func toResponse(resp wasmResponse) any {
	b, _ := json.Marshal(resp)
	return string(b)
}

func renderFromMap(_ js.Value, args []js.Value) any {
	if len(args) < 1 {
		return toResponse(wasmResponse{OK: false, Error: "missing request json"})
	}

	var req renderMapRequest
	if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
		return toResponse(wasmResponse{OK: false, Error: err.Error()})
	}
	if req.Context == nil {
		req.Context = map[string]any{}
	}
	if req.Files == nil {
		req.Files = map[string]string{}
	}

	env := nunchucks.Configure(nunchucks.ConfigOptions{Loader: nunchucks.MemoryLoader(req.Files)})
	out, err := env.Render(req.Template, req.Context)
	if err != nil {
		return toResponse(wasmResponse{OK: false, Error: err.Error()})
	}

	return toResponse(wasmResponse{OK: true, Output: out})
}

func renderString(_ js.Value, args []js.Value) any {
	if len(args) < 1 {
		return toResponse(wasmResponse{OK: false, Error: "missing request json"})
	}

	var req renderStringRequest
	if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
		return toResponse(wasmResponse{OK: false, Error: err.Error()})
	}
	if req.Context == nil {
		req.Context = map[string]any{}
	}

	env := nunchucks.Configure(nunchucks.ConfigOptions{})
	out, err := env.RenderString(req.Source, req.Context)
	if err != nil {
		return toResponse(wasmResponse{OK: false, Error: err.Error()})
	}

	return toResponse(wasmResponse{OK: true, Output: out})
}

func main() {
	api := js.Global().Get("Object").New()
	api.Set("renderFromMap", js.FuncOf(renderFromMap))
	api.Set("renderString", js.FuncOf(renderString))
	js.Global().Set("NunchucksWasm", api)

	select {}
}
