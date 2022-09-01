package app

import (
	"freedom-sentry/config"
	"freedom-sentry/mediawiki"
	"freedom-sentry/suppressor"
	"log"
)

type changeHandlerFunc func([]mediawiki.Revision) error

func createChangesHandler(subhandlers []changeHandlerFunc, changeProcessor <-chan []mediawiki.Revision) func() {
	handler := func(changes []mediawiki.Revision) {
		for _, subhandler := range subhandlers {
			_ = subhandler(changes)
		}
	}

	return func() {
		for {
			select {
			case changes := <-changeProcessor:
				handler(changes)
			}
		}
	}
}

func createHandlerChangeForSuppressor(pageRepo suppressor.SuppressedPageRepository, revSuppressor suppressor.RevisionSuppressor) changeHandlerFunc {
	return func(changes []mediawiki.Revision) error {
		list, err := pageRepo.GetAll()
		if err != nil {
			log.Println("failed to get suppression list:", err)
			return err
		}

		indexedList := make(map[string]bool, len(list))
		for _, title := range list {
			indexedList[title] = true
		}

		revs := make([]mediawiki.Revision, 0, len(changes))
		for _, rev := range changes {
			if _, inList := indexedList[rev.Title]; !inList {
				continue
			}

			revs = append(revs, rev)
		}

		err = revSuppressor.SuppressRevisions(revs)
		if err != nil {
			log.Println("failed to suppress revisions", revs)
			return err
		}

		return nil
	}
}

func createHandlerForListUpdate(listUpdatedChan chan bool) changeHandlerFunc {
	var lastSeenListRev mediawiki.RevisionId

	return func(changes []mediawiki.Revision) error {
		for _, rev := range changes {
			if rev.Title != config.GetSuppressionListName() {
				continue
			}

			if rev.Id == lastSeenListRev {
				continue
			}

			lastSeenListRev = rev.Id

			listUpdatedChan <- true
		}

		return nil
	}
}
