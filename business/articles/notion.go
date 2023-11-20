package articles

import (
	"fmt"
	"github.com/so-heil/goblog/foundation/notion"
	"net/http"
	"time"
)

func (na notionArticle) ToArticle() Article {
	var article Article

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

func (nas *notionArticlesResponse) ToArticles() []Article {
	articles := make([]Article, len(nas.Results))
	for i, na := range nas.Results {
		articles[i] = na.ToArticle()
	}
	return articles
}

type notionArticlesResponse struct {
	Results []notionArticle `json:"results"`
}

type NotionArticleStore struct {
	notionClient *notion.Client
	databaseID   string
}

func NewNotionArticleStore(nc *notion.Client, databaseID string) *NotionArticleStore {
	return &NotionArticleStore{notionClient: nc, databaseID: databaseID}
}

func (nas *NotionArticleStore) Articles() ([]Article, error) {
	var nArticles notionArticlesResponse
	if err := nas.notionClient.Request(http.MethodPost, fmt.Sprintf("/databases/%s/query", nas.databaseID), nil, &nArticles); err != nil {
		return nil, fmt.Errorf("retiriving articles: %w", err)
	}
	return nArticles.ToArticles(), nil
}
