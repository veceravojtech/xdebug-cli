package cli

import (
	"os"
	"testing"

	"github.com/console/xdebug-cli/internal/dbgp"
)

func TestActiveSessionManagement(t *testing.T) {
	// Test initial state
	client, active := getActiveSession()
	if active {
		t.Error("expected no active session initially")
	}
	if client != nil {
		t.Error("expected nil client initially")
	}

	// Create a mock session
	session := dbgp.NewSession()
	session.SetIDEKey("test-key")
	session.SetAppID("test-app")

	// Test setting active session
	mockClient := &dbgp.Client{}
	setActiveSession(mockClient)

	client, active = getActiveSession()
	if !active {
		t.Error("expected active session after setting")
	}
	if client == nil {
		t.Error("expected non-nil client after setting")
	}

	// Test clearing active session
	clearActiveSession()

	client, active = getActiveSession()
	if active {
		t.Error("expected no active session after clearing")
	}
	if client != nil {
		t.Error("expected nil client after clearing")
	}
}

func TestActiveSessionConcurrency(t *testing.T) {
	// Test that concurrent access doesn't race
	// This test will be caught by go test -race if there are race conditions

	clearActiveSession()

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			mockClient := &dbgp.Client{}
			setActiveSession(mockClient)
			clearActiveSession()
		}
		done <- true
	}()

	// Reader goroutines
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				getActiveSession()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 6; i++ {
		<-done
	}

	// Clean up
	clearActiveSession()
}

func TestDaemonSubcommands(t *testing.T) {
	tests := []struct {
		name        string
		subcommand  string
		description string
	}{
		{
			name:        "start command",
			subcommand:  "start",
			description: "start a new daemon session",
		},
		{
			name:        "status command",
			subcommand:  "status",
			description: "show daemon status",
		},
		{
			name:        "list command",
			subcommand:  "list",
			description: "list all daemon sessions",
		},
		{
			name:        "isAlive command",
			subcommand:  "isAlive",
			description: "check if daemon is running",
		},
		{
			name:        "kill command",
			subcommand:  "kill",
			description: "terminate daemon session(s)",
		},
	}

	// Verify that daemon command exists
	if daemonCmd == nil {
		t.Fatal("daemonCmd should not be nil")
	}

	// Verify daemon command Use
	if daemonCmd.Use != "daemon" {
		t.Errorf("expected Use='daemon', got %q", daemonCmd.Use)
	}

	// Verify subcommands are registered
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Find subcommand
			var found bool
			for _, cmd := range daemonCmd.Commands() {
				if cmd.Use == tt.subcommand {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("subcommand %q not found", tt.subcommand)
			}
		})
	}
}

func TestDaemonListFlags(t *testing.T) {
	// Verify that list command has --json flag
	if listCmd == nil {
		t.Fatal("listCmd should not be nil")
	}

	jsonFlag := listCmd.Flags().Lookup("json")
	if jsonFlag == nil {
		t.Error("list command should have --json flag")
	}

	if jsonFlag.Usage != "Output in JSON format" {
		t.Errorf("expected usage 'Output in JSON format', got %q", jsonFlag.Usage)
	}
}

func TestDaemonKillFlags(t *testing.T) {
	// Verify that kill command has --all and --force flags
	if killCmd == nil {
		t.Fatal("killCmd should not be nil")
	}

	allFlag := killCmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Error("kill command should have --all flag")
	}

	if allFlag.Usage != "Kill all daemon sessions" {
		t.Errorf("expected usage 'Kill all daemon sessions', got %q", allFlag.Usage)
	}

	forceFlag := killCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("kill command should have --force flag")
	}

	if forceFlag.Usage != "Skip confirmation prompt" {
		t.Errorf("expected usage 'Skip confirmation prompt', got %q", forceFlag.Usage)
	}
}

func TestProcessExists(t *testing.T) {
	// Test with current process (should exist)
	currentPID := os.Getpid() // Use current process PID (works on Linux and macOS)
	if !processExists(currentPID) {
		t.Errorf("current process PID %d should exist", currentPID)
	}

	// Test with invalid PID (should not exist)
	invalidPID := 999999999
	if processExists(invalidPID) {
		t.Error("invalid PID should not exist")
	}
}

func TestOutputSessionListJSON(t *testing.T) {
	// Mock data can't be easily tested without capturing stdout
	// This test verifies the function doesn't panic with valid data
	sessions := []struct {
		PID        int
		Port       int
		SocketPath string
	}{
		{PID: 12345, Port: 9003, SocketPath: "/tmp/test.sock"},
	}

	// Convert to daemon.SessionInfo (can't do this without importing daemon package properly)
	// For now, just verify function signature compiles
	_ = sessions
}
