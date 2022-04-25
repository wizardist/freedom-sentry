package revisiondelete

import (
	"freedom-sentry/mediawiki"
	"freedom-sentry/util"
	"reflect"
	"testing"
)

func TestRevisionDelete_ToActionPayload(t *testing.T) {
	tests := []struct {
		name   string
		action RevisionDelete
		want   map[string]interface{}
	}{
		{
			name:   "Empty",
			action: RevisionDelete{},
			want: map[string]interface{}{
				"action": actionName,
				"type":   Type(""),
				"ids":    util.CreateNilSlice[mediawiki.RevisionId](),
			},
		},
		{
			name: "Suppress user and comment for revisions",
			action: RevisionDelete{
				Type:        "revision",
				Revisions:   []mediawiki.RevisionId{"42", "1337"},
				HideDetails: []string{"user", "comment"},
				Suppress:    mediawiki.TextBoolYes,
			},
			want: map[string]interface{}{
				"action":   actionName,
				"type":     TypeRevision,
				"ids":      []mediawiki.RevisionId{"42", "1337"},
				"hide":     []string{"user", "comment"},
				"suppress": mediawiki.TextBoolYes,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.action.ToActionPayload(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToActionPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}
