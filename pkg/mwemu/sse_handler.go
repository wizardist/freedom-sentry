package mwemu

import (
	"fmt"
	"net/http"
	"time"
)

// handleSSE handles Server-Sent Events (SSE) for EventStreams
func (e *Emulator) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Parse 'since' parameter for historical playback
	since := r.URL.Query().Get("since")
	if since != "" {
		e.t.Logf("SSE client connected with since parameter: %s", since)
		// For now, we'll ignore historical events
		// In a more complete implementation, we'd replay events from that timestamp
		_, err := time.Parse(time.RFC3339, since)
		if err != nil {
			e.t.Logf("Invalid since parameter: %v", err)
		}
	} else {
		e.t.Log("SSE client connected")
	}

	// Register this client
	eventChan := e.registerSSEClient()
	defer e.unregisterSSEClient(eventChan)

	// Stream events to the client
	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				// Channel closed, client should disconnect
				return
			}

			// Write SSE event
			if event.ID != "" {
				fmt.Fprintf(w, "id: %s\n", event.ID)
			}
			fmt.Fprintf(w, "data: %s\n\n", event.Data)
			flusher.Flush()

		case <-r.Context().Done():
			// Client disconnected
			e.t.Log("SSE client disconnected")
			return
		}
	}
}
