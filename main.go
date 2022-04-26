package main

import (
	"freedom-sentry/config"
	"freedom-sentry/mediawiki"
	"freedom-sentry/mediawiki/action/query"
	"freedom-sentry/suppressor"
	"log"
	"net/http"
	"os"
	"strings"
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
	repo := suppressor.NewRepository(api)
	suppressedPagesStr, err := repo.GetLatestPageContent(os.Getenv(config.EnvSuppressionListName))
	if err != nil {
		log.Fatalln(err)
	}

	suppressedPages := strings.Split(suppressedPagesStr, "\n")

	pageSuppressor := suppressor.NewPageSuppressor(repo, suppressor.NewRevisionSuppressor(api))

	for _, pageName := range suppressedPages {
		err = pageSuppressor.SuppressPageByName(pageName)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
