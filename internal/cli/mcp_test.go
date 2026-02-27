package cli

import "testing"

func TestMcpCommand(t *testing.T) {
	if mcpCmd == nil {
		t.Fatal("mcpCmd should not be nil")
	}
	if mcpCmd.Use != "mcp" {
		t.Errorf("expected Use=%q, got %q", "mcp", mcpCmd.Use)
	}
	if mcpCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
	if mcpCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestMcpCommandProperties(t *testing.T) {
	if !mcpCmd.SilenceUsage {
		t.Error("SilenceUsage should be true")
	}
	if !mcpCmd.SilenceErrors {
		t.Error("SilenceErrors should be true")
	}
}

func TestMcpCommandRegistration(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "mcp" {
			found = true
			break
		}
	}
	if !found {
		t.Error("mcp command should be registered with root command")
	}
}

func TestResolveXdebugBinary(t *testing.T) {
	path, err := resolveXdebugBinary()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty binary path")
	}
}
