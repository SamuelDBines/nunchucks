package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	nunchucks "github.com/SamuelDBines/nunjucks/go"
)

type playgroundRenderRequest struct {
	Template string            `json:"template"`
	Files    map[string]string `json:"files"`
	Context  map[string]any    `json:"context"`
}

type playgroundRenderResponse struct {
	OK      bool              `json:"ok"`
	Outputs map[string]string `json:"outputs,omitempty"`
	Error   string            `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func renderPlayground(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, playgroundRenderResponse{OK: false, Error: "method not allowed"})
		return
	}

	var req playgroundRenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, playgroundRenderResponse{OK: false, Error: err.Error()})
		return
	}
	if req.Files == nil {
		req.Files = map[string]string{}
	}
	if req.Context == nil {
		req.Context = map[string]any{}
	}
	if strings.TrimSpace(req.Template) == "" {
		req.Template = "app.njk"
	}

	env := nunchucks.Configure(nunchucks.ConfigOptions{Loader: nunchucks.MemoryLoader(req.Files)})

	outputs := map[string]string{}
	for name, content := range req.Files {
		if strings.HasSuffix(strings.ToLower(name), ".njk") {
			out, err := env.Render(name, req.Context)
			if err != nil {
				outputs[name] = "[render error] " + err.Error()
				continue
			}
			outputs[name] = out
			continue
		}
		outputs[name] = content
	}

	if _, ok := outputs[req.Template]; !ok {
		if out, err := env.Render(req.Template, req.Context); err == nil {
			outputs[req.Template] = out
		}
	}

	writeJSON(w, http.StatusOK, playgroundRenderResponse{OK: true, Outputs: outputs})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/playground/render", withCORS(renderPlayground))

	addr := ":8090"
	log.Printf("nunchucks playground server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
