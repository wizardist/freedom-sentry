package mediawiki

import (
	"fmt"
)

type RevisionId string

type Revision struct {
	Id           RevisionId
	IsSuppressed bool

	Content string
}

func RevisionIdFromAny(v interface{}) RevisionId {
	return RevisionId(fmt.Sprintf("%.f", v))
}
