package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// BotProcess manages the external bot binary lifecycle
type BotProcess struct {
	cmd    *exec.Cmd
	ctx    context.Context
	cancel context.CancelFunc
}

// StartBot builds and runs the bot binary as an external process
func StartBot(t *testing.T, backend Backend) *BotProcess {
	t.Helper()

	// Build binary (Go caches compilation for fast rebuilds)
	binPath := filepath.Join("bin", "freedom-sentry-test")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = getRepoRoot()
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build bot binary: %v", err)
	}

	// Prepare environment with backend endpoints
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, filepath.Join(getRepoRoot(), binPath))
	cmd.Env = append(os.Environ(),
		"API_ENDPOINT="+backend.APIEndpoint(),
		"EVENTSTREAMS_URL="+backend.EventStreamsEndpoint(),
		"WIKI_DOMAIN="+backend.WikiDomain(),
		"LIST_NAME="+backend.ListPageName(),
		"ACCESS_TOKEN=test-token",
	)

	// Capture output for debugging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = getRepoRoot()

	// Start process
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("Failed to start bot: %v", err)
	}

	return &BotProcess{
		cmd:    cmd,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Stop terminates the bot process gracefully
func (bp *BotProcess) Stop() {
	bp.cancel()
	bp.cmd.Wait()
}

// WaitForReady waits for the bot to be fully initialized
func WaitForReady(duration time.Duration) {
	time.Sleep(duration)
}

// getRepoRoot returns the repository root directory
func getRepoRoot() string {
	// E2E tests are in e2e/ directory, so go up one level
	if wd, err := os.Getwd(); err == nil {
		if filepath.Base(wd) == "e2e" {
			return filepath.Dir(wd)
		}
	}
	return "."
}
