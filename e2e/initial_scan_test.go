package e2e

import (
	"testing"
	"time"
)

func TestInitialFullScan(t *testing.T) {
	// Setup backend with existing pages/revisions
	backend := NewEmulatorBackend(t, Config{
		InitialList:  []string{"European Bison"},
		WikiDomain:   "en.wikipedia.org",
		ListPageName: "Wikipedia:Suppression_list",
		Pages: map[string]*Page{
			"European Bison": {
				Title:     "European Bison",
				PageID:    100,
				Namespace: 0,
				Revisions: []*Revision{
					{
						RevID:     1001,
						ParentID:  1000,
						User:      "User1",
						Timestamp: time.Now().Add(-1 * time.Hour),
						Comment:   "First edit",
						Content:   "Content 1",
					},
					{
						RevID:     1002,
						ParentID:  1001,
						User:      "User2",
						Timestamp: time.Now().Add(-30 * time.Minute),
						Comment:   "Second edit",
						Content:   "Content 2",
					},
				},
			},
		},
	})
	defer backend.Close()

	// Start bot (initial scan should run by default)
	bot := StartBot(t, backend)
	defer bot.Stop()

	// Verify both existing revisions are suppressed
	backend.AssertRevisionSuppressed(t, 1001, 10*time.Second)
	backend.AssertRevisionSuppressed(t, 1002, 10*time.Second)
}
