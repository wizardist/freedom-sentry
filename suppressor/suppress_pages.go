package suppressor

import (
	"freedom-sentry/revision"
	"log"
)

type PageSuppressor interface {
	SuppressPageByName(name string) error
}

type pageSuppressorImpl struct {
	revRepo       revision.Repository
	revSuppressor RevisionSuppressor
}

func NewPageSuppressor(revRepo revision.Repository, revSuppressor RevisionSuppressor) PageSuppressor {
	return &pageSuppressorImpl{
		revRepo:       revRepo,
		revSuppressor: revSuppressor,
	}
}

func (ps pageSuppressorImpl) SuppressPageByName(name string) error {
	log.Println("retrieving revisions for page:", name)
	revs, err := ps.revRepo.GetRevisionsByPageName(name)
	if err != nil {
		log.Println("failed to retrieve revisions for page:", err)
		return err
	}

	err = ps.revSuppressor.SuppressRevisions(revs)

	return err
}
