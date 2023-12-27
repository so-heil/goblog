package pages

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/a-h/templ"
	"github.com/so-heil/goblog/business/articles"
	"github.com/so-heil/goblog/business/templates/components/breadcrumb"
	"github.com/so-heil/goblog/business/templates/pages/about"
	"github.com/so-heil/goblog/business/templates/pages/blog"
	"github.com/so-heil/goblog/business/templates/pages/notfound"
)

const BlogPageID = "blog_page"
const AboutPageID = "about_page"
const NotFoundPageID = "404_page"

type BlogPage struct {
	articles          []articles.Article
	anyArticleUpdated bool
}

func (bp *BlogPage) ID() string {
	return BlogPageID
}

func (bp *BlogPage) Version() time.Time {
	return time.Now()
}

func (bp *BlogPage) IsUpdated() bool {
	return bp.anyArticleUpdated
}

func (bp *BlogPage) Render() (templ.Component, error) {
	blogArticles := make([]blog.Article, len(bp.articles))
	for i, article := range bp.articles {
		blogArticles[i] = toBlogArticle(article)
	}

	page := blog.BlogPage([]breadcrumb.Link{{
		Title: "BLOG",
		Href:  "/blog",
	}}, blogArticles)

	return page, nil
}

type ArticlePage struct {
	article  articles.Article
	provider Provider
	versions map[string]time.Time
}

func (ap *ArticlePage) ID() string {
	return ap.article.Slug
}

func (ap *ArticlePage) Version() time.Time {
	return ap.article.LastEditedTime
}

func (ap *ArticlePage) IsUpdated() bool {
	aboutUpdate, ok := ap.versions[ap.article.Slug]
	return !ok || aboutUpdate != ap.article.LastEditedTime
}

func (ap *ArticlePage) Render() (templ.Component, error) {
	sections, err := ap.provider.Content(ap.article.ID)
	if err != nil {
		return nil, fmt.Errorf("retrieve article content from provider: %w", err)
	}

	components := make([]templ.Component, len(sections))
	headings := make([]string, len(sections))

	for i := 0; i < len(sections); i++ {
		components[i] = sections[i].Component
		headings[i] = sections[i].Title
	}

	page := blog.ArticlePage([]breadcrumb.Link{{
		Title: ap.article.Title,
		Href:  fmt.Sprintf("/blog/%s", ap.article.Slug),
	}}, toBlogArticle(ap.article), components, headings)

	return page, nil
}

type AboutPage struct {
	data     *About
	provider Provider
	versions map[string]time.Time
}

func (ap *AboutPage) ID() string {
	return AboutPageID
}

func (ap *AboutPage) Version() time.Time {
	return ap.data.LastEditedTime
}

func (ap *AboutPage) IsUpdated() bool {
	aboutUpdate, ok := ap.versions[AboutPageID]
	return !ok || aboutUpdate != ap.data.LastEditedTime
}

func (ap *AboutPage) Render() (templ.Component, error) {
	sections, err := ap.provider.Content(ap.data.ID)
	if err != nil {
		return nil, fmt.Errorf("retrieve about content from provider: %w", err)
	}

	componenets := make([]templ.Component, len(sections))
	for i := 0; i < len(sections); i++ {
		componenets[i] = sections[i].Component
	}

	page := about.AboutPage([]breadcrumb.Link{{
		Title: "About",
		Href:  "/",
	}}, about.About{
		Title:    ap.data.Title,
		SubTitle: ap.data.SubTitle,
		Content:  componenets,
	})

	return page, nil
}

type NotFoundPage struct {
	versions map[string]time.Time
}

func (n *NotFoundPage) IsUpdated() bool {
	_, ok := n.versions[NotFoundPageID]
	return !ok
}

func (n *NotFoundPage) Render() (templ.Component, error) {
	return notfound.NotFoundPage(), nil
}

func (n *NotFoundPage) ID() string {
	return NotFoundPageID
}

func (n *NotFoundPage) Version() time.Time {
	return time.Now()
}

// build renders the page and returns the rendered content that can be stored
func build(page Page) ([]byte, error) {
	component, err := page.Render()
	if err != nil {
		return nil, fmt.Errorf("render page: %w", err)
	}

	content, err := pageContent(component)
	if err != nil {
		return nil, fmt.Errorf("page content: %w", err)
	}

	return content, nil
}

func toBlogArticle(a articles.Article) blog.Article {
	return blog.Article{
		Title:     a.Title,
		Excerpt:   a.Excerpt,
		WrittenAt: a.WrittenAt,
		Slug:      a.Slug,
	}
}

// pageContent renders the component into a buffer and returns the result
func pageContent(page templ.Component) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := page.Render(context.Background(), buf); err != nil {
		return nil, fmt.Errorf("render page: %w", err)
	}

	return buf.Bytes(), nil
}
