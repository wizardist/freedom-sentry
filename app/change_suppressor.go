package app

import (
	"freedom-sentry/mediawiki"
	"freedom-sentry/suppressor"
	"time"
)

func scheduleRecentChangeSuppressor(pageRepo suppressor.SuppressedPageRepository, revSuppressor suppressor.RevisionSuppressor, listUpdatedChan chan bool, revRepo suppressor.RevisionRepository) {
	changeProcessor := make(chan []mediawiki.Revision)

	subhandlers := []changeHandlerFunc{
		createHandlerForListUpdate(listUpdatedChan),
		createHandlerChangeForSuppressor(pageRepo, revSuppressor),
	}
	handleChanges := createChangesHandler(subhandlers, changeProcessor)

	go handleChanges()

	lastProcessed := time.Now().Add(-30 * time.Minute)

	for range time.Tick(5 * time.Second) {
		lastProcessed, _ = scanChanges(revRepo, lastProcessed, changeProcessor)
	}
}
