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

func TestBatchingLatency(t *testing.T) {
	// Setup backend with a page for latency testing
	backend := NewEmulatorBackend(t, Config{
		InitialList:  []string{"Test Page"},
		WikiDomain:   "en.wikipedia.org",
		ListPageName: "Wikipedia:Suppression_list",
		Pages: map[string]*Page{
			"Test Page": {
				Title:     "Test Page",
				PageID:    200,
				Namespace: 0,
				Revisions: []*Revision{
					{
						RevID:     2001,
						ParentID:  2000,
						User:      "LatencyTestUser",
						Timestamp: time.Now(),
						Content:   "Test content",
					},
				},
			},
		},
	})
	defer backend.Close()

	// Start bot with default settings (1s period, 50 batch size)
	bot := StartBot(t, backend, nil)
	defer bot.Stop()

	// Emit a single edit and measure latency
	startTime := time.Now()
	edit := Edit{
		Title:     "Test Page",
		RevID:     2001,
		OldRevID:  2000,
		User:      "LatencyTestUser",
		Timestamp: time.Now(),
		Bot:       false,
		Namespace: 0,
	}
	if err := backend.EmitEdit(edit); err != nil {
		t.Fatalf("Failed to emit edit: %v", err)
	}

	// Assert suppression happens within 1.5 seconds (1s period + 500ms buffer)
	backend.AssertRevisionSuppressed(t, edit.RevID, 1500*time.Millisecond)
	suppressionLatency := time.Since(startTime)

	t.Logf("Suppression latency: %v", suppressionLatency)

	// Verify latency is within acceptable bounds
	if suppressionLatency > 1500*time.Millisecond {
		t.Errorf("Suppression latency too high: %v > 1.5s", suppressionLatency)
	}
}

func TestConfigurableBatchSize(t *testing.T) {
	// Setup backend with multiple revisions
	backend := NewEmulatorBackend(t, Config{
		InitialList:  []string{"Batch Size Test"},
		WikiDomain:   "en.wikipedia.org",
		ListPageName: "Wikipedia:Suppression_list",
		Pages: map[string]*Page{
			"Batch Size Test": {
				Title:     "Batch Size Test",
				PageID:    300,
				Namespace: 0,
				Revisions: []*Revision{
					{RevID: 3001, ParentID: 3000, User: "User1", Timestamp: time.Now(), Content: "Content 1"},
					{RevID: 3002, ParentID: 3001, User: "User2", Timestamp: time.Now(), Content: "Content 2"},
					{RevID: 3003, ParentID: 3002, User: "User3", Timestamp: time.Now(), Content: "Content 3"},
				},
			},
		},
	})
	defer backend.Close()

	// Start bot with small batch size for faster triggering
	extraEnv := map[string]string{
		config.EnvBatchingSuppressorSize:   "2",
		config.EnvBatchingSuppressorPeriod: "1s",
	}
	bot := StartBot(t, backend, extraEnv)
	defer bot.Stop()

	// Emit three edits
	edits := []Edit{
		{Title: "Batch Size Test", RevID: 3001, OldRevID: 3000, User: "User1", Timestamp: time.Now(), Namespace: 0},
		{Title: "Batch Size Test", RevID: 3002, OldRevID: 3001, User: "User2", Timestamp: time.Now(), Namespace: 0},
		{Title: "Batch Size Test", RevID: 3003, OldRevID: 3002, User: "User3", Timestamp: time.Now(), Namespace: 0},
	}

	for _, edit := range edits {
		if err := backend.EmitEdit(edit); err != nil {
			t.Fatalf("Failed to emit edit %d: %v", edit.RevID, err)
		}
	}

	// Wait for batching to complete
	time.Sleep(1500 * time.Millisecond)

	// All revisions should be suppressed
	for _, edit := range edits {
		backend.AssertRevisionSuppressed(t, edit.RevID, 500*time.Millisecond)
	}

	// With batch size 2, we expect 2 batches (2 + 1)
	backend.AssertSuppressionBatchCount(t, 2)
}
