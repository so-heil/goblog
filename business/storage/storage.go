package storage

import (
	"bytes"
	"context"
	"fmt"
	"github.com/a-h/templ"
	"github.com/allegro/bigcache/v3"
	"github.com/so-heil/goblog/business/articles"
	"github.com/so-heil/goblog/business/templates/components/breadcrumb"
	"github.com/so-heil/goblog/business/templates/pages/blog"
	"log"
	"time"
)

const Key = "blog"

type Storage struct {
	cache   *bigcache.BigCache
	np      *articles.NotionProvider
	records map[string]time.Time
	log     *log.Logger
	errLog  *log.Logger
}

func New(cache *bigcache.BigCache, np *articles.NotionProvider, log *log.Logger, errLog *log.Logger) *Storage {
	return &Storage{
		cache:   cache,
		np:      np,
		records: make(map[string]time.Time),
		log:     log,
		errLog:  errLog,
	}
}

type result struct {
	slug       string
	lastEdited time.Time
	result     bool
}

func filterSlugless(articles []articles.Article) []articles.Article {
	var p int
	for i := 0; i < len(articles); i++ {
		if articles[i].Slug != "" {
			articles[p] = articles[i]
			p++
		}
	}
	return articles[:p]
}

func (s *Storage) savePage(id string, content templ.Component) {
	buf := new(bytes.Buffer)

	// any error in this step would cause loss of integrity
	if err := content.Render(context.Background(), buf); err != nil {
		panic(err)
	}
	if err := s.cache.Set(id, buf.Bytes()); err != nil {
		panic(err)
	}
}

func (s *Storage) deletePage(id string) error {
	if err := s.cache.Delete(id); err != nil {
		return fmt.Errorf("delete page: %w", err)
	}

	log.Printf("deleted page[%s], total stored pages[%d]\n", id, s.cache.Len())
	return nil
}

func (s *Storage) Page(id string) ([]byte, error) {
	data, err := s.cache.Get(id)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Storage) updateArticlePage(article articles.Article, rc chan<- result) {
	res := result{
		slug:       article.Slug,
		lastEdited: article.LastEditedTime,
		result:     false,
	}

	sections, err := s.np.Content(article.ID)
	if err != nil {
		s.errLog.Printf("get article[%s] content: %s\n", article.Slug, err)
		rc <- res
		return
	}

	components := make([]templ.Component, len(sections))
	headings := make([]string, len(sections))

	for i := 0; i < len(sections); i++ {
		components[i] = sections[i].Component
		headings[i] = sections[i].Title
	}

	page := blog.Article([]breadcrumb.Link{{
		Title: "BLOG",
		Href:  "/blog",
	}, {
		Title: "ARTICLE",
		Href:  fmt.Sprintf("/blog/%s", article.Slug),
	}}, article, components, headings)

	s.savePage(fmt.Sprintf("%s_%s", Key, article.Slug), page)

	log.Printf("stored article[%s] page, total stored pages[%d]\n", article.Slug, s.cache.Len())
	res.result = true
	rc <- res
}

func (s *Storage) updateBlogPage(articles []articles.Article) {
	page := blog.Blog([]breadcrumb.Link{{
		Title: "BLOG",
		Href:  "/blog",
	}}, articles)

	log.Printf("stored blog page, total stored pages[%d]\n", s.cache.Len())
	s.savePage(Key, page)
}

func (s *Storage) Update() bool {
	artcls, err := s.np.Articles()
	if err != nil {
		s.errLog.Printf("get articles: %s\n", err)
		return false
	}

	artcls = filterSlugless(artcls)

	// used to wait for result of all initiated goroutine and return the final result
	rc := make(chan result, 10)
	// used as the counting semaphore to control the number of goroutines
	sem := make(chan struct{}, 10)

	unmatched := make(map[string]struct{})
	for slug, _ := range s.records {
		unmatched[slug] = struct{}{}
	}

	var expected int
	for _, article := range artcls {
		delete(unmatched, article.Slug)
		if article.Slug == "" {
			continue
		}

		lastUpdate, ok := s.records[article.Slug]
		// if we don't have the record or updated, rebuild the page
		if !ok || lastUpdate != article.LastEditedTime {
			// acquire token
			sem <- struct{}{}
			expected++
			go func(article articles.Article) {
				// release token
				defer func() {
					<-sem
				}()
				s.updateArticlePage(article, rc)
			}(article)
		}
	}

	for slug, _ := range unmatched {
		if err := s.deletePage(slug); err != nil {
			panic(err)
		}
	}

	var numStored int
	success := true

	if expected > 0 || len(unmatched) > 0 {
		s.updateBlogPage(artcls)
		numStored++
	}

	for i := 0; i < expected; i++ {
		res := <-rc
		if res.result {
			s.records[res.slug] = res.lastEdited
			numStored++
		} else {
			success = false
		}
	}

	s.log.Printf("update: updated %d, deleted %d, now have %d, complete: %t\n", numStored, len(unmatched), s.Len(), success)
	return success
}

func (s *Storage) Len() int {
	return s.cache.Len()
}
