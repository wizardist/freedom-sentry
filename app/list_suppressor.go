package app

import (
	"freedom-sentry/config"
	"freedom-sentry/suppressor"
	"log"
	"time"
)

func scheduleListSuppressor(pageRepo suppressor.PageRepository, pageSuppressor suppressor.PageSuppressor) {
	if !config.IsInitFullscanSkipped() {
		suppressList(pageRepo, pageSuppressor)
	}

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
