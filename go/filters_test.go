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
