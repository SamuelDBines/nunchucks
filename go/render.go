package nunchucks

import (
	"fmt"
	"regexp"
	"strings"
)

var exprRe = regexp.MustCompile(`\{\{\s*([\s\S]*?)\s*\}\}`)
var stmtRe = regexp.MustCompile(`\{%\s*([A-Za-z_][A-Za-z0-9_]*)\b([\s\S]*?)%\}`)

type MacroDef struct {
	Args []string
	Body string
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneMacros(in map[string]MacroDef) map[string]MacroDef {
	out := make(map[string]MacroDef, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func (e *Env) renderString(src string, ctx map[string]any) (string, error) {
	if ctx == nil {
		ctx = map[string]any{}
	}
	return e.renderWithState(src, ctx, map[string]any{}, map[string]MacroDef{})
}

func (e *Env) renderWithState(src string, ctx, vars map[string]any, macros map[string]MacroDef) (string, error) {
	out := src
	var err error

	out, raws, err := extractRawBlocks(out)
	if err != nil {
		return "", err
	}

	out, err = e.applyImports(out, ctx, vars, macros)
	if err != nil {
		return "", err
	}

	out, err = e.applyMacroDefs(out, ctx, vars, macros)
	if err != nil {
		return "", err
	}

	out, err = applySets(out, vars, ctx)
	if err != nil {
		return "", err
	}

	out, err = e.applyForLoops(out, vars, ctx, macros)
	if err != nil {
		return "", err
	}

	out, err = e.applyIfElse(out, vars, ctx, macros)
	if err != nil {
		return "", err
	}

	out, err = e.applyFilterBlocks(out, vars, ctx, macros)
	if err != nil {
		return "", err
	}

	out, err = e.applyCallBlocks(out, vars, ctx, macros)
	if err != nil {
		return "", err
	}

	out = exprRe.ReplaceAllStringFunc(out, func(m string) string {
		mm := exprRe.FindStringSubmatch(m)
		if len(mm) < 2 {
			return ""
		}
		return fmt.Sprint(evalExpr(mm[1], vars, ctx))
	})

	out = stmtRe.ReplaceAllString(out, "")
	out = restoreRawBlocks(out, raws)
	return out, nil
}

func extractRawBlocks(src string) (string, map[string]string, error) {
	out := src
	store := map[string]string{}
	index := 0

	handle := func(tag string) error {
		openRe := regexp.MustCompile(`\{%\s*` + tag + `\s*%\}`)
		closeRe := regexp.MustCompile(`\{%\s*end` + tag + `\s*%\}`)
		for {
			open := openRe.FindStringIndex(out)
			if open == nil {
				return nil
			}
			close := closeRe.FindStringIndex(out[open[1]:])
			if close == nil {
				return fmt.Errorf("missing end%s", tag)
			}
			bodyStart := open[1]
			bodyEnd := open[1] + close[0]
			closeEnd := open[1] + close[1]
			key := fmt.Sprintf("@@NUNCHUCKS_RAW_%d@@", index)
			index++
			store[key] = out[bodyStart:bodyEnd]
			out = out[:open[0]] + key + out[closeEnd:]
		}
	}

	if err := handle("raw"); err != nil {
		return "", nil, err
	}
	if err := handle("verbatim"); err != nil {
		return "", nil, err
	}

	return out, store, nil
}

func restoreRawBlocks(src string, store map[string]string) string {
	out := src
	for k, v := range store {
		out = strings.ReplaceAll(out, k, v)
	}
	return out
}

func parseArgsList(s string) []string {
	parts := splitArgs(strings.TrimSpace(s))
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func (e *Env) registerMacro(name string, def MacroDef, ctx, vars map[string]any, macros map[string]MacroDef) {
	macros[name] = def
	vars[name] = TemplateFunc(func(args []any, caller string) (any, error) {
		localVars := cloneMap(vars)
		for i, argName := range def.Args {
			if i < len(args) {
				localVars[argName] = args[i]
			} else {
				localVars[argName] = nil
			}
		}
		localVars["caller"] = TemplateFunc(func(_ []any, _ string) (any, error) {
			return caller, nil
		})
		return e.renderWithState(def.Body, ctx, localVars, cloneMacros(macros))
	})
}

func (e *Env) applyMacroDefs(src string, ctx, vars map[string]any, macros map[string]MacroDef) (string, error) {
	openRe := regexp.MustCompile(`\{%\s*macro\s+([A-Za-z_][A-Za-z0-9_]*)\s*\((.*?)\)\s*%\}`)
	closeRe := regexp.MustCompile(`\{%\s*endmacro\s*%\}`)
	out := src

	for {
		open := openRe.FindStringSubmatchIndex(out)
		if open == nil {
			return out, nil
		}
		close := closeRe.FindStringIndex(out[open[1]:])
		if close == nil {
			return "", fmt.Errorf("missing endmacro")
		}

		name := strings.TrimSpace(out[open[2]:open[3]])
		argsRaw := strings.TrimSpace(out[open[4]:open[5]])
		bodyStart := open[1]
		bodyEnd := open[1] + close[0]
		closeEnd := open[1] + close[1]

		def := MacroDef{Args: parseArgsList(argsRaw), Body: out[bodyStart:bodyEnd]}
		e.registerMacro(name, def, ctx, vars, macros)
		out = out[:open[0]] + out[closeEnd:]
	}
}

func (e *Env) applyImports(src string, ctx, vars map[string]any, macros map[string]MacroDef) (string, error) {
	importRe := regexp.MustCompile(`\{%\s*import\s+(["'][^"']+["'])\s+as\s+([A-Za-z_][A-Za-z0-9_]*)\s*%\}`)
	out := src

	for {
		m := importRe.FindStringSubmatchIndex(out)
		if m == nil {
			return out, nil
		}

		fileTok := out[m[2]:m[3]]
		alias := strings.TrimSpace(out[m[4]:m[5]])
		name := unquote(fileTok)

		importSrc, err := e.readTemplate(name)
		if err != nil {
			return "", err
		}
		resolved, err := e.resolveIncludes(importSrc, map[string]bool{name: true})
		if err != nil {
			return "", err
		}

		localMacros := map[string]MacroDef{}
		namespace := map[string]any{}
		_, err = e.applyMacroDefs(resolved, ctx, namespace, localMacros)
		if err != nil {
			return "", err
		}
		vars[alias] = namespace

		for k, v := range localMacros {
			macros[alias+"."+k] = v
		}

		out = out[:m[0]] + out[m[1]:]
	}
}

func applySets(src string, vars, ctx map[string]any) (string, error) {
	setRe := regexp.MustCompile(`\{%\s*set\s+([\s\S]*?)%\}`)
	out := src
	for {
		m := setRe.FindStringSubmatchIndex(out)
		if m == nil {
			return out, nil
		}
		body := strings.TrimSpace(out[m[2]:m[3]])
		parts := splitArgs(body)
		for _, part := range parts {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) != 2 {
				continue
			}
			k := strings.TrimSpace(kv[0])
			v := evalExpr(strings.TrimSpace(kv[1]), vars, ctx)
			vars[k] = v
		}
		out = out[:m[0]] + out[m[1]:]
	}
}

func (e *Env) applyForLoops(src string, vars, ctx map[string]any, macros map[string]MacroDef) (string, error) {
	forOpenRe := regexp.MustCompile(`\{%\s*for\s+([A-Za-z_][A-Za-z0-9_]*)\s+in\s+([\s\S]*?)%\}`)
	endForRe := regexp.MustCompile(`\{%\s*endfor\s*%\}`)
	out := src

	for {
		open := forOpenRe.FindStringSubmatchIndex(out)
		if open == nil {
			return out, nil
		}

		close := endForRe.FindStringIndex(out[open[1]:])
		if close == nil {
			return "", fmt.Errorf("missing endfor")
		}
		closeStart := open[1] + close[0]
		closeEnd := open[1] + close[1]

		varName := strings.TrimSpace(out[open[2]:open[3]])
		expr := strings.TrimSpace(out[open[4]:open[5]])
		body := out[open[1]:closeStart]

		iter := evalExpr(expr, vars, ctx)
		tmp := ""

		renderItem := func(item any) error {
			nextVars := cloneMap(vars)
			nextVars[varName] = item
			chunk, err := e.renderWithState(body, ctx, nextVars, cloneMacros(macros))
			if err != nil {
				return err
			}
			tmp += chunk
			return nil
		}

		switch vv := iter.(type) {
		case []any:
			for _, item := range vv {
				if err := renderItem(item); err != nil {
					return "", err
				}
			}
		case map[string]any:
			for _, item := range vv {
				if err := renderItem(item); err != nil {
					return "", err
				}
			}
		case nil:
			// no-op
		default:
			if err := renderItem(vv); err != nil {
				return "", err
			}
		}

		out = out[:open[0]] + tmp + out[closeEnd:]
	}
}

func (e *Env) applyIfElse(src string, vars, ctx map[string]any, macros map[string]MacroDef) (string, error) {
	ifOpenRe := regexp.MustCompile(`\{%\s*if\s+([\s\S]*?)%\}`)
	elifRe := regexp.MustCompile(`\{%\s*elif\s+([\s\S]*?)%\}`)
	elseRe := regexp.MustCompile(`\{%\s*else\s*%\}`)
	endIfRe := regexp.MustCompile(`\{%\s*endif\s*%\}`)

	out := src
	for {
		open := ifOpenRe.FindStringSubmatchIndex(out)
		if open == nil {
			return out, nil
		}

		cond := strings.TrimSpace(out[open[2]:open[3]])
		rest := out[open[1]:]
		endif := endIfRe.FindStringIndex(rest)
		if endif == nil {
			return "", fmt.Errorf("missing endif")
		}

		endAbsStart := open[1] + endif[0]
		endAbsEnd := open[1] + endif[1]
		middle := out[open[1]:endAbsStart]

		type branch struct {
			cond *string
			body string
		}
		branches := []branch{}

		cursor := 0
		firstCuts := []struct {
			idx  int
			isE  bool
			cond string
		}{{idx: len(middle), isE: true, cond: ""}}
		for _, m := range elifRe.FindAllStringSubmatchIndex(middle, -1) {
			firstCuts = append(firstCuts, struct {
				idx  int
				isE  bool
				cond string
			}{idx: m[0], isE: false, cond: strings.TrimSpace(middle[m[2]:m[3]])})
		}
		if m := elseRe.FindStringIndex(middle); m != nil {
			firstCuts = append(firstCuts, struct {
				idx  int
				isE  bool
				cond string
			}{idx: m[0], isE: true, cond: ""})
		}

		minIdx := len(middle)
		for _, c := range firstCuts {
			if c.idx < minIdx {
				minIdx = c.idx
			}
		}
		firstBody := middle[cursor:minIdx]
		c := cond
		branches = append(branches, branch{cond: &c, body: firstBody})

		segments := []struct {
			kind  string
			cond  string
			start int
			end   int
		}{}

		for _, m := range elifRe.FindAllStringSubmatchIndex(middle, -1) {
			segments = append(segments, struct {
				kind  string
				cond  string
				start int
				end   int
			}{kind: "elif", cond: strings.TrimSpace(middle[m[2]:m[3]]), start: m[0], end: m[1]})
		}
		if m := elseRe.FindStringIndex(middle); m != nil {
			segments = append(segments, struct {
				kind  string
				cond  string
				start int
				end   int
			}{kind: "else", cond: "", start: m[0], end: m[1]})
		}

		for i := 0; i < len(segments); i++ {
			start := segments[i].end
			end := len(middle)
			for j := 0; j < len(segments); j++ {
				if segments[j].start > segments[i].start && segments[j].start < end {
					end = segments[j].start
				}
			}
			if segments[i].kind == "elif" {
				cc := segments[i].cond
				branches = append(branches, branch{cond: &cc, body: middle[start:end]})
			} else {
				branches = append(branches, branch{cond: nil, body: middle[start:end]})
			}
		}

		chosen := ""
		for _, b := range branches {
			if b.cond == nil {
				chosen = b.body
				break
			}
			if evalCond(*b.cond, vars, ctx) {
				chosen = b.body
				break
			}
		}

		chosenOut, err := e.renderWithState(chosen, ctx, cloneMap(vars), cloneMacros(macros))
		if err != nil {
			return "", err
		}

		out = out[:open[0]] + chosenOut + out[endAbsEnd:]
	}
}

func (e *Env) applyFilterBlocks(src string, vars, ctx map[string]any, macros map[string]MacroDef) (string, error) {
	openRe := regexp.MustCompile(`\{%\s*filter\s+([\s\S]*?)%\}`)
	closeRe := regexp.MustCompile(`\{%\s*endfilter\s*%\}`)
	out := src

	for {
		open := openRe.FindStringSubmatchIndex(out)
		if open == nil {
			return out, nil
		}
		close := closeRe.FindStringIndex(out[open[1]:])
		if close == nil {
			return "", fmt.Errorf("missing endfilter")
		}

		filterExpr := strings.TrimSpace(out[open[2]:open[3]])
		bodyStart := open[1]
		bodyEnd := open[1] + close[0]
		closeEnd := open[1] + close[1]
		body := out[bodyStart:bodyEnd]

		rendered, err := e.renderWithState(body, ctx, cloneMap(vars), cloneMacros(macros))
		if err != nil {
			return "", err
		}

		filtered := any(rendered)
		for _, part := range strings.Split(filterExpr, "|") {
			filtered = applyFilter(strings.TrimSpace(part), filtered)
		}

		out = out[:open[0]] + fmt.Sprint(filtered) + out[closeEnd:]
	}
}

func (e *Env) applyCallBlocks(src string, vars, ctx map[string]any, macros map[string]MacroDef) (string, error) {
	openRe := regexp.MustCompile(`\{%\s*call\s+([\s\S]*?)%\}`)
	closeRe := regexp.MustCompile(`\{%\s*endcall\s*%\}`)
	out := src

	for {
		open := openRe.FindStringSubmatchIndex(out)
		if open == nil {
			return out, nil
		}
		close := closeRe.FindStringIndex(out[open[1]:])
		if close == nil {
			return "", fmt.Errorf("missing endcall")
		}

		callExpr := strings.TrimSpace(out[open[2]:open[3]])
		bodyStart := open[1]
		bodyEnd := open[1] + close[0]
		closeEnd := open[1] + close[1]
		body := out[bodyStart:bodyEnd]

		name, args, ok := parseCallableExpr(callExpr)
		if !ok {
			return "", fmt.Errorf("invalid call expression: %s", callExpr)
		}

		callerBody, err := e.renderWithState(body, ctx, cloneMap(vars), cloneMacros(macros))
		if err != nil {
			return "", err
		}
		called, err := invokeCallable(name, args, callerBody, vars, ctx)
		if err != nil {
			return "", err
		}

		out = out[:open[0]] + fmt.Sprint(called) + out[closeEnd:]
	}
}
