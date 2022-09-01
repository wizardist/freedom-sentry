package query

import (
	"freedom-sentry/mediawiki"
	"freedom-sentry/util"
	"golang.org/x/exp/slices"
	"reflect"
	"testing"
	"time"
)

func TestRecentChangesQueryList_ToListPayload(t *testing.T) {
	tests := []struct {
		name string
		list RecentChangesQueryList
		want map[string]interface{}
	}{
		{
			name: "Key test",
			list: RecentChangesQueryList{
				Start:      util.WithoutErr(time.Parse(time.RFC3339, "2022-04-20T12:13:14Z")),
				Direction:  "newer",
				Properties: []string{"title", "timestamp", "ids", "user"},
				Show:       []string{"!bot", "minor"},
				Limit:      5000,
				Types:      []string{"edit", "log"},
				TopOnly:    true,
			},
			want: map[string]interface{}{
				"list":      "recentchanges",
				"rcstart":   "2022-04-20T12:13:14Z",
				"rcdir":     "newer",
				"rcprop":    []string{"title", "timestamp", "ids", "user"},
				"rcshow":    []string{"!bot", "minor"},
				"rclimit":   5000,
				"rctype":    []string{"edit", "log"},
				"rctoponly": true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.list.ToListPayload(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToListPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecentChangesQueryList_setResponse(t *testing.T) {
	tests := []struct {
		name     string
		json     map[string]interface{}
		expected []mediawiki.Revision
		wantErr  bool
	}{
		{
			name:    "No recentchanges",
			json:    map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "Recentchanges not a slice",
			json: map[string]interface{}{
				"recentchanges": interface{}(map[string]interface{}{"some": "thing"}),
			},
			wantErr: true,
		},
		{
			name: "Empty with production structure from json.Unmarshal",
			json: map[string]interface{}{
				"recentchanges": interface{}([]interface{}{}),
			},
		},
		{
			name: "Revision",
			json: map[string]interface{}{
				// Value is an interface{}-wrapped slice of interface{}-wrapped maps
				"recentchanges": interface{}([]interface{}{
					map[string]interface{}{
						"type":       "edit",
						"ns":         0,
						"title":      "Test title",
						"pageid":     float64(42),
						"revid":      float64(73),
						"old_revid":  float64(72),
						"timestamp":  "2022-04-20T12:13:14Z",
						"suppressed": "",
					},
					map[string]interface{}{
						"type":       "edit",
						"ns":         0,
						"title":      "Test title",
						"pageid":     float64(42),
						"revid":      float64(73),
						"old_revid":  float64(72),
						"timestamp":  "malformed time",
						"suppressed": "",
					},
					map[string]interface{}{
						"type":      "edit",
						"ns":        0,
						"title":     "Test title",
						"pageid":    float64(42),
						"revid":     float64(74),
						"old_revid": float64(73),
					},
				}),
			},
			expected: []mediawiki.Revision{
				{
					Id:           "73",
					IsSuppressed: true,
					Timestamp:    util.WithoutErr(time.Parse(time.RFC3339, "2022-04-20T12:13:14Z")),
					Title:        "Test title",
				},
				{
					Id:           "73",
					IsSuppressed: true,
					Title:        "Test title",
				},
				{
					Id:    "74",
					Title: "Test title",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RecentChangesQueryList{}
			if err := r.setResponse(tt.json); (err != nil) != tt.wantErr {
				t.Errorf("setResponse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if !slices.Equal(r.GetRecentChanges(), tt.expected) {
				t.Errorf("GetRecentChanges() got = %v, wanted = %v", r.GetRecentChanges(), tt.expected)
			}
		})
	}
}
