package mwemu

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// Emulator provides both MediaWiki API and EventStreams SSE endpoints for testing
type Emulator struct {
	t *testing.T

	// HTTP servers
	apiServer *httptest.Server
	sseServer *httptest.Server

	// State
	state *State

	// SSE client management
	sseClients   map[chan sseEvent]bool
	sseClientsMu sync.Mutex

	// Configuration
	config Config
}

// sseEvent represents an event to be sent over SSE
type sseEvent struct {
	Event string
	ID    string
	Data  string
}

// New creates a new MediaWiki emulator for testing
func New(t *testing.T, config Config) *Emulator {
	t.Helper()

	// Set defaults
	if config.User.Name == "" {
		config.User = UserConfig{
			Name:   "TestBot",
			UserID: 12345,
			Rights: []string{"read", "edit", "suppressrevision"},
		}
	}
	if config.WikiDomain == "" {
		config.WikiDomain = "en.wikipedia.org"
	}
	if config.ListPageName == "" {
		config.ListPageName = "Wikipedia:Suppression_list"
	}

	emu := &Emulator{
		t:          t,
		config:     config,
		sseClients: make(map[chan sseEvent]bool),
	}

	// Initialize state
	emu.state = NewState(config)

	// Ensure suppression list page exists in Pages
	if emu.state.Pages == nil {
		emu.state.Pages = make(map[string]*Page)
	}
	if _, exists := emu.state.Pages[config.ListPageName]; !exists {
		// Create initial revision with suppression list content
		initialContent := strings.Join(config.InitialList, "\n")
		emu.state.Pages[config.ListPageName] = &Page{
			Title:     config.ListPageName,
			Namespace: 4, // Wikipedia namespace
			PageID:    1,
			Revisions: []*Revision{
				{
					RevID:     1,
					ParentID:  0,
					User:      "System",
					Timestamp: time.Now(),
					Comment:   "Initial suppression list",
					Content:   initialContent,
				},
			},
		}
	}

	// Create API server
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/", emu.handleAPI)
	emu.apiServer = httptest.NewServer(apiMux)

	// Create SSE server
	sseMux := http.NewServeMux()
	sseMux.HandleFunc("/", emu.handleSSE)
	emu.sseServer = httptest.NewServer(sseMux)

	t.Logf("MediaWiki emulator started - API: %s, SSE: %s", emu.apiServer.URL, emu.sseServer.URL)

	return emu
}

// Close shuts down the emulator servers
func (e *Emulator) Close() {
	e.t.Helper()

	e.apiServer.Close()
	e.sseServer.Close()

	// Close all SSE client channels
	e.sseClientsMu.Lock()
	for ch := range e.sseClients {
		close(ch)
	}
	e.sseClients = make(map[chan sseEvent]bool)
	e.sseClientsMu.Unlock()

	e.t.Log("MediaWiki emulator closed")
}

// APIEndpoint returns the URL of the MediaWiki API endpoint
func (e *Emulator) APIEndpoint() string {
	return e.apiServer.URL
}

// EventStreamsEndpoint returns the URL of the EventStreams SSE endpoint
func (e *Emulator) EventStreamsEndpoint() string {
	return e.sseServer.URL
}

// EmitEdit broadcasts an edit event to all connected SSE clients
func (e *Emulator) EmitEdit(edit Edit) {
	// Increment event sequence
	seq := e.state.IncrementEventSequence()

	// Add to recent changes
	if edit.Timestamp.IsZero() {
		edit.Timestamp = time.Now()
	}

	e.state.AddRecentChange(recentChangeEntry{
		Type:      "edit",
		Namespace: edit.Namespace,
		Title:     edit.Title,
		PageID:    e.getPageID(edit.Title),
		RevID:     edit.RevID,
		OldRevID:  edit.OldRevID,
		User:      edit.User,
		Timestamp: edit.Timestamp,
	})

	// Create EventStreams JSON format
	eventData := map[string]interface{}{
		"server_name": e.config.WikiDomain,
		"type":        "edit",
		"namespace":   edit.Namespace,
		"title":       edit.Title,
		"user":        edit.User,
		"bot":         edit.Bot,
		"timestamp":   edit.Timestamp.Unix(),
		"revision": map[string]interface{}{
			"new": edit.RevID,
			"old": edit.OldRevID,
		},
		"meta": map[string]interface{}{
			"dt": edit.Timestamp.UTC().Format("2006-01-02T15:04:05Z"),
		},
	}

	dataJSON, err := json.Marshal(eventData)
	if err != nil {
		e.t.Fatalf("Failed to marshal edit event: %v", err)
	}

	// Create checkpoint (Kafka offset in EventStreams format)
	checkpoint := []map[string]interface{}{
		{
			"topic":     "test.recentchange",
			"partition": 0,
			"offset":    seq,
			"timestamp": edit.Timestamp.UnixMilli(),
		},
	}
	checkpointJSON, err := json.Marshal(checkpoint)
	if err != nil {
		e.t.Fatalf("Failed to marshal checkpoint: %v", err)
	}

	// Create SSE event
	event := sseEvent{
		Event: "message",
		ID:    string(checkpointJSON),
		Data:  string(dataJSON),
	}

	// Broadcast to all connected clients
	e.sseClientsMu.Lock()
	defer e.sseClientsMu.Unlock()

	for client := range e.sseClients {
		select {
		case client <- event:
		default:
			// Client buffer full, skip (slow consumer)
			e.t.Logf("Warning: SSE client buffer full, event dropped")
		}
	}

	e.t.Logf("Emitted edit event: %s (rev %d)", edit.Title, edit.RevID)
}

// getPageID returns the page ID for a given title
func (e *Emulator) getPageID(title string) int64 {
	page := e.state.GetPageByTitle(title)
	if page != nil {
		return page.PageID
	}
	return 0
}

// sendJSON sends a JSON response
func (e *Emulator) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		e.t.Errorf("Failed to encode JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// sendError sends a MediaWiki-style error response
func (e *Emulator) sendError(w http.ResponseWriter, code, info string) {
	e.sendJSON(w, map[string]interface{}{
		"error": map[string]interface{}{
			"code": code,
			"info": info,
		},
	})
}

// registerSSEClient registers a new SSE client
func (e *Emulator) registerSSEClient() chan sseEvent {
	e.sseClientsMu.Lock()
	defer e.sseClientsMu.Unlock()

	ch := make(chan sseEvent, 10)
	e.sseClients[ch] = true
	return ch
}

// unregisterSSEClient removes an SSE client
func (e *Emulator) unregisterSSEClient(ch chan sseEvent) {
	e.sseClientsMu.Lock()
	defer e.sseClientsMu.Unlock()

	delete(e.sseClients, ch)
	close(ch)
}
