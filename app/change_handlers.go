package app

import (
	"log/slog"

	"github.com/wizardist/freedom-sentry/config"
	"github.com/wizardist/freedom-sentry/mediawiki"
	"github.com/wizardist/freedom-sentry/suppressor"
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
			slog.Error("failed to get suppression list", "error", err)
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

		slog.Debug("processing revisions for suppression",
			"total_changes", len(changes),
			"list_size", len(list),
			"matching_revisions", len(revs))

		err = revSuppressor.SuppressRevisions(revs)
		if err != nil {
			slog.Error("failed to suppress revisions", "error", err, "revision_count", len(revs))
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
