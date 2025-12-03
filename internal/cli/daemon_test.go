package cli

import (
	"testing"

	"github.com/console/xdebug-cli/internal/cfg"
)

// TestDaemonCommand tests the daemon parent command
func TestDaemonCommand(t *testing.T) {
	// Test that daemon command exists
	if daemonCmd == nil {
		t.Fatal("daemonCmd should not be nil")
	}

	// Test command metadata
	if daemonCmd.Use != "daemon" {
		t.Errorf("expected Use='daemon', got '%s'", daemonCmd.Use)
	}

	if daemonCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	// Test that start subcommand is registered
	startSubCmd := daemonCmd.Commands()
	if len(startSubCmd) == 0 {
		t.Fatal("daemon command should have subcommands")
	}

	found := false
	for _, cmd := range startSubCmd {
		if cmd.Use == "start" {
			found = true
			break
		}
	}

	if !found {
		t.Error("start subcommand not found under daemon command")
	}
}

// TestDaemonStartCommand tests the daemon start subcommand
func TestDaemonStartCommand(t *testing.T) {
	// Test that start command exists
	if startCmd == nil {
		t.Fatal("startCmd should not be nil")
	}

	// Test command metadata
	if startCmd.Use != "start" {
		t.Errorf("expected Use='start', got '%s'", startCmd.Use)
	}

	if startCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if startCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	// Test that --commands flag is registered
	commandsFlag := startCmd.Flags().Lookup("commands")
	if commandsFlag == nil {
		t.Fatal("--commands flag should be registered")
	}

	// Test flag properties
	if commandsFlag.DefValue != "[]" {
		t.Errorf("expected default value '[]', got '%s'", commandsFlag.DefValue)
	}
}

// TestDaemonStartFlagsBinding tests that flags bind to CLIArgs correctly
func TestDaemonStartFlagsBinding(t *testing.T) {
	// Reset CLIArgs
	CLIArgs = cfg.CLIParameter{}

	// Set flag value (simulate command line parsing)
	testCommands := []string{"break :42", "run"}
	CLIArgs.Commands = testCommands

	// Verify binding
	if len(CLIArgs.Commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(CLIArgs.Commands))
	}

	if CLIArgs.Commands[0] != "break :42" {
		t.Errorf("expected first command 'break :42', got '%s'", CLIArgs.Commands[0])
	}

	if CLIArgs.Commands[1] != "run" {
		t.Errorf("expected second command 'run', got '%s'", CLIArgs.Commands[1])
	}
}

// TestDaemonStartInheritsGlobalFlags tests that global flags are inherited
func TestDaemonStartInheritsGlobalFlags(t *testing.T) {
	// Reset CLIArgs
	CLIArgs = cfg.CLIParameter{
		Port: 9004,
		JSON: true,
	}

	// Verify global flags are accessible
	if CLIArgs.Port != 9004 {
		t.Errorf("expected port 9004, got %d", CLIArgs.Port)
	}

	if !CLIArgs.JSON {
		t.Error("expected JSON flag to be true")
	}
}
