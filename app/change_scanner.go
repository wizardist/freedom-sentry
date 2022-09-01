package app

import (
	"freedom-sentry/mediawiki"
	"freedom-sentry/suppressor"
	"log"
	"time"
)

func scanChanges(repo suppressor.RevisionRepository, since time.Time, changeProcessor chan<- []mediawiki.Revision) (time.Time, error) {
	changes, err := repo.GetRecentChanges(since)
	if err != nil {
		log.Println("failed to get recent changes since", since, "error:", err)
		return time.Time{}, err
	}

	mostRecent := since

	if len(changes) == 0 {
		return mostRecent, nil
	}

	changeProcessor <- changes

	for _, change := range changes {
		if change.Timestamp.After(mostRecent) {
			mostRecent = change.Timestamp.Add(time.Second)
		}
	}

	return mostRecent, nil
}
