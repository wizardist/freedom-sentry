package mediawiki

import (
	"fmt"
	"time"
)

type RevisionId string

type Revision struct {
	Id           RevisionId
	IsSuppressed bool

	Title   string
	Content string

	Timestamp time.Time
}

func RevisionIdFromAny(v interface{}) RevisionId {
	switch val := v.(type) {
	case int64:
		return RevisionId(fmt.Sprintf("%d", val))
	case int:
		return RevisionId(fmt.Sprintf("%d", val))
	case float64:
		return RevisionId(fmt.Sprintf("%.f", val))
	default:
		return RevisionId(fmt.Sprintf("%v", val))
	}
}
