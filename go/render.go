package nunchucks

import (
	"fmt"
	"regexp"
	"strings"
)

var exprRe = regexp.MustCompile(`\{\{\s*([\s\S]*?)\s*\}\}`)
var stmtRe = regexp.MustCompile(`\{%\s*([A-Za-z_][A-Za-z0-9_]*)\b([\s\S]*?)%\}`)

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func (e *Env) renderString(src string, ctx map[string]any) (string, error) {
	if ctx == nil {
		ctx = map[string]any{}
	}
	vars := map[string]any{}
	out := src

	var err error
	out, err = applySets(out, vars, ctx)
	if err != nil {
		return "", err
	}

	out, err = applyForLoops(out, vars, ctx)
	if err != nil {
		return "", err
	}

	out, err = applyIfElse(out, vars, ctx)
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
	return out, nil
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
		parts := strings.Split(body, ",")
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

func applyForLoops(src string, vars, ctx map[string]any) (string, error) {
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
			chunk, err := applySets(body, nextVars, ctx)
			if err != nil {
				return err
			}
			chunk, err = applyForLoops(chunk, nextVars, ctx)
			if err != nil {
				return err
			}
			chunk, err = applyIfElse(chunk, nextVars, ctx)
			if err != nil {
				return err
			}
			chunk = exprRe.ReplaceAllStringFunc(chunk, func(m string) string {
				mm := exprRe.FindStringSubmatch(m)
				if len(mm) < 2 {
					return ""
				}
				return fmt.Sprint(evalExpr(mm[1], nextVars, ctx))
			})
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

func applyIfElse(src string, vars, ctx map[string]any) (string, error) {
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
		}{
			{idx: len(middle), isE: true, cond: ""},
		}
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

		chosenOut, err := applyIfElse(chosen, vars, ctx)
		if err != nil {
			return "", err
		}

		out = out[:open[0]] + chosenOut + out[endAbsEnd:]
	}
}
