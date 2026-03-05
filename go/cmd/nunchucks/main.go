package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	nunchucks "github.com/SamuelDBines/nunjucks/go"
)

var version = "dev"

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
	fmt.Fprintln(os.Stderr, "  -data string         JSON context object (default \"{}\")")
	fmt.Fprintln(os.Stderr, "  -global value        global template, repeatable")
	fmt.Fprintln(os.Stderr, "  -global-head value   global head template, repeatable")
	fmt.Fprintln(os.Stderr, "  -global-foot value   global foot template, repeatable")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Example:")
	fmt.Fprintf(os.Stderr, "  %s precompile -views ./views -out ./public -data '{\"title\":\"Hello\"}'\n", name)
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
	return env.PrecompileDir(*outDir, ctx)
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
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
