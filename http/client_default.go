package http

import "net/http"

type defaultClient struct {
	client *http.Client
}

func (c *defaultClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}
