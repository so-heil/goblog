package frontend

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/so-heil/goblog/business/assets"
	"github.com/so-heil/goblog/business/pages"
)

type Frontend struct {
	store      pages.Store
	assetFiles *assets.Assets
}

func New(storer pages.Store, assetFiles *assets.Assets) *Frontend {
	return &Frontend{store: storer, assetFiles: assetFiles}
}

// Routes registers paths to mux
func (frontend *Frontend) Routes(mux *http.ServeMux) {
	mux.Handle("/static/", frontend.assetFiles)
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

	w.Header().Set("Content-Type", "text/html")
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

// SSG is a Static Site Generator that runs a static build of the website putting the resulting static files in dir
func (frontend *Frontend) SSG(dir string, perm os.FileMode) error {
	// copy asset files to static directory
	if err := frontend.assetFiles.RecursiveCopy(dir, perm); err != nil {
		return fmt.Errorf("recursive copy assets: %w", err)
	}

	type staticPage struct {
		id           string
		relativePath string
	}

	staticPages := []staticPage{
		{pages.AboutPageID, "index.html"},
		{pages.AboutPageID, "about.html"},
		{pages.BlogPageID, "blog/index.html"},
	}

	// add article pages to static pages
	for slug, _ := range frontend.store.Versions() {
		if slug != pages.AboutPageID && slug != pages.BlogPageID {
			staticPages = append(staticPages, staticPage{slug, fmt.Sprintf("blog/%s.html", slug)})
		}
	}

	for _, p := range staticPages {
		if err := frontend.putStaticPage(p.id, filepath.Join(dir, p.relativePath), perm); err != nil {
			return fmt.Errorf("put static page %s: %w", p.id, err)
		}
	}

	return nil
}

func (frontend *Frontend) putStaticPage(id string, path string, perm os.FileMode) error {
	// create parent directory if not exists
	if err := os.MkdirAll(filepath.Dir(path), perm); err != nil {
		return fmt.Errorf("make path dir: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create target static page file: %w", err)
	}
	defer f.Close()

	page, err := frontend.store.Load(id)
	if err != nil {
		return fmt.Errorf("load page for static generation: %w", err)
	}

	if _, err := f.Write(page); err != nil {
		return fmt.Errorf("write static page in target file: %w", err)
	}

	return nil
}
