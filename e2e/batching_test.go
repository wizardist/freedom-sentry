package e2e

import (
	"testing"
	"time"

	"github.com/wizardist/freedom-sentry/config"
)

func TestBatching(t *testing.T) {
	// Setup backend with pages that will receive edits
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
						User:      "TestUser1",
						Timestamp: time.Now(),
						Content:   "New content 1",
					},
					{
						RevID:     1002,
						ParentID:  1001,
						User:      "TestUser2",
						Timestamp: time.Now(),
						Content:   "New content 2",
					},
				},
			},
		},
	})
	defer backend.Close()

	// Start bot with a short batching period
	extraEnv := map[string]string{
		config.EnvBatchingSuppressorPeriod: "2s",
	}
	bot := StartBot(t, backend, extraEnv)
	defer bot.Stop()

	// Emit two edits in quick succession
	edit1 := Edit{
		Title:     "European Bison",
		RevID:     1001,
		OldRevID:  1000,
		User:      "TestUser1",
		Timestamp: time.Now(),
		Bot:       false,
		Namespace: 0,
	}
	if err := backend.EmitEdit(edit1); err != nil {
		t.Fatalf("Failed to emit edit 1: %v", err)
	}

	edit2 := Edit{
		Title:     "European Bison",
		RevID:     1002,
		OldRevID:  1001,
		User:      "TestUser2",
		Timestamp: time.Now(),
		Bot:       false,
		Namespace: 0,
	}
	if err := backend.EmitEdit(edit2); err != nil {
		t.Fatalf("Failed to emit edit 2: %v", err)
	}

	// Wait for the batching period to elapse plus a small buffer
	time.Sleep(2500 * time.Millisecond)

	// Assert that both revisions were suppressed
	backend.AssertRevisionSuppressed(t, edit1.RevID, 1*time.Second)
	backend.AssertRevisionSuppressed(t, edit2.RevID, 1*time.Second)

	// Assert that the suppressions happened in a single batch
	backend.AssertSuppressionBatchCount(t, 1)
}
