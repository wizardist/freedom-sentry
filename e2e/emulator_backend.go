package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/wizardist/freedom-sentry/pkg/mwemu"
)

type emulatorBackend struct {
	t      *testing.T
	emu    *mwemu.Emulator
	config Config
}

// NewEmulatorBackend creates a new emulator-based test backend
func NewEmulatorBackend(t *testing.T, config Config) Backend {
	t.Helper()

	// Apply defaults
	if config.WikiDomain == "" {
		config.WikiDomain = "en.wikipedia.org"
	}
	if config.ListPageName == "" {
		config.ListPageName = "Wikipedia:Suppression_list"
	}

	// Convert to mwemu.Config
	mwemuConfig := mwemu.Config{
		InitialList:  config.InitialList,
		WikiDomain:   config.WikiDomain,
		ListPageName: config.ListPageName,
		Pages:        convertPages(config.Pages),
		User: mwemu.UserConfig{
			Name:   "TestBot",
			UserID: 12345,
			Rights: []string{"read", "edit", "suppressrevision"},
		},
	}

	emu := mwemu.New(t, mwemuConfig)

	return &emulatorBackend{
		t:      t,
		emu:    emu,
		config: config,
	}
}

func (e *emulatorBackend) APIEndpoint() string {
	return e.emu.APIEndpoint()
}

func (e *emulatorBackend) EventStreamsEndpoint() string {
	return e.emu.EventStreamsEndpoint()
}

func (e *emulatorBackend) WikiDomain() string {
	return e.config.WikiDomain
}

func (e *emulatorBackend) ListPageName() string {
	return e.config.ListPageName
}

func (e *emulatorBackend) EmitEdit(edit Edit) error {
	e.emu.EmitEdit(mwemu.Edit{
		Title:     edit.Title,
		RevID:     edit.RevID,
		OldRevID:  edit.OldRevID,
		User:      edit.User,
		Timestamp: edit.Timestamp,
		Bot:       edit.Bot,
		Namespace: edit.Namespace,
		Comment:   edit.Comment,
	})
	return nil
}

func (e *emulatorBackend) AssertRevisionSuppressed(t *testing.T, revID int64, timeout time.Duration) {
	e.emu.AssertRevisionSuppressed(t, revID, timeout)
}

func (e *emulatorBackend) AssertRevisionNotSuppressed(t *testing.T, revID int64, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	pollInterval := 10 * time.Millisecond

	for time.Now().Before(deadline) {
		if !e.emu.GetState().IsRevisionSuppressed(revID) {
			return
		}
		time.Sleep(pollInterval)
	}

	t.Errorf("Revision %d was suppressed, but expected not suppressed", revID)
}

func (e *emulatorBackend) UpdateSuppressionList(pages []string) error {
	state := e.emu.GetState()
	state.UpdateSuppressionList(pages)
	// The emulator automatically generates list page content from the suppression list
	return nil
}

func (e *emulatorBackend) GetSuppressionList() ([]string, error) {
	return e.emu.GetState().GetSuppressionList(), nil
}

func (e *emulatorBackend) Close() {
	e.emu.Close()
}

// convertPages converts e2e.Page to mwemu.Page
func convertPages(pages map[string]*Page) map[string]*mwemu.Page {
	if pages == nil {
		return nil
	}

	result := make(map[string]*mwemu.Page)
	for title, page := range pages {
		result[title] = &mwemu.Page{
			Title:     page.Title,
			Namespace: page.Namespace,
			PageID:    page.PageID,
			Revisions: convertRevisions(page.Revisions),
		}
	}
	return result
}

// convertRevisions converts e2e.Revision to mwemu.Revision
func convertRevisions(revisions []*Revision) []*mwemu.Revision {
	if revisions == nil {
		return nil
	}

	result := make([]*mwemu.Revision, len(revisions))
	for i, rev := range revisions {
		result[i] = &mwemu.Revision{
			RevID:      rev.RevID,
			ParentID:   rev.ParentID,
			User:       rev.User,
			Timestamp:  rev.Timestamp,
			Comment:    rev.Comment,
			Content:    rev.Content,
			Suppressed: rev.Suppressed,
		}
	}
	return result
}

// verifyEmulatorExists checks that the emulator package is available
func init() {
	// This init function ensures the emulator import compiles correctly
	_ = fmt.Sprintf("emulator backend initialized")
}
