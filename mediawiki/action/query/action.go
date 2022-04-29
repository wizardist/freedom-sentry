package query

import "golang.org/x/exp/maps"

type Property interface {
	ToPropertyPayload() map[string]interface{}

	setResponse(json map[string]interface{}) error
}

type Meta interface {
	ToMetaPayload() map[string]interface{}

	setResponse(json map[string]interface{}) error
}

type List interface {
	ToListPayload() map[string]interface{}

	setResponse(json map[string]interface{}) error
}

type Query struct {
	Properties      []Property
	Meta            []Meta
	List            []List
	PageNames       []string
	FollowRedirects bool
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
		// FIXME: "prop" is overwritten, must be piped
		maps.Copy(payload, p.ToPropertyPayload())
	}

	for _, m := range a.Meta {
		// FIXME: "meta" is overwritten, must be piped
		maps.Copy(payload, m.ToMetaPayload())
	}

	for _, l := range a.List {
		// FIXME: "list" is overwritten, must be piped
		maps.Copy(payload, l.ToListPayload())
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

	for _, l := range a.List {
		err := l.setResponse(query)
		if err != nil {
			return err
		}
	}

	return nil
}
