package nunchucks

import (
	"regexp"
	"strings"
	"testing"
)

type testLoader struct {
	files map[string]string
}

func (m *testLoader) TypeName() string { return "memory" }

func (m *testLoader) Source(name string) LoaderResponse {
	if _, ok := m.files[name]; !ok {
		return LoaderResponse{Err: "missing template", Res: name}
	}
	return LoaderResponse{Res: name}
}

func (m *testLoader) Read(name string) LoaderResponse {
	v, ok := m.files[name]
	if !ok {
		return LoaderResponse{Err: "missing template", Res: name}
	}
	return LoaderResponse{Res: v}
}

func TestRenderExprSetIfFor(t *testing.T) {
	env := Configure(ConfigOptions{Loader: &testLoader{files: map[string]string{}}})
	src := `{% set title = "Hello" %}
{{ title | upper }}
{% if user %}yes{% else %}no{% endif %}
{% for item in items %}[{{ item }}]{% endfor %}`

	out, err := env.RenderString(src, map[string]any{
		"user":  true,
		"items": []any{"a", "b"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "\nHELLO\nyes\n[a][b]"
	if out != want {
		t.Fatalf("unexpected output\nwant: %q\n got: %q", want, out)
	}
}

func TestRenderExtendsAndInclude(t *testing.T) {
	files := map[string]string{
		"base.njk":  "<h1>{% block title %}Base{% endblock %}</h1> {% block content %}C{% endblock %}",
		"child.njk": `{% extends "base.njk" %}{% block title %}Child{% endblock %}{% block content %}Body {% include "part.njk" %}{% endblock %}`,
		"part.njk":  `{{ msg | lower }}`,
	}
	env := Configure(ConfigOptions{Loader: &testLoader{files: files}})
	out, err := env.Render("child.njk", map[string]any{"msg": "LOUD"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "<h1>Child</h1> Body loud"
	if out != want {
		t.Fatalf("unexpected output\nwant: %q\n got: %q", want, out)
	}
}

func compactWhitespace(s string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

func TestMacroAndCall(t *testing.T) {
	src := `{% macro wrap(cls) %}<div class="{{ cls }}">{{ caller() }}</div>{% endmacro %}
{% call wrap("box") %}inside {{ name }}{% endcall %}`
	env := Configure(ConfigOptions{Loader: &testLoader{files: map[string]string{}}})
	out, err := env.RenderString(src, map[string]any{"name": "sam"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `<div class="box">inside sam</div>`
	if compactWhitespace(out) != compactWhitespace(want) {
		t.Fatalf("unexpected output\nwant: %q\n got: %q", want, out)
	}
}

func TestImportAndMacroCall(t *testing.T) {
	files := map[string]string{
		"main.njk":   `{% import "macros.njk" as ui %}{{ ui.badge(name) }}`,
		"macros.njk": `{% macro badge(label) %}<b>{{ label | upper }}</b>{% endmacro %}`,
	}
	env := Configure(ConfigOptions{Loader: &testLoader{files: files}})
	out, err := env.Render("main.njk", map[string]any{"name": "ok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `<b>OK</b>`
	if compactWhitespace(out) != compactWhitespace(want) {
		t.Fatalf("unexpected output\nwant: %q\n got: %q", want, out)
	}
}

func TestRawVerbatimAndFilterBlock(t *testing.T) {
	src := `{% raw %}{{ untouched }}{% endraw %}
{% verbatim %}{% if nope %}x{% endif %}{% endverbatim %}
{% filter upper %}hello {{ name }}{% endfilter %}`
	env := Configure(ConfigOptions{Loader: &testLoader{files: map[string]string{}}})
	out, err := env.RenderString(src, map[string]any{"name": "sam"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "{{ untouched }}") {
		t.Fatalf("raw block was not preserved: %q", out)
	}
	if !strings.Contains(out, "{% if nope %}x{% endif %}") {
		t.Fatalf("verbatim block was not preserved: %q", out)
	}
	if !strings.Contains(out, "HELLO SAM") {
		t.Fatalf("filter block not applied: %q", out)
	}
}
