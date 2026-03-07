package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	nunchucks "github.com/SamuelDBines/nunjucks/go"
)

var version = "dev"

const (
	ansiRed   = "\033[31m"
	ansiReset = "\033[0m"
)

type stringListFlag []string

func (s *stringListFlag) String() string {
	return fmt.Sprint([]string(*s))
}

func (s *stringListFlag) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func executableName() string {
	name := filepath.Base(os.Args[0])
	if name == "." || name == string(filepath.Separator) || name == "" {
		return "nunchucks"
	}
	return name
}

func printRootUsage() {
	name := executableName()
	fmt.Fprintf(os.Stderr, "%s renders Nunchucks templates from the command line.\n\n", name)
	fmt.Fprintf(os.Stderr, "Usage:\n  %s <command> [options]\n\n", name)
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  render      Render one template to stdout")
	fmt.Fprintln(os.Stderr, "  precompile  Render a views directory to static output")
	fmt.Fprintln(os.Stderr, "  version     Print CLI version information")
	fmt.Fprintln(os.Stderr, "  help        Show general help or help for a command")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintf(os.Stderr, "  %s render -views ./views -template index.njk -data '{\"title\":\"Hello\"}'\n", name)
	fmt.Fprintf(os.Stderr, "  %s precompile -views ./views -out ./public\n", name)
	fmt.Fprintf(os.Stderr, "  %s precompile -views ./views -out ./public --watch\n", name)
	fmt.Fprintf(os.Stderr, "  %s help render\n", name)
	fmt.Fprintf(os.Stderr, "  %s version\n", name)
}

func printRenderUsage() {
	name := executableName()
	fmt.Fprintf(os.Stderr, "Usage:\n  %s render [options]\n\n", name)
	fmt.Fprintln(os.Stderr, "Render a single template and write the result to stdout.")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Options:")
	fmt.Fprintln(os.Stderr, "  -views string        templates directory (default \"views\")")
	fmt.Fprintln(os.Stderr, "  -template string     template path relative to views (required)")
	fmt.Fprintln(os.Stderr, "  -data string         JSON context object (default \"{}\")")
	fmt.Fprintln(os.Stderr, "  -global value        global template, repeatable")
	fmt.Fprintln(os.Stderr, "  -global-head value   global head template, repeatable")
	fmt.Fprintln(os.Stderr, "  -global-foot value   global foot template, repeatable")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Example:")
	fmt.Fprintf(os.Stderr, "  %s render -views ./views -template index.njk -data '{\"user\":{\"name\":\"sam\"}}'\n", name)
}

func printPrecompileUsage() {
	name := executableName()
	fmt.Fprintf(os.Stderr, "Usage:\n  %s precompile [options]\n\n", name)
	fmt.Fprintln(os.Stderr, "Render a views directory to static files on disk.")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Options:")
	fmt.Fprintln(os.Stderr, "  -views string        templates directory (default \"views\")")
	fmt.Fprintln(os.Stderr, "  -out string          output directory (default \"public\")")
	fmt.Fprintln(os.Stderr, "  -out-format string   output naming: preserve or html (default \"preserve\")")
	fmt.Fprintln(os.Stderr, "  -watch               rerender when template files change")
	fmt.Fprintln(os.Stderr, "  -interval duration   polling interval for watch mode (default 1s)")
	fmt.Fprintln(os.Stderr, "  -data string         JSON context object (default \"{}\")")
	fmt.Fprintln(os.Stderr, "  -global value        global template, repeatable")
	fmt.Fprintln(os.Stderr, "  -global-head value   global head template, repeatable")
	fmt.Fprintln(os.Stderr, "  -global-foot value   global foot template, repeatable")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Example:")
	fmt.Fprintf(os.Stderr, "  %s precompile -views ./views -out ./public -data '{\"title\":\"Hello\"}'\n", name)
	fmt.Fprintf(os.Stderr, "  %s precompile -views ./views -out ./public -out-format html\n", name)
	fmt.Fprintf(os.Stderr, "  %s precompile -views ./views -out ./public --watch -interval 750ms\n", name)
}

func printVersionUsage() {
	fmt.Fprintln(os.Stderr, "Usage:\n  nunchucks version")
}

func printHelp(command string) error {
	switch command {
	case "", "help", "-h", "--help":
		printRootUsage()
		return nil
	case "render":
		printRenderUsage()
		return nil
	case "precompile":
		printPrecompileUsage()
		return nil
	case "version":
		printVersionUsage()
		return nil
	default:
		return fmt.Errorf("unknown help topic: %s", command)
	}
}

func printVersion() {
	fmt.Fprintf(os.Stdout, "nunchucks %s\n", version)
}

func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	term := strings.TrimSpace(os.Getenv("TERM"))
	return term != "" && term != "dumb"
}

func errorText(msg string) string {
	if !colorEnabled() {
		return msg
	}
	return ansiRed + msg + ansiReset
}

func printError(err error) {
	fmt.Fprintln(os.Stderr, errorText("error: "+err.Error()))
}

func parseData(raw string) (map[string]any, error) {
	if raw == "" {
		return map[string]any{}, nil
	}
	out := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	return out, nil
}

type fileState struct {
	modTime time.Time
	size    int64
}

func scanTemplateState(views string) (map[string]fileState, error) {
	state := map[string]fileState{}
	err := filepath.WalkDir(views, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !nunchucks.IsTemplateFile(d.Name()) {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(views, path)
		if err != nil {
			return err
		}
		state[filepath.ToSlash(rel)] = fileState{
			modTime: info.ModTime().UTC(),
			size:    info.Size(),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return state, nil
}

func diffTemplateState(previous, current map[string]fileState) []string {
	changed := make([]string, 0)
	seen := make(map[string]struct{}, len(previous)+len(current))

	for name, prev := range previous {
		seen[name] = struct{}{}
		next, ok := current[name]
		if !ok || !prev.modTime.Equal(next.modTime) || prev.size != next.size {
			changed = append(changed, name)
		}
	}
	for name := range current {
		if _, ok := seen[name]; !ok {
			changed = append(changed, name)
		}
	}

	sort.Strings(changed)
	return changed
}

func removeDeletedOutputs(outDir string, previous, current map[string]fileState, opts nunchucks.PrecompileOptions) error {
	for name := range previous {
		if _, ok := current[name]; ok {
			continue
		}
		target := nunchucksOutputPath(name, opts)
		if err := os.Remove(filepath.Join(outDir, filepath.FromSlash(target))); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func nunchucksOutputPath(rel string, opts nunchucks.PrecompileOptions) string {
	if strings.EqualFold(strings.TrimSpace(opts.OutputFormat), "html") && strings.EqualFold(filepath.Ext(rel), ".njk") {
		return strings.TrimSuffix(rel, filepath.Ext(rel)) + ".html"
	}
	return rel
}

func precompileOnce(env *nunchucks.Env, outDir string, ctx map[string]any, opts nunchucks.PrecompileOptions) error {
	return env.PrecompileDirWithOptions(outDir, ctx, opts)
}

func watchPrecompile(env *nunchucks.Env, views, outDir string, ctx map[string]any, interval time.Duration, opts nunchucks.PrecompileOptions) error {
	if interval <= 0 {
		return fmt.Errorf("-interval must be greater than 0")
	}
	current, err := scanTemplateState(views)
	if err != nil {
		return err
	}

	if err := precompileOnce(env, outDir, ctx, opts); err != nil {
		printError(err)
	}

	fmt.Fprintf(os.Stderr, "watching %s and writing to %s every %s\n", views, outDir, interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	for {
		select {
		case <-signals:
			fmt.Fprintln(os.Stderr, "watch stopped")
			return nil
		case <-ticker.C:
			next, err := scanTemplateState(views)
			if err != nil {
				printError(err)
				continue
			}

			changed := diffTemplateState(current, next)
			if len(changed) == 0 {
				continue
			}

			if err := precompileOnce(env, outDir, ctx, opts); err != nil {
				printError(err)
				current = next
				continue
			}
			if err := removeDeletedOutputs(outDir, current, next, opts); err != nil {
				printError(err)
				current = next
				continue
			}

			fmt.Fprintf(os.Stderr, "recompiled %d changed file(s): %s\n", len(changed), strings.Join(changed, ", "))
			current = next
		}
	}
}

func runRender(args []string) error {
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = printRenderUsage

	views := fs.String("views", "views", "templates directory")
	template := fs.String("template", "", "template path relative to views")
	data := fs.String("data", "{}", "JSON context object")
	var globalTemplates stringListFlag
	var globalHeadTemplates stringListFlag
	var globalFootTemplates stringListFlag
	fs.Var(&globalTemplates, "global", "global template (repeatable)")
	fs.Var(&globalHeadTemplates, "global-head", "global head template (repeatable)")
	fs.Var(&globalFootTemplates, "global-foot", "global foot template (repeatable)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *template == "" {
		return fmt.Errorf("-template is required")
	}
	ctx, err := parseData(*data)
	if err != nil {
		return fmt.Errorf("invalid -data JSON: %w", err)
	}
	env := nunchucks.Configure(nunchucks.ConfigOptions{
		Path:                *views,
		GlobalTemplates:     []string(globalTemplates),
		GlobalHeadTemplates: []string(globalHeadTemplates),
		GlobalFootTemplates: []string(globalFootTemplates),
	})
	out, err := env.Render(*template, ctx)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(os.Stdout, out)
	return err
}

func runPrecompile(args []string) error {
	fs := flag.NewFlagSet("precompile", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = printPrecompileUsage

	views := fs.String("views", "views", "templates directory")
	outDir := fs.String("out", "public", "output directory")
	outFormat := fs.String("out-format", "preserve", "output naming: preserve or html")
	watch := fs.Bool("watch", false, "rerender when template files change")
	interval := fs.Duration("interval", time.Second, "polling interval for watch mode")
	data := fs.String("data", "{}", "JSON context object")
	var globalTemplates stringListFlag
	var globalHeadTemplates stringListFlag
	var globalFootTemplates stringListFlag
	fs.Var(&globalTemplates, "global", "global template (repeatable)")
	fs.Var(&globalHeadTemplates, "global-head", "global head template (repeatable)")
	fs.Var(&globalFootTemplates, "global-foot", "global foot template (repeatable)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	ctx, err := parseData(*data)
	if err != nil {
		return fmt.Errorf("invalid -data JSON: %w", err)
	}
	env := nunchucks.Configure(nunchucks.ConfigOptions{
		Path:                *views,
		GlobalTemplates:     []string(globalTemplates),
		GlobalHeadTemplates: []string(globalHeadTemplates),
		GlobalFootTemplates: []string(globalFootTemplates),
	})
	precompileOpts := nunchucks.PrecompileOptions{
		OutputFormat: *outFormat,
	}
	if *watch {
		return watchPrecompile(env, *views, *outDir, ctx, *interval, precompileOpts)
	}
	return env.PrecompileDirWithOptions(*outDir, ctx, precompileOpts)
}

func main() {
	if len(os.Args) < 2 {
		printRootUsage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "render":
		err = runRender(os.Args[2:])
	case "precompile":
		err = runPrecompile(os.Args[2:])
	case "version", "--version", "-version":
		printVersion()
		return
	case "-h", "--help", "help":
		topic := ""
		if len(os.Args) > 2 {
			topic = os.Args[2]
		}
		err = printHelp(topic)
		if err == nil {
			return
		}
	default:
		printRootUsage()
		err = fmt.Errorf("unknown command: %s", os.Args[1])
	}

	if err != nil {
		printError(err)
		os.Exit(1)
	}
}
