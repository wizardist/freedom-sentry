package suppressor

import (
	"log"
)

type PageSuppressor interface {
	SuppressPageByName(name string) error
}

func NewPageSuppressor(revRepo RevisionRepository, revSuppressor RevisionSuppressor) PageSuppressor {
	return &pageSuppressorImpl{
		revRepo:       revRepo,
		revSuppressor: revSuppressor,
	}
}

type pageSuppressorImpl struct {
	revRepo       RevisionRepository
	revSuppressor RevisionSuppressor
}

func (ps pageSuppressorImpl) SuppressPageByName(name string) error {
	log.Println("retrieving revisions for page:", name)
	revs, err := ps.revRepo.GetAllByPageName(name)
	if err != nil {
		log.Println("failed to retrieve revisions for page:", err)
		return err
	}

	err = ps.revSuppressor.SuppressRevisions(revs)

	return err
}
