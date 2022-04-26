package suppressor

import (
	"freedom-sentry/mediawiki"
	"freedom-sentry/mediawiki/action/query"
)

type RevisionRepository interface {
	GetAllByPageName(name string) ([]mediawiki.Revision, error)
	GetLatestPageContent(name string) (string, error)
}

type RepositoryImpl struct {
	api mediawiki.Api
}

func NewRepository(api mediawiki.Api) RevisionRepository {
	return &RepositoryImpl{api: api}
}

func (rr *RepositoryImpl) GetAllByPageName(name string) ([]mediawiki.Revision, error) {
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

func (rr *RepositoryImpl) GetLatestPageContent(name string) (string, error) {
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