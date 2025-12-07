package mwemu

import (
	"testing"
	"time"
)

// AssertRevisionSuppressed polls for a revision to be suppressed within the timeout
func (e *Emulator) AssertRevisionSuppressed(t *testing.T, revID int64, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	pollInterval := 10 * time.Millisecond

	for time.Now().Before(deadline) {
		if e.state.IsRevisionSuppressed(revID) {
			t.Logf("Revision %d successfully suppressed", revID)
			return
		}
		time.Sleep(pollInterval)
	}

	t.Errorf("Revision %d was not suppressed within %v", revID, timeout)
}

// AssertNoSuppressionCalls verifies that no revisiondelete calls were made during the duration
func (e *Emulator) AssertNoSuppressionCalls(t *testing.T, duration time.Duration) {
	t.Helper()

	initialCount := e.countAPICalls("revisiondelete")
	time.Sleep(duration)
	finalCount := e.countAPICalls("revisiondelete")

	if finalCount > initialCount {
		t.Errorf("Expected no suppression calls, but %d calls were made", finalCount-initialCount)
	} else {
		t.Logf("No suppression calls made during %v", duration)
	}
}

// AssertAPICallCount verifies the number of calls to a specific API action
func (e *Emulator) AssertAPICallCount(t *testing.T, action string, expected int) {
	t.Helper()

	actual := e.countAPICalls(action)
	if actual != expected {
		t.Errorf("Expected %d calls to action=%s, got %d", expected, action, actual)
	} else {
		t.Logf("API call count for action=%s: %d (as expected)", action, actual)
	}
}

// AssertAPICallEventually waits for a specific number of API calls within the timeout
func (e *Emulator) AssertAPICallEventually(t *testing.T, action string, expected int, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	pollInterval := 10 * time.Millisecond

	for time.Now().Before(deadline) {
		actual := e.countAPICalls(action)
		if actual >= expected {
			t.Logf("API call count for action=%s reached %d within %v", action, actual, time.Since(deadline.Add(-timeout)))
			return
		}
		time.Sleep(pollInterval)
	}

	actual := e.countAPICalls(action)
	t.Errorf("Expected at least %d calls to action=%s within %v, got %d", expected, action, timeout, actual)
}

// countAPICalls returns the number of calls to a specific action
func (e *Emulator) countAPICalls(action string) int {
	history := e.state.GetAPICallHistory()
	count := 0
	for _, call := range history {
		if call.Action == action {
			count++
		}
	}
	return count
}

// GetAPICallHistory returns the full API call history
func (e *Emulator) GetAPICallHistory() []APICall {
	return e.state.GetAPICallHistory()
}

// GetState returns the emulator's state for custom assertions
func (e *Emulator) GetState() *State {
	return e.state
}

// AssertPageExists verifies that a page exists in the emulator
func (e *Emulator) AssertPageExists(t *testing.T, title string) {
	t.Helper()

	page := e.state.GetPageByTitle(title)
	if page == nil {
		t.Errorf("Page %q does not exist in emulator", title)
	} else {
		t.Logf("Page %q exists with %d revisions", title, len(page.Revisions))
	}
}

// AssertRevisionCount verifies the number of revisions for a page
func (e *Emulator) AssertRevisionCount(t *testing.T, title string, expected int) {
	t.Helper()

	page := e.state.GetPageByTitle(title)
	if page == nil {
		t.Errorf("Page %q does not exist in emulator", title)
		return
	}

	actual := len(page.Revisions)
	if actual != expected {
		t.Errorf("Expected %d revisions for page %q, got %d", expected, title, actual)
	} else {
		t.Logf("Page %q has %d revisions (as expected)", title, actual)
	}
}
