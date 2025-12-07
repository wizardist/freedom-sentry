package app

import (
	"log/slog"

	"github.com/wizardist/freedom-sentry/mediawiki"
	"github.com/wizardist/freedom-sentry/mediawiki/action/query"
)

func acquireCsrfTokenFn(api mediawiki.Api) (mediawiki.Token, error) {
	tokensQm := &query.TokensMetaQuery{
		Type: []string{"csrf"},
	}
	a := query.Query{
		Meta: []query.Meta{tokensQm},
	}

	slog.Debug("requesting CSRF token")

	err := api.Execute(a)
	if err != nil {
		slog.Error("failed to retrieve CSRF token", "error", err)
		return "", err
	}

	csrfToken := tokensQm.GetTokens().Csrf

	slog.Debug("acquired CSRF token")

	return mediawiki.Token(csrfToken), nil
}
