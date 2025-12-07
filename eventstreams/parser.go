package eventstreams

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/wizardist/freedom-sentry/mediawiki"
)

// RecentChangeEvent represents the EventStreams recentchange event schema
type RecentChangeEvent struct {
	ServerName string `json:"server_name"`
	Type       string `json:"type"`
	Namespace  int    `json:"namespace"`
	Title      string `json:"title"`
	User       string `json:"user"`
	Bot        bool   `json:"bot"`
	Timestamp  int64  `json:"timestamp"`
	Revision   struct {
		RevID    int64 `json:"rev_id"`
		OldRevID int64 `json:"old_rev_id"`
	} `json:"revision"`
	Meta struct {
		Dt string `json:"dt"` // ISO8601 timestamp
	} `json:"meta"`
}

// ParseRecentChange unmarshals JSON data into RecentChangeEvent
func ParseRecentChange(data []byte) (*RecentChangeEvent, error) {
	var event RecentChangeEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return &event, nil
}

// ToRevision converts EventStreams event to mediawiki.Revision
func (e *RecentChangeEvent) ToRevision() (mediawiki.Revision, error) {
	// Parse timestamp - EventStreams provides Unix timestamp in seconds
	var timestamp time.Time
	if e.Timestamp > 0 {
		timestamp = time.Unix(e.Timestamp, 0).UTC()
	} else if e.Meta.Dt != "" {
		// Fallback to meta.dt if timestamp field is not available
		var err error
		timestamp, err = time.Parse(time.RFC3339, e.Meta.Dt)
		if err != nil {
			return mediawiki.Revision{}, fmt.Errorf("failed to parse timestamp: %w", err)
		}
	}

	return mediawiki.Revision{
		Id:           mediawiki.RevisionIdFromAny(e.Revision.RevID),
		IsSuppressed: false, // EventStreams doesn't include suppressed revisions
		Title:        e.Title,
		Content:      "", // Not available in recent changes stream
		Timestamp:    timestamp,
	}, nil
}

// ShouldProcess filters events based on criteria matching current API polling
func (e *RecentChangeEvent) ShouldProcess() bool {
	// Only process edits (exclude logs, page creations, etc.)
	if e.Type != "edit" {
		return false
	}

	// Exclude bot edits
	if e.Bot {
		return false
	}

	return true
}
