package nunchucks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PrecompileOptions struct {
	OutputFormat string
}

func IsTemplateFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".njk", ".html", ".txt", ".yaml", ".yml", ".json", ".xml", ".css", ".js":
		return true
	default:
		return false
	}
}

func outputPathFor(rel string, opts PrecompileOptions) string {
	if strings.EqualFold(strings.TrimSpace(opts.OutputFormat), "html") && strings.EqualFold(filepath.Ext(rel), ".njk") {
		return strings.TrimSuffix(rel, filepath.Ext(rel)) + ".html"
	}
	return rel
}

// PrecompileDir renders templates from the configured views path into outDir.
func (e *Env) PrecompileDir(outDir string, ctx map[string]any) error {
	return e.PrecompileDirWithOptions(outDir, ctx, PrecompileOptions{})
}

// PrecompileDirWithOptions renders templates from the configured views path into outDir.
func (e *Env) PrecompileDirWithOptions(outDir string, ctx map[string]any, opts PrecompileOptions) error {
	if strings.TrimSpace(outDir) == "" {
		return fmt.Errorf("outDir is required")
	}
	if strings.TrimSpace(e.basePath) == "" {
		return fmt.Errorf("env base path is empty")
	}
	if format := strings.TrimSpace(opts.OutputFormat); format != "" && !strings.EqualFold(format, "preserve") && !strings.EqualFold(format, "html") {
		return fmt.Errorf("unsupported output format: %s", format)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	return filepath.WalkDir(e.basePath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !IsTemplateFile(d.Name()) {
			return nil
		}

		rel, err := filepath.Rel(e.basePath, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		rendered, err := e.Render(rel, ctx)
		if err != nil {
			return fmt.Errorf("render %s: %w", rel, err)
		}

		dst := filepath.Join(outDir, filepath.FromSlash(outputPathFor(rel, opts)))
		if strings.TrimSpace(rendered) == "" {
			_ = os.Remove(dst)
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dst, []byte(rendered), 0o644)
	})
}
