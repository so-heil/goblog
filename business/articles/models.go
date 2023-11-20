package articles

import "time"

type notionArticle struct {
	Properties struct {
		Slug struct {
			RichText []struct {
				PlainText string `json:"plain_text"`
			} `json:"rich_text"`
		} `json:"Slug"`
		WrittenAt struct {
			Date *struct {
				Start string `json:"start"`
			} `json:"date"`
		} `json:"WrittenAt"`
		Excerpt struct {
			RichText []struct {
				PlainText string `json:"plain_text"`
			} `json:"rich_text"`
		} `json:"Excerpt"`
		Title struct {
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"Title"`
	} `json:"properties"`
}

type Article struct {
	Title     string    `json:"title"`
	Excerpt   string    `json:"excerpt"`
	WrittenAt time.Time `json:"written_at"`
	Slug      string    `json:"slug"`
}
