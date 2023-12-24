// Package repository provides an in-memory repository for articles
package repository

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/so-heil/goblog/business/pages"
)

const prefix = "repository_article_"

func key(id string) []byte {
	return []byte(fmt.Sprintf("%s%s", prefix, id))
}

type Repository struct {
	db       *badger.DB
	versions sync.Map
}

func New(db *badger.DB) *Repository {
	return &Repository{db: db}
}

func (repo *Repository) Store(id string, content []byte, version time.Time) error {
	if err := repo.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key(id), content)
		return err
	}); err != nil {
		return fmt.Errorf("store article content: %w", err)
	}
	repo.versions.Store(id, version)

	return nil
}

func (repo *Repository) Load(id string) ([]byte, error) {
	var content []byte
	err := repo.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key(id))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return pages.ErrArticleNotFound
			}

			return fmt.Errorf("get article[%s] from db: %w", id, err)
		}

		if content, err = item.ValueCopy(nil); err != nil {
			return fmt.Errorf("value copy item: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("retrieve from db: %w", err)
	}

	return content, nil
}

func (repo *Repository) Delete(id string) error {
	err := repo.db.Update(func(txn *badger.Txn) error {
		if err := txn.Delete(key(id)); err != nil {
			return fmt.Errorf("delete article[%s]: %w", id, err)
		}
		return nil
	})
	repo.versions.Delete(id)

	return err
}

func (repo *Repository) Versions() map[string]time.Time {
	versions := make(map[string]time.Time)
	repo.versions.Range(func(key, value any) bool {
		slug := key.(string)
		version := value.(time.Time)
		versions[slug] = version
		return true
	})
	return versions
}
