package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/dgraph-io/badger/v4"
	"github.com/so-heil/goblog/business/notionprovider"
	"github.com/so-heil/goblog/business/pages"
	"github.com/so-heil/goblog/business/repository"
	"github.com/so-heil/goblog/cmd/website/frontend"
	"github.com/so-heil/goblog/foundation/notion"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

type config struct {
	NotionAPIKey            string        `env:"NOTION_API_KEY"`
	NotionArticleDatabaseID string        `env:"NOTION_ARTICLE_DATABASE_ID"`
	BadgerDBPath            string        `env:"BADGER_DB_PATH" envDefault:"/tmp/badger"`
	MaxSeedWorkers          int           `env:"MAX_SEED_WORKERS" envDefault:"10"`
	SeedInterval            time.Duration `env:"UPDATE_INTERVAL" envDefault:"60s"`
	ListenAddress           string        `env:"LISTEN_ADDRESS" envDefault:":3000"`
}

func run() error {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		return fmt.Errorf("startup: parse config from env: %w", err)
	}

	notionClient := notion.NewClient(cfg.NotionAPIKey)
	provider := notionprovider.NewProvider(notionClient, cfg.NotionArticleDatabaseID)

	db, err := badger.Open(badger.DefaultOptions(cfg.BadgerDBPath))
	if err != nil {
		return fmt.Errorf("startup: open badger db: %w", err)
	}
	storer := repository.New(db)

	// initial seed of storer, a successful seed is required to start the app
	if err := pages.UpdateStore(provider, storer, cfg.MaxSeedWorkers); err != nil {
		return fmt.Errorf("startup: initial seed: %w", err)
	}

	// start the update of storer on interval
	go func() {
		t := time.NewTicker(cfg.SeedInterval)
		for range t.C {
			if pages.UpdateStore(provider, storer, cfg.MaxSeedWorkers) != nil {
				fmt.Printf("seed: %s\n", err)
			}
		}
	}()

	mux := http.NewServeMux()
	fe := frontend.New(storer)
	fe.Routes(mux)

	fmt.Printf("Starting web server on %s\n", cfg.ListenAddress)
	if err := http.ListenAndServe(cfg.ListenAddress, mux); err != nil {
		return fmt.Errorf("shutdown: website server: %w", err)
	}

	return nil
}
