package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	nunchucks "github.com/SamuelDBines/nunjucks/go"
)

func usage() {
	fmt.Fprintln(os.Stderr, "nunchucks <command> [options]")
	fmt.Fprintln(os.Stderr, "commands:")
	fmt.Fprintln(os.Stderr, "  render      Render one template to stdout")
	fmt.Fprintln(os.Stderr, "  precompile  Render all templates from views to output directory")
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
	views := fs.String("views", "views", "templates directory")
	template := fs.String("template", "", "template path relative to views")
	data := fs.String("data", "{}", "JSON context object")
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
	env := nunchucks.Configure(nunchucks.ConfigOptions{Path: *views})
	out, err := env.Render(*template, ctx)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(os.Stdout, out)
	return err
}

func runPrecompile(args []string) error {
	fs := flag.NewFlagSet("precompile", flag.ContinueOnError)
	views := fs.String("views", "views", "templates directory")
	outDir := fs.String("out", "public", "output directory")
	data := fs.String("data", "{}", "JSON context object")
	if err := fs.Parse(args); err != nil {
		return err
	}
	ctx, err := parseData(*data)
	if err != nil {
		return fmt.Errorf("invalid -data JSON: %w", err)
	}
	env := nunchucks.Configure(nunchucks.ConfigOptions{Path: *views})
	return env.PrecompileDir(*outDir, ctx)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "render":
		err = runRender(os.Args[2:])
	case "precompile":
		err = runPrecompile(os.Args[2:])
	case "-h", "--help", "help":
		usage()
		return
	default:
		usage()
		err = fmt.Errorf("unknown command: %s", os.Args[1])
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
