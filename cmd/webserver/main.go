package main

import (
	"fmt"
	"github.com/a-h/templ"
	"log"
	"net/http"

	"github.com/so-heil/goblog/business/frontend/assets"
	"github.com/so-heil/goblog/business/templates/pages/hello"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	component := hello.Hello("World!")

	http.Handle("/", templ.Handler(component))
	http.Handle("/static/", assets.StaticHandler)

	fmt.Println("Starting web server on :3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		return fmt.Errorf("http server: %w", err)
	}

	return nil
}
