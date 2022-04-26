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
	suppressedPages, err := pageRepo.GetAllSuppressed()

	pageSuppressor := suppressor.NewPageSuppressor(revRepo, suppressor.NewRevisionSuppressor(api))

	for _, pageName := range suppressedPages {
		err = pageSuppressor.SuppressPageByName(pageName)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
