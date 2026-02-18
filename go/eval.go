package nunchucks

import (
	"fmt"
	"strconv"
	"strings"
)

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

func evalExpr(expr string, vars, ctx map[string]any) any {
	parts := strings.Split(expr, "|")
	if len(parts) == 0 {
		return ""
	}

	base := strings.TrimSpace(parts[0])
	var value any
	if lit, ok := parseLiteral(base); ok {
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
