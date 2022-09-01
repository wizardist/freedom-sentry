package app

import (
	"fmt"
	"freedom-sentry/config"
	"freedom-sentry/http"
	"freedom-sentry/mediawiki"
	"freedom-sentry/mediawiki/action/query"
	"freedom-sentry/suppressor"
	"os"
)

func Run() {
	apiEndpoint := os.Getenv(config.EnvApiEndpoint)

	api := mediawiki.NewApi(apiEndpoint, http.DefaultClient, acquireCsrfTokenFn)

	validateAccess(api)

	revRepo := suppressor.NewRepository(api)

	revSuppressor := suppressor.NewRevisionSuppressor(api)
	pageSuppressor := suppressor.NewPageSuppressor(revRepo, revSuppressor)

	listUpdatedChan := make(chan bool)

	pageRepo, listPurgeChan := suppressor.NewPageRepository(revRepo, config.GetSuppressionListName())

	go func() {
		for {
			select {
			case <-listUpdatedChan:
				listPurgeChan <- true
				suppressList(pageRepo, pageSuppressor)
			}
		}
	}()

	done := make(chan bool)

	go scheduleListSuppressor(pageRepo, pageSuppressor)
	go scheduleRecentChangeSuppressor(pageRepo, revSuppressor, listUpdatedChan, revRepo)

	<-done
}

// validateAccess panics if the given access credentials do not provide suppression capability.
func validateAccess(api mediawiki.Api) {
	userinfoQuery := query.UserinfoMetaQuery{Properties: []string{"rights"}}
	action := query.Query{Meta: []query.Meta{&userinfoQuery}}

	err := api.Execute(action)
	if err != nil {
		panic(fmt.Errorf("failed to retrieve user access rights: %w", err))
	}

	userinfo := userinfoQuery.GetUserinfo()

	for _, right := range userinfo.Rights {
		if right == "suppressrevision" {
			return
		}
	}

	panic("the configured access token doesn't allow suppressing revisions")
}
