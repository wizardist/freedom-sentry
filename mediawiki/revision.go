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
	return RevisionId(fmt.Sprintf("%.f", v))
}
