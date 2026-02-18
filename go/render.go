package nunchucks

import (
	"fmt"
	"regexp"
	"strings"
)

var exprRe = regexp.MustCompile(`\{\{\s*([\s\S]*?)\s*\}\}`)
var stmtRe = regexp.MustCompile(`\{%\s*([A-Za-z_][A-Za-z0-9_]*)\b([\s\S]*?)%\}`)

type MacroDef struct {
	Params []MacroParam
	Body   string
}

type MacroParam struct {
	Name       string
	Default    string
	HasDefault bool
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

	out, err = e.applyIncludes(out, ctx, vars, macros)
	if err != nil {
		return "", err
	}

	out, err = e.applyFromImports(out, ctx, vars, macros)
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

type includeSpec struct {
	Name          string
	IgnoreMissing bool
	WithContext   bool
}

func parseIncludeSpec(inner string) (includeSpec, bool) {
	s := strings.TrimSpace(inner)
	if !strings.HasPrefix(s, "include ") {
		return includeSpec{}, false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(s, "include "))
	quote := byte(0)
	if strings.HasPrefix(rest, "\"") {
		quote = '"'
	} else if strings.HasPrefix(rest, "'") {
		quote = '\''
	} else {
		return includeSpec{}, false
	}
	end := 1
	for end < len(rest) && rest[end] != quote {
		if rest[end] == '\\' {
			end += 2
			continue
		}
		end++
	}
	if end >= len(rest) {
		return includeSpec{}, false
	}
	name := rest[1:end]
	flags := strings.ToLower(strings.TrimSpace(rest[end+1:]))

	spec := includeSpec{Name: name, WithContext: true}
	if strings.Contains(flags, "ignore missing") {
		spec.IgnoreMissing = true
	}
	if strings.Contains(flags, "without context") {
		spec.WithContext = false
	}
	if strings.Contains(flags, "with context") {
		spec.WithContext = true
	}
	return spec, true
}

func (e *Env) applyIncludes(src string, ctx, vars map[string]any, macros map[string]MacroDef) (string, error) {
	re := regexp.MustCompile(`\{%\s*include\s+[\s\S]*?%\}`)
	out := src
	for {
		m := re.FindStringIndex(out)
		if m == nil {
			return out, nil
		}
		raw := out[m[0]:m[1]]
		inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(raw, "{%"), "%}"))
		spec, ok := parseIncludeSpec(inner)
		if !ok {
			return "", fmt.Errorf("invalid include statement: %s", raw)
		}

		includeSrc, err := e.readTemplate(spec.Name)
		if err != nil {
			if spec.IgnoreMissing {
				out = out[:m[0]] + out[m[1]:]
				continue
			}
			return "", err
		}

		var incCtx map[string]any
		var incVars map[string]any
		if spec.WithContext {
			incCtx = ctx
			incVars = cloneMap(vars)
		} else {
			incCtx = map[string]any{}
			incVars = map[string]any{}
		}

		rendered, err := e.renderWithState(includeSrc, incCtx, incVars, cloneMacros(macros))
		if err != nil {
			return "", err
		}
		out = out[:m[0]] + rendered + out[m[1]:]
	}
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

func parseMacroParams(s string) []MacroParam {
	parts := splitArgs(strings.TrimSpace(s))
	out := make([]MacroParam, 0, len(parts))
	for _, part := range parts {
		t := strings.TrimSpace(part)
		if t == "" {
			continue
		}
		if k, v, ok := splitTopLevelAssign(t); ok {
			out = append(out, MacroParam{Name: k, Default: v, HasDefault: true})
			continue
		}
		out = append(out, MacroParam{Name: t})
	}
	return out
}

func (e *Env) registerMacro(name string, def MacroDef, ctx, vars map[string]any, macros map[string]MacroDef) {
	macros[name] = def
	vars[name] = TemplateFunc(func(args []any, kwargs map[string]any, caller string) (any, error) {
		localVars := cloneMap(vars)
		for i, p := range def.Params {
			if i < len(args) {
				localVars[p.Name] = args[i]
				continue
			}
			if v, ok := kwargs[p.Name]; ok {
				localVars[p.Name] = v
				continue
			}
			if p.HasDefault {
				localVars[p.Name] = evalExpr(p.Default, localVars, ctx)
			} else {
				localVars[p.Name] = nil
			}
		}
		localVars["caller"] = TemplateFunc(func(_ []any, _ map[string]any, _ string) (any, error) {
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

		def := MacroDef{Params: parseMacroParams(argsRaw), Body: out[bodyStart:bodyEnd]}
		e.registerMacro(name, def, ctx, vars, macros)
		out = out[:open[0]] + out[closeEnd:]
	}
}

func parseImportedNames(s string) [][2]string {
	out := [][2]string{}
	for _, p := range splitArgs(s) {
		t := strings.TrimSpace(p)
		if t == "" {
			continue
		}
		re := regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)(?:\s+as\s+([A-Za-z_][A-Za-z0-9_]*))?$`)
		m := re.FindStringSubmatch(t)
		if m == nil {
			continue
		}
		local := m[1]
		if strings.TrimSpace(m[2]) != "" {
			local = m[2]
		}
		out = append(out, [2]string{m[1], local})
	}
	return out
}

func parseImportStmt(inner string) (file string, alias string, flags string, ok bool) {
	re := regexp.MustCompile(`^import\s+(".*?"|'.*?')\s+as\s+([A-Za-z_][A-Za-z0-9_]*)([\s\S]*)$`)
	m := re.FindStringSubmatch(strings.TrimSpace(inner))
	if m == nil {
		return "", "", "", false
	}
	return unquote(strings.TrimSpace(m[1])), strings.TrimSpace(m[2]), strings.TrimSpace(m[3]), true
}

func parseFromImportStmt(inner string) (file string, imports string, flags string, ok bool) {
	re := regexp.MustCompile(`^from\s+(".*?"|'.*?')\s+import\s+([\s\S]+?)(?:\s+(with\s+context|without\s+context))?$`)
	m := re.FindStringSubmatch(strings.TrimSpace(inner))
	if m == nil {
		return "", "", "", false
	}
	return unquote(strings.TrimSpace(m[1])), strings.TrimSpace(m[2]), strings.TrimSpace(m[3]), true
}

func parseContextModeFlags(flags string, defaultWithContext bool) bool {
	f := strings.ToLower(strings.TrimSpace(flags))
	withContext := defaultWithContext
	if strings.Contains(f, "without context") {
		withContext = false
	}
	if strings.Contains(f, "with context") {
		withContext = true
	}
	return withContext
}

func (e *Env) loadMacrosFromTemplate(name string, ctx map[string]any, withContext bool) (map[string]any, map[string]MacroDef, error) {
	importSrc, err := e.readTemplate(name)
	if err != nil {
		return nil, nil, err
	}
	resolved, err := e.resolveIncludes(importSrc, map[string]bool{name: true})
	if err != nil {
		return nil, nil, err
	}
	localMacros := map[string]MacroDef{}
	namespace := map[string]any{}
	macroCtx := map[string]any{}
	if withContext {
		macroCtx = ctx
	}
	_, err = e.applyMacroDefs(resolved, macroCtx, namespace, localMacros)
	if err != nil {
		return nil, nil, err
	}
	return namespace, localMacros, nil
}

func (e *Env) applyImports(src string, ctx, vars map[string]any, macros map[string]MacroDef) (string, error) {
	importRe := regexp.MustCompile(`\{%\s*import\s+[\s\S]*?%\}`)
	out := src

	for {
		m := importRe.FindStringSubmatchIndex(out)
		if m == nil {
			return out, nil
		}
		raw := out[m[0]:m[1]]
		inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(raw, "{%"), "%}"))
		name, alias, flags, ok := parseImportStmt(inner)
		if !ok {
			return "", fmt.Errorf("invalid import statement: %s", raw)
		}
		withContext := parseContextModeFlags(flags, false)

		namespace, localMacros, err := e.loadMacrosFromTemplate(name, ctx, withContext)
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

func (e *Env) applyFromImports(src string, ctx, vars map[string]any, macros map[string]MacroDef) (string, error) {
	fromRe := regexp.MustCompile(`\{%\s*from\s+[\s\S]*?%\}`)
	out := src
	for {
		m := fromRe.FindStringSubmatchIndex(out)
		if m == nil {
			return out, nil
		}
		raw := out[m[0]:m[1]]
		inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(raw, "{%"), "%}"))
		name, importSpec, flags, ok := parseFromImportStmt(inner)
		if !ok {
			return "", fmt.Errorf("invalid from-import statement: %s", raw)
		}
		withContext := parseContextModeFlags(flags, false)

		namespace, localMacros, err := e.loadMacrosFromTemplate(name, ctx, withContext)
		if err != nil {
			return "", err
		}
		for _, pair := range parseImportedNames(importSpec) {
			remote := pair[0]
			local := pair[1]
			if fn, ok := namespace[remote]; ok {
				vars[local] = fn
			}
			if def, ok := localMacros[remote]; ok {
				macros[local] = def
			}
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

		renderItem := func(item any, idx int, length int) error {
			nextVars := cloneMap(vars)
			nextVars[varName] = item
			nextVars["loop"] = map[string]any{
				"index":     idx + 1,
				"index0":    idx,
				"revindex":  length - idx,
				"revindex0": length - idx - 1,
				"first":     idx == 0,
				"last":      idx == length-1,
				"length":    length,
			}
			chunk, err := e.renderWithState(body, ctx, nextVars, cloneMacros(macros))
			if err != nil {
				return err
			}
			tmp += chunk
			return nil
		}

		switch vv := iter.(type) {
		case []any:
			for i, item := range vv {
				if err := renderItem(item, i, len(vv)); err != nil {
					return "", err
				}
			}
		case map[string]any:
			arr := make([]any, 0, len(vv))
			for _, item := range vv {
				arr = append(arr, item)
			}
			for i, item := range arr {
				if err := renderItem(item, i, len(arr)); err != nil {
					return "", err
				}
			}
		case nil:
			// no-op
		default:
			if err := renderItem(vv, 0, 1); err != nil {
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
			name, argExprs := parseFilterSpec(strings.TrimSpace(part))
			args := make([]any, 0, len(argExprs))
			for _, expr := range argExprs {
				args = append(args, evalExpr(expr, vars, ctx))
			}
			filtered = applyFilter(name, filtered, args)
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
