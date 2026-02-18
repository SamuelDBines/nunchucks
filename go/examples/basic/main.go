package main

import (
	"fmt"
	"log"

	nunchucks "github.com/SamuelDBines/nunjucks/go"
)

func main() {
	env := nunchucks.Configure(nunchucks.ConfigOptions{Path: "examples/basic/views"})

	ctx := map[string]any{
		"siteName": "Nunchucks",
		"footer":   "POWERED BY GO",
		"user": map[string]any{
			"name": "sam",
		},
		"items": []any{
			map[string]any{"label": "one", "show": true},
			map[string]any{"label": "two", "show": false},
			map[string]any{"label": "three", "show": true},
		},
	}

	compiled, err := env.Compile("index.njk")
	if err != nil {
		log.Fatal(err)
	}

	html, err := env.Render("index.njk", ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== COMPILED TEMPLATE ===")
	fmt.Println(compiled)
	fmt.Println("=== RENDERED OUTPUT ===")
	fmt.Println(html)
}
