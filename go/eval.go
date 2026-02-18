package nunchucks

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type TemplateFunc func(args []any, caller string) (any, error)

var callableExprRe = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)*)\s*\((.*)\)$`)

func getPath(data map[string]any, path string) (any, bool) {
	parts := strings.Split(path, ".")
	var cur any = data
	for _, p := range parts {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		v, exists := m[p]
		if !exists {
			return nil, false
		}
		cur = v
	}
	return cur, true
}

func parseLiteral(s string) (any, bool) {
	t := strings.TrimSpace(s)
	if t == "" {
		return "", true
	}
	if (strings.HasPrefix(t, `"`) && strings.HasSuffix(t, `"`)) || (strings.HasPrefix(t, "'") && strings.HasSuffix(t, "'")) {
		return t[1 : len(t)-1], true
	}
	switch t {
	case "true":
		return true, true
	case "false":
		return false, true
	case "null", "nil":
		return nil, true
	}
	if i, err := strconv.Atoi(t); err == nil {
		return i, true
	}
	if f, err := strconv.ParseFloat(t, 64); err == nil {
		return f, true
	}
	return nil, false
}

func splitArgs(src string) []string {
	out := []string{}
	cur := strings.Builder{}
	depth := 0
	quote := byte(0)
	esc := false

	for i := 0; i < len(src); i++ {
		ch := src[i]
		if esc {
			cur.WriteByte(ch)
			esc = false
			continue
		}
		if quote != 0 {
			if ch == '\\' {
				esc = true
				cur.WriteByte(ch)
				continue
			}
			cur.WriteByte(ch)
			if ch == quote {
				quote = 0
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			quote = ch
			cur.WriteByte(ch)
			continue
		}
		switch ch {
		case '(', '[', '{':
			depth++
			cur.WriteByte(ch)
		case ')', ']', '}':
			if depth > 0 {
				depth--
			}
			cur.WriteByte(ch)
		case ',':
			if depth == 0 {
				t := strings.TrimSpace(cur.String())
				if t != "" {
					out = append(out, t)
				}
				cur.Reset()
				continue
			}
			cur.WriteByte(ch)
		default:
			cur.WriteByte(ch)
		}
	}
	t := strings.TrimSpace(cur.String())
	if t != "" {
		out = append(out, t)
	}
	return out
}

func parseCallableExpr(src string) (string, []string, bool) {
	s := strings.TrimSpace(src)
	m := callableExprRe.FindStringSubmatch(s)
	if m == nil {
		return "", nil, false
	}
	name := strings.TrimSpace(m[1])
	argsRaw := strings.TrimSpace(m[2])
	if argsRaw == "" {
		return name, nil, true
	}
	return name, splitArgs(argsRaw), true
}

func resolveIdent(name string, vars, ctx map[string]any) any {
	key := strings.TrimSpace(name)
	if key == "" {
		return ""
	}

	if v, ok := vars[key]; ok {
		return v
	}
	if v, ok := ctx[key]; ok {
		return v
	}

	if strings.Contains(key, ".") {
		if v, ok := getPath(vars, key); ok {
			return v
		}
		if v, ok := getPath(ctx, key); ok {
			return v
		}
	}

	return ""
}

func applyFilter(name string, v any) any {
	switch strings.TrimSpace(name) {
	case "lower":
		return strings.ToLower(fmt.Sprint(v))
	case "upper":
		return strings.ToUpper(fmt.Sprint(v))
	default:
		return v
	}
}

func invokeCallable(name string, argExprs []string, caller string, vars, ctx map[string]any) (any, error) {
	raw := resolveIdent(name, vars, ctx)
	args := make([]any, 0, len(argExprs))
	for _, a := range argExprs {
		args = append(args, evalExpr(a, vars, ctx))
	}

	switch fn := raw.(type) {
	case TemplateFunc:
		return fn(args, caller)
	case func(...any) any:
		return fn(args...), nil
	default:
		return "", fmt.Errorf("%s is not callable", name)
	}
}

func evalExpr(expr string, vars, ctx map[string]any) any {
	parts := strings.Split(expr, "|")
	if len(parts) == 0 {
		return ""
	}

	base := strings.TrimSpace(parts[0])
	var value any

	if name, args, ok := parseCallableExpr(base); ok {
		called, err := invokeCallable(name, args, "", vars, ctx)
		if err == nil {
			value = called
		} else {
			value = ""
		}
	} else if lit, ok := parseLiteral(base); ok {
		value = lit
	} else {
		value = resolveIdent(base, vars, ctx)
	}

	for i := 1; i < len(parts); i++ {
		value = applyFilter(parts[i], value)
	}

	return value
}

func truthy(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return strings.TrimSpace(x) != ""
	case int:
		return x != 0
	case int64:
		return x != 0
	case float64:
		return x != 0
	case []any:
		return len(x) > 0
	case map[string]any:
		return len(x) > 0
	default:
		return true
	}
}

func evalCond(cond string, vars, ctx map[string]any) bool {
	s := strings.TrimSpace(cond)
	if s == "" {
		return false
	}

	if strings.HasPrefix(s, "not ") {
		return !truthy(evalExpr(strings.TrimSpace(strings.TrimPrefix(s, "not ")), vars, ctx))
	}

	if strings.Contains(s, " and ") {
		for _, p := range strings.Split(s, " and ") {
			if !truthy(evalExpr(strings.TrimSpace(p), vars, ctx)) {
				return false
			}
		}
		return true
	}

	if strings.Contains(s, " or ") {
		for _, p := range strings.Split(s, " or ") {
			if truthy(evalExpr(strings.TrimSpace(p), vars, ctx)) {
				return true
			}
		}
		return false
	}

	return truthy(evalExpr(s, vars, ctx))
}
