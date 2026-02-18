package main

import (
	"fmt"
	"log"

	nunchucks "github.com/SamuelDBines/nunjucks/go"
)

func main() {
	env := nunchucks.Configure(nunchucks.ConfigOptions{Path: "examples/advanced/views"})
	out, err := env.Render("index.njk", map[string]any{
		"user": map[string]any{"name": "sam"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(out)
}
