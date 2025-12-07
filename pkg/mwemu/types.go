package mwemu

import "time"

// Config defines the initial configuration for the MediaWiki emulator
type Config struct {
	// InitialList is the initial set of page titles in the suppression list
	InitialList []string

	// Pages defines the pages and their revisions available in the emulator
	Pages map[string]*Page

	// User configuration for the authenticated user
	User UserConfig

	// WikiDomain is the server_name for EventStreams events (e.g., "en.wikipedia.org")
	WikiDomain string

	// ListPageName is the name of the suppression list page (e.g., "Wikipedia:Suppression_list")
	ListPageName string
}

// UserConfig defines the authenticated user's properties
type UserConfig struct {
	Name   string
	UserID int64
	Rights []string
}

// Page represents a MediaWiki page with its revisions
type Page struct {
	Title     string
	Namespace int
	PageID    int64
	Revisions []*Revision
}

// Revision represents a single revision of a page
type Revision struct {
	RevID      int64
	ParentID   int64
	User       string
	Timestamp  time.Time
	Comment    string
	Content    string
	Suppressed bool

	// Suppression details
	UserHidden    bool
	CommentHidden bool
}

// APICall records an API request for test verification
type APICall struct {
	Timestamp time.Time
	Action    string
	Params    map[string]string
}

// Edit represents an edit event to be emitted via EventStreams
type Edit struct {
	Title     string
	RevID     int64
	OldRevID  int64
	User      string
	Timestamp time.Time
	Bot       bool
	Namespace int
	Comment   string
}
