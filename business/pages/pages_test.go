package pages_test

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/so-heil/goblog/business/notionprovider"
	"github.com/so-heil/goblog/business/pages"
	"github.com/so-heil/goblog/business/repository"
	"github.com/so-heil/goblog/foundation/notion"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomSlug(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestUpdateStore(t *testing.T) {
	// setup provider
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

	// setup storer
	dir, err := os.MkdirTemp("", "badger-test")
	if err != nil {
		t.Fatalf("mkdir temp: %s", err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatalf("remove temp dir: %s", err)
		}
	}(dir)

	db, err := badger.Open(badger.DefaultOptions(dir))
	if err != nil {
		t.Fatalf("open badger db with temp dir: %s", err)
	}
	defer func(db *badger.DB) {
		err := db.Close()
		if err != nil {
			t.Fatalf("close db: %s", err)
		}
	}(db)
	s := repository.New(db)

	atcls, err := p.Articles()
	if err != nil {
		t.Fatalf("should get articles from provider: %s", err)
	}

	if len(atcls) == 0 {
		t.Fatalf("should have at least 1 article in provider articles")
	}

	sample := atcls[0]

	if err := pages.UpdateStore(p, s, runtime.NumCPU()); err != nil {
		t.Fatalf("initial seed: %s", err)
	}

	initVersions := s.Versions()
	if len(initVersions) < 1 {
		t.Fatal("should have at least one article seeded")

	}

	sampleContent, err := s.Load(sample.Slug)
	if err != nil {
		t.Fatalf("should load sample article content: %s", err)
	}

	if len(sampleContent) == 0 {
		t.Fatal("sample article content should not be empty")
	}

	time.Sleep(time.Second)
	if err := pages.UpdateStore(p, s, runtime.NumCPU()); err != nil {
		t.Fatalf("initial seed: %s", err)
	}

	if !reflect.DeepEqual(s.Versions(), initVersions) {
		t.Fatal("versions should remained the same after second seed")
	}

	newSlug := randomSlug(10)
	updateBody := fmt.Sprintf(`{
	"properties": {
		"Slug": {
			"rich_text": [
				{
					 "type": "text",
					 "text": {
						  "content": "%s",
						  "link": null
					 },
					 "annotations": {
						  "bold": false,
						  "italic": false,
						  "strikethrough": false,
						  "underline": false,
						  "code": false,
						  "color": "default"
					 },
					 "plain_text": "%s",
					 "href": null
				}
			]
		}
	}
}`, newSlug, newSlug)

	if err := client.Request(http.MethodPatch, fmt.Sprintf("/pages/%s", sample.ID), strings.NewReader(updateBody), &struct{}{}); err != nil {
		t.Fatalf("request update failed: %s", err)
	}
	time.Sleep(time.Second)

	if err := pages.UpdateStore(p, s, runtime.NumCPU()); err != nil {
		t.Fatalf("initial seed: %s", err)
	}

	if reflect.DeepEqual(s.Versions(), initVersions) {
		t.Fatal("versions should not remain the same after update")
	}

	if _, err := s.Load(newSlug); err != nil {
		t.Fatal("newSlug article should be in storer")
	}

	if _, err := s.Load(sample.Slug); err != nil {
		if !errors.Is(err, pages.ErrArticleNotFound) {
			t.Fatal("old slug should not be in storer")
		}
	}

	if len(s.Versions()) != len(initVersions) {
		t.Fatalf("articles cound should remained the same after update")
	}
}
