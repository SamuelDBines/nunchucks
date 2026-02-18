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

func TestFromImportMacro(t *testing.T) {
	files := map[string]string{
		"main.njk":   `{% from "macros.njk" import badge %}{{ badge("sam") }}`,
		"macros.njk": `{% macro badge(label) %}<b>{{ label | upper }}</b>{% endmacro %}`,
	}
	env := Configure(ConfigOptions{Loader: &testLoader{files: files}})
	out, err := env.Render("main.njk", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `<b>SAM</b>`
	if compactWhitespace(out) != compactWhitespace(want) {
		t.Fatalf("unexpected output\nwant: %q\n got: %q", want, out)
	}
}

func TestMacroDefaultAndNamedArgs(t *testing.T) {
	src := `{% macro greet(name="friend", salutation="hello") %}{{ salutation }} {{ name }}{% endmacro %}
{{ greet() }}
{{ greet("sam") }}
{{ greet(name="mike", salutation="yo") }}`
	env := Configure(ConfigOptions{Loader: &testLoader{files: map[string]string{}}})
	out, err := env.RenderString(src, map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := compactWhitespace(out)
	if !strings.Contains(s, "hello friend") {
		t.Fatalf("missing default output: %q", out)
	}
	if !strings.Contains(s, "hello sam") {
		t.Fatalf("missing positional output: %q", out)
	}
	if !strings.Contains(s, "yo mike") {
		t.Fatalf("missing named output: %q", out)
	}
}

func TestExpressionPrecedenceComparisonsAndTernary(t *testing.T) {
	env := Configure(ConfigOptions{Loader: &testLoader{files: map[string]string{}}})
	src := `{{ 1 + 2 * 3 }}
{{ (1 + 2) * 3 }}
{{ 10 > 3 and 2 < 5 }}
{{ "yes" if user.isAdmin else "no" }}
{{ 10 / 2 + 1 }}`
	out, err := env.RenderString(src, map[string]any{
		"user": map[string]any{"isAdmin": true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := compactWhitespace(out)
	for _, must := range []string{"7", "9", "true", "yes", "6"} {
		if !strings.Contains(s, must) {
			t.Fatalf("expected %q in output, got %q", must, out)
		}
	}
}

func TestIncludeIgnoreMissingAndWithoutContext(t *testing.T) {
	files := map[string]string{
		"main.njk": `{% include "part.njk" without context %}
{% include "missing.njk" ignore missing %}
{% include "part.njk" with context %}`,
		"part.njk": `Hello {{ name | default("guest", true) }}`,
	}
	env := Configure(ConfigOptions{Loader: &testLoader{files: files}})
	out, err := env.Render("main.njk", map[string]any{"name": "sam"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := compactWhitespace(out)
	if !strings.Contains(s, "Hello guest") {
		t.Fatalf("expected without-context include to hide name, got: %q", out)
	}
	if !strings.Contains(s, "Hello sam") {
		t.Fatalf("expected with-context include to see name, got: %q", out)
	}
}

func TestLoopMetadata(t *testing.T) {
	env := Configure(ConfigOptions{Loader: &testLoader{files: map[string]string{}}})
	src := `{% for item in items %}[{{ loop.index0 }}:{{ loop.first }}:{{ loop.last }}:{{ item }}]{% endfor %}`
	out, err := env.RenderString(src, map[string]any{
		"items": []any{"a", "b", "c"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `[0:true:false:a][1:false:false:b][2:false:true:c]`
	if compactWhitespace(out) != compactWhitespace(want) {
		t.Fatalf("unexpected output\nwant: %q\n got: %q", want, out)
	}
}

func TestImportContextModes(t *testing.T) {
	files := map[string]string{
		"main.njk": `{% import "macros.njk" as a %}
{% import "macros.njk" as b with context %}
{{ a.showUser() }}|{{ b.showUser() }}`,
		"macros.njk": `{% macro showUser() %}{{ user.name | default("guest", true) }}{% endmacro %}`,
	}
	env := Configure(ConfigOptions{Loader: &testLoader{files: files}})
	out, err := env.Render("main.njk", map[string]any{"user": map[string]any{"name": "sam"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := compactWhitespace(out)
	if !strings.Contains(s, "guest|sam") {
		t.Fatalf("expected default import context behavior guest|sam, got %q", out)
	}
}

func TestFromImportContextModes(t *testing.T) {
	files := map[string]string{
		"main.njk": `{% from "macros.njk" import showUser as one %}
{% from "macros.njk" import showUser as two with context %}
{{ one() }}|{{ two() }}`,
		"macros.njk": `{% macro showUser() %}{{ user.name | default("guest", true) }}{% endmacro %}`,
	}
	env := Configure(ConfigOptions{Loader: &testLoader{files: files}})
	out, err := env.Render("main.njk", map[string]any{"user": map[string]any{"name": "sam"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := compactWhitespace(out)
	if !strings.Contains(s, "guest|sam") {
		t.Fatalf("expected from-import context behavior guest|sam, got %q", out)
	}
}
