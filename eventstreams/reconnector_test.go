package eventstreams

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewReconnector(t *testing.T) {
	tests := []struct {
		name        string
		client      *Client
		backoff     *ExponentialBackoff
		wantDefault bool
	}{
		{
			name:   "With provided backoff config",
			client: NewClient("https://example.org/stream", nil),
			backoff: &ExponentialBackoff{
				Initial: 2 * time.Second,
				Max:     120 * time.Second,
				Factor:  3.0,
				Jitter:  0.5,
			},
			wantDefault: false,
		},
		{
			name:        "With nil backoff - should use default",
			client:      NewClient("https://example.org/stream", nil),
			backoff:     nil,
			wantDefault: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReconnector(tt.client, tt.backoff)

			if r == nil {
				t.Fatal("NewReconnector returned nil")
			}

			if r.client != tt.client {
				t.Error("Client not set correctly")
			}

			if r.backoff == nil {
				t.Fatal("Backoff config is nil")
			}

			if tt.wantDefault {
				if r.backoff.Initial != 1*time.Second {
					t.Errorf("Default Initial = %v, want 1s", r.backoff.Initial)
				}
				if r.backoff.Max != 60*time.Second {
					t.Errorf("Default Max = %v, want 60s", r.backoff.Max)
				}
				if r.backoff.Factor != 2.0 {
					t.Errorf("Default Factor = %v, want 2.0", r.backoff.Factor)
				}
				if r.backoff.Jitter != 0.25 {
					t.Errorf("Default Jitter = %v, want 0.25", r.backoff.Jitter)
				}
			}
		})
	}
}

func TestReconnector_calculateBackoff(t *testing.T) {
	tests := []struct {
		name     string
		backoff  *ExponentialBackoff
		attempt  int
		wantMin  time.Duration
		wantMax  time.Duration
		checkCap bool
	}{
		{
			name: "Attempt 0 - initial backoff",
			backoff: &ExponentialBackoff{
				Initial: 1 * time.Second,
				Max:     60 * time.Second,
				Factor:  2.0,
				Jitter:  0.25,
			},
			attempt: 0,
			wantMin: 750 * time.Millisecond,  // 1s - 25%
			wantMax: 1250 * time.Millisecond, // 1s + 25%
		},
		{
			name: "Attempt 1 - exponential growth",
			backoff: &ExponentialBackoff{
				Initial: 1 * time.Second,
				Max:     60 * time.Second,
				Factor:  2.0,
				Jitter:  0.25,
			},
			attempt: 1,
			wantMin: 1500 * time.Millisecond, // 2s - 25%
			wantMax: 2500 * time.Millisecond, // 2s + 25%
		},
		{
			name: "Attempt 2 - further exponential growth",
			backoff: &ExponentialBackoff{
				Initial: 1 * time.Second,
				Max:     60 * time.Second,
				Factor:  2.0,
				Jitter:  0.25,
			},
			attempt: 2,
			wantMin: 3000 * time.Millisecond, // 4s - 25%
			wantMax: 5000 * time.Millisecond, // 4s + 25%
		},
		{
			name: "Large attempt - capped at max",
			backoff: &ExponentialBackoff{
				Initial: 1 * time.Second,
				Max:     10 * time.Second,
				Factor:  2.0,
				Jitter:  0.25,
			},
			attempt:  10, // Would be 1024s without cap
			wantMin:  7500 * time.Millisecond,  // 10s - 25%
			wantMax:  12500 * time.Millisecond, // 10s + 25%
			checkCap: true,
		},
		{
			name: "Zero jitter",
			backoff: &ExponentialBackoff{
				Initial: 5 * time.Second,
				Max:     60 * time.Second,
				Factor:  2.0,
				Jitter:  0.0,
			},
			attempt: 0,
			wantMin: 5 * time.Second,
			wantMax: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient("https://example.org/stream", nil)
			r := NewReconnector(client, tt.backoff)

			// Run multiple times to check jitter variation
			for i := 0; i < 5; i++ {
				backoff := r.calculateBackoff(tt.attempt)

				if backoff < 0 {
					t.Errorf("Backoff is negative: %v", backoff)
				}

				if backoff < tt.wantMin {
					t.Errorf("Backoff %v is less than minimum %v", backoff, tt.wantMin)
				}

				if backoff > tt.wantMax {
					t.Errorf("Backoff %v is greater than maximum %v", backoff, tt.wantMax)
				}

				if tt.checkCap && backoff > tt.backoff.Max*125/100 {
					t.Errorf("Backoff %v exceeds cap with jitter %v", backoff, tt.backoff.Max*125/100)
				}
			}
		})
	}
}

func TestReconnector_Run_ContextCancellation(t *testing.T) {
	client := NewClient("https://example.org/stream", nil)
	r := NewReconnector(client, nil)

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately
	cancel()

	handler := func(event Event) error {
		return nil
	}

	err := r.Run(ctx, nil, handler)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

func TestReconnector_Run_HandlerErrors(t *testing.T) {
	// Create a test server that sends one event then closes
	server := newTestSSEServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		flusher := w.(http.Flusher)
		fmt.Fprint(w, "data: {\"test\":1}\n\n")
		flusher.Flush()

		time.Sleep(50 * time.Millisecond)
	})
	defer server.Close()

	client := NewClient(server.URL, nil)
	r := NewReconnector(client, &ExponentialBackoff{
		Initial: 10 * time.Millisecond,
		Max:     100 * time.Millisecond,
		Factor:  2.0,
		Jitter:  0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	handlerCallCount := 0
	handler := func(event Event) error {
		handlerCallCount++
		// Return error - should be logged but processing continues
		return errors.New("handler error")
	}

	r.Run(ctx, nil, handler)

	// Handler should have been called at least once
	if handlerCallCount == 0 {
		t.Error("Handler was never called")
	}
}

// Helper function to create a test SSE server
func newTestSSEServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func TestReconnector_calculateBackoff_NonNegative(t *testing.T) {
	// Test with extreme jitter to ensure non-negative even in edge cases
	backoff := &ExponentialBackoff{
		Initial: 100 * time.Millisecond,
		Max:     1 * time.Second,
		Factor:  2.0,
		Jitter:  1.0, // 100% jitter
	}

	client := NewClient("https://example.org/stream", nil)
	r := NewReconnector(client, backoff)

	// Run many iterations to catch edge cases
	for attempt := 0; attempt < 10; attempt++ {
		for i := 0; i < 100; i++ {
			result := r.calculateBackoff(attempt)
			if result < 0 {
				t.Errorf("Attempt %d iteration %d: got negative backoff %v", attempt, i, result)
			}
		}
	}
}
