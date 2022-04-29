package query

import (
	"errors"
	"freedom-sentry/mediawiki"
	"time"
)

type RecentChangesQueryList struct {
	Start      time.Time
	Direction  string
	Properties []string
	Show       []string
	Limit      int
	Types      []string

	recentChanges []mediawiki.Revision
}

func (r RecentChangesQueryList) ToListPayload() map[string]interface{} {
	return map[string]interface{}{
		"list":    "recentchanges",
		"rcstart": r.Start.Format(time.RFC3339),
		"rcdir":   r.Direction,
		"rcprop":  r.Properties,
		"rcshow":  r.Show,
		"rclimit": r.Limit,
		"rctype":  r.Types,
	}
}

func (r *RecentChangesQueryList) setResponse(json map[string]interface{}) error {
	rawRevs, ok := jsonMapValueToSliceOfMaps(json["recentchanges"])
	if !ok {
		return errors.New("response does not contain `recentchanges` or invalid structure")
	}

	revs := make([]mediawiki.Revision, len(rawRevs))

	for i, rawRev := range rawRevs {

		rev := mediawiki.Revision{
			Id: mediawiki.RevisionIdFromAny(rawRev["revid"]),
		}

		_, rev.IsSuppressed = rawRev["suppressed"]

		timestampStr, ok := rawRev["timestamp"].(string)
		if ok {
			if timestamp, err := time.Parse(time.RFC3339, timestampStr); err == nil {
				rev.Timestamp = timestamp
			}
		}

		rev.Title, _ = rawRev["title"].(string)

		revs[i] = rev
	}

	r.recentChanges = revs

	return nil
}

func jsonMapValueToSliceOfMaps(json interface{}) ([]map[string]interface{}, bool) {
	rawSlice, ok := json.([]interface{})
	if !ok {
		return nil, false
	}

	typedSlice := make([]map[string]interface{}, len(rawSlice))
	for i, raw := range rawSlice {
		typedSlice[i], ok = raw.(map[string]interface{})
		if !ok {
			return nil, false
		}
	}

	return typedSlice, true
}

func (r RecentChangesQueryList) GetRecentChanges() []mediawiki.Revision {
	return r.recentChanges
}
