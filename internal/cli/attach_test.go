package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/console/xdebug-cli/internal/daemon"
)

func TestAttachCommand(t *testing.T) {
	// Verify that attach command exists
	if attachCmd == nil {
		t.Fatal("attachCmd should not be nil")
	}

	// Verify command properties
	if attachCmd.Use != "attach" {
		t.Errorf("expected Use='attach', got %q", attachCmd.Use)
	}

	if attachCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if attachCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestAttachCommandFlags(t *testing.T) {
	// Verify --commands flag exists
	commandsFlag := attachCmd.Flags().Lookup("commands")
	if commandsFlag == nil {
		t.Fatal("--commands flag should be registered")
	}

	if commandsFlag.Value.Type() != "stringArray" {
		t.Errorf("expected --commands to be stringArray, got %s", commandsFlag.Value.Type())
	}

	// Verify --json flag is available (inherited from root)
	jsonFlag := attachCmd.InheritedFlags().Lookup("json")
	if jsonFlag == nil {
		t.Error("--json flag should be available (inherited from root)")
	}

	// Verify --port flag is available (inherited from root)
	portFlag := attachCmd.InheritedFlags().Lookup("port")
	if portFlag == nil {
		t.Error("--port flag should be available (inherited from root)")
	}
}

func TestAttachCommandRegistration(t *testing.T) {
	// Verify attach command is registered with root command
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "attach" {
			found = true
			break
		}
	}

	if !found {
		t.Error("attach command should be registered with root command")
	}
}

func TestRunAttachCmd_NoCommands(t *testing.T) {
	// Save original commands and restore after test
	originalCommands := CLIArgs.Commands
	defer func() { CLIArgs.Commands = originalCommands }()

	// Set empty commands
	CLIArgs.Commands = []string{}

	// Should return error when no commands provided
	err := runAttachCmd()
	if err == nil {
		t.Error("expected error when no commands provided")
	}

	expectedMsg := "--commands flag is required"
	if err != nil && err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("expected error message to start with %q, got %q", expectedMsg, err.Error())
	}
}

func TestRunAttachCmd_NoRegistry(t *testing.T) {
	// Save original values
	originalCommands := CLIArgs.Commands
	originalPort := CLIArgs.Port
	defer func() {
		CLIArgs.Commands = originalCommands
		CLIArgs.Port = originalPort
	}()

	// Set test values
	CLIArgs.Commands = []string{"run"}
	CLIArgs.Port = 9999 // Port that doesn't exist in registry

	// Create a temporary registry directory for testing
	tmpDir, err := os.MkdirTemp("", "xdebug-cli-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set HOME to temp dir so registry uses it
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Should return error when no daemon is running
	err = runAttachCmd()
	if err == nil {
		t.Error("expected error when no daemon is running")
	}

	expectedMsg := "no daemon running on port"
	if err != nil && err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("expected error message to start with %q, got %q", expectedMsg, err.Error())
	}
}

func TestRunAttachCmd_Integration(t *testing.T) {
	// This is a placeholder for integration testing
	// Real integration tests would require:
	// 1. Starting an actual daemon
	// 2. Registering it in the session registry
	// 3. Testing attach with real IPC communication
	// 4. Verifying command execution results

	// For now, we'll test the registry interaction
	tmpDir, err := os.MkdirTemp("", "xdebug-cli-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set HOME to temp dir
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create registry
	registry, err := daemon.NewSessionRegistry()
	if err != nil {
		t.Fatalf("failed to create registry: %v", err)
	}

	// Add a test session
	socketPath := filepath.Join(tmpDir, ".xdebug-cli", "daemon-9003.sock")
	err = registry.Add(daemon.SessionInfo{
		PID:        os.Getpid(), // Use current process PID
		Port:       9003,
		SocketPath: socketPath,
	})
	if err != nil {
		t.Fatalf("failed to add session to registry: %v", err)
	}

	// Verify the session can be retrieved
	session, err := registry.Get(9003)
	if err != nil {
		t.Errorf("failed to get session: %v", err)
	}

	if session.Port != 9003 {
		t.Errorf("expected port 9003, got %d", session.Port)
	}

	if session.SocketPath != socketPath {
		t.Errorf("expected socket path %q, got %q", socketPath, session.SocketPath)
	}
}

func TestAttachCommandExamples(t *testing.T) {
	// Verify that the Long description contains usage examples
	examples := []string{
		"xdebug-cli attach --commands",
		"context local",
		"print \\$myVar",
		"run",
		"--json",
	}

	for _, example := range examples {
		if !contains(attachCmd.Long, example) {
			t.Errorf("Long description should contain example: %q", example)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
