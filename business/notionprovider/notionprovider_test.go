package notionprovider_test

import (
	"os"
	"testing"

	"github.com/so-heil/goblog/business/notionprovider"
	"github.com/so-heil/goblog/foundation/notion"
)

func TestProvider(t *testing.T) {
	apiKey := os.Getenv("NOTION_TEST_API_KEY")
	databaseID := os.Getenv("NOTION_TEST_DATABASE_ID")

	if apiKey == "" {
		t.Fatal("notion test api key should be provided as NOTION_TEST_API_KEY environment variable")
	}

	if databaseID == "" {
		t.Fatal("notion test database id should be provided as NOTION_TEST_DATABASE_ID environment variable")
	}

	client := notion.NewClient(apiKey)
	p := notionprovider.NewProvider(client, databaseID)

	articles, err := p.Articles()
	if err != nil {
		t.Fatalf("should fetch pages: %s", err)
	}
	if len(articles) == 0 {
		t.Fatal("should return at least one article")
	}

	sample := articles[0]
	blocks, err := p.Content(sample.ID)
	if err != nil {
		t.Fatalf("should get article content: article id: %s: %s", sample.ID, err)
	}
	if len(blocks) == 0 {
		t.Fatal("sample article should have at least one section block")
	}
}
