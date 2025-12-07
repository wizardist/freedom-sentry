package http

import (
	"log/slog"
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
			slog.Debug("HTTP request successful",
				"method", req.Method,
				"url", req.URL.String(),
				"status", resp.StatusCode)
			return resp, nil
		}

		if resp != nil && resp.StatusCode >= 400 {
			return resp, err
		}

		sleepTime := 2 ^ (maxAttempts - attempts)
		attempts--

		slog.Warn("HTTP request attempt failed, retrying",
			"method", req.Method,
			"url", req.URL.String(),
			"attempt", maxAttempts-attempts,
			"max_attempts", maxAttempts,
			"backoff_seconds", sleepTime,
			"error", err)
	}

	return resp, err
}
