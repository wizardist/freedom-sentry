package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type ratelimitClient struct {
	client  Client
	limiter *rate.Limiter
}

func (c *ratelimitClient) Do(req *http.Request) (*http.Response, error) {
	ctx := context.Background()
	err := c.limiter.Wait(ctx)
	if err != nil {
		slog.Error("rate limiter wait failed",
			"method", req.Method,
			"url", req.URL.String(),
			"error", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp != nil && (resp.StatusCode == http.StatusTooManyRequests || isRateLimited(resp)) {
		slog.Warn("rate limited by server, backing off",
			"method", req.Method,
			"url", req.URL.String(),
			"status", resp.StatusCode,
			"backoff_seconds", 3)
		reservation := c.limiter.ReserveN(time.Now().Add(3*time.Second), c.limiter.Burst())
		time.Sleep(reservation.Delay())
	}

	slog.Debug("HTTP request executed",
		"method", req.Method,
		"url", req.URL.String(),
		"status", resp.StatusCode)

	return resp, nil
}

func isRateLimited(resp *http.Response) bool {
	errorHeaderValue := resp.Header.Get("MediaWiki-API-Error")

	return errorHeaderValue == "ratelimited"
}
