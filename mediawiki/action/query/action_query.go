package query

import (
	"errors"
	"freedom-sentry/mediawiki"
	"golang.org/x/exp/maps"
)

type Query struct {
	Properties      []Property
	Meta            []Meta
	PageNames       []string
	FollowRedirects bool
}

type Property interface {
	ToPropertyPayload() map[string]interface{}

	setResponse(json map[string]interface{}) error
}

type Meta interface {
	ToMetaPayload() map[string]interface{}

	setResponse(json map[string]interface{}) error
}

func (Query) IsWriteAction() bool {
	return false
}

func (a Query) ToActionPayload() map[string]interface{} {
	payload := map[string]interface{}{
		"action": "query",
	}

	if len(a.PageNames) > 0 {
		payload["titles"] = a.PageNames
	}

	if a.FollowRedirects {
		payload["redirects"] = true
	}

	for _, p := range a.Properties {
		// FIXME: "prop" is overwritten
		maps.Copy(payload, p.ToPropertyPayload())
	}

	for _, m := range a.Meta {
		// FIXME: "meta" is overwritten
		maps.Copy(payload, m.ToMetaPayload())
	}

	return payload
}

func (a Query) SetResponse(payload map[string]interface{}) error {
	var query map[string]interface{}
	var ok bool

	if query, ok = payload["query"].(map[string]interface{}); !ok {
		return nil
	}

	for _, p := range a.Properties {
		err := p.setResponse(query)
		if err != nil {
			return err
		}
	}

	for _, m := range a.Meta {
		err := m.setResponse(query)
		if err != nil {
			return err
		}
	}

	return nil
}

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

type TokensQueryMeta struct {
	// FIXME: Only supports CSRF token requests

	Type []string

	tokens struct {
		Csrf string
	}
}

func (qm TokensQueryMeta) ToMetaPayload() map[string]interface{} {
	return map[string]interface{}{
		"meta": "tokens",
		"type": qm.Type,
	}
}

func (qm TokensQueryMeta) GetTokens() struct{ Csrf string } {
	return qm.tokens
}

func (qm *TokensQueryMeta) setResponse(json map[string]interface{}) error {
	tokens := json["tokens"].(map[string]interface{})

	qm.tokens.Csrf = tokens["csrftoken"].(string)

	return nil
}
