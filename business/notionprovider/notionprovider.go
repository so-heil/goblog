// Package notionprovider is an article provider that works with notion API to fetch articles and their content
package notionprovider

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/so-heil/goblog/business/articles"
	"github.com/so-heil/goblog/business/pages"
	"github.com/so-heil/goblog/business/templates/components/elements"
	"github.com/so-heil/goblog/business/templates/components/toc"
	"github.com/so-heil/goblog/foundation/notion"
)

// Provider implements pages.Provider giving access to articles through notionprovider API stored in a database
type Provider struct {
	notionClient *notion.Client
	databaseID   string
}

func NewProvider(nc *notion.Client, databaseID string) *Provider {
	return &Provider{notionClient: nc, databaseID: databaseID}
}

// Articles looks for articles in Provider database
func (np *Provider) Articles() ([]articles.Article, error) {
	var nr struct {
		LastEditedTime time.Time       `json:"last_edited_time"`
		Results        []notionArticle `json:"results"`
	}

	filter := `{
		"filter": {
			"property": "Type",
			"select": {
				"equals": "Article"
			}
		}
	}`
	if err := np.notionClient.Request(http.MethodPost, fmt.Sprintf("/databases/%s/query", np.databaseID), strings.NewReader(filter), &nr); err != nil {
		return nil, fmt.Errorf("retiriving articles: %w", err)
	}

	articles := make([]articles.Article, len(nr.Results))
	for i, na := range nr.Results {
		articles[i] = na.toArticle()
	}
	return articles, nil
}

func (np *Provider) AboutPage() (pages.About, error) {
	var nr struct {
		LastEditedTime time.Time       `json:"last_edited_time"`
		Results        []notionArticle `json:"results"`
	}

	filter := fmt.Sprintf(`{
		"filter": {
			"property": "Slug",
			"rich_text": {
				"equals": "%s"
			}
		}
	}`, pages.AboutPageID)
	if err := np.notionClient.Request(http.MethodPost, fmt.Sprintf("/databases/%s/query", np.databaseID), strings.NewReader(filter), &nr); err != nil {
		return pages.About{}, fmt.Errorf("retiriving pages: %w", err)
	}

	if len(nr.Results) != 1 {
		return pages.About{}, fmt.Errorf("about page request should have 1 result in response")
	}

	aboutArticle := nr.Results[0].toArticle()
	data := pages.About{
		ID:             aboutArticle.ID,
		Title:          aboutArticle.Title,
		SubTitle:       aboutArticle.Excerpt,
		LastEditedTime: aboutArticle.LastEditedTime,
	}

	return data, nil
}

// Content looks at every notionprovider content blocks and create the corresponding pages.SectionBlock for each
func (np *Provider) Content(articleID string) ([]pages.SectionBlock, error) {
	var br struct {
		Results []notionBlock `json:"results"`
	}
	if err := np.notionClient.Request(http.MethodGet, fmt.Sprintf("/blocks/%s/children?page_size=100", articleID), nil, &br); err != nil {
		return nil, fmt.Errorf("retiriving block with id %s childern: %w", articleID, err)
	}

	var blocks []templ.Component
	var sblock []pages.SectionBlock
	var sectionTitle string
	for _, nblock := range br.Results {
		switch nblock.Type {
		case "heading_1":
			blocks = append(blocks, elements.Heading1(nblock.Heading1.RichText.toString()))
		case "heading_2":
			blocks = append(blocks, elements.Heading2(nblock.Heading2.RichText.toString()))
			if sectionTitle == "" {
				sectionTitle = nblock.Heading2.RichText.toString()
			}
		case "heading_3":
			blocks = append(blocks, elements.Heading3(nblock.Heading3.RichText.toString()))
		case "bulleted_list_item":
			blocks = append(blocks, elements.ListItem(nblock.BulletedListItem.RichText.toString()))
		case "quote":
			blocks = append(blocks, elements.Quote(nblock.Quote.RichText.toString()))
		case "image":
			blocks = append(blocks, elements.Image(nblock.Image.File.Url, nblock.Image.Caption.toString()))
		case "paragraph":
			p := make([]elements.P, len(nblock.Paragraph.RichText))
			for i, text := range nblock.Paragraph.RichText {
				if text.Annotations.Code {
					fmt.Println(text)
				}
				p[i] = elements.P{
					Content: text.PlainText,
					Bold:    text.Annotations.Bold,
					Italic:  text.Annotations.Italic,
					Code:    text.Annotations.Code,
					Link:    text.Href,
				}
			}
			blocks = append(blocks, elements.Paragraph(p))
		case "code":
			s := new(strings.Builder)
			for _, text := range nblock.Code.RichText {
				s.WriteString(text.Text.Content)
				s.WriteString("\n")
			}
			blocks = append(blocks, elements.Code(s.String(), nblock.Code.Language))
		case "divider":
			cblock := pages.SectionBlock{
				Title:     sectionTitle,
				Component: elements.Section(toc.ElementID(sectionTitle), blocks),
			}
			sblock = append(sblock, cblock)
			blocks = nil
			sectionTitle = ""
		}
	}

	if len(blocks) > 0 {
		cblock := pages.SectionBlock{
			Title:     sectionTitle,
			Component: elements.Section(toc.ElementID(sectionTitle), blocks),
		}
		sblock = append(sblock, cblock)
	}

	return sblock, nil
}
