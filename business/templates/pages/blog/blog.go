package blog

import "time"

type Article struct {
	Title     string
	Excerpt   string
	WrittenAt time.Time
	Slug      string
}
