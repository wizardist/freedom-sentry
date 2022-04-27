package suppressor

import (
	"errors"
	"freedom-sentry/mediawiki"
	"reflect"
	"strings"
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

type mockSuppressor struct {
	called      bool
	callCount   int
	revs        []mediawiki.Revision
	throwError  bool
	callHistory string
}

func (m *mockSuppressor) SuppressRevisions(revs []mediawiki.Revision) error {
	m.called = true
	m.revs = revs
	m.addHistory(revs)

	if m.throwError {
		return errors.New("dummy error")
	}

	return nil
}

func (m *mockSuppressor) addHistory(revs []mediawiki.Revision) {
	m.callCount++

	batch := make([]string, len(revs))
	for i, rev := range revs {
		batch[i] = string(rev.Id)
	}

	if m.callCount > 1 {
		m.callHistory += "|"
	}

	m.callHistory += strings.Join(batch, ",")
}

func Test_filteringRevisionSuppressor_SuppressRevisions(t *testing.T) {
	tests := []struct {
		name     string
		revs     []mediawiki.Revision
		expected []mediawiki.Revision
		wantErr  bool
	}{
		{
			name:     "Nils",
			revs:     nil,
			expected: []mediawiki.Revision{},
			wantErr:  false,
		},
		{
			name:     "Empties",
			revs:     []mediawiki.Revision{},
			expected: []mediawiki.Revision{},
			wantErr:  false,
		},
		{
			name: "Nothing to suppress",
			revs: []mediawiki.Revision{
				{Id: "1", IsSuppressed: true},
				{Id: "2", IsSuppressed: true},
			},
			expected: []mediawiki.Revision{},
			wantErr:  false,
		},
		{
			name: "Some to suppress",
			revs: []mediawiki.Revision{
				{Id: "1", IsSuppressed: false},
				{Id: "2", IsSuppressed: true},
			},
			expected: []mediawiki.Revision{
				{Id: "1", IsSuppressed: false},
			},
			wantErr: false,
		},
		{
			name: "All are suppressed",
			revs: []mediawiki.Revision{
				{Id: "1", IsSuppressed: false},
				{Id: "2", IsSuppressed: false},
			},
			expected: []mediawiki.Revision{
				{Id: "1", IsSuppressed: false},
				{Id: "2", IsSuppressed: false},
			},
			wantErr: false,
		},
		{
			name: "Error from fn",
			revs: []mediawiki.Revision{
				{Id: "1", IsSuppressed: false},
				{Id: "2", IsSuppressed: false},
			},
			expected: []mediawiki.Revision{
				{Id: "1", IsSuppressed: false},
				{Id: "2", IsSuppressed: false},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suppressor := &mockSuppressor{
				throwError: tt.wantErr,
			}

			rs := filteringRevisionSuppressor{suppressor: suppressor}
			err := rs.SuppressRevisions(tt.revs)
			if (err != nil) != tt.wantErr {
				t.Errorf("SuppressRevisions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(tt.expected, suppressor.revs) {
				t.Errorf("SuppressRevisions(), bad action: got = %v, want %v", suppressor.revs, tt.expected)
			}
		})
	}
}
