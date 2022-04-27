package http

import (
	"context"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"time"
)

type ratelimitClient struct {
	client  Client
	limiter *rate.Limiter
}

func (c *ratelimitClient) Do(req *http.Request) (*http.Response, error) {
	ctx := context.Background()
	err := c.limiter.Wait(ctx)
	if err != nil {
		log.Printf("failed to rate limit during %s %s", req.Method, req.URL.String())
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		reservation := c.limiter.ReserveN(time.Now().Add(3*time.Second), c.limiter.Burst())
		time.Sleep(reservation.Delay())
	}

	return resp, nil
}
