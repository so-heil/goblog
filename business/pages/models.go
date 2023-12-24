package pages

import (
	"errors"
	"time"

	"github.com/a-h/templ"
	"github.com/so-heil/goblog/business/articles"
)

var ErrArticleNotFound = errors.New("article not found")

// SectionBlock is a section containing content in form of templ.Component
type SectionBlock struct {
	Title     string
	Component templ.Component
}

// About contains a about page data that provider should return
// The content of About page should be accessible by Provider.Content via its ID
type About struct {
	ID             string
	Title          string
	SubTitle       string
	LastEditedTime time.Time
}

// Provider is any type that can provide the website content
type Provider interface {
	// Articles should return all articles accessed by the provider
	Articles() ([]articles.Article, error)
	// Content shpuld return the corresponding page content as SectionBlocks
	Content(id string) ([]SectionBlock, error)
	// AboutPage returns the about page data as AboutData
	AboutPage() (About, error)
}

// Store is any type that can store and retrieve website pages
type Store interface {
	// Store can store an Article content and track it's version for later use
	Store(id string, content []byte, version time.Time) error
	// Load should load the requested article content
	Load(id string) ([]byte, error)
	// Delete shpuld delete an article and its version from Store
	Delete(id string) error
	// Versions should return a map of all present articles in Store with their corresponding version
	Versions() map[string]time.Time
}

// Page represents a website page
type Page interface {
	// IsUpdated reports that if page has been updated since the last update
	IsUpdated() bool
	// Render return the page HTML content as a templ.Component
	Render() (templ.Component, error)
	// ID returns the page's unique identifier used in Store
	ID() string
	// Version returns the current version (update) of the page
	Version() time.Time
}
