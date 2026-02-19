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

func builtinGlobals() map[string]any {
	return map[string]any{
		"range": TemplateFunc(func(args []any, _ map[string]any, _ string) (any, error) {
			start := 0
			stop := 0
			step := 1

			switch len(args) {
			case 1:
				stop = toInt(args[0], 0)
			case 2:
				start = toInt(args[0], 0)
				stop = toInt(args[1], 0)
			default:
				if len(args) >= 3 {
					start = toInt(args[0], 0)
					stop = toInt(args[1], 0)
					step = toInt(args[2], 1)
				}
			}

			if step == 0 {
				step = 1
			}

			out := []any{}
			if step > 0 {
				for i := start; i < stop; i += step {
					out = append(out, i)
				}
				return out, nil
			}
			for i := start; i > stop; i += step {
				out = append(out, i)
			}
			return out, nil
		}),
	}
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
