package nunchucks

import "strings"

// ConfigOptions controls environment setup for rendering templates.
type ConfigOptions struct {
	Path   string
	Loader Loader
}

// Env is the Go renderer environment.
type Env struct {
	basePath string
	loader   Loader
}

// Configure builds a rendering environment using options similar to the TS API.
func Configure(opts ConfigOptions) *Env {
	path := strings.TrimSpace(opts.Path)
	if path == "" {
		path = "views"
	}

	ldr := opts.Loader
	if ldr == nil {
		ldr = FileSystemLoader(path)
	}

	return &Env{basePath: path, loader: ldr}
}

// Render loads and renders a template file with the provided context.
func (e *Env) Render(name string, ctx map[string]any) (string, error) {
	src, err := e.compileTemplate(name)
	if err != nil {
		return "", err
	}
	return e.renderString(src, ctx)
}

// Compile resolves includes/extends into a compiled template string.
func (e *Env) Compile(name string) (string, error) {
	return e.compileTemplate(name)
}

// RenderString renders a string template with the provided context.
func (e *Env) RenderString(src string, ctx map[string]any) (string, error) {
	return e.renderString(src, ctx)
}
