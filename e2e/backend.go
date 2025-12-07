package e2e

import (
	"testing"
	"time"
)

// Backend represents a test backend (emulator or live MediaWiki).
// This abstraction allows tests to work with either the MediaWiki emulator
// or a live MediaWiki instance without modification.
type Backend interface {
	// Configuration
	APIEndpoint() string
	EventStreamsEndpoint() string
	WikiDomain() string
	ListPageName() string

	// Event Emission
	EmitEdit(edit Edit) error

	// Assertions (with polling for eventual consistency)
	AssertRevisionSuppressed(t *testing.T, revID int64, timeout time.Duration)
	AssertRevisionNotSuppressed(t *testing.T, revID int64, timeout time.Duration)

	// State Management
	UpdateSuppressionList(pages []string) error
	GetSuppressionList() ([]string, error)

	// Cleanup
	Close()
}

// Edit represents an edit event to be emitted to the bot
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

// Config contains configuration for backend initialization
type Config struct {
	InitialList  []string
	WikiDomain   string
	ListPageName string
	Pages        map[string]*Page
}

// Page represents a wiki page with revisions
type Page struct {
	Title     string
	PageID    int64
	Namespace int
	Revisions []*Revision
}

// Revision represents a page revision
type Revision struct {
	RevID      int64
	ParentID   int64
	User       string
	Timestamp  time.Time
	Comment    string
	Content    string
	Suppressed bool
}
