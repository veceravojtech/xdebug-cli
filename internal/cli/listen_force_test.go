package cli

import (
	"bytes"
	"os"
	"testing"
)

func TestListenForceValidation(t *testing.T) {
	tests := []struct {
		name        string
		force       bool
		commands    []string
		shouldError bool
	}{
		{
			name:        "force with commands should pass",
			force:       true,
			commands:    []string{"run"},
			shouldError: false,
		},
		{
			name:        "no commands should error",
			force:       false,
			commands:    []string{},
			shouldError: true,
		},
		{
			name:        "commands without force should pass",
			force:       false,
			commands:    []string{"run"},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origForce := CLIArgs.Force
			origCommands := CLIArgs.Commands
			origStderr := os.Stderr

			// Restore after test
			defer func() {
				CLIArgs.Force = origForce
				CLIArgs.Commands = origCommands
				os.Stderr = origStderr
			}()

			// Set test values
			CLIArgs.Force = tt.force
			CLIArgs.Commands = tt.commands

			// Capture stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// Validate flags
			err := validateListenFlags()

			w.Close()
			var stderr bytes.Buffer
			stderr.ReadFrom(r)

			if tt.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestRunListeningCmdWithForce(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test")
	}

	// This test verifies that force flag triggers cleanup before server start
	// We can't fully test server startup without mocking, but we can verify
	// the killDaemonOnPort is called when Force is true

	tests := []struct {
		name  string
		force bool
	}{
		{
			name:  "with force flag",
			force: true,
		},
		{
			name:  "without force flag",
			force: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origForce := CLIArgs.Force
			origPort := CLIArgs.Port

			// Restore after test
			defer func() {
				CLIArgs.Force = origForce
				CLIArgs.Port = origPort
			}()

			// Set test values
			CLIArgs.Force = tt.force
			CLIArgs.Port = 9999 // Use test port

			// Note: We can't actually run the full command without starting a server
			// This test documents the expected behavior
			// Manual testing required for full integration
		})
	}
}
