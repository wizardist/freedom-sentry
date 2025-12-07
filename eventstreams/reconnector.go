package eventstreams

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"time"
)

// ExponentialBackoff configuration for reconnection strategy
type ExponentialBackoff struct {
	Initial time.Duration // Initial backoff duration
	Max     time.Duration // Maximum backoff duration
	Factor  float64       // Exponential factor (typically 2.0)
	Jitter  float64       // Jitter factor (0.0-1.0, typically 0.25 for ±25%)
}

// Reconnector manages connection resilience with automatic reconnection
type Reconnector struct {
	client  *Client
	backoff *ExponentialBackoff
}

// NewReconnector creates a new reconnector with the given client and backoff configuration
func NewReconnector(client *Client, backoff *ExponentialBackoff) *Reconnector {
	if backoff == nil {
		// Default backoff configuration
		backoff = &ExponentialBackoff{
			Initial: 1 * time.Second,
			Max:     60 * time.Second,
			Factor:  2.0,
			Jitter:  0.25,
		}
	}

	return &Reconnector{
		client:  client,
		backoff: backoff,
	}
}

// Run maintains connection, auto-reconnecting on failures
// The eventHandler is called for each received event
// Returns error only if context is cancelled
func (r *Reconnector) Run(ctx context.Context, since *time.Time, eventHandler func(Event) error) error {
	attempt := 0

	for {
		select {
		case <-ctx.Done():
			slog.Info("EventStreams shutting down", "reason", "context_cancelled")
			return ctx.Err()
		default:
		}

		// Try to connect
		stream, err := r.client.Connect(ctx, since)
		if err != nil {
			backoff := r.calculateBackoff(attempt)
			slog.Error("EventStreams connection failed", "error", err, "attempt", attempt+1, "backoff", backoff)

			// Wait before retrying
			select {
			case <-time.After(backoff):
				attempt++
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Connection successful, reset attempt counter
		attempt = 0
		slog.Info("EventStreams connected successfully")

		// After first successful connection, don't use 'since' parameter on reconnect
		// Instead, rely on checkpoint-based resume
		since = nil

		// Process events from the stream
		err = r.processStream(ctx, stream, eventHandler)
		stream.Close()

		if err != nil {
			if ctx.Err() != nil {
				// Context was cancelled, exit gracefully
				return ctx.Err()
			}

			// Connection lost, will reconnect
			slog.Warn("EventStreams connection lost, reconnecting", "error", err)
		}
	}
}

// processStream handles events from an active stream
func (r *Reconnector) processStream(ctx context.Context, stream *EventStream, eventHandler func(Event) error) error {
	for {
		select {
		case event, ok := <-stream.Events():
			if !ok {
				// Channel closed, stream ended
				return fmt.Errorf("event channel closed")
			}

			if err := eventHandler(event); err != nil {
				slog.Error("event handler failed", "error", err)
				// Continue processing despite handler errors
			}

		case err, ok := <-stream.Errors():
			if !ok {
				// Error channel closed
				return fmt.Errorf("error channel closed")
			}
			return fmt.Errorf("stream error: %w", err)

		case <-stream.Done():
			return fmt.Errorf("stream finished unexpectedly")

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// calculateBackoff returns the next backoff duration with jitter
func (r *Reconnector) calculateBackoff(attempt int) time.Duration {
	// Calculate base backoff: Initial * (Factor ^ attempt)
	backoff := float64(r.backoff.Initial) * math.Pow(r.backoff.Factor, float64(attempt))

	// Cap at maximum
	if backoff > float64(r.backoff.Max) {
		backoff = float64(r.backoff.Max)
	}

	// Apply jitter: backoff * (1 + random(-Jitter, +Jitter))
	jitterRange := backoff * r.backoff.Jitter
	jitter := (rand.Float64()*2 - 1) * jitterRange // Random value in [-jitterRange, +jitterRange]
	backoff += jitter

	// Ensure non-negative
	if backoff < 0 {
		backoff = 0
	}

	return time.Duration(backoff)
}
