package eventstreams

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Checkpoint represents the last processed event position in the Kafka stream
type Checkpoint struct {
	Topic     string `json:"topic"`
	Partition int    `json:"partition"`
	Offset    int64  `json:"offset,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// Event represents a parsed SSE event
type Event struct {
	ID   string          // Raw event ID (JSON array)
	Data json.RawMessage // Unparsed JSON data
}

// Client manages SSE connection to WikiMedia EventStreams
type Client struct {
	baseURL    string
	httpClient *http.Client
	checkpoint *Checkpoint
	mu         sync.RWMutex
}

// EventStream represents an active SSE connection
type EventStream struct {
	events chan Event
	errors chan error
	done   chan struct{}
	cancel context.CancelFunc
}

// NewClient creates a new EventStreams client
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 0, // No timeout for long-lived SSE connections
		}
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// Connect establishes an SSE connection with optional since parameter for lookback
func (c *Client) Connect(ctx context.Context, since *time.Time) (*EventStream, error) {
	url := c.baseURL
	if since != nil {
		// EventStreams expects ISO8601 format
		url = fmt.Sprintf("%s?since=%s", url, since.UTC().Format(time.RFC3339))
	}

	slog.Debug("establishing EventStreams connection", "url", url, "since", since)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	streamCtx, cancel := context.WithCancel(ctx)
	stream := &EventStream{
		events: make(chan Event, 10),
		errors: make(chan error, 1),
		done:   make(chan struct{}),
		cancel: cancel,
	}

	go c.readEvents(streamCtx, resp.Body, stream)

	slog.Info("EventStreams connected", "url", url)
	return stream, nil
}

// GetCheckpoint returns the last stored checkpoint
func (c *Client) GetCheckpoint() *Checkpoint {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.checkpoint == nil {
		return nil
	}
	// Return a copy to prevent external modification
	cp := *c.checkpoint
	return &cp
}

// readEvents reads SSE events from the response body
func (c *Client) readEvents(ctx context.Context, body io.ReadCloser, stream *EventStream) {
	defer func() {
		body.Close()
		close(stream.events)
		close(stream.errors)
		close(stream.done)
	}()

	scanner := bufio.NewScanner(body)
	var eventID string
	var dataLines []string

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				stream.errors <- fmt.Errorf("scanner error: %w", err)
			}
			return
		}

		line := scanner.Text()

		// Empty line signals end of event
		if line == "" {
			if len(dataLines) > 0 {
				data := strings.Join(dataLines, "\n")
				event := Event{
					ID:   eventID,
					Data: json.RawMessage(data),
				}

				slog.Debug("EventStreams event received",
					"event_id", eventID,
					"data_size", len(data))

				// Update checkpoint after receiving event (before sending to channel)
				if eventID != "" {
					c.updateCheckpoint(eventID)
				}

				select {
				case stream.events <- event:
				case <-ctx.Done():
					return
				}

				// Reset for next event
				eventID = ""
				dataLines = nil
			}
			continue
		}

		// Parse SSE fields
		if strings.HasPrefix(line, "id:") {
			eventID = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
		} else if strings.HasPrefix(line, "data:") {
			dataLine := strings.TrimPrefix(line, "data:")
			// EventStreams uses 'data: ' with a space
			if strings.HasPrefix(dataLine, " ") {
				dataLine = dataLine[1:]
			}
			dataLines = append(dataLines, dataLine)
		}
		// Ignore other SSE fields like "event:", "retry:", comments (":...")
	}
}

// updateCheckpoint stores the checkpoint from event ID
func (c *Client) updateCheckpoint(eventID string) {
	// EventStreams event IDs are JSON arrays: [{"topic":"...","partition":0,"offset":123}]
	// We'll parse and store the first element
	var checkpoints []Checkpoint
	if err := json.Unmarshal([]byte(eventID), &checkpoints); err != nil {
		slog.Error("failed to parse checkpoint from event ID", "error", err, "id", eventID)
		return
	}

	if len(checkpoints) == 0 {
		return
	}

	c.mu.Lock()
	c.checkpoint = &checkpoints[0]
	c.mu.Unlock()

	slog.Debug("checkpoint updated", "topic", checkpoints[0].Topic, "partition", checkpoints[0].Partition, "offset", checkpoints[0].Offset)
}

// Events returns the read-only channel of incoming events
func (s *EventStream) Events() <-chan Event {
	return s.events
}

// Errors returns the read-only channel of connection errors
func (s *EventStream) Errors() <-chan error {
	return s.errors
}

// Done returns a channel that's closed when the stream is finished
func (s *EventStream) Done() <-chan struct{} {
	return s.done
}

// Close terminates the SSE connection
func (s *EventStream) Close() {
	slog.Debug("closing EventStreams connection")
	s.cancel()
}
