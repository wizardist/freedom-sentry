package eventstreams

import (
	"testing"
	"time"

	"github.com/wizardist/freedom-sentry/mediawiki"
)

func TestParseRecentChange(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    *RecentChangeEvent
		wantErr bool
	}{
		{
			name: "Valid edit event with all fields",
			input: []byte(`{
				"server_name": "en.wikipedia.org",
				"type": "edit",
				"namespace": 0,
				"title": "Test Article",
				"user": "TestUser",
				"bot": false,
				"timestamp": 1670000000,
				"revision": {
					"new": 123456,
					"old": 123455
				},
				"meta": {
					"dt": "2022-12-02T20:53:20Z"
				}
			}`),
			want: &RecentChangeEvent{
				ServerName: "en.wikipedia.org",
				Type:       "edit",
				Namespace:  0,
				Title:      "Test Article",
				User:       "TestUser",
				Bot:        false,
				Timestamp:  1670000000,
				Revision: struct {
					New int64 `json:"new"`
					Old int64 `json:"old"`
				}{
					New: 123456,
					Old: 123455,
				},
				Meta: struct {
					Dt string `json:"dt"`
				}{
					Dt: "2022-12-02T20:53:20Z",
				},
			},
			wantErr: false,
		},
		{
			name:  "Valid bot edit",
			input: []byte(`{"server_name":"en.wikipedia.org","type":"edit","namespace":0,"title":"Page","user":"BotUser","bot":true,"timestamp":1670000000,"revision":{"new":100,"old":99},"meta":{"dt":"2022-12-02T20:53:20Z"}}`),
			want: &RecentChangeEvent{
				ServerName: "en.wikipedia.org",
				Type:       "edit",
				Namespace:  0,
				Title:      "Page",
				User:       "BotUser",
				Bot:        true,
				Timestamp:  1670000000,
			},
			wantErr: false,
		},
		{
			name:  "Log event",
			input: []byte(`{"server_name":"en.wikipedia.org","type":"log","namespace":0,"title":"Page","user":"Admin","bot":false,"timestamp":1670000000,"revision":{"new":0,"old":0},"meta":{"dt":"2022-12-02T20:53:20Z"}}`),
			want: &RecentChangeEvent{
				ServerName: "en.wikipedia.org",
				Type:       "log",
				Namespace:  0,
				Title:      "Page",
				User:       "Admin",
				Bot:        false,
				Timestamp:  1670000000,
			},
			wantErr: false,
		},
		{
			name:  "New page event",
			input: []byte(`{"server_name":"en.wikipedia.org","type":"new","namespace":0,"title":"New Page","user":"Creator","bot":false,"timestamp":1670000000,"revision":{"new":1,"old":0},"meta":{"dt":"2022-12-02T20:53:20Z"}}`),
			want: &RecentChangeEvent{
				ServerName: "en.wikipedia.org",
				Type:       "new",
				Namespace:  0,
				Title:      "New Page",
				User:       "Creator",
				Bot:        false,
				Timestamp:  1670000000,
			},
			wantErr: false,
		},
		{
			name:  "Categorize event",
			input: []byte(`{"server_name":"en.wikipedia.org","type":"categorize","namespace":14,"title":"Category:Test","user":"User","bot":false,"timestamp":1670000000,"revision":{"new":0,"old":0},"meta":{"dt":"2022-12-02T20:53:20Z"}}`),
			want: &RecentChangeEvent{
				ServerName: "en.wikipedia.org",
				Type:       "categorize",
				Namespace:  14,
				Title:      "Category:Test",
				User:       "User",
				Bot:        false,
				Timestamp:  1670000000,
			},
			wantErr: false,
		},
		{
			name:  "Unicode in title and user",
			input: []byte(`{"server_name":"ja.wikipedia.org","type":"edit","namespace":0,"title":"日本語","user":"ユーザー","bot":false,"timestamp":1670000000,"revision":{"new":999,"old":998},"meta":{"dt":"2022-12-02T20:53:20Z"}}`),
			want: &RecentChangeEvent{
				ServerName: "ja.wikipedia.org",
				Type:       "edit",
				Namespace:  0,
				Title:      "日本語",
				User:       "ユーザー",
				Bot:        false,
				Timestamp:  1670000000,
			},
			wantErr: false,
		},
		{
			name:  "Large revision IDs",
			input: []byte(`{"server_name":"en.wikipedia.org","type":"edit","namespace":0,"title":"Page","user":"User","bot":false,"timestamp":1670000000,"revision":{"new":9223372036854775807,"old":9223372036854775806},"meta":{"dt":"2022-12-02T20:53:20Z"}}`),
			want: &RecentChangeEvent{
				ServerName: "en.wikipedia.org",
				Type:       "edit",
				Namespace:  0,
				Title:      "Page",
				User:       "User",
				Bot:        false,
				Timestamp:  1670000000,
			},
			wantErr: false,
		},
		{
			name:  "Zero revision IDs",
			input: []byte(`{"server_name":"en.wikipedia.org","type":"edit","namespace":0,"title":"Page","user":"User","bot":false,"timestamp":1670000000,"revision":{"new":0,"old":0},"meta":{"dt":"2022-12-02T20:53:20Z"}}`),
			want: &RecentChangeEvent{
				ServerName: "en.wikipedia.org",
				Type:       "edit",
				Namespace:  0,
				Title:      "Page",
				User:       "User",
				Bot:        false,
				Timestamp:  1670000000,
			},
			wantErr: false,
		},
		{
			name:    "Malformed JSON",
			input:   []byte(`{invalid json`),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Empty JSON",
			input:   []byte(`{}`),
			want:    &RecentChangeEvent{},
			wantErr: false,
		},
		{
			name:    "Invalid JSON type",
			input:   []byte(`"not an object"`),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRecentChange(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRecentChange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Compare basic fields
			if got.ServerName != tt.want.ServerName {
				t.Errorf("ServerName = %v, want %v", got.ServerName, tt.want.ServerName)
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Namespace != tt.want.Namespace {
				t.Errorf("Namespace = %v, want %v", got.Namespace, tt.want.Namespace)
			}
			if got.Title != tt.want.Title {
				t.Errorf("Title = %v, want %v", got.Title, tt.want.Title)
			}
			if got.User != tt.want.User {
				t.Errorf("User = %v, want %v", got.User, tt.want.User)
			}
			if got.Bot != tt.want.Bot {
				t.Errorf("Bot = %v, want %v", got.Bot, tt.want.Bot)
			}
		})
	}
}

func TestRecentChangeEvent_ToRevision(t *testing.T) {
	tests := []struct {
		name    string
		event   *RecentChangeEvent
		want    mediawiki.Revision
		wantErr bool
	}{
		{
			name: "Timestamp from Unix field",
			event: &RecentChangeEvent{
				Title:     "Test Page",
				Timestamp: 1670000000,
				Revision: struct {
					New int64 `json:"new"`
					Old int64 `json:"old"`
				}{New: 123},
			},
			want: mediawiki.Revision{
				Id:           mediawiki.RevisionId("123"),
				IsSuppressed: false,
				Title:        "Test Page",
				Content:      "",
				Timestamp:    time.Unix(1670000000, 0).UTC(),
			},
			wantErr: false,
		},
		{
			name: "Timestamp from meta.dt field",
			event: &RecentChangeEvent{
				Title:     "Test Page",
				Timestamp: 0,
				Revision: struct {
					New int64 `json:"new"`
					Old int64 `json:"old"`
				}{New: 456},
				Meta: struct {
					Dt string `json:"dt"`
				}{Dt: "2022-12-02T20:53:20Z"},
			},
			want: mediawiki.Revision{
				Id:           mediawiki.RevisionId("456"),
				IsSuppressed: false,
				Title:        "Test Page",
				Content:      "",
				Timestamp:    time.Date(2022, 12, 2, 20, 53, 20, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "Both timestamps present - Unix takes precedence",
			event: &RecentChangeEvent{
				Title:     "Test Page",
				Timestamp: 1670000000,
				Revision: struct {
					New int64 `json:"new"`
					Old int64 `json:"old"`
				}{New: 789},
				Meta: struct {
					Dt string `json:"dt"`
				}{Dt: "2099-01-01T00:00:00Z"},
			},
			want: mediawiki.Revision{
				Id:           mediawiki.RevisionId("789"),
				IsSuppressed: false,
				Title:        "Test Page",
				Content:      "",
				Timestamp:    time.Unix(1670000000, 0).UTC(),
			},
			wantErr: false,
		},
		{
			name: "Invalid meta.dt timestamp",
			event: &RecentChangeEvent{
				Title:     "Test Page",
				Timestamp: 0,
				Revision: struct {
					New int64 `json:"new"`
					Old int64 `json:"old"`
				}{New: 999},
				Meta: struct {
					Dt string `json:"dt"`
				}{Dt: "invalid-timestamp"},
			},
			wantErr: true,
		},
		{
			name: "Missing both timestamps",
			event: &RecentChangeEvent{
				Title:     "Test Page",
				Timestamp: 0,
				Revision: struct {
					New int64 `json:"new"`
					Old int64 `json:"old"`
				}{New: 111},
			},
			want: mediawiki.Revision{
				Id:           mediawiki.RevisionId("111"),
				IsSuppressed: false,
				Title:        "Test Page",
				Content:      "",
				Timestamp:    time.Time{}, // Zero time
			},
			wantErr: false,
		},
		{
			name: "IsSuppressed always false",
			event: &RecentChangeEvent{
				Title:     "Suppressed Page",
				Timestamp: 1670000000,
				Revision: struct {
					New int64 `json:"new"`
					Old int64 `json:"old"`
				}{New: 222},
			},
			want: mediawiki.Revision{
				Id:           mediawiki.RevisionId("222"),
				IsSuppressed: false,
				Title:        "Suppressed Page",
				Content:      "",
				Timestamp:    time.Unix(1670000000, 0).UTC(),
			},
			wantErr: false,
		},
		{
			name: "Content always empty",
			event: &RecentChangeEvent{
				Title:     "Test Page",
				Timestamp: 1670000000,
				Revision: struct {
					New int64 `json:"new"`
					Old int64 `json:"old"`
				}{New: 333},
			},
			want: mediawiki.Revision{
				Id:           mediawiki.RevisionId("333"),
				IsSuppressed: false,
				Title:        "Test Page",
				Content:      "",
				Timestamp:    time.Unix(1670000000, 0).UTC(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.event.ToRevision()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToRevision() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if string(got.Id) != string(tt.want.Id) {
				t.Errorf("Id = %s, want %s", got.Id, tt.want.Id)
			}
			if got.IsSuppressed != tt.want.IsSuppressed {
				t.Errorf("IsSuppressed = %v, want %v", got.IsSuppressed, tt.want.IsSuppressed)
			}
			if got.Title != tt.want.Title {
				t.Errorf("Title = %v, want %v", got.Title, tt.want.Title)
			}
			if got.Content != tt.want.Content {
				t.Errorf("Content = %v, want %v", got.Content, tt.want.Content)
			}
			if !got.Timestamp.Equal(tt.want.Timestamp) {
				t.Errorf("Timestamp = %v, want %v", got.Timestamp, tt.want.Timestamp)
			}
		})
	}
}

func TestRecentChangeEvent_ShouldProcess(t *testing.T) {
	tests := []struct {
		name  string
		event *RecentChangeEvent
		want  bool
	}{
		{
			name: "Type=edit, Bot=false - should process",
			event: &RecentChangeEvent{
				Type: "edit",
				Bot:  false,
			},
			want: true,
		},
		{
			name: "Type=log - should not process",
			event: &RecentChangeEvent{
				Type: "log",
				Bot:  false,
			},
			want: false,
		},
		{
			name: "Type=new - should not process",
			event: &RecentChangeEvent{
				Type: "new",
				Bot:  false,
			},
			want: false,
		},
		{
			name: "Type=categorize - should not process",
			event: &RecentChangeEvent{
				Type: "categorize",
				Bot:  false,
			},
			want: false,
		},
		{
			name: "Bot=true - should not process",
			event: &RecentChangeEvent{
				Type: "edit",
				Bot:  true,
			},
			want: false,
		},
		{
			name: "Type=edit, Bot=true - should not process",
			event: &RecentChangeEvent{
				Type: "edit",
				Bot:  true,
			},
			want: false,
		},
		{
			name: "Missing type field",
			event: &RecentChangeEvent{
				Type: "",
				Bot:  false,
			},
			want: false,
		},
		{
			name: "Unknown type",
			event: &RecentChangeEvent{
				Type: "unknown",
				Bot:  false,
			},
			want: false,
		},
		{
			name: "Missing bot field defaults to false",
			event: &RecentChangeEvent{
				Type: "edit",
				Bot:  false, // Default value
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.event.ShouldProcess(); got != tt.want {
				t.Errorf("ShouldProcess() = %v, want %v", got, tt.want)
			}
		})
	}
}
