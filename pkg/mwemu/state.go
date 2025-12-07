package mwemu

import (
	"strconv"
	"sync"
	"time"
)

// State maintains the emulator's internal state with thread-safe access
type State struct {
	mu sync.RWMutex

	// Pages indexed by title
	Pages map[string]*Page

	// Suppression list (page titles that should have revisions suppressed)
	SuppressionList []string

	// API call history for test verification
	APICallHistory []APICall

	// CSRF token for write operations
	CSRFToken string

	// Event sequence counter for SSE events
	EventSequence int64

	// Recent changes history (for API queries)
	RecentChanges []recentChangeEntry
}

// recentChangeEntry stores a recent change for API queries
type recentChangeEntry struct {
	Type      string
	Namespace int
	Title     string
	PageID    int64
	RevID     int64
	OldRevID  int64
	User      string
	Timestamp time.Time
}

// NewState creates a new State with the given configuration
func NewState(config Config) *State {
	return &State{
		Pages:           config.Pages,
		SuppressionList: config.InitialList,
		APICallHistory:  []APICall{},
		CSRFToken:       "test-csrf-token-" + time.Now().Format("20060102150405"),
		EventSequence:   0,
		RecentChanges:   []recentChangeEntry{},
	}
}

// SuppressRevision marks a revision as suppressed
func (s *State) SuppressRevision(revID int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, page := range s.Pages {
		for _, rev := range page.Revisions {
			if rev.RevID == revID {
				rev.Suppressed = true
				rev.UserHidden = true
				rev.CommentHidden = true
				return true
			}
		}
	}
	return false
}

// IsRevisionSuppressed checks if a revision is marked as suppressed
func (s *State) IsRevisionSuppressed(revID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, page := range s.Pages {
		for _, rev := range page.Revisions {
			if rev.RevID == revID {
				return rev.Suppressed
			}
		}
	}
	return false
}

// GetPageByTitle retrieves a page by its title
func (s *State) GetPageByTitle(title string) *Page {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Pages[title]
}

// GetRecentChanges returns recent changes filtered by start time and limit
func (s *State) GetRecentChanges(start string, limitStr string) []interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var startTime time.Time
	if start != "" {
		var err error
		startTime, err = time.Parse(time.RFC3339, start)
		if err != nil {
			return []interface{}{}
		}
	}

	limit := 5000
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	result := []interface{}{}
	for _, rc := range s.RecentChanges {
		// Filter by start time (if provided)
		if !startTime.IsZero() && rc.Timestamp.Before(startTime) {
			continue
		}

		entry := map[string]interface{}{
			"type":       rc.Type,
			"ns":         rc.Namespace,
			"title":      rc.Title,
			"pageid":     rc.PageID,
			"revid":      rc.RevID,
			"old_revid":  rc.OldRevID,
			"user":       rc.User,
			"timestamp":  rc.Timestamp.UTC().Format("2006-01-02T15:04:05Z"),
		}

		result = append(result, entry)

		if len(result) >= limit {
			break
		}
	}

	return result
}

// AddRecentChange adds a change to the recent changes list
func (s *State) AddRecentChange(rc recentChangeEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.RecentChanges = append(s.RecentChanges, rc)
}

// RecordAPICall records an API call for test verification
func (s *State) RecordAPICall(action string, params map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.APICallHistory = append(s.APICallHistory, APICall{
		Timestamp: time.Now(),
		Action:    action,
		Params:    params,
	})
}

// GetAPICallHistory returns a copy of the API call history
func (s *State) GetAPICallHistory() []APICall {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history := make([]APICall, len(s.APICallHistory))
	copy(history, s.APICallHistory)
	return history
}

// GetSuppressionList returns a copy of the suppression list
func (s *State) GetSuppressionList() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]string, len(s.SuppressionList))
	copy(list, s.SuppressionList)
	return list
}

// UpdateSuppressionList updates the suppression list (simulates editing the list page)
func (s *State) UpdateSuppressionList(newList []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.SuppressionList = newList
}

// GetCSRFToken returns the current CSRF token
func (s *State) GetCSRFToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.CSRFToken
}

// IncrementEventSequence increments and returns the new event sequence number
func (s *State) IncrementEventSequence() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.EventSequence++
	return s.EventSequence
}
