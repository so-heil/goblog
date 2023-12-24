package frontend

import (
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/so-heil/goblog/business/assets"
	"github.com/so-heil/goblog/business/pages"
)

type Frontend struct {
	store pages.Store
}

func New(storer pages.Store) *Frontend {
	return &Frontend{store: storer}
}

// Routes registers paths to mux
func (frontend *Frontend) Routes(mux *http.ServeMux) {
	mux.Handle("/static/", assets.StaticHandler)
	mux.HandleFunc("/about", frontend.aboutPage)
	mux.HandleFunc("/blog", frontend.blogPage)
	mux.HandleFunc("/blog/", frontend.articlePage)
	mux.HandleFunc("/", frontend.root)
}

func (frontend *Frontend) blogPage(w http.ResponseWriter, r *http.Request) {
	if err := frontend.handlePage(pages.BlogPageID, w); err != nil {
		internalError(err, w, r)
	}
}

func (frontend *Frontend) aboutPage(w http.ResponseWriter, r *http.Request) {
	if err := frontend.handlePage(pages.AboutPageID, w); err != nil {
		internalError(err, w, r)
	}
}

func (frontend *Frontend) notFound(w http.ResponseWriter, r *http.Request) {
	if err := frontend.handlePage(pages.NotFoundPageID, w); err != nil {
		internalError(err, w, r)
	}
}

func (frontend *Frontend) articlePage(w http.ResponseWriter, r *http.Request) {
	slug := path.Base(r.URL.Path)
	if slug == "blog" {
		frontend.blogPage(w, r)
		return
	}

	if err := frontend.handlePage(slug, w); err != nil {
		if errors.Is(err, pages.ErrArticleNotFound) {
			frontend.notFound(w, r)
			return
		}
		internalError(err, w, r)
	}
}

func (frontend *Frontend) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		frontend.aboutPage(w, r)
		return
	}

	frontend.notFound(w, r)
}

func (frontend *Frontend) handlePage(id string, w http.ResponseWriter) error {
	page, err := frontend.store.Load(id)
	if err != nil {
		return fmt.Errorf("handlePage: load page %s: %s", id, err)
	}

	if _, err := w.Write(page); err != nil {
		return fmt.Errorf("handlePage: write reponse: %w", err)
	}

	return nil
}

func internalError(err error, w http.ResponseWriter, r *http.Request) {
	fmt.Printf("ERROR: internal server error: page: %s: %s\n", r.URL.String(), err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("something went wrong"))
}
