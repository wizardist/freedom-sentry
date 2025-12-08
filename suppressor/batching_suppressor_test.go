package suppressor

import (
	"testing"
	"time"

	"github.com/wizardist/freedom-sentry/mediawiki"
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

func Test_batchingSuppressor_TimeBasedDrain(t *testing.T) {
	t.Run("drains buffer after period", func(t *testing.T) {
		standard := &mockSuppressor{}
		batching := &batchingSuppressor{
			period:     50 * time.Millisecond,
			size:       5,
			suppressor: standard,
		}

		revs := []mediawiki.Revision{{Id: "1"}}
		if err := batching.SuppressRevisions(revs); err != nil {
			t.Errorf("SuppressRevisions() must not throw error")
			return
		}

		time.Sleep(55 * time.Millisecond)

		if standard.callHistory != "1" {
			t.Errorf("SuppressRevisions() call pattern [%s], expected [1]", standard.callHistory)
		}
	})
}

func Test_batchingSuppressor_ConfigurableBatchSize(t *testing.T) {
	tests := []struct {
		name           string
		batchSize      int
		revisions      [][]mediawiki.Revision
		suppressValues string
		forceDrain     int
	}{
		{
			name:      "batch size 1 triggers immediately",
			batchSize: 1,
			revisions: [][]mediawiki.Revision{
				{{Id: "1"}},
				{{Id: "2"}},
				{{Id: "3"}},
			},
			suppressValues: "1|2|3",
		},
		{
			name:      "batch size 3 groups correctly",
			batchSize: 3,
			revisions: [][]mediawiki.Revision{
				{{Id: "1"}, {Id: "2"}},
				{{Id: "3"}, {Id: "4"}},
			},
			suppressValues: "1,2,3|4",
			forceDrain:     1,
		},
		{
			name:      "batch size 10 waits for more",
			batchSize: 10,
			revisions: [][]mediawiki.Revision{
				{{Id: "1"}, {Id: "2"}, {Id: "3"}},
				{{Id: "4"}, {Id: "5"}},
			},
			suppressValues: "1,2,3,4,5",
			forceDrain:     1,
		},
		{
			name:      "large batch size 100",
			batchSize: 100,
			revisions: [][]mediawiki.Revision{
				{{Id: "1"}, {Id: "2"}, {Id: "3"}},
			},
			suppressValues: "1,2,3",
			forceDrain:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			standard := &mockSuppressor{}
			batching := &batchingSuppressor{
				size:       tt.batchSize,
				suppressor: standard,
			}

			for _, revs := range tt.revisions {
				if err := batching.SuppressRevisions(revs); err != nil {
					t.Errorf("SuppressRevisions() must not throw error")
					return
				}
			}

			for i := 0; i < tt.forceDrain; i++ {
				batching.forceDrainRequest <- true
			}

			time.Sleep(time.Millisecond)

			if tt.suppressValues != standard.callHistory {
				t.Errorf("SuppressRevisions() call pattern [%s], expected [%s]", standard.callHistory, tt.suppressValues)
			}
		})
	}
}
