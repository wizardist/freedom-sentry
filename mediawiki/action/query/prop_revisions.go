package query

import (
	"errors"
	"freedom-sentry/mediawiki"
)

type RevisionsQueryProperty struct {
	Properties []string
	Limit      int

	revisions []mediawiki.Revision
}

func (qp RevisionsQueryProperty) ToPropertyPayload() map[string]interface{} {
	payload := map[string]interface{}{
		"prop":    "revisions",
		"rvprop":  qp.Properties,
		"rvlimit": qp.Limit,
	}

	return payload
}

func (qp RevisionsQueryProperty) GetRevisions() []mediawiki.Revision {
	return qp.revisions
}

func (qp *RevisionsQueryProperty) setResponse(payload map[string]interface{}) error {
	invalidPayloadErr := errors.New("invalid revisions payload")

	if !isValidRevisionsPayload(payload) {
		return invalidPayloadErr
	}

	for _, page := range payload["pages"].(map[string]interface{}) {
		revs := page.(map[string]interface{})["revisions"].([]interface{})
		qp.revisions = parsePagesPayload(revs)

		break
	}

	return nil
}

func parsePagesPayload(revs []interface{}) []mediawiki.Revision {
	revisions := make([]mediawiki.Revision, len(revs))

	for i, trev := range revs {
		rev := trev.(map[string]interface{})

		_, isSuppressed := rev["suppressed"]

		revision := mediawiki.Revision{
			Id:           mediawiki.RevisionIdFromAny(rev["revid"]),
			IsSuppressed: isSuppressed,
		}

		if content, ok := rev["*"].(string); ok {
			revision.Content = content
		}

		revisions[i] = revision
	}
	return revisions
}

func isValidRevisionsPayload(payload map[string]interface{}) bool {
	if len(payload) == 0 {
		return false
	}

	if _, ok := payload["pages"]; !ok {
		return false
	}

	pages, ok := payload["pages"].(map[string]interface{})
	if !ok {
		return false
	}

	for _, page := range pages {
		revisions, ok := page.(map[string]interface{})["revisions"].([]interface{})
		if !ok {
			return false
		}

		for _, rev := range revisions {
			if _, ok := rev.(map[string]interface{})["revid"]; !ok {
				return false
			}
		}

		break // Only take the first page in case API returns more than one, which is unlikely
	}

	return true
}
