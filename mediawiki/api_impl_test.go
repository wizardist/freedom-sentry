package mediawiki

import (
	"errors"
	"freedom-sentry/config"
	"freedom-sentry/util"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type mockClient struct {
	request    *http.Request
	response   *http.Response
	throwError bool
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	m.request = req

	var err error
	if m.throwError {
		err = errors.New("dummy error")
	}

	return m.response, err
}

type mockTokenFn struct {
	token      string
	throwError bool
}

func (t *mockTokenFn) tokenFn(Api) (string, error) {
	if t.throwError {
		return "", errors.New("dummy error")
	}

	return t.token, nil
}

type throwingReadCloser struct{}

func (m throwingReadCloser) Read([]byte) (n int, err error) {
	return 0, errors.New("dummy error")
}

func (m throwingReadCloser) Close() error {
	return nil
}

func assertEqualExecuteRequests(t *testing.T, expected, actual *http.Request) {
	actualUrl := actual.URL.String()
	expectedUrl := expected.URL.String()

	if actualUrl != expectedUrl {
		t.Errorf("request URLs don't match, got = %v, want = %v", actualUrl, expectedUrl)
	}

	actualBody, _ := ioutil.ReadAll(actual.Body)
	expectedBody, _ := ioutil.ReadAll(expected.Body)

	if !slices.Equal(actualBody, expectedBody) {
		t.Errorf("request bodies don't match, got = '%s', want = '%s'", actualBody, expectedBody)
	}

	if !maps.EqualFunc(actual.Header, expected.Header, func(v1, v2 []string) bool {
		return slices.Equal(v1, v2)
	}) {
		t.Errorf("headers don't match, got = %v, want = %v", actual.Header, expected.Header)
	}
}

type dummyAction struct {
	isWrite  bool
	payload  map[string]interface{}
	response map[string]interface{}

	throwError bool
}

func (d dummyAction) IsWriteAction() bool {
	return d.isWrite
}

func (d dummyAction) ToActionPayload() map[string]interface{} {
	return d.payload
}

func (d *dummyAction) SetResponse(json map[string]interface{}) error {
	if d.throwError {
		return errors.New("dummy error")
	}

	d.response = json

	return nil
}

const expectedToken = "expected-token"
const expectedDestination = "https://example.org/"

func Test_apiImpl_Execute(t *testing.T) {

	t.Setenv(config.EnvAccessToken, expectedToken)
	tests := []struct {
		name            string
		destination     string
		action          Action
		expectedRequest *http.Request
		response        *http.Response
		writeToken      string
		wantErr         bool
		wantClientErr   bool
		wantTokenErr    bool
	}{
		{
			name:         "Acquiring write token fails",
			action:       &dummyAction{isWrite: true},
			wantTokenErr: true,
			wantErr:      true,
		},
		{
			name:        "Invalid endpoint",
			destination: "\u007f",
			action:      &dummyAction{},
			wantErr:     true,
		},
		{
			name:          "Client fails",
			destination:   expectedDestination,
			action:        &dummyAction{},
			wantErr:       true,
			wantClientErr: true,
		},
		{
			name:        "Body read fails",
			destination: expectedDestination,
			action:      &dummyAction{},
			response:    &http.Response{Body: &throwingReadCloser{}},
			wantErr:     true,
		},
		{
			name:        "Invalid JSON",
			destination: expectedDestination,
			action:      &dummyAction{},
			response:    &http.Response{Body: io.NopCloser(strings.NewReader("{invalid json}"))},
			wantErr:     true,
		},
		{
			name:        "SetResponse fails",
			destination: expectedDestination,
			action:      &dummyAction{throwError: true},
			response:    &http.Response{Body: io.NopCloser(strings.NewReader(`{"test": 42}`))},
			wantErr:     true,
		},
		{
			name:        "Payload with booleans, slices and other",
			destination: expectedDestination,
			action: &dummyAction{
				payload: map[string]interface{}{
					"truebool":    true,
					"falsebool":   false,
					"stringslice": []string{"one", "two"},
					"ids":         []RevisionId{"42", "1337"},
					"other":       42,
					"empty":       []string{},
				},
			},
			expectedRequest: &http.Request{
				URL: util.WithoutErr(url.Parse(expectedDestination)),
				Header: map[string][]string{
					"User-Agent":    {userAgent},
					"Content-Type":  {"application/x-www-form-urlencoded"},
					"Authorization": {"Bearer " + expectedToken},
				},
				Body: createBodyForValues(map[string]string{
					"format":      "json",
					"truebool":    "1",
					"stringslice": "one|two",
					"ids":         "42|1337",
					"other":       "42",
					"empty":       "",
				}),
			},
			response: &http.Response{Body: io.NopCloser(strings.NewReader(`{"test": 42}`))},
		},
		{
			name:        "Writing action is authorized by a token",
			destination: expectedDestination,
			action: &dummyAction{
				isWrite: true,
				payload: map[string]interface{}{
					"pay": "load",
				},
			},
			writeToken: "write-token",
			expectedRequest: &http.Request{
				URL: util.WithoutErr(url.Parse(expectedDestination)),
				Header: map[string][]string{
					"User-Agent":    {userAgent},
					"Content-Type":  {"application/x-www-form-urlencoded"},
					"Authorization": {"Bearer " + expectedToken},
				},
				Body: createBodyForValues(map[string]string{
					"format": "json",
					"pay":    "load",
					"token":  "write-token",
				}),
			},
			response: &http.Response{Body: io.NopCloser(strings.NewReader(`{"test": 42}`))},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockClient{
				throwError: tt.wantClientErr,
				response:   tt.response,
			}
			tokenFn := &mockTokenFn{
				token:      tt.writeToken,
				throwError: tt.wantTokenErr,
			}
			api := NewApi(tt.destination, client, tokenFn.tokenFn)

			if err := api.Execute(tt.action); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			assertEqualExecuteRequests(t, tt.expectedRequest, client.request)
		})
	}
}

func createBodyForValues(values map[string]string) io.ReadCloser {
	urlValues := url.Values{}

	for k, v := range values {
		urlValues.Set(k, v)
	}

	return io.NopCloser(strings.NewReader(urlValues.Encode()))
}
