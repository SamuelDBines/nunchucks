package nunchucks

import (
	"os"
	"path/filepath"
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

func TestClientBlockTransformsToModuleScript(t *testing.T) {
	env := Configure(ConfigOptions{Loader: &testLoader{files: map[string]string{}}})
	src := `{% client %}
{% fetch | '/api/users/' + userID | as users | %}
console.log(users.length);
{% endclient %}`

	out, err := env.RenderString(src, map[string]any{"userID": 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, `<script type="module" data-nunchucks-client>`) {
		t.Fatalf("missing client module script wrapper: %q", out)
	}
	if !strings.Contains(out, `const __nc_fetch_res_0 = await fetch("/api/users/42");`) {
		t.Fatalf("missing fetch transform output: %q", out)
	}
	if !strings.Contains(out, `const users = await __nc_fetch_res_0.json();`) {
		t.Fatalf("missing json assignment output: %q", out)
	}
}

func TestClientBlockFetchTextMode(t *testing.T) {
	env := Configure(ConfigOptions{Loader: &testLoader{files: map[string]string{}}})
	src := `{% client %}
{% fetch | '/status.txt' | as statusText | text | %}
console.log(statusText);
{% endclient %}`

	out, err := env.RenderString(src, map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, `const __nc_fetch_res_0 = await fetch("/status.txt");`) {
		t.Fatalf("missing fetch call output: %q", out)
	}
	if !strings.Contains(out, `const statusText = await __nc_fetch_res_0.text();`) {
		t.Fatalf("missing text mode output: %q", out)
	}
}

func TestClientBlockStateStatement(t *testing.T) {
	env := Configure(ConfigOptions{Loader: &testLoader{files: map[string]string{}}})
	src := `{% client %}
{% state app | { pressed: false, userId: userID } | %}
app.pressed = !app.pressed;
{% endclient %}`

	out, err := env.RenderString(src, map[string]any{"userID": 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, `window.__nunchucks = window.__nunchucks || { state: {} };`) {
		t.Fatalf("missing nunchucks state bootstrap: %q", out)
	}
	if !strings.Contains(out, `window.__nunchucks.state["app"] = {"pressed":false,"userId":42};`) {
		t.Fatalf("missing state payload transform: %q", out)
	}
	if !strings.Contains(out, `const app = window.__nunchucks.state["app"];`) {
		t.Fatalf("missing state variable alias: %q", out)
	}
}

func TestInlineClientEventBindingTransform(t *testing.T) {
	env := Configure(ConfigOptions{Loader: &testLoader{files: map[string]string{}}})
	src := `{% client %}{% state app | { pressed: false } | %}{% endclient %}
<button id="btn" onClick={{ app.pressed = true }}>Press</button>`

	out, err := env.RenderString(src, map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, `data-nc-onclick="__nc_evt_0"`) {
		t.Fatalf("missing inline event attribute transform: %q", out)
	}
	if !strings.Contains(out, `<script type="module" data-nunchucks-client-events>`) {
		t.Fatalf("missing inline event runtime script: %q", out)
	}
	if !strings.Contains(out, `with (window.__nunchucks.state) { app.pressed = true; }`) {
		t.Fatalf("missing inline event expression runtime: %q", out)
	}
}

func TestGlobalTemplatesInjectAcrossAllFiles(t *testing.T) {
	files := map[string]string{
		"a.njk":       `<div>A</div>`,
		"b.njk":       `<section>B</section>`,
		"globals.njk": `<style>.shared{color:red}</style>`,
	}
	env := Configure(ConfigOptions{
		Loader:          &testLoader{files: files},
		GlobalTemplates: []string{"globals.njk"},
	})

	outA, err := env.Render("a.njk", nil)
	if err != nil {
		t.Fatalf("unexpected error rendering a.njk: %v", err)
	}
	outB, err := env.Render("b.njk", nil)
	if err != nil {
		t.Fatalf("unexpected error rendering b.njk: %v", err)
	}

	if !strings.Contains(outA, `<style>.shared{color:red}</style>`) || !strings.Contains(outA, `<div>A</div>`) {
		t.Fatalf("expected global styles in a.njk output, got %q", outA)
	}
	if !strings.Contains(outB, `<style>.shared{color:red}</style>`) || !strings.Contains(outB, `<section>B</section>`) {
		t.Fatalf("expected global styles in b.njk output, got %q", outB)
	}
}

func TestGlobalTemplatesProvideMacrosWithoutLayout(t *testing.T) {
	files := map[string]string{
		"main.njk":    `{{ badge("sam") }}`,
		"globals.njk": `{% macro badge(name) %}<b>{{ name | upper }}</b>{% endmacro %}`,
	}
	env := Configure(ConfigOptions{
		Loader:          &testLoader{files: files},
		GlobalTemplates: []string{"globals.njk"},
	})

	out, err := env.Render("main.njk", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if compactWhitespace(out) != `<b>SAM</b>` {
		t.Fatalf("expected global macro output, got %q", out)
	}
}

func TestGlobalHeadAndFootTemplateInjection(t *testing.T) {
	files := map[string]string{
		"page.njk":     `<!doctype html><html><head><title>{{ title }}</title></head><body><main>Hello</main></body></html>`,
		"links.njk":    `<link rel="stylesheet" href="/assets/app.css?v={{ version }}">`,
		"scripts.njk":  `<script type="module" src="/assets/app.js?v={{ version }}"></script>`,
		"scripts2.njk": `<script>window.__build="{{ version }}"</script>`,
	}
	env := Configure(ConfigOptions{
		Loader:              &testLoader{files: files},
		GlobalHeadTemplates: []string{"links.njk"},
		GlobalFootTemplates: []string{"scripts.njk", "scripts2.njk"},
	})

	out, err := env.Render("page.njk", map[string]any{"title": "Demo", "version": "42"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	headNeedle := `<link rel="stylesheet" href="/assets/app.css?v=42">`
	footNeedleA := `<script type="module" src="/assets/app.js?v=42"></script>`
	footNeedleB := `<script>window.__build="42"</script>`

	headIdx := strings.Index(out, headNeedle)
	headCloseIdx := strings.Index(strings.ToLower(out), "</head>")
	if headIdx < 0 || headCloseIdx < 0 || headIdx > headCloseIdx {
		t.Fatalf("expected global head links before </head>, got %q", out)
	}

	bodyCloseIdx := strings.Index(strings.ToLower(out), "</body>")
	footIdxA := strings.Index(out, footNeedleA)
	footIdxB := strings.Index(out, footNeedleB)
	if footIdxA < 0 || footIdxB < 0 || bodyCloseIdx < 0 || footIdxA > bodyCloseIdx || footIdxB > bodyCloseIdx {
		t.Fatalf("expected global scripts before </body>, got %q", out)
	}
}

func TestCustomDelimitersRenderString(t *testing.T) {
	env := Configure(ConfigOptions{
		Loader:        &testLoader{files: map[string]string{}},
		VariableStart: "[[",
		VariableEnd:   "]]",
		BlockStart:    "[%",
		BlockEnd:      "%]",
	})

	src := `[% set title = "Hello" %][[ title | upper ]]`
	out, err := env.RenderString(src, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if compactWhitespace(out) != "HELLO" {
		t.Fatalf("expected custom delimiter render, got %q", out)
	}
}

func TestCustomDelimitersRenderTemplates(t *testing.T) {
	files := map[string]string{
		"base.njk":  `<h1>[% block title %]Base[% endblock %]</h1><main>[% block body %]x[% endblock %]</main>`,
		"child.njk": `[% extends "base.njk" %][% block title %][[ pageTitle ]][% endblock %][% block body %]Hello [[ user.name ]][% endblock %]`,
	}
	env := Configure(ConfigOptions{
		Loader:        &testLoader{files: files},
		VariableStart: "[[",
		VariableEnd:   "]]",
		BlockStart:    "[%",
		BlockEnd:      "%]",
	})

	out, err := env.Render("child.njk", map[string]any{
		"pageTitle": "Docs",
		"user":      map[string]any{"name": "sam"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := compactWhitespace(out)
	want := `<h1>Docs</h1><main>Hello sam</main>`
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestPrecompileDirWithHTMLFormat(t *testing.T) {
	viewsDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(viewsDir, "index.njk"), []byte(`<h1>{{ title }}</h1>`), 0o644); err != nil {
		t.Fatalf("write index.njk: %v", err)
	}
	if err := os.WriteFile(filepath.Join(viewsDir, "feed.json"), []byte(`{"title":"{{ title }}"}`), 0o644); err != nil {
		t.Fatalf("write feed.json: %v", err)
	}
	env := Configure(ConfigOptions{Path: viewsDir})
	outDir := t.TempDir()
	if err := env.PrecompileDirWithOptions(outDir, map[string]any{"title": "Hello"}, PrecompileOptions{OutputFormat: "html"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	htmlPath := filepath.Join(outDir, "index.html")
	if _, err := os.Stat(htmlPath); err != nil {
		t.Fatalf("expected html output at %s: %v", htmlPath, err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "index.njk")); !os.IsNotExist(err) {
		t.Fatalf("expected .njk output to be omitted, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "feed.json")); err != nil {
		t.Fatalf("expected non-njk output to keep extension: %v", err)
	}
}
