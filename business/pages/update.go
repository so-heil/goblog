package pages

import (
	"errors"
	"fmt"
	"log"
)

// UpdateStore seeds the Store with all absent and outdated article pages and blog page from the Provider with the specified maxWorkers as concurrent workers
func UpdateStore(provider Provider, storer Store, maxWorkers int) error {
	atcls, err := provider.Articles()
	if err != nil {
		return fmt.Errorf("get provider articles: %w", err)
	}
	aboutData, err := provider.AboutPage()
	if err != nil {
		return fmt.Errorf("get provider about page: %w", err)
	}

	// counting semaphore to control the number of workers
	sem := make(chan struct{}, maxWorkers)
	// receive worker errors as its result from this channel
	workerErrs := make(chan error, maxWorkers)
	// used to track total number of workers initiated
	var workers int
	// contains all pages stored after this update, used to delete pages that no longer exist
	pagesAfterUpdate := make(map[string]struct{})

	// if a page is updated build the updated version and store it concurrently
	updateIfChanged := func(page Page) {
		pagesAfterUpdate[page.ID()] = struct{}{}
		if page.IsUpdated() {
			workers++
			go func() {
				sem <- struct{}{}
				var wErr error

				defer func() {
					<-sem
					workerErrs <- wErr
				}()

				id := page.ID()
				version := page.Version()
				content, err := build(page)
				if err != nil {
					wErr = fmt.Errorf("build page[%s:%s]: %w", id, version.String(), err)
					return
				}

				if err := storer.Store(page.ID(), content, page.Version()); err != nil {
					wErr = fmt.Errorf("store page[%s:%s]: %w", id, version, err)
					return
				}
			}()
		}
	}

	versionsBeforeUpdate := storer.Versions()
	for _, article := range atcls {
		if article.Slug == "" {
			continue
		}

		updateIfChanged(&ArticlePage{
			article:  article,
			provider: provider,
			versions: versionsBeforeUpdate,
		})
	}
	updateIfChanged(&AboutPage{
		data:     &aboutData,
		provider: provider,
		versions: versionsBeforeUpdate,
	})
	updateIfChanged(&NotFoundPage{versions: versionsBeforeUpdate})
	updateIfChanged(&BlogPage{
		articles:          atcls,
		anyArticleUpdated: workers > 0 || len(pagesAfterUpdate) != len(versionsBeforeUpdate)-1,
	})

	var buildErr error

	var deleted int
	for id := range versionsBeforeUpdate {
		if _, ok := pagesAfterUpdate[id]; !ok {
			if err := storer.Delete(id); err != nil {
				// an early return after initialization of workers would cause goroutine leak
				buildErr = err
				continue
			}
			deleted++
		}
	}

	for i := 0; i < workers; i++ {
		err := <-workerErrs
		if err != nil {
			buildErr = errors.New(buildErr.Error() + "\n" + err.Error())
			continue
		}
	}
	if buildErr != nil {
		return buildErr
	}

	log.Printf("updateStore: updated+added %d, deleted %d, total currently stored: %d\n", workers, deleted, len(storer.Versions()))
	return nil
}
