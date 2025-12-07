package app

import (
	"log/slog"
	"time"

	"github.com/wizardist/freedom-sentry/config"
	"github.com/wizardist/freedom-sentry/suppressor"
)

func scheduleListSuppressor(pageRepo suppressor.SuppressedPageRepository, pageSuppressor suppressor.PageSuppressor) {
	if !config.IsInitFullscanSkipped() {
		suppressList(pageRepo, pageSuppressor)
	}

	for range time.Tick(15 * time.Minute) {
		suppressList(pageRepo, pageSuppressor)
	}
}

func suppressList(pageRepo suppressor.SuppressedPageRepository, pageSuppressor suppressor.PageSuppressor) {
	slog.Info("running scheduled suppression job")

	suppressedPages, err := pageRepo.GetAll()
	slog.Info("found suppressed pages", "count", len(suppressedPages), "pages", suppressedPages)
	for _, pageName := range suppressedPages {
		err = pageSuppressor.SuppressPageByName(pageName)
		if err != nil {
			slog.Error("failed to suppress page revisions", "page", pageName, "error", err)
		}
	}
}
