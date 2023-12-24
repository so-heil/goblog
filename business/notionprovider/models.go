package notionprovider

import (
	"strings"
	"time"

	"github.com/so-heil/goblog/business/articles"
)

type (
	textContent struct {
		Text string `json:"plain_text"`
	}
	textContents  []textContent
	notionArticle struct {
		LastEditedTime time.Time `json:"last_edited_time"`
		ID             string    `json:"id"`
		Properties     struct {
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
			Type struct {
				Select struct {
					Name string `json:"name"`
				} `json:"select"`
			} `json:"Type"`
		} `json:"properties"`
	}
	textBlock struct {
		RichText textContents `json:"rich_text"`
	}
	paragraph struct {
		RichText []struct {
			Annotations struct {
				Bold          bool `json:"bold"`
				Italic        bool `json:"italic"`
				Strikethrough bool `json:"strikethrough"`
				Underline     bool `json:"underline"`
				Code          bool `json:"code"`
			} `json:"annotations"`
			PlainText string  `json:"plain_text"`
			Href      *string `json:"href"`
		} `json:"rich_text"`
	}
	image struct {
		Caption textContents `json:"caption"`
		File    struct {
			Url        string    `json:"url"`
			ExpiryTime time.Time `json:"expiry_time"`
		} `json:"file"`
	}
	code struct {
		RichText []struct {
			Type string `json:"type"`
			Text struct {
				Content string `json:"content"`
			} `json:"text"`
		} `json:"rich_text"`
		Language string `json:"language"`
	}
	notionBlock struct {
		Type             string     `json:"type"`
		Heading1         *textBlock `json:"heading_1"`
		Heading2         *textBlock `json:"heading_2"`
		Heading3         *textBlock `json:"heading_3"`
		Quote            *textBlock `json:"quote"`
		Paragraph        *paragraph `json:"paragraph"`
		Image            *image     `json:"image"`
		Code             *code      `json:"code"`
		BulletedListItem *textBlock `json:"bulleted_list_item"`
	}
)

func (na *notionArticle) toArticle() articles.Article {
	article := articles.Article{
		ID:             na.ID,
		LastEditedTime: na.LastEditedTime,
	}

	if len(na.Properties.Title.Title) > 0 {
		article.Title = na.Properties.Title.Title[0].PlainText
	}

	if len(na.Properties.Excerpt.RichText) > 0 {
		article.Excerpt = na.Properties.Excerpt.RichText[0].PlainText
	}

	if na.Properties.WrittenAt.Date != nil {
		wa, err := time.Parse("2006-01-02", na.Properties.WrittenAt.Date.Start)
		if err == nil {
			article.WrittenAt = wa
		}
	}

	if len(na.Properties.Slug.RichText) > 0 {
		article.Slug = na.Properties.Slug.RichText[0].PlainText
	}

	return article
}

func (tb textContents) toString() string {
	var s strings.Builder
	for _, text := range tb {
		s.WriteString(text.Text)
	}
	return s.String()
}
