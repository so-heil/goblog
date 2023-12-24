package articles

import (
	"errors"
	"time"
)

var ErrArticleNotFound = errors.New("article not found")

// Article is a blog article
type Article struct {
	ID             string    `json:"id"`
	LastEditedTime time.Time `json:"last_edited_time"`
	Title          string    `json:"title"`
	Excerpt        string    `json:"excerpt"`
	WrittenAt      time.Time `json:"written_at"`
	Slug           string    `json:"slug"`
}
