package mediawiki

import (
	"encoding/json"
	"fmt"
	"freedom-sentry/config"
	"freedom-sentry/http"
	"freedom-sentry/util"
	"io/ioutil"
	"log"
	gohttp "net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
)

const userAgent = "FreedomSentry/1"
const writeTokenKey = "token"

type Token string
type TokenRequestFn func(api Api) (Token, error)

type apiImpl struct {
	httpClient http.Client
	endpoint   string
	tokenFn    TokenRequestFn
}

func NewApi(endpoint string, client http.Client, tokenFn TokenRequestFn) Api {
	return &apiImpl{
		httpClient: client,
		endpoint:   endpoint,
		tokenFn:    tokenFn,
	}
}

func (api *apiImpl) Execute(action Action) error {
	payload := action.ToActionPayload()

	log.Println("executing action", payload["action"])

	if action.IsWriteAction() {
		log.Println("injecting token for a write action")
		err := api.injectToken(payload)
		if err != nil {
			return err
		}
	}

	request, err := api.createRequest(payload)
	if err != nil {
		return err
	}

	resp, err := api.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer util.Close(resp.Body)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var respJson map[string]interface{}
	err = json.Unmarshal(body, &respJson)
	if err != nil {
		return err
	}

	err = action.SetResponse(respJson)

	return err
}

func (api *apiImpl) createRequest(payload map[string]interface{}) (*gohttp.Request, error) {
	data := payloadToUrlValues(payload)
	data.Set("format", "json")

	request, err := gohttp.NewRequest(gohttp.MethodPost, api.endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", "Bearer "+os.Getenv(config.EnvAccessToken))

	return request, nil
}

func (api *apiImpl) injectToken(payload map[string]interface{}) error {
	token, err := api.tokenFn(api)
	if err != nil {
		return err
	}

	payload[writeTokenKey] = token

	return nil
}

func payloadToUrlValues(payload map[string]interface{}) url.Values {
	data := url.Values{}

	for k, v := range payload {
		switch v.(type) {
		case bool:
			if v.(bool) {
				data.Set(k, "1")
			}
		default:
			rt := reflect.TypeOf(v)
			if rt.Kind() == reflect.Slice {
				data.Set(k, joinInterfaceSlice(v, listSeparator))
			} else {
				data.Set(k, fmt.Sprint(v))
			}
		}
	}

	return data
}

func joinInterfaceSlice(slice interface{}, separator string) string {
	rv := reflect.ValueOf(slice)

	rvLen := rv.Len()

	if rvLen == 0 {
		return ""
	}

	values := make([]string, rvLen)
	strlen := 0

	for i := 0; i < rv.Len(); i++ {
		iv := rv.Index(i)
		v := iv.String()
		values[i] = v
		strlen += len(v)
	}

	builder := strings.Builder{}
	builder.Grow(strlen + rvLen - 1)
	builder.WriteString(values[0])

	for _, v := range values[1:] {
		builder.WriteString(separator)
		builder.WriteString(v)
	}

	return builder.String()
}
