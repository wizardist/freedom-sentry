package e2e

import (
	"testing"
	"time"
)

func TestBasicSuppression(t *testing.T) {
	tests := []struct {
		name             string
		initialList      []string
		editToEmit       Edit
		expectSuppressed bool
	}{
		{
			name:        "Edit on listed page gets suppressed",
			initialList: []string{"European Bison"},
			editToEmit: Edit{
				Title:     "European Bison",
				RevID:     1001,
				OldRevID:  1000,
				User:      "TestUser",
				Timestamp: time.Now(),
				Bot:       false,
				Namespace: 0,
			},
			expectSuppressed: true,
		},
		{
			name:        "Edit on unlisted page is not suppressed",
			initialList: []string{"European Bison"},
			editToEmit: Edit{
				Title:     "Eurasian Lynx",
				RevID:     2001,
				OldRevID:  2000,
				User:      "TestUser",
				Timestamp: time.Now(),
				Bot:       false,
				Namespace: 0,
			},
			expectSuppressed: false,
		},
		{
			name:        "Bot edits are filtered and not suppressed",
			initialList: []string{"European Bison"},
			editToEmit: Edit{
				Title:     "European Bison",
				RevID:     1002,
				OldRevID:  1001,
				User:      "BotUser",
				Timestamp: time.Now(),
				Bot:       true,
				Namespace: 0,
			},
			expectSuppressed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup backend with pages that will receive edits
			backend := NewEmulatorBackend(t, Config{
				InitialList:  tt.initialList,
				WikiDomain:   "en.wikipedia.org",
				ListPageName: "Wikipedia:Suppression_list",
				Pages: map[string]*Page{
					"European Bison": {
						Title:     "European Bison",
						PageID:    100,
						Namespace: 0,
						Revisions: []*Revision{
							{
								RevID:     1000,
								ParentID:  999,
								User:      "OldUser",
								Timestamp: time.Now().Add(-1 * time.Hour),
								Content:   "Initial content",
							},
							{
								RevID:     1001,
								ParentID:  1000,
								User:      "TestUser",
								Timestamp: time.Now(),
								Content:   "New content",
							},
							{
								RevID:     1002,
								ParentID:  1001,
								User:      "BotUser",
								Timestamp: time.Now(),
								Content:   "Bot edit",
							},
						},
					},
					"Eurasian Lynx": {
						Title:     "Eurasian Lynx",
						PageID:    200,
						Namespace: 0,
						Revisions: []*Revision{
							{
								RevID:     2000,
								ParentID:  1999,
								User:      "OldUser",
								Timestamp: time.Now().Add(-1 * time.Hour),
								Content:   "Initial content",
							},
							{
								RevID:     2001,
								ParentID:  2000,
								User:      "TestUser",
								Timestamp: time.Now(),
								Content:   "New content",
							},
						},
					},
				},
			})
			defer backend.Close()

			// Start bot
			bot := StartBot(t, backend, nil)
			defer bot.Stop()

			// Emit edit
			if err := backend.EmitEdit(tt.editToEmit); err != nil {
				t.Fatalf("Failed to emit edit: %v", err)
			}

			// Assert suppression
			if tt.expectSuppressed {
				backend.AssertRevisionSuppressed(t, tt.editToEmit.RevID, 10*time.Second)
			} else {
				backend.AssertRevisionNotSuppressed(t, tt.editToEmit.RevID, 5*time.Second)
			}
		})
	}
}
