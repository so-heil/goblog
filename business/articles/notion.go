package articles

import (
	"fmt"
	"github.com/a-h/templ"
	"github.com/so-heil/goblog/business/templates/components/elements"
	"github.com/so-heil/goblog/business/templates/components/toc"
	"github.com/so-heil/goblog/foundation/notion"
	"net/http"
	"strings"
	"time"
)

type notionArticle struct {
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
	} `json:"properties"`
}

func (na notionArticle) toArticle() Article {
	article := Article{
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

type textContent struct {
	Text string `json:"plain_text"`
}

type textContents []textContent

func (tb textContents) toString() string {
	var s strings.Builder
	for _, text := range tb {
		s.WriteString(text.Text)
	}
	return s.String()
}

type (
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
	NotionBlock struct {
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

type NotionProvider struct {
	notionClient *notion.Client
	databaseID   string
}

func NewNotionProvider(nc *notion.Client, databaseID string) *NotionProvider {
	return &NotionProvider{notionClient: nc, databaseID: databaseID}
}

func (np *NotionProvider) Articles() ([]Article, error) {
	var nr struct {
		LastEditedTime time.Time       `json:"last_edited_time"`
		Results        []notionArticle `json:"results"`
	}
	if err := np.notionClient.Request(http.MethodPost, fmt.Sprintf("/databases/%s/query", np.databaseID), nil, &nr); err != nil {
		return nil, fmt.Errorf("retiriving articles: %w", err)
	}

	articles := make([]Article, len(nr.Results))
	for i, na := range nr.Results {
		articles[i] = na.toArticle()
	}
	return articles, nil
}

// Content looks at every notion content blocks and create the corresponding SectionBlock for each
func (np *NotionProvider) Content(pageID string) ([]SectionBlock, error) {
	var br struct {
		Results []NotionBlock `json:"results"`
	}
	if err := np.notionClient.Request(http.MethodGet, fmt.Sprintf("/blocks/%s/children?page_size=100", pageID), nil, &br); err != nil {
		return nil, fmt.Errorf("retiriving block with id %s childern: %w", pageID, err)
	}

	var blocks []templ.Component
	var sblock []SectionBlock
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
			cblock := SectionBlock{
				Title:     sectionTitle,
				Component: elements.Section(toc.ElementID(sectionTitle), blocks),
			}
			sblock = append(sblock, cblock)
			blocks = nil
			sectionTitle = ""
		}
	}

	if len(blocks) > 0 {
		cblock := SectionBlock{
			Title:     sectionTitle,
			Component: elements.Section(toc.ElementID(sectionTitle), blocks),
		}
		sblock = append(sblock, cblock)
	}

	return sblock, nil
}
