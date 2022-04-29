package main

import (
	"freedom-sentry/config"
	"freedom-sentry/http"
	"freedom-sentry/mediawiki"
	"freedom-sentry/mediawiki/action/query"
	"freedom-sentry/suppressor"
	"log"
	"os"
	"time"
)

func main() {
	tokenFn := func(api mediawiki.Api) (string, error) {
		tokensQm := &query.TokensQueryMeta{
			Type: []string{"csrf"},
		}
		a := query.Query{
			Meta: []query.Meta{tokensQm},
		}

		log.Println("requesting a new CSRF token")

		err := api.Execute(a)
		if err != nil {
			log.Println("failed to retrieve a CSRF token:", err)
			return "", err
		}

		csrfToken := tokensQm.GetTokens().Csrf

		log.Println("acquired a new CSRF token:", csrfToken)

		return csrfToken, nil
	}

	api := mediawiki.NewApi(os.Getenv(config.EnvApiEndpoint), http.DefaultClient, tokenFn)
	revRepo := suppressor.NewRepository(api)
	pageRepo := suppressor.NewPageRepository(revRepo, os.Getenv(config.EnvSuppressionListName))

	revSuppressor := suppressor.NewRevisionSuppressor(api)
	pageSuppressor := suppressor.NewPageSuppressor(revRepo, revSuppressor)

	done := make(chan bool)

	go scheduleListSuppressor(pageRepo, pageSuppressor)
	go scheduleRecentChangeSuppressor(revRepo, pageRepo, revSuppressor)

	<-done
}

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

func scheduleListSuppressor(pageRepo suppressor.PageRepository, pageSuppressor suppressor.PageSuppressor) {
	suppressList(pageRepo, pageSuppressor)

	for range time.Tick(15 * time.Minute) {
		suppressList(pageRepo, pageSuppressor)
	}
}

func suppressList(pageRepo suppressor.PageRepository, pageSuppressor suppressor.PageSuppressor) {
	log.Println("running a new suppression job")

	suppressedPages, err := pageRepo.GetAllSuppressed()
	for _, pageName := range suppressedPages {
		err = pageSuppressor.SuppressPageByName(pageName)
		if err != nil {
			log.Printf("failed to suppress [%s] revisions: %v", pageName, err)
		}
	}
}
