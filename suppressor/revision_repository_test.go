package suppressor

import (
	"freedom-sentry/mediawiki"
	"freedom-sentry/util"
	"reflect"
	"testing"
	"time"
)

func Test_revRepoImpl_GetRecentChanges(t *testing.T) {
	expectedTime := util.WithoutErr(time.Parse(time.RFC3339, "2022-04-20T12:13:14Z"))

	tests := []struct {
		name            string
		expectedPayload map[string]interface{}
		apiReturn       mediawiki.Action
		wantApiErr      bool
		want            []mediawiki.Revision
		wantErr         bool
	}{
		{
			name:       "Will return API error",
			want:       []mediawiki.Revision{},
			wantErr:    true,
			wantApiErr: true,
		},
		{
			name: "Will return no changes",
			expectedPayload: map[string]interface{}{
				"action":  "query",
				"list":    "recentchanges",
				"rcstart": "2022-04-20T12:13:14Z",
				"rcdir":   "newer",
				"rcshow":  []string{"!bot"},
				"rclimit": 5000,
				"rcprop":  []string{"title", "timestamp", "ids", "user"},
				"rctype":  []string{"edit"},
			},
			want: util.CreateNilSlice[mediawiki.Revision](),
		},
		// FIXME: Not tested
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &mockApi{
				executeThrowError: tt.wantApiErr,
			}
			rr := &revRepoImpl{
				api: api,
			}

			got, err := rr.GetRecentChanges(expectedTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRecentChanges() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if !api.executeCalled {
				t.Errorf("GetRecentChanges() must call API")
				return
			}

			if !reflect.DeepEqual(tt.expectedPayload, api.executeAction.ToActionPayload()) {
				t.Errorf("GetRecentChanges() called API = %v, wanted %v", api.executeAction.ToActionPayload(), tt.expectedPayload)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRecentChanges() got = %v, want %v", got, tt.want)
			}
		})
	}
}
