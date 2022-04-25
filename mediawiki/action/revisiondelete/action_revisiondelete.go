package revisiondelete

import "freedom-sentry/mediawiki"

type Type string

const TypeRevision Type = "revision"
const actionName = "revisiondelete"

type RevisionDelete struct {
	// Type of revision deletion being performed
	Type Type
	// Identifiers for the revisions to be deleted
	Revisions []mediawiki.RevisionId
	// What to hide for each revision
	HideDetails []string
	// Whether to suppress data from administrators as well as others
	Suppress mediawiki.TextBool
}

func (RevisionDelete) IsWriteAction() bool {
	return true
}

func (a RevisionDelete) ToActionPayload() map[string]interface{} {
	payload := map[string]interface{}{
		"action": actionName,
		"type":   a.Type,
		"ids":    a.Revisions,
	}

	if len(a.HideDetails) > 0 {
		payload["hide"] = a.HideDetails
	}

	if a.Suppress != "" {
		payload["suppress"] = a.Suppress
	}

	return payload
}

func (a RevisionDelete) SetResponse(payload map[string]interface{}) error {
	// TODO: Provide a detailed response

	return nil
}
