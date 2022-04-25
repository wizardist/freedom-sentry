package suppressor

import (
	"freedom-sentry/mediawiki"
	"freedom-sentry/mediawiki/action/revisiondelete"
	"log"
)

type RevisionSuppressor interface {
	SuppressRevisions(revs []mediawiki.Revision) error
}

type revisionSuppressorImpl struct {
	api mediawiki.Api
}

func NewRevisionSuppressor(api mediawiki.Api) RevisionSuppressor {
	return &revisionSuppressorImpl{api: api}
}

// Update the list of suppressed pages in a separate process
// During run
// - For each article
//   - Get its revisions, filter to only include non-suppressed
//   - Suppress revision author and comment
// - If an article is not found or moved, message back to the list manager to update the source list in the wiki

func (rs revisionSuppressorImpl) SuppressRevisions(revs []mediawiki.Revision) error {
	ids := make([]mediawiki.RevisionId, 0, len(revs))

	for _, rev := range revs {
		if rev.IsSuppressed {
			continue
		}

		ids = append(ids, rev.Id)
	}

	if len(ids) == 0 {
		log.Println("nothing to suppress")
		return nil
	}

	log.Printf("suppressing %d/%d revisions", len(ids), len(revs))

	return rs.api.Execute(getActionForRevisions(ids))
}

func getActionForRevisions(revs []mediawiki.RevisionId) revisiondelete.RevisionDelete {
	return revisiondelete.RevisionDelete{
		Type:        "revision",
		Revisions:   revs,
		HideDetails: []string{"user", "comment"},
		Suppress:    mediawiki.TextBoolYes,
	}
}
