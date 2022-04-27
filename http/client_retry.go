package http

import (
	"log"
	"net/http"
)

type retryClient struct {
	client Client
}

func (c *retryClient) Do(req *http.Request) (*http.Response, error) {
	const maxAttempts = 2

	attempts := maxAttempts

	var resp *http.Response
	var err error

	for attempts > 0 {
		resp, err = c.client.Do(req)
		if err == nil {
			return resp, nil
		}

		if resp != nil && resp.StatusCode >= 400 {
			return resp, err
		}

		sleepTime := 2 ^ (maxAttempts - attempts)
		attempts--

		log.Printf("failed attempt %d to request %s %s, waiting %d s", maxAttempts-attempts, req.Method, req.URL, sleepTime)
	}

	return resp, err
}
