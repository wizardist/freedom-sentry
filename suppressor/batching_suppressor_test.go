package suppressor

import (
	"freedom-sentry/mediawiki"
	"testing"
	"time"
)

func Test_batchingSuppressor_SuppressRevisions(t *testing.T) {
	type args [][]mediawiki.Revision

	tests := []struct {
		name           string
		invocations    args
		suppressValues string
		forceDrain     int
	}{
		{
			name: "Will not call API if nothing ever passed",
			invocations: [][]mediawiki.Revision{
				{}, {}, {},
			},
			forceDrain: 15, // Will tolerate draining on empty buffer
		},
		{
			name: "Will call API once",
			invocations: [][]mediawiki.Revision{
				{mediawiki.Revision{Id: "1"}},
			},
			suppressValues: "1",
			forceDrain:     1,
		},
		{
			name: "Will call API once with exactly the buffer size",
			invocations: [][]mediawiki.Revision{
				{mediawiki.Revision{Id: "1"}},
				{mediawiki.Revision{Id: "2"}, mediawiki.Revision{Id: "3"}},
				{mediawiki.Revision{Id: "4"}, mediawiki.Revision{Id: "5"}},
			},
			suppressValues: "1,2,3,4,5",
		},
		{
			name: "Will call API twice when a single invocation is larger",
			invocations: [][]mediawiki.Revision{
				{
					mediawiki.Revision{Id: "1"},
					mediawiki.Revision{Id: "2"},
					mediawiki.Revision{Id: "3"},
					mediawiki.Revision{Id: "4"},
					mediawiki.Revision{Id: "5"},
					mediawiki.Revision{Id: "6"},
					mediawiki.Revision{Id: "7"},
				},
			},
			suppressValues: "1,2,3,4,5|6,7",
			forceDrain:     1,
		},
		{
			name: "Will call API twice when overflowing",
			invocations: [][]mediawiki.Revision{
				{mediawiki.Revision{Id: "1"}},
				{mediawiki.Revision{Id: "2"}, mediawiki.Revision{Id: "3"}},
				{mediawiki.Revision{Id: "4"}, mediawiki.Revision{Id: "5"}},
				{mediawiki.Revision{Id: "6"}, mediawiki.Revision{Id: "7"}},
			},
			suppressValues: "1,2,3,4,5|6,7",
			forceDrain:     1,
		},
		{
			name: "Will call API three times when overflowing",
			invocations: [][]mediawiki.Revision{
				{
					mediawiki.Revision{Id: "1"},
					mediawiki.Revision{Id: "2"},
					mediawiki.Revision{Id: "3"},
					mediawiki.Revision{Id: "4"},
					mediawiki.Revision{Id: "5"},
					mediawiki.Revision{Id: "6"},
					mediawiki.Revision{Id: "7"},
				},
				{
					mediawiki.Revision{Id: "8"},
				},
				{
					mediawiki.Revision{Id: "9"},
					mediawiki.Revision{Id: "10"},
				},
				{
					mediawiki.Revision{Id: "11"},
					mediawiki.Revision{Id: "12"},
				},
				{
					mediawiki.Revision{Id: "13"},
					mediawiki.Revision{Id: "14"},
				},
			},
			suppressValues: "1,2,3,4,5|6,7,8,9,10|11,12,13,14",
			forceDrain:     2, // Testing more force drains when the buffer is empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			standard := &mockSuppressor{}
			batching := &batchingSuppressor{
				size:       5,
				suppressor: standard,
			}

			for _, revs := range tt.invocations {
				if err := batching.SuppressRevisions(revs); err != nil {
					t.Errorf("SuppressRevisions() must not throw error")
					return
				}
			}

			// Wait until force drain executes
			for i := 0; i < tt.forceDrain; i++ {
				batching.forceDrainRequest <- true
			}

			// Goroutines might not complete as fast
			time.Sleep(time.Millisecond)

			if tt.suppressValues != standard.callHistory {
				t.Errorf("SuppressRevisions() call pattern [%s], expected [%s]", standard.callHistory, tt.suppressValues)
			}
		})
	}
}
