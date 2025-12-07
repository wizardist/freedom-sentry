package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/wizardist/freedom-sentry/config"
	"github.com/wizardist/freedom-sentry/eventstreams"
	"github.com/wizardist/freedom-sentry/mediawiki"
	"github.com/wizardist/freedom-sentry/suppressor"
)

func scheduleRecentChangeSuppressor(pageRepo suppressor.SuppressedPageRepository, revSuppressor suppressor.RevisionSuppressor, listUpdatedChan chan bool) {
	changeProcessor := make(chan []mediawiki.Revision, 100)

	subhandlers := []changeHandlerFunc{
		createHandlerForListUpdate(listUpdatedChan),
		createHandlerChangeForSuppressor(pageRepo, revSuppressor),
	}
	handleChanges := createChangesHandler(subhandlers, changeProcessor)

	go handleChanges()

	// Initialize EventStreams client
	eventStreamsURL := config.GetEventStreamsURL()
	wikiDomain := config.GetWikiDomain()

	client := eventstreams.NewClient(eventStreamsURL, &http.Client{
		Timeout: 0, // No timeout for SSE
	})

	reconnector := eventstreams.NewReconnector(client, &eventstreams.ExponentialBackoff{
		Initial: 1 * time.Second,
		Max:     60 * time.Second,
		Factor:  2.0,
		Jitter:  0.25,
	})

	// Start 30 minutes in the past
	since := time.Now().Add(-30 * time.Minute)

	ctx := context.Background()

	// Run reconnector (blocks forever)
	err := reconnector.Run(ctx, &since, func(event eventstreams.Event) error {
		return handleEvent(event, wikiDomain, changeProcessor)
	})

	if err != nil {
		slog.Error("EventStreams reconnector stopped", "error", err)
	}
}

func handleEvent(
	event eventstreams.Event,
	wikiDomain string,
	changeProcessor chan<- []mediawiki.Revision,
) error {
	rc, err := eventstreams.ParseRecentChange(event.Data)
	if err != nil {
		slog.Error("failed to parse event", "error", err)
		return nil // Don't fail on parse errors
	}

	// Filter by wiki domain
	if wikiDomain != "" && rc.ServerName != wikiDomain {
		return nil
	}

	// Apply filters (bot, type)
	if !rc.ShouldProcess() {
		return nil
	}

	rev, err := rc.ToRevision()
	if err != nil {
		slog.Error("failed to convert to revision", "error", err)
		return nil
	}

	// Send single revision wrapped in slice (matches current interface)
	changeProcessor <- []mediawiki.Revision{rev}

	return nil
}
