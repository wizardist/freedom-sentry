package query

import (
	"errors"
	"freedom-sentry/mediawiki"
	"reflect"
	"testing"
)

func TestQuery_ToActionPayload(t *testing.T) {
	tests := []struct {
		name  string
		query Query
		want  map[string]interface{}
	}{
		{
			name: "Redirect ON",
			query: Query{
				FollowRedirects: true,
			},
			want: map[string]interface{}{
				"action":    "query",
				"redirects": true,
			},
		},
		{
			name: "prop=revisions",
			query: Query{
				Properties: []Property{
					&RevisionsQueryProperty{
						Properties: []string{"ids", "timestamp", "user"},
						Limit:      5000,
					},
				},
				PageNames: []string{"Test 1", "Test 2"},
			},
			want: map[string]interface{}{
				"action":  "query",
				"prop":    "revisions",
				"titles":  []string{"Test 1", "Test 2"}, // Revisions don't support multiple pages in fact
				"rvprop":  []string{"ids", "timestamp", "user"},
				"rvlimit": 5000,
			},
		},
		{
			name: "meta=tokens",
			query: Query{
				Meta: []Meta{
					&TokensQueryMeta{
						Type: []string{"csrf"},
					},
				},
			},
			want: map[string]interface{}{
				"action": "query",
				"meta":   "tokens",
				"type":   []string{"csrf"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.query.ToActionPayload(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToActionPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockProperty struct {
	throwError  bool
	responseSet bool
}

func (qp mockProperty) ToPropertyPayload() map[string]interface{} { return nil }
func (qp *mockProperty) setResponse(map[string]interface{}) error {
	if qp.throwError {
		return errors.New("dummy error")
	}

	qp.responseSet = true
	return nil
}

type mockMeta struct {
	throwError  bool
	responseSet bool
}

func (qm mockMeta) ToMetaPayload() map[string]interface{} { return nil }
func (qm *mockMeta) setResponse(map[string]interface{}) error {
	if qm.throwError {
		return errors.New("dummy error")
	}

	qm.responseSet = true
	return nil
}

func TestQuery_SetResponse(t *testing.T) {
	tests := []struct {
		name         string
		props        []Property
		meta         []Meta
		payload      map[string]interface{}
		wantPropsSet bool
		wantMetaSet  bool
		wantErr      bool
	}{
		{
			name: "Empty",
			props: []Property{
				&mockProperty{}, &mockProperty{},
			},
			payload: map[string]interface{}{
				"batchcomplete": "",
			},
		},
		{
			name: "With property error",
			props: []Property{
				&mockProperty{throwError: true},
			},
			payload: map[string]interface{}{
				"query": map[string]interface{}{
					"pages": map[string]interface{}{
						"42": map[string]interface{}{
							"pageid": 42,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "With meta error",
			meta: []Meta{
				&mockMeta{throwError: true},
			},
			payload: map[string]interface{}{
				"batchcomplete": "",
				"query": map[string]interface{}{
					"tokens": map[string]interface{}{
						"csrftoken": "token",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Two properties with data",
			props: []Property{
				&mockProperty{}, &mockProperty{},
			},
			payload: map[string]interface{}{
				"batchcomplete": "",
				"query": map[string]interface{}{
					"pages": map[string]interface{}{
						"42": map[string]interface{}{
							"pageid": 42,
						},
					},
				},
			},
			wantPropsSet: true,
		},
		{
			name: "Meta",
			meta: []Meta{
				&mockMeta{}, &mockMeta{},
			},
			payload: map[string]interface{}{
				"batchcomplete": "",
				"query": map[string]interface{}{
					"tokens": map[string]interface{}{
						"csrftoken": "token",
					},
				},
			},
			wantMetaSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Query{
				Properties: tt.props,
				Meta:       tt.meta,
			}

			if err := a.SetResponse(tt.payload); (err != nil) != tt.wantErr {
				t.Errorf("SetResponse() error = %v, wantErr %v", err, tt.wantErr)
			}

			for _, p := range a.Properties {
				if p.(*mockProperty).responseSet != tt.wantPropsSet {
					t.Errorf("Property.responseSet = %v but wanted %v", p.(*mockProperty).responseSet, tt.wantPropsSet)
				}
			}
			for _, m := range a.Meta {
				if m.(*mockMeta).responseSet != tt.wantMetaSet {
					t.Errorf("Meta.responseSet = %v but wanted %v", m.(*mockMeta).responseSet, tt.wantMetaSet)
				}
			}
		})
	}
}

func TestRevisionsQueryProperty_GetRevisions(t *testing.T) {
	tests := []struct {
		name    string
		payload map[string]interface{}
		want    []mediawiki.Revision
		wantErr bool
	}{
		{
			name:    "Empty payload",
			payload: map[string]interface{}{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "No pages object",
			payload: map[string]interface{}{
				"something": nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "pages is not an object",
			payload: map[string]interface{}{
				"pages": "not an object",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "No revisions object in page",
			payload: map[string]interface{}{
				"pages": map[string]interface{}{
					"42": map[string]interface{}{},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Revisions is not a collection",
			payload: map[string]interface{}{
				"pages": map[string]interface{}{
					"42": map[string]interface{}{
						"revisions": "not an object",
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Revision without revid",
			payload: map[string]interface{}{
				"pages": map[string]interface{}{
					"42": map[string]interface{}{
						"revisions": []interface{}{
							map[string]interface{}{
								// no "revid"
							},
						},
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "List of revisions for one page",
			payload: map[string]interface{}{
				"pages": map[string]interface{}{
					"42": map[string]interface{}{
						"pageid": 42,
						"ns":     0,
						"title":  "Dummy Title",
						"revisions": []interface{}{ // Comes like that from json.Unmarshal()
							map[string]interface{}{
								"revid":     float64(1337), // Comes like that from json.Unmarshal()
								"parentid":  73,
								"timestamp": "2022-04-20T12:13:14Z",
							},
							map[string]interface{}{
								"revid":      float64(73),
								"parentid":   0,
								"timestamp":  "2022-04-20T12:13:14Z",
								"userhidden": "",
								"suppressed": "",
							},
						},
					},
				},
			},
			want: []mediawiki.Revision{
				{
					Id:           "1337",
					IsSuppressed: false,
				},
				{
					Id:           "73",
					IsSuppressed: true,
				},
			},
			wantErr: false,
		},
		{
			name: "Revision with content",
			payload: map[string]interface{}{
				"pages": map[string]interface{}{
					"42": map[string]interface{}{
						"pageid": 42,
						"ns":     0,
						"title":  "Dummy Title",
						"revisions": []interface{}{ // Comes like that from json.Unmarshal()
							map[string]interface{}{
								"revid": float64(1337), // Comes like that from json.Unmarshal()
								"*":     "page contents",
							},
						},
					},
				},
			},
			want: []mediawiki.Revision{
				{
					Id:      "1337",
					Content: "page contents",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qp := RevisionsQueryProperty{}
			err := qp.setResponse(tt.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("setResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got := qp.GetRevisions(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRevisions() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}

func TestTokensQueryMeta_GetTokens(t *testing.T) {
	type fields struct {
		Type   []string
		tokens struct {
			Csrf string
		}
	}
	tests := []struct {
		name    string
		types   []string
		payload map[string]interface{}
		want    struct{ Csrf string }
		wantErr bool
	}{
		{
			name:  "CSRF",
			types: []string{"csrf"},
			payload: map[string]interface{}{
				"tokens": map[string]interface{}{
					"csrftoken": interface{}("tokenvalue"),
				},
			},
			want: struct {
				Csrf string
			}{
				Csrf: "tokenvalue",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qm := TokensQueryMeta{
				Type: tt.types,
			}
			err := qm.setResponse(tt.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("setResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got := qm.GetTokens(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTokens() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuery_IsWriteAction(t *testing.T) {
	t.Run("Must be false", func(t *testing.T) {
		if (Query{}).IsWriteAction() {
			t.Errorf("must be false")
		}
	})
}
