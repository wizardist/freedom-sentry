package app

import (
	"freedom-sentry/mediawiki"
	"freedom-sentry/mediawiki/action/query"
	"log"
)

func acquireCsrfTokenFn(api mediawiki.Api) (mediawiki.Token, error) {
	tokensQm := &query.TokensMetaQuery{
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

	return mediawiki.Token(csrfToken), nil
}
