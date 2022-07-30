package app

import (
	"freedom-sentry/mediawiki"
	"freedom-sentry/suppressor"
	"log"
	"time"
)

func scheduleRecentChangeSuppressor(revRepo suppressor.RevisionRepository, pageRepo suppressor.PageRepository, revSuppressor suppressor.RevisionSuppressor) {
	lastProcessed := suppressRecentChanges(time.Now().Add(-30*time.Minute), revRepo, pageRepo, revSuppressor)

	for range time.Tick(5 * time.Second) {
		lastProcessed = suppressRecentChanges(lastProcessed, revRepo, pageRepo, revSuppressor)
	}
}

func suppressRecentChanges(since time.Time, repo suppressor.RevisionRepository, pageRepo suppressor.PageRepository, revSuppressor suppressor.RevisionSuppressor) time.Time {
	changes, err := repo.GetRecentChanges(since)
	if err != nil {
		log.Println("failed to get recent changes since", since, "error:", err)
		return time.Time{}
	}

	mostRecent := since
	list, err := pageRepo.GetAllSuppressed()
	if err != nil {
		log.Println("failed to get suppression list:", err)
		return mostRecent
	}

	indexedList := make(map[string]bool, len(list))
	for _, title := range list {
		indexedList[title] = true
	}

	revs := make([]mediawiki.Revision, 0, len(changes))
	for _, rev := range changes {
		if rev.Timestamp.After(mostRecent) {
			mostRecent = rev.Timestamp
		}

		if _, inList := indexedList[rev.Title]; !inList {
			continue
		}

		revs = append(revs, rev)
	}

	err = revSuppressor.SuppressRevisions(revs)
	if err != nil {
		log.Println("failed to suppress revisions", revs)
	}

	return mostRecent
}
