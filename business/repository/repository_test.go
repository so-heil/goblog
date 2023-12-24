package repository_test

import (
	"crypto/rand"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/so-heil/goblog/business/pages"
	"github.com/so-heil/goblog/business/repository"
)

func TestRepository(t *testing.T) {
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

	content := make([]byte, 10*1024)
	if _, err := rand.Read(content); err != nil {
		t.Fatalf("create random content: %s", err)
	}

	testID := "test_id"
	if err := s.Store(testID, content, time.Now()); err != nil {
		t.Fatalf("store content: %s", err)
	}

	versions := s.Versions()
	if len(versions) != 1 {
		t.Errorf("should have 1 record, has: %d", len(versions))
	}

	for id := range s.Versions() {
		if id != testID {
			t.Errorf("only id in versions should be: %s, is %s", testID, id)
		}
	}

	content2 := make([]byte, 10*1024)
	if _, err := rand.Read(content2); err != nil {
		t.Fatalf("create random content: %s", err)
	}

	testID2 := "test_id2"
	if err := s.Store(testID2, content2, time.Now()); err != nil {
		t.Fatalf("store content: %s", err)
	}

	if len(s.Versions()) != 2 {
		t.Errorf("should have 2 record, has: %d", len(versions))
	}

	retrieved, err := s.Load(testID)
	if err != nil {
		t.Fatalf("retrieve content: %s", err)
	}

	if !reflect.DeepEqual(retrieved, content) {
		t.Error("same content should be retrieved from db")
	}

	if _, err := s.Load("some_random_id"); err != nil {
		if !errors.Is(err, pages.ErrArticleNotFound) {
			t.Fatalf("should yeild not found error, got: %s", err)
		}
	}

	if err := s.Delete(testID); err != nil {
		t.Fatalf("delete test 1 content: %s", err)
	}

	if err := s.Delete(testID2); err != nil {
		t.Fatalf("delete test 2 content: %s", err)
	}

	if _, err := s.Load(testID); err != nil {
		if !errors.Is(err, pages.ErrArticleNotFound) {
			t.Fatalf("should yeild not found error, got: %s", err)
		}
	}

	if len(s.Versions()) != 0 {
		t.Errorf("should have 0 record, has: %d", len(versions))
	}
}
