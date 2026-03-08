package nunchucks

import (
	"strings"
	"testing"
)

func TestRenderValidatesPropsContract(t *testing.T) {
	files := map[string]string{
		"page.njk": `{# @props
title: string
count?: int
#}
<h1>{{ title }}</h1>`,
	}
	env := Configure(ConfigOptions{Loader: &testLoader{files: files}})

	out, err := env.Render("page.njk", map[string]any{"title": "Docs"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "<h1>Docs</h1>") {
		t.Fatalf("unexpected output: %q", out)
	}

	_, err = env.Render("page.njk", map[string]any{})
	if err == nil || !strings.Contains(err.Error(), `missing required prop "title"`) {
		t.Fatalf("expected missing title validation error, got %v", err)
	}
}

func TestRenderValidatesNamedParamsContract(t *testing.T) {
	files := map[string]string{
		"page.njk": `{# @params user
name: string
age: int = default(0)
#}
{# @props
title: string
user?: user
items: list
#}
<h1>{{ title }}</h1>`,
	}
	env := Configure(ConfigOptions{Loader: &testLoader{files: files}})

	_, err := env.Render("page.njk", map[string]any{
		"title": "Docs",
		"items": []any{"a", "b"},
		"user": map[string]any{
			"name": "Sam",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = env.Render("page.njk", map[string]any{
		"title": "Docs",
		"items": []any{"a", "b"},
		"user": map[string]any{
			"name": 99,
		},
	})
	if err == nil || !strings.Contains(err.Error(), `prop "user.name" should be string`) {
		t.Fatalf("expected user.name validation error, got %v", err)
	}
}

func TestRenderAppliesContractDefaults(t *testing.T) {
	files := map[string]string{
		"page.njk": `{# @params user
name: string
age: int = default(0)
#}
{# @props
title: string = 'Base'
user?: user
#}
<title>{{ title }}</title><p>{{ user.age }}</p>`,
	}
	env := Configure(ConfigOptions{Loader: &testLoader{files: files}})

	out, err := env.Render("page.njk", map[string]any{
		"user": map[string]any{
			"name": "Sam",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "<title>Base</title>") {
		t.Fatalf("expected title default to render, got %q", out)
	}
	if !strings.Contains(out, "<p>0</p>") {
		t.Fatalf("expected nested default to render, got %q", out)
	}
}

func TestIncludeValidatesPartialPropsAgainstCurrentScope(t *testing.T) {
	files := map[string]string{
		"page.njk": `{% set title = "Home" %}{% include "partial.njk" %}`,
		"partial.njk": `{# @props
title: string
#}
<header>{{ title }}</header>`,
	}
	env := Configure(ConfigOptions{Loader: &testLoader{files: files}})

	out, err := env.Render("page.njk", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "<header>Home</header>") {
		t.Fatalf("unexpected output: %q", out)
	}

	files["page.njk"] = `{% include "partial.njk" without context %}`
	env = Configure(ConfigOptions{Loader: &testLoader{files: files}})
	_, err = env.Render("page.njk", nil)
	if err == nil || !strings.Contains(err.Error(), `missing required prop "title"`) {
		t.Fatalf("expected include contract validation error, got %v", err)
	}
}

func TestIncludeAppliesContractDefaults(t *testing.T) {
	files := map[string]string{
		"page.njk": `{% include "partial.njk" %}`,
		"partial.njk": `{# @props
title: string = 'Header'
#}
<header>{{ title }}</header>`,
	}
	env := Configure(ConfigOptions{Loader: &testLoader{files: files}})

	out, err := env.Render("page.njk", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "<header>Header</header>") {
		t.Fatalf("expected include default to render, got %q", out)
	}
}
