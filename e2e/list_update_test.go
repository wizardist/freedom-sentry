package e2e

import (
	"testing"
	"time"
)

func TestListUpdateDetection(t *testing.T) {
	// Setup backend with initial list and pages
	backend := NewEmulatorBackend(t, Config{
		InitialList:  []string{"European Bison"},
		WikiDomain:   "en.wikipedia.org",
		ListPageName: "Wikipedia:Suppression_list",
		Pages: map[string]*Page{
			"European Bison": {
				Title:     "European Bison",
				PageID:    100,
				Namespace: 0,
				Revisions: []*Revision{},
			},
			"Eurasian Lynx": {
				Title:     "Eurasian Lynx",
				PageID:    200,
				Namespace: 0,
				Revisions: []*Revision{
					{
						RevID:     2001,
						ParentID:  2000,
						User:      "TestUser",
						Timestamp: time.Now(),
						Content:   "First edit",
					},
					{
						RevID:     2002,
						ParentID:  2001,
						User:      "TestUser",
						Timestamp: time.Now(),
						Content:   "Second edit",
					},
				},
			},
		},
	})
	defer backend.Close()

	// Start bot
	bot := StartBot(t, backend)
	defer bot.Stop()

	// Wait for bot to fully initialize and connect to EventStreams
	WaitForReady(2 * time.Second)

	// Emit edit to unlisted page (should NOT be suppressed)
	edit1 := Edit{
		Title:     "Eurasian Lynx",
		RevID:     2001,
		OldRevID:  2000,
		User:      "TestUser",
		Timestamp: time.Now(),
		Bot:       false,
		Namespace: 0,
	}
	backend.EmitEdit(edit1)
	backend.AssertRevisionNotSuppressed(t, edit1.RevID, 5*time.Second)

	// Update suppression list to include "Eurasian Lynx"
	err := backend.UpdateSuppressionList([]string{"European Bison", "Eurasian Lynx"})
	if err != nil {
		t.Fatalf("Failed to update list: %v", err)
	}

	// Emit edit event for list page update to trigger bot
	listEdit := Edit{
		Title:     backend.ListPageName(),
		RevID:     9001,
		OldRevID:  9000,
		User:      "AdminUser",
		Timestamp: time.Now(),
		Bot:       false,
		Namespace: 4, // Wikipedia namespace
	}
	backend.EmitEdit(listEdit)

	// Wait for bot to detect and process list update
	time.Sleep(2 * time.Second)

	// Emit another edit to Eurasian Lynx (should NOW be suppressed)
	edit2 := Edit{
		Title:     "Eurasian Lynx",
		RevID:     2002,
		OldRevID:  2001,
		User:      "TestUser",
		Timestamp: time.Now(),
		Bot:       false,
		Namespace: 0,
	}
	backend.EmitEdit(edit2)
	backend.AssertRevisionSuppressed(t, edit2.RevID, 10*time.Second)
}
