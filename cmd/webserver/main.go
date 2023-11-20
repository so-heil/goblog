package main

import (
	"encoding/json"
	"fmt"
	"github.com/a-h/templ"
	"github.com/so-heil/goblog/business/articles"
	"github.com/so-heil/goblog/business/frontend/assets"
	"github.com/so-heil/goblog/business/templates/components/breadcrumb"
	"github.com/so-heil/goblog/business/templates/pages/blog"
	"github.com/so-heil/goblog/business/templates/pages/home"
	"github.com/so-heil/goblog/foundation/notion"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func articlesG(w http.ResponseWriter, r *http.Request) {
	nc := notion.NewClient(os.Getenv("NOTION_API_KEY"))
	nas := articles.NewNotionArticleStore(nc, "e0d22df6-5be5-4651-9fdc-02b368b99e0f")
	articles, err := nas.Articles()
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	json.NewEncoder(w).Encode(articles)
}

func run() error {
	nc := notion.NewClient(os.Getenv("NOTION_API_KEY"))
	nas := articles.NewNotionArticleStore(nc, "e0d22df6-5be5-4651-9fdc-02b368b99e0f")

	artcls, err := nas.Articles()
	if err != nil {
		return err
	}
	cont, heads, err := nas.NotionContent("ad198904-1e30-499b-b78f-e2b6029a9df5")
	if err != nil {
		return err
	}

	http.Handle("/static/", assets.StaticHandler)

	http.Handle("/", templ.Handler(home.Home([]breadcrumb.Link{{
		Title: "ABOUT",
		Href:  "/",
	}})))
	http.Handle("/blog", templ.Handler(blog.Blog([]breadcrumb.Link{{
		Title: "BLOG",
		Href:  "/blog",
	}}, artcls)))
	http.Handle("/blog/", templ.Handler(blog.BlogArticle([]breadcrumb.Link{{
		Title: "BLOG",
		Href:  "/blog",
	}, {
		Title: strings.ToUpper(artcls[2].Title),
		Href:  fmt.Sprintf("/blog/%s", artcls[2].Slug),
	}}, artcls[2], cont, heads)))

	fmt.Println("Starting web server on :3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		return fmt.Errorf("http server: %w", err)
	}

	return nil
}
