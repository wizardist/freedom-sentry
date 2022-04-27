package suppressor

import (
	"freedom-sentry/mediawiki"
	"freedom-sentry/mediawiki/action/revisiondelete"
	"log"
	"time"
)

type RevisionSuppressor interface {
	SuppressRevisions(revs []mediawiki.Revision) error
}

type revisionSuppressorImpl struct {
	api mediawiki.Api
}

func (rs revisionSuppressorImpl) SuppressRevisions(revs []mediawiki.Revision) error {
	if len(revs) == 0 {
		log.Println("nothing to suppress")
		return nil
	}

	ids := make([]mediawiki.RevisionId, 0, len(revs))

	for _, rev := range revs {
		ids = append(ids, rev.Id)
	}

	log.Printf("suppressing %d revisions", len(ids))

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

func NewRevisionSuppressor(api mediawiki.Api) RevisionSuppressor {
	return &filteringRevisionSuppressor{
		suppressor: &batchingSuppressor{
			period: 5 * time.Second,
			size:   500,
			suppressor: &revisionSuppressorImpl{
				api: api,
			},
		},
		/*suppressor: &revisionSuppressorImpl{
			api: api,
		},*/
	}
}

type filteringRevisionSuppressor struct {
	suppressor RevisionSuppressor
}

// TODO: If an article is not found or moved, message back to the list manager to update the source list in the wiki

func (rs filteringRevisionSuppressor) SuppressRevisions(revs []mediawiki.Revision) error {
	filtered := make([]mediawiki.Revision, 0, len(revs))

	for _, rev := range revs {
		if rev.IsSuppressed {
			continue
		}

		filtered = append(filtered, rev)
	}

	return rs.suppressor.SuppressRevisions(filtered)
}
