package suppressor

import (
	"errors"
	"freedom-sentry/mediawiki"
	"freedom-sentry/mediawiki/action/revisiondelete"
	"reflect"
	"testing"
)

type mockApi struct {
	executeCalled     bool
	executeAction     mediawiki.Action
	executeThrowError bool
}

func (m *mockApi) Execute(action mediawiki.Action) error {
	m.executeCalled = true
	m.executeAction = action

	if m.executeThrowError {
		return errors.New("dummy error")
	}

	return nil
}

func TestSuppressRevisions(t *testing.T) {
	tests := []struct {
		name         string
		revs         []mediawiki.Revision
		expected     []mediawiki.RevisionId
		wantSuppress bool
		wantErr      bool
	}{
		{
			name:     "Nils",
			revs:     nil,
			expected: []mediawiki.RevisionId{},
			wantErr:  false,
		},
		{
			name:     "Empties",
			revs:     []mediawiki.Revision{},
			expected: []mediawiki.RevisionId{},
			wantErr:  false,
		},
		{
			name: "Nothing to suppress",
			revs: []mediawiki.Revision{
				{Id: "1", IsSuppressed: true},
				{Id: "2", IsSuppressed: true},
			},
			expected: []mediawiki.RevisionId{},
			wantErr:  false,
		},
		{
			name: "Some to suppress",
			revs: []mediawiki.Revision{
				{Id: "1", IsSuppressed: false},
				{Id: "2", IsSuppressed: true},
			},
			expected:     []mediawiki.RevisionId{"1"},
			wantSuppress: true,
			wantErr:      false,
		},
		{
			name: "All are suppressed",
			revs: []mediawiki.Revision{
				{Id: "1", IsSuppressed: false},
				{Id: "2", IsSuppressed: false},
			},
			expected:     []mediawiki.RevisionId{"1", "2"},
			wantSuppress: true,
			wantErr:      false,
		},
		{
			name: "Error from fn",
			revs: []mediawiki.Revision{
				{Id: "1", IsSuppressed: false},
				{Id: "2", IsSuppressed: false},
			},
			expected:     []mediawiki.RevisionId{"1", "2"},
			wantSuppress: true,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &mockApi{
				executeThrowError: tt.wantErr,
			}

			rs := revisionSuppressorImpl{api: api}
			err := rs.SuppressRevisions(tt.revs)
			if (err != nil) != tt.wantErr {
				t.Errorf("SuppressRevisions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantSuppress != api.executeCalled {
				t.Errorf("Had to be suppressed but the API was not invoked")
				return
			}

			expectedAction := revisiondelete.RevisionDelete{
				Type:        "revision",
				Revisions:   tt.expected,
				HideDetails: []string{"user", "comment"},
				Suppress:    mediawiki.TextBoolYes,
			}

			if tt.wantSuppress && !reflect.DeepEqual(expectedAction, api.executeAction) {
				t.Errorf("SuppressRevisions(), bad action: got = %v, want %v", api.executeAction, expectedAction)
			}
		})
	}
}
