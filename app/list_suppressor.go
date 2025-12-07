package app

import (
	"log"
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
	log.Println("running a new suppression job")

	suppressedPages, err := pageRepo.GetAll()
	log.Println("found suppressed pages", suppressedPages)
	for _, pageName := range suppressedPages {
		err = pageSuppressor.SuppressPageByName(pageName)
		if err != nil {
			log.Printf("failed to suppress [%s] revisions: %v", pageName, err)
		}
	}
}
