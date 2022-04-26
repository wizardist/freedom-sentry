package main

import (
	"freedom-sentry/config"
	"freedom-sentry/mediawiki"
	"freedom-sentry/mediawiki/action/query"
	"freedom-sentry/suppressor"
	"log"
	"net/http"
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

	client := &http.Client{}
	client.Timeout = 15 * time.Second

	api := mediawiki.NewApi(os.Getenv(config.EnvApiEndpoint), client, tokenFn)
	revRepo := suppressor.NewRepository(api)
	pageRepo := suppressor.NewPageRepository(revRepo, os.Getenv(config.EnvSuppressionListName))

	pageSuppressor := suppressor.NewPageSuppressor(revRepo, suppressor.NewRevisionSuppressor(api))

	runSuppression(pageRepo, pageSuppressor)

	for range time.Tick(15 * time.Minute) {
		runSuppression(pageRepo, pageSuppressor)
	}
}

func runSuppression(pageRepo suppressor.PageRepository, pageSuppressor suppressor.PageSuppressor) {
	log.Println("running a new suppression job")

	suppressedPages, err := pageRepo.GetAllSuppressed()
	for _, pageName := range suppressedPages {
		err = pageSuppressor.SuppressPageByName(pageName)
		if err != nil {
			log.Printf("failed to suppress [%s] revisions: %v", pageName, err)
		}
	}
}
