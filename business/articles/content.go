package articles

import (
	"fmt"
	"github.com/a-h/templ"
	"github.com/so-heil/goblog/business/templates/components/elements"
	"github.com/so-heil/goblog/business/templates/components/toc"
	"net/http"
	"strings"
	"time"
)

type TextContent struct {
	Text string `json:"plain_text"`
}

type Contents []TextContent

func (tb Contents) String() string {
	var s strings.Builder
	for _, text := range tb {
		s.WriteString(text.Text)
	}
	return s.String()
}

type (
	TextBlock struct {
		RichText Contents `json:"rich_text"`
	}
	Paragraph struct {
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
	Image struct {
		Caption Contents `json:"caption"`
		File    struct {
			Url        string    `json:"url"`
			ExpiryTime time.Time `json:"expiry_time"`
		} `json:"file"`
	}
	Code struct {
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
		Heading1         *TextBlock `json:"heading_1"`
		Heading2         *TextBlock `json:"heading_2"`
		Heading3         *TextBlock `json:"heading_3"`
		Quote            *TextBlock `json:"quote"`
		Paragraph        *Paragraph `json:"paragraph"`
		Image            *Image     `json:"image"`
		Code             *Code      `json:"code"`
		BulletedListItem *TextBlock `json:"bulleted_list_item"`
	}
	BlocksResponse struct {
		Results []NotionBlock `json:"results"`
	}
)

// NotionContent looks at every notion content blocks and create the corresponding components, returns components and related sections title
func (nas *NotionArticleStore) NotionContent(pageID string) ([]templ.Component, []string, error) {
	var br BlocksResponse
	if err := nas.notionClient.Request(http.MethodGet, fmt.Sprintf("/blocks/%s/children?page_size=100", pageID), nil, &br); err != nil {
		return nil, nil, fmt.Errorf("retiriving block with id %s childern: %w", pageID, err)
	}

	var blocks []templ.Component
	var sections []templ.Component
	var sectionTitles []string
	var sectionTitle string
	for _, nblock := range br.Results {
		switch nblock.Type {
		case "heading_1":
			blocks = append(blocks, elements.Heading1(nblock.Heading1.RichText.String()))
		case "heading_2":
			blocks = append(blocks, elements.Heading2(nblock.Heading2.RichText.String()))
			if sectionTitle == "" {
				sectionTitle = nblock.Heading2.RichText.String()
			}
		case "heading_3":
			blocks = append(blocks, elements.Heading3(nblock.Heading3.RichText.String()))
		case "bulleted_list_item":
			blocks = append(blocks, elements.ListItem(nblock.BulletedListItem.RichText.String()))
		case "quote":
			blocks = append(blocks, elements.Quote(nblock.Quote.RichText.String()))
		case "image":
			blocks = append(blocks, elements.Image(nblock.Image.File.Url, nblock.Image.Caption.String()))
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
			sections = append(sections, elements.Section(toc.ElementID(sectionTitle), blocks))
			sectionTitles = append(sectionTitles, sectionTitle)
			blocks = nil
			sectionTitle = ""
		}
	}

	if len(blocks) > 0 {
		sections = append(sections, elements.Section(toc.ElementID(sectionTitle), blocks))
		sectionTitles = append(sectionTitles, sectionTitle)
	}

	return sections, sectionTitles, nil
}
