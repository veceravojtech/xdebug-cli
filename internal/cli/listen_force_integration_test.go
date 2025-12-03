package cli

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/console/xdebug-cli/internal/daemon"
)

func TestKillDaemonOnPort(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test")
	}

	// Test 1: Kill non-existent daemon (should show warning)
	t.Run("kill non-existent daemon", func(t *testing.T) {
		killDaemonOnPort(8888) // Port with no daemon
		// Should return without error
	})

	// Test 2: Kill stale daemon (PID doesn't exist)
	t.Run("kill stale daemon", func(t *testing.T) {
		registry, err := daemon.NewSessionRegistry()
		if err != nil {
			t.Fatalf("Failed to create registry: %v", err)
		}

		testPort := 9998
		stalePID := 99999 // PID that likely doesn't exist

		sessionInfo := daemon.SessionInfo{
			PID:        stalePID,
			Port:       testPort,
			SocketPath: fmt.Sprintf("/tmp/test-xdebug-cli-%d.sock", testPort),
			StartedAt:  time.Now(),
		}

		// Add to registry
		err = registry.Add(sessionInfo)
		if err != nil {
			t.Fatalf("Failed to add session: %v", err)
		}

		// Clean up after test
		defer registry.Remove(testPort)

		// Should detect stale process and clean up
		killDaemonOnPort(testPort)
		// Should return without error
	})
}
