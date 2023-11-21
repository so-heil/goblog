package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/a-h/templ"
	"github.com/allegro/bigcache/v3"
	"github.com/so-heil/goblog/business/articles"
	"github.com/so-heil/goblog/business/frontend/assets"
	"github.com/so-heil/goblog/business/storage"
	"github.com/so-heil/goblog/business/templates/components/breadcrumb"
	"github.com/so-heil/goblog/business/templates/pages/home"
	"github.com/so-heil/goblog/foundation/notion"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

type Webapp struct {
	bc *storage.Storage
}

func (wa Webapp) BlogPage(w http.ResponseWriter, r *http.Request) {
	fmt.Println(wa.bc.Len())
	cached, err := wa.bc.Page(storage.Key)

	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	if _, err := w.Write(cached); err != nil {
		w.Write([]byte(err.Error()))
		return
	}
}

func (wa Webapp) ArticlePage(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	slug := parts[len(parts)-1]

	fmt.Println(slug)
	cached, err := wa.bc.Page(fmt.Sprintf("%s_%s", storage.Key, slug))
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	if _, err := w.Write(cached); err != nil {
		w.Write([]byte(err.Error()))
		return
	}
}

func run() error {
	nc := notion.NewClient(os.Getenv("NOTION_API_KEY"))
	nas := articles.NewNotionProvider(nc, "e0d22df6-5be5-4651-9fdc-02b368b99e0f")
	config := bigcache.DefaultConfig(10 * time.Minute)
	config.CleanWindow = 0
	cache, err := bigcache.New(context.Background(), config)
	logger := log.New(os.Stdout, "", log.LstdFlags)
	errLogger := log.New(os.Stdout, "error", log.LstdFlags)

	bc := storage.New(cache, nas, logger, errLogger)
	success := bc.Update()
	if !success {
		return errors.New("storing page failed")
	}

	wa := Webapp{bc: bc}

	if err != nil {
		return err
	}

	http.Handle("/static/", assets.StaticHandler)

	http.Handle("/", templ.Handler(home.Home([]breadcrumb.Link{{
		Title: "ABOUT",
		Href:  "/",
	}})))

	http.HandleFunc("/blog", wa.BlogPage)
	http.HandleFunc("/blog/", wa.ArticlePage)

	go func() {
		t := time.NewTicker(1 * time.Minute)
		for range t.C {
			bc.Update()
		}

	}()
	fmt.Println("Starting web server on :3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		return fmt.Errorf("http server: %w", err)
	}

	return nil
}
