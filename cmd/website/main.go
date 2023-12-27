package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/dgraph-io/badger/v4"
	"github.com/so-heil/goblog/business/assets"
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

func run() error {
	a, err := newApp()
	if err != nil {
		return fmt.Errorf("new app: %w", err)
	}

	if len(os.Args) <= 1 {
		return fmt.Errorf("not enough arguments: usage: goblog (serve|static)")
	}

	switch os.Args[1] {
	case "serve":
		return a.startWebServer()
	case "static":
		return a.startSSG()
	default:
		return fmt.Errorf("wrong usage: usage: goblog (serve|static)")
	}
}

type config struct {
	NotionAPIKey            string        `env:"NOTION_API_KEY"`
	NotionArticleDatabaseID string        `env:"NOTION_ARTICLE_DATABASE_ID"`
	BadgerDBPath            string        `env:"BADGER_DB_PATH" envDefault:"/tmp/badger"`
	MaxSeedWorkers          int           `env:"MAX_SEED_WORKERS" envDefault:"10"`
	SeedInterval            time.Duration `env:"UPDATE_INTERVAL" envDefault:"60s"`
	ListenAddress           string        `env:"LISTEN_ADDRESS" envDefault:":3000"`
	DBInMemory              bool          `env:"DB_IN_MEMORY" envDefault:"false"`
	SSGPath                 string        `env:"SSG_PATH" envDefault:""`
}

type app struct {
	fe       *frontend.Frontend
	provider pages.Provider
	store    pages.Store
	cfg      *config
}

func newApp() (*app, error) {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("startup: parse config from env: %w", err)
	}

	notionClient := notion.NewClient(cfg.NotionAPIKey)
	provider := notionprovider.NewProvider(notionClient, cfg.NotionArticleDatabaseID)

	var options badger.Options
	if cfg.DBInMemory {
		options = badger.DefaultOptions("")
		options.InMemory = true
	} else {
		options = badger.DefaultOptions(cfg.BadgerDBPath)
	}

	db, err := badger.Open(options)
	if err != nil {
		return nil, fmt.Errorf("startup: open badger db: %w", err)
	}
	store := repository.New(db)

	assetFiles := assets.New()
	fe := frontend.New(store, assetFiles)

	return &app{
		fe:       fe,
		cfg:      &cfg,
		provider: provider,
		store:    store,
	}, nil
}

func (a *app) updateStore() error {
	if err := pages.UpdateStore(a.provider, a.store, a.cfg.MaxSeedWorkers); err != nil {
		return fmt.Errorf("update store: %w", err)
	}
	return nil
}

func (a *app) startSSG() error {
	target := a.cfg.SSGPath

	if target == "" {
		fmt.Println("SSGPath is not configured, creating a temp dir as SSG target")
		tmp, err := os.MkdirTemp("", "ssg")
		if err != nil {
			return fmt.Errorf("mkdir temp: %w", err)
		}
		target = tmp
	}

	// make sure the target exists
	if err := os.MkdirAll(target, 0777); err != nil {
		return fmt.Errorf("mkdirall target: %w", err)
	}

	fmt.Println("starting seeding store with provider data")
	if err := a.updateStore(); err != nil {
		return fmt.Errorf("initial store seed: %w", err)
	}
	fmt.Println("seed successful")

	fmt.Printf("starting SSG to %s\n", target)
	if err := a.fe.SSG(target, 0777); err != nil {
		return fmt.Errorf("frontend SSG: %w", err)
	}
	fmt.Printf("successfully generated static site to %s\n", target)
	return nil
}

func (a *app) startWebServer() error {
	if err := a.updateStore(); err != nil {
		return fmt.Errorf("initial store seed: %w", err)
	}

	// start goroutine to keep store updated
	go func() {
		t := time.NewTicker(a.cfg.SeedInterval)
		for range t.C {
			if err := a.updateStore(); err != nil {
				fmt.Printf("update store: %s\n", err)
			}
		}
	}()

	mux := http.NewServeMux()
	a.fe.Routes(mux)

	fmt.Printf("Starting web server on %s\n", a.cfg.ListenAddress)
	if err := http.ListenAndServe(a.cfg.ListenAddress, mux); err != nil {
		return fmt.Errorf("shutdown: website server: %w", err)
	}

	return nil
}
