package eventstreams

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name       string
		baseURL    string
		httpClient *http.Client
		wantNil    bool
	}{
		{
			name:       "With provided HTTP client",
			baseURL:    "https://example.org/stream",
			httpClient: &http.Client{Timeout: 30 * time.Second},
			wantNil:    false,
		},
		{
			name:       "With nil HTTP client - should use default",
			baseURL:    "https://example.org/stream",
			httpClient: nil,
			wantNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.baseURL, tt.httpClient)
			if (client == nil) != tt.wantNil {
				t.Errorf("NewClient() = nil: %v, want nil: %v", client == nil, tt.wantNil)
			}
			if client != nil && client.baseURL != tt.baseURL {
				t.Errorf("baseURL = %v, want %v", client.baseURL, tt.baseURL)
			}
			if client != nil && client.httpClient == nil {
				t.Error("httpClient should not be nil")
			}
		})
	}
}

func TestClient_Connect(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		since       *time.Time
		wantErr     bool
		checkURL    bool
		expectedURL string
	}{
		{
			name: "Successful connection without since",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				w.WriteHeader(http.StatusOK)
			},
			since:   nil,
			wantErr: false,
		},
		{
			name: "Successful connection with since parameter",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Verify since parameter in URL
				since := r.URL.Query().Get("since")
				if since == "" {
					t.Error("Expected 'since' query parameter")
				}
				w.Header().Set("Content-Type", "text/event-stream")
				w.WriteHeader(http.StatusOK)
			},
			since:    func() *time.Time { t := time.Date(2022, 12, 1, 10, 0, 0, 0, time.UTC); return &t }(),
			wantErr:  false,
			checkURL: true,
		},
		{
			name: "Non-200 status code",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			since:   nil,
			wantErr: true,
		},
		{
			name: "404 Not Found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			since:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewClient(server.URL, nil)
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			stream, err := client.Connect(ctx, tt.since)
			if (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && stream != nil {
				stream.Close()
			}
		})
	}
}

func TestClient_Connect_Headers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if accept := r.Header.Get("Accept"); accept != "text/event-stream" {
			t.Errorf("Accept header = %v, want text/event-stream", accept)
		}
		if cacheControl := r.Header.Get("Cache-Control"); cacheControl != "no-cache" {
			t.Errorf("Cache-Control header = %v, want no-cache", cacheControl)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stream, err := client.Connect(ctx, nil)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer stream.Close()
}

func TestClient_Connect_ContextCancellation(t *testing.T) {
	// Create a server that never responds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	_, err := client.Connect(ctx, nil)
	if err == nil {
		t.Error("Expected error when context is cancelled")
	}
}

func TestClient_ReadEvents(t *testing.T) {
	tests := []struct {
		name      string
		sseData   string
		wantCount int
		checkData bool
	}{
		{
			name: "Single event with ID and data",
			sseData: "id: 123\n" +
				"data: {\"test\":\"value\"}\n" +
				"\n",
			wantCount: 1,
			checkData: true,
		},
		{
			name: "Multi-line data field",
			sseData: "id: 456\n" +
				"data: {\"line1\":\"value1\",\n" +
				"data: \"line2\":\"value2\"}\n" +
				"\n",
			wantCount: 1,
		},
		{
			name: "Event with just data, no ID",
			sseData: "data: {\"noId\":true}\n" +
				"\n",
			wantCount: 1,
		},
		{
			name: "Event with just ID, no data",
			sseData: "id: 789\n" +
				"\n",
			wantCount: 0, // No data, event not sent
		},
		{
			name: "Multiple events in sequence",
			sseData: "id: 1\n" +
				"data: {\"event\":1}\n" +
				"\n" +
				"id: 2\n" +
				"data: {\"event\":2}\n" +
				"\n" +
				"id: 3\n" +
				"data: {\"event\":3}\n" +
				"\n",
			wantCount: 3,
		},
		{
			name: "Comment lines (should be ignored)",
			sseData: ": this is a comment\n" +
				"id: 100\n" +
				"data: {\"test\":\"data\"}\n" +
				": another comment\n" +
				"\n",
			wantCount: 1,
		},
		{
			name: "Unknown fields (should be ignored)",
			sseData: "id: 200\n" +
				"event: message\n" +
				"retry: 5000\n" +
				"data: {\"test\":\"data\"}\n" +
				"\n",
			wantCount: 1,
		},
		{
			name:      "Empty SSE stream",
			sseData:   "",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				w.WriteHeader(http.StatusOK)

				flusher := w.(http.Flusher)
				fmt.Fprint(w, tt.sseData)
				flusher.Flush()
			}))
			defer server.Close()

			client := NewClient(server.URL, nil)
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			stream, err := client.Connect(ctx, nil)
			if err != nil {
				t.Fatalf("Connect() error = %v", err)
			}
			defer stream.Close()

			// Collect events
			events := []Event{}
			timeout := time.After(200 * time.Millisecond)

		collectLoop:
			for {
				select {
				case event, ok := <-stream.Events():
					if !ok {
						break collectLoop
					}
					events = append(events, event)
				case <-timeout:
					break collectLoop
				}
			}

			if len(events) != tt.wantCount {
				t.Errorf("Got %d events, want %d", len(events), tt.wantCount)
			}

			if tt.checkData && len(events) > 0 {
				if events[0].ID != "123" {
					t.Errorf("Event ID = %s, want 123", events[0].ID)
				}
				if string(events[0].Data) != `{"test":"value"}` {
					t.Errorf("Event Data = %s, want {\"test\":\"value\"}", events[0].Data)
				}
			}
		})
	}
}

func TestClient_UpdateCheckpoint(t *testing.T) {
	tests := []struct {
		name      string
		eventID   string
		wantTopic string
		wantErr   bool
	}{
		{
			name:      "Valid checkpoint JSON array",
			eventID:   `[{"topic":"test.recentchange","partition":0,"offset":123}]`,
			wantTopic: "test.recentchange",
			wantErr:   false,
		},
		{
			name:    "Invalid JSON - should log error but not crash",
			eventID: `{invalid json}`,
			wantErr: true,
		},
		{
			name:    "Empty array - should not panic",
			eventID: `[]`,
			wantErr: true, // No checkpoint to extract
		},
		{
			name:      "Multiple checkpoints - should use first",
			eventID:   `[{"topic":"first","partition":0,"offset":1},{"topic":"second","partition":1,"offset":2}]`,
			wantTopic: "first",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				w.WriteHeader(http.StatusOK)

				flusher := w.(http.Flusher)
				fmt.Fprintf(w, "id: %s\n", tt.eventID)
				fmt.Fprint(w, "data: {}\n\n")
				flusher.Flush()
			}))
			defer server.Close()

			client := NewClient(server.URL, nil)
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			stream, err := client.Connect(ctx, nil)
			if err != nil {
				t.Fatalf("Connect() error = %v", err)
			}
			defer stream.Close()

			// Wait for event to be processed
			time.Sleep(100 * time.Millisecond)

			checkpoint := client.GetCheckpoint()
			if tt.wantErr {
				if checkpoint != nil {
					t.Error("Expected nil checkpoint for invalid input")
				}
			} else {
				if checkpoint == nil {
					t.Fatal("Expected checkpoint, got nil")
				}
				if checkpoint.Topic != tt.wantTopic {
					t.Errorf("Checkpoint topic = %s, want %s", checkpoint.Topic, tt.wantTopic)
				}
			}
		})
	}
}

func TestClient_GetCheckpoint(t *testing.T) {
	t.Run("Returns nil when no checkpoint set", func(t *testing.T) {
		client := NewClient("https://example.org/stream", nil)
		checkpoint := client.GetCheckpoint()
		if checkpoint != nil {
			t.Error("Expected nil checkpoint, got non-nil")
		}
	})

	t.Run("Returns copy, not reference", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)

			flusher := w.(http.Flusher)
			fmt.Fprint(w, "id: [{\"topic\":\"test\",\"partition\":0,\"offset\":100}]\n")
			fmt.Fprint(w, "data: {}\n\n")
			flusher.Flush()
		}))
		defer server.Close()

		client := NewClient(server.URL, nil)
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		stream, err := client.Connect(ctx, nil)
		if err != nil {
			t.Fatalf("Connect() error = %v", err)
		}
		defer stream.Close()

		// Wait for checkpoint to be set
		time.Sleep(100 * time.Millisecond)

		checkpoint1 := client.GetCheckpoint()
		if checkpoint1 == nil {
			t.Fatal("Expected checkpoint, got nil")
		}

		// Modify the returned checkpoint
		checkpoint1.Offset = 999
		checkpoint1.Topic = "modified"

		// Get checkpoint again and verify it wasn't modified
		checkpoint2 := client.GetCheckpoint()
		if checkpoint2 == nil {
			t.Fatal("Expected checkpoint, got nil")
		}

		if checkpoint2.Offset == 999 || checkpoint2.Topic == "modified" {
			t.Error("Checkpoint was modified - not returning a copy")
		}

		if checkpoint2.Offset != 100 || checkpoint2.Topic != "test" {
			t.Errorf("Checkpoint values incorrect: offset=%d topic=%s", checkpoint2.Offset, checkpoint2.Topic)
		}
	})
}

func TestClient_Events_Channels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		flusher := w.(http.Flusher)
		fmt.Fprint(w, "data: {\"test\":1}\n\n")
		flusher.Flush()

		// Close connection after sending one event
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stream, err := client.Connect(ctx, nil)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer stream.Close()

	// Read from events channel
	select {
	case event, ok := <-stream.Events():
		if !ok {
			t.Error("Events channel closed unexpectedly before event")
		}
		if len(event.Data) == 0 {
			t.Error("Expected event data")
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Timeout waiting for event")
	}

	// Wait for stream to close
	select {
	case <-stream.Done():
		// Expected
	case <-time.After(500 * time.Millisecond):
		t.Error("Timeout waiting for Done channel")
	}
}
