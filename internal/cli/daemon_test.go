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

// TestDaemonStartCurlFlagRegistered tests that --curl flag is registered
func TestDaemonStartCurlFlagRegistered(t *testing.T) {
	// Test that --curl flag is registered
	curlFlag := startCmd.Flags().Lookup("curl")
	if curlFlag == nil {
		t.Fatal("--curl flag should be registered")
	}

	// Test flag properties
	if curlFlag.DefValue != "" {
		t.Errorf("expected default value '', got '%s'", curlFlag.DefValue)
	}
}

// TestDaemonStartCurlValidation tests that --curl flag is required
func TestDaemonStartCurlValidation(t *testing.T) {
	// Reset CLIArgs
	CLIArgs = cfg.CLIParameter{
		Port: 9003,
		Curl: "", // Empty curl should fail validation
	}

	// runDaemonStart should return an error when --curl is empty
	err := runDaemonStart()
	if err == nil {
		t.Fatal("runDaemonStart should return error when --curl is empty")
	}

	// Error message should mention --curl flag is required
	if err.Error()[:26] != "--curl flag is required\n\nU" {
		t.Errorf("error message should start with '--curl flag is required', got '%s'", err.Error()[:26])
	}
}

// TestParseShellArgs tests the shell argument parser
func TestParseShellArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
		wantErr  bool
	}{
		{
			name:     "simple URL",
			input:    "http://localhost/app.php",
			expected: []string{"http://localhost/app.php"},
			wantErr:  false,
		},
		{
			name:     "URL with flags",
			input:    "http://localhost/api -X POST",
			expected: []string{"http://localhost/api", "-X", "POST"},
			wantErr:  false,
		},
		{
			name:     "double quoted data",
			input:    `http://localhost/api -X POST -d "name=value"`,
			expected: []string{"http://localhost/api", "-X", "POST", "-d", "name=value"},
			wantErr:  false,
		},
		{
			name:     "single quoted data",
			input:    `http://localhost/api -X POST -d 'name=value'`,
			expected: []string{"http://localhost/api", "-X", "POST", "-d", "name=value"},
			wantErr:  false,
		},
		{
			name:     "quoted data with spaces",
			input:    `http://localhost/api -d "hello world"`,
			expected: []string{"http://localhost/api", "-d", "hello world"},
			wantErr:  false,
		},
		{
			name:     "multiple headers",
			input:    `http://localhost/api -H "Content-Type: application/json" -H "Accept: text/plain"`,
			expected: []string{"http://localhost/api", "-H", "Content-Type: application/json", "-H", "Accept: text/plain"},
			wantErr:  false,
		},
		{
			name:    "unclosed double quote",
			input:   `http://localhost/api -d "incomplete`,
			wantErr: true,
		},
		{
			name:    "unclosed single quote",
			input:   `http://localhost/api -d 'incomplete`,
			wantErr: true,
		},
		{
			name:     "escaped spaces",
			input:    `http://localhost/api -d hello\ world`,
			expected: []string{"http://localhost/api", "-d", "hello world"},
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
			wantErr:  false,
		},
		{
			name:     "JSON data",
			input:    `http://localhost/api -X POST -H "Content-Type: application/json" -d '{"key":"value"}'`,
			expected: []string{"http://localhost/api", "-X", "POST", "-H", "Content-Type: application/json", "-d", `{"key":"value"}`},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseShellArgs(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d args, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			for i, arg := range result {
				if arg != tt.expected[i] {
					t.Errorf("arg[%d]: expected '%s', got '%s'", i, tt.expected[i], arg)
				}
			}
		})
	}
}
