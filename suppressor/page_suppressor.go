package suppressor

import (
	"log/slog"
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
	slog.Debug("retrieving page revisions", "page", name)
	revs, err := ps.revRepo.GetAllByPageName(name)
	if err != nil {
		slog.Error("failed to retrieve page revisions", "page", name, "error", err)
		return err
	}

	slog.Info("retrieved page revisions", "page", name, "count", len(revs))

	err = ps.revSuppressor.SuppressRevisions(revs)

	return err
}
