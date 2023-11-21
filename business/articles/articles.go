package articles

import (
	"github.com/a-h/templ"
	"time"
)

type Article struct {
	ID             string    `json:"id"`
	LastEditedTime time.Time `json:"last_edited_time"`
	Title          string    `json:"title"`
	Excerpt        string    `json:"excerpt"`
	WrittenAt      time.Time `json:"written_at"`
	Slug           string    `json:"slug"`
}

type SectionBlock struct {
	Title     string
	Component templ.Component
}
