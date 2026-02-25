package main

import (
	"fmt"
	"log"

	nunchucks "github.com/SamuelDBines/nunjucks/go"
)

func main() {
	env := nunchucks.Configure(nunchucks.ConfigOptions{Path: "examples/client/views"})

	out, err := env.Render("index.njk", map[string]any{
		"pageTitle": "Client Fetch Demo",
		"userID":    42,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(out)
}

