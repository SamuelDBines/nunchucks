package nunchucks

import (
	"strings"
	"testing"
)

func TestFilterBuiltinsAndArgs(t *testing.T) {
	env := Configure(ConfigOptions{Loader: &testLoader{files: map[string]string{}}})
	src := `
{{ missing | default("fallback") }}
{{ html | escape }}
{{ text | replace("world", "go") }}
{{ nums | join("-") }}
{{ nums | sum }}
{{ letters | sort() | join(",") }}
{{ name | trim | title }}
{{ phrase | truncate(8, false, "...") }}
{{ url | urlencode }}
`
	ctx := map[string]any{
		"html":    "<b>x</b>",
		"text":    "hello world",
		"nums":    []any{1, 2, 3},
		"letters": []any{"b", "a", "C"},
		"name":    "  hello world  ",
		"phrase":  "alpha beta gamma",
		"url":     "a b",
	}

	out, err := env.RenderString(src, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mustContain := []string{
		"fallback",
		"&lt;b&gt;x&lt;/b&gt;",
		"hello go",
		"1-2-3",
		"6",
		"a,b,C",
		"Hello World",
		"alpha...",
		"a+b",
	}

	for _, m := range mustContain {
		if !strings.Contains(out, m) {
			t.Fatalf("expected output to contain %q, got:\n%s", m, out)
		}
	}
}

func TestFilterCollectionParitySet(t *testing.T) {
	env := Configure(ConfigOptions{Loader: &testLoader{files: map[string]string{}}})
	src := `
{{ users | selectattr("active") | length }}
{{ users | rejectattr("active") | length }}
{{ nums | select(2) | length }}
{{ nums | reject(2) | length }}
{{ data | dictsort("key") | length }}
{{ users | groupby("role") | length }}
{{ users | sort(false, false, "name") | first | dump }}
`
	ctx := map[string]any{
		"nums": []any{1, 2, 2, 3},
		"data": map[string]any{"b": 2, "a": 1},
		"users": []any{
			map[string]any{"name": "Zed", "active": true, "role": "dev"},
			map[string]any{"name": "Amy", "active": false, "role": "ops"},
			map[string]any{"name": "Bob", "active": true, "role": "dev"},
		},
	}
	out, err := env.RenderString(src, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, must := range []string{"2", "1", "\"name\":\"Amy\""} {
		if !strings.Contains(out, must) {
			t.Fatalf("expected output to contain %q, got:\n%s", must, out)
		}
	}
}

func TestExpressionIsAndIn(t *testing.T) {
	env := Configure(ConfigOptions{Loader: &testLoader{files: map[string]string{}}})
	src := `
{{ 2 in nums }}
{{ 5 not in nums }}
{{ role is "admin" }}
{{ role is not "member" }}
{{ role is string }}
{{ missing is undefined }}
{{ role is defined }}
{{ missing is not defined }}
{{ nums is iterable }}
{{ 2 < 3 < 4 }}
`
	out, err := env.RenderString(src, map[string]any{
		"nums": []any{1, 2, 3},
		"role": "admin",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := compactWhitespace(out)
	if !strings.Contains(s, "true true true true true true true true true") {
		t.Fatalf("unexpected output: %q", out)
	}
}
