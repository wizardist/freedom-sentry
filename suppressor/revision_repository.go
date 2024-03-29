package suppressor

import (
	"freedom-sentry/mediawiki"
	"freedom-sentry/mediawiki/action/query"
	"time"
)

type RevisionRepository interface {
	GetAllByPageName(name string) ([]mediawiki.Revision, error)
	GetLatestPageContent(name string) (string, error)
	GetRecentChanges(since time.Time) ([]mediawiki.Revision, error)
}

func NewRepository(api mediawiki.Api) RevisionRepository {
	return &revRepoImpl{api: api}
}

type revRepoImpl struct {
	api mediawiki.Api
}

func (rr *revRepoImpl) GetAllByPageName(name string) ([]mediawiki.Revision, error) {
	revProp := &query.RevisionsQueryProperty{
		Properties: []string{"ids", "timestamp", "user"},
		Limit:      5000,
	}

	q := query.Query{
		Properties:      []query.Property{revProp},
		PageNames:       []string{name},
		FollowRedirects: true,
	}

	err := rr.api.Execute(q)

	return revProp.GetRevisions(), err
}

func (rr *revRepoImpl) GetLatestPageContent(name string) (string, error) {
	revProp := &query.RevisionsQueryProperty{
		Properties: []string{"ids", "content"},
		Limit:      1,
	}

	q := query.Query{
		Properties:      []query.Property{revProp},
		PageNames:       []string{name},
		FollowRedirects: true,
	}

	err := rr.api.Execute(q)
	if err != nil {
		return "", err
	}

	revisions := revProp.GetRevisions()
	if len(revisions) == 0 {
		return "", nil
	}

	return revisions[0].Content, nil
}

func (rr *revRepoImpl) GetRecentChanges(since time.Time) ([]mediawiki.Revision, error) {
	changes := query.RecentChangesQueryList{
		Start:      since,
		Direction:  "newer",
		Properties: []string{"title", "timestamp", "ids", "user"},
		Show:       []string{"!bot"},
		Limit:      5000,
		Types:      []string{"edit"},
		TopOnly:    true,
	}
	action := query.Query{
		List: []query.List{&changes},
	}
	err := rr.api.Execute(action)

	return changes.GetRecentChanges(), err
}
