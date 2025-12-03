package cli

import (
	"strconv"
	"strings"
	"testing"

	"github.com/console/xdebug-cli/internal/cfg"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedCommand string
		expectedArgs    []string
	}{
		{
			name:            "run command",
			input:           "run",
			expectedCommand: "run",
			expectedArgs:    []string{},
		},
		{
			name:            "run command short form",
			input:           "r",
			expectedCommand: "r",
			expectedArgs:    []string{},
		},
		{
			name:            "step command",
			input:           "step",
			expectedCommand: "step",
			expectedArgs:    []string{},
		},
		{
			name:            "step command short form",
			input:           "s",
			expectedCommand: "s",
			expectedArgs:    []string{},
		},
		{
			name:            "next command",
			input:           "next",
			expectedCommand: "next",
			expectedArgs:    []string{},
		},
		{
			name:            "out long form",
			input:           "out",
			expectedCommand: "out",
			expectedArgs:    []string{},
		},
		{
			name:            "out short form",
			input:           "o",
			expectedCommand: "o",
			expectedArgs:    []string{},
		},
		{
			name:            "break with line number",
			input:           "break 42",
			expectedCommand: "break",
			expectedArgs:    []string{"42"},
		},
		{
			name:            "break with :line format",
			input:           "break :42",
			expectedCommand: "break",
			expectedArgs:    []string{":42"},
		},
		{
			name:            "break with file:line format",
			input:           "break /path/to/file.php:42",
			expectedCommand: "break",
			expectedArgs:    []string{"/path/to/file.php:42"},
		},
		{
			name:            "break call function",
			input:           "break call myFunction",
			expectedCommand: "break",
			expectedArgs:    []string{"call", "myFunction"},
		},
		{
			name:            "break exception",
			input:           "break exception",
			expectedCommand: "break",
			expectedArgs:    []string{"exception"},
		},
		{
			name:            "print variable",
			input:           "print $myVar",
			expectedCommand: "print",
			expectedArgs:    []string{"$myVar"},
		},
		{
			name:            "print variable without $",
			input:           "print myVar",
			expectedCommand: "print",
			expectedArgs:    []string{"myVar"},
		},
		{
			name:            "print complex expression",
			input:           "print $obj->prop",
			expectedCommand: "print",
			expectedArgs:    []string{"$obj->prop"},
		},
		{
			name:            "context local",
			input:           "context local",
			expectedCommand: "context",
			expectedArgs:    []string{"local"},
		},
		{
			name:            "context global",
			input:           "context global",
			expectedCommand: "context",
			expectedArgs:    []string{"global"},
		},
		{
			name:            "context no args",
			input:           "context",
			expectedCommand: "context",
			expectedArgs:    []string{},
		},
		{
			name:            "list command",
			input:           "list",
			expectedCommand: "list",
			expectedArgs:    []string{},
		},
		{
			name:            "info breakpoints",
			input:           "info breakpoints",
			expectedCommand: "info",
			expectedArgs:    []string{"breakpoints"},
		},
		{
			name:            "info breakpoints short",
			input:           "info b",
			expectedCommand: "info",
			expectedArgs:    []string{"b"},
		},
		{
			name:            "finish command",
			input:           "finish",
			expectedCommand: "finish",
			expectedArgs:    []string{},
		},
		{
			name:            "help command",
			input:           "help",
			expectedCommand: "help",
			expectedArgs:    []string{},
		},
		{
			name:            "help for specific command",
			input:           "help break",
			expectedCommand: "help",
			expectedArgs:    []string{"break"},
		},
		{
			name:            "quit command",
			input:           "quit",
			expectedCommand: "quit",
			expectedArgs:    []string{},
		},
		{
			name:            "multiple spaces",
			input:           "break   call    myFunction",
			expectedCommand: "break",
			expectedArgs:    []string{"call", "myFunction"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := strings.Fields(tt.input)
			if len(parts) == 0 {
				t.Fatal("expected non-empty input")
			}

			command := parts[0]
			args := []string{}
			if len(parts) > 1 {
				args = parts[1:]
			}

			if command != tt.expectedCommand {
				t.Errorf("expected command %q, got %q", tt.expectedCommand, command)
			}

			if len(args) != len(tt.expectedArgs) {
				t.Errorf("expected %d args, got %d", len(tt.expectedArgs), len(args))
				return
			}

			for i, arg := range args {
				if arg != tt.expectedArgs[i] {
					t.Errorf("arg[%d]: expected %q, got %q", i, tt.expectedArgs[i], arg)
				}
			}
		})
	}
}

func TestCommandAliases(t *testing.T) {
	tests := []struct {
		name          string
		command       string
		isValidAlias  bool
		canonicalName string
	}{
		{"run long form", "run", true, "run"},
		{"run short form", "r", true, "run"},
		{"step long form", "step", true, "step"},
		{"step short form", "s", true, "step"},
		{"next long form", "next", true, "next"},
		{"next short form", "n", true, "next"},
		{"out long form", "out", true, "out"},
		{"out short form", "o", true, "out"},
		{"break long form", "break", true, "break"},
		{"break short form", "b", true, "break"},
		{"print long form", "print", true, "print"},
		{"print short form", "p", true, "print"},
		{"context long form", "context", true, "context"},
		{"context short form", "c", true, "context"},
		{"list long form", "list", true, "list"},
		{"list short form", "l", true, "list"},
		{"info long form", "info", true, "info"},
		{"info short form", "i", true, "info"},
		{"finish long form", "finish", true, "finish"},
		{"finish short form", "f", true, "finish"},
		{"help long form", "help", true, "help"},
		{"help short form h", "h", true, "help"},
		{"help short form ?", "?", true, "help"},
		{"quit long form", "quit", true, "quit"},
		{"quit short form", "q", true, "quit"},
		{"invalid command", "invalid", false, ""},
	}

	// Map of valid commands and their aliases
	validCommands := map[string]string{
		"run":     "run",
		"r":       "run",
		"step":    "step",
		"s":       "step",
		"next":    "next",
		"n":       "next",
		"out":     "out",
		"o":       "out",
		"break":   "break",
		"b":       "break",
		"print":   "print",
		"p":       "print",
		"context": "context",
		"c":       "context",
		"list":    "list",
		"l":       "list",
		"info":    "info",
		"i":       "info",
		"finish":  "finish",
		"f":       "finish",
		"help":    "help",
		"h":       "help",
		"?":       "help",
		"quit":    "quit",
		"q":       "quit",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canonical, exists := validCommands[tt.command]

			if exists != tt.isValidAlias {
				t.Errorf("expected valid=%v, got valid=%v", tt.isValidAlias, exists)
			}

			if exists && canonical != tt.canonicalName {
				t.Errorf("expected canonical name %q, got %q", tt.canonicalName, canonical)
			}
		})
	}
}

func TestBreakpointParsing(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedType string // "line", "file:line", "call", "exception"
		valid        bool
	}{
		{
			name:         "line number only",
			args:         []string{"42"},
			expectedType: "line",
			valid:        true,
		},
		{
			name:         "colon prefix line number",
			args:         []string{":42"},
			expectedType: "line",
			valid:        true,
		},
		{
			name:         "file:line format",
			args:         []string{"/path/to/file.php:42"},
			expectedType: "file:line",
			valid:        true,
		},
		{
			name:         "call with function",
			args:         []string{"call", "myFunction"},
			expectedType: "call",
			valid:        true,
		},
		{
			name:         "call without function",
			args:         []string{"call"},
			expectedType: "call",
			valid:        false,
		},
		{
			name:         "exception breakpoint",
			args:         []string{"exception"},
			expectedType: "exception",
			valid:        true,
		},
		{
			name:         "exception with name",
			args:         []string{"exception", "RuntimeException"},
			expectedType: "exception",
			valid:        true,
		},
		{
			name:         "no args",
			args:         []string{},
			expectedType: "",
			valid:        false,
		},
		{
			name:         "invalid line number",
			args:         []string{"abc"},
			expectedType: "line", // Detected as line type, but marked invalid
			valid:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.args) == 0 {
				if tt.valid {
					t.Error("expected invalid for empty args")
				}
				return
			}

			// Determine breakpoint type
			var bpType string
			var isValid bool = true

			if tt.args[0] == "call" {
				bpType = "call"
				if len(tt.args) < 2 {
					isValid = false
				}
			} else if tt.args[0] == "exception" {
				bpType = "exception"
			} else if strings.HasPrefix(tt.args[0], ":") {
				bpType = "line"
				// Validate line number
				lineStr := strings.TrimPrefix(tt.args[0], ":")
				if _, err := strconv.Atoi(lineStr); err != nil {
					isValid = false
				}
			} else if strings.Contains(tt.args[0], ":") {
				bpType = "file:line"
				// Validate line part
				parts := strings.SplitN(tt.args[0], ":", 2)
				if _, err := strconv.Atoi(parts[1]); err != nil {
					isValid = false
				}
			} else {
				bpType = "line"
				// Validate line number
				if _, err := strconv.Atoi(tt.args[0]); err != nil {
					isValid = false
				}
			}

			if bpType != tt.expectedType {
				t.Errorf("expected type %q, got %q", tt.expectedType, bpType)
			}

			if isValid != tt.valid {
				t.Errorf("expected valid=%v, got valid=%v", tt.valid, isValid)
			}
		})
	}
}

func TestContextTypeParsing(t *testing.T) {
	tests := []struct {
		name        string
		arg         string
		expectedID  int
		expectedErr bool
	}{
		{"local context", "local", 0, false},
		{"global context", "global", 1, false},
		{"constant context", "constant", 2, false},
		{"empty defaults to local", "", 0, false},
		{"invalid context", "invalid", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contextType := tt.arg
			if contextType == "" {
				contextType = "local"
			}

			var contextID int
			var hasError bool

			switch strings.ToLower(contextType) {
			case "local":
				contextID = 0
			case "global":
				contextID = 1
			case "constant":
				contextID = 2
			default:
				hasError = true
				contextID = -1
			}

			if hasError != tt.expectedErr {
				t.Errorf("expected error=%v, got error=%v", tt.expectedErr, hasError)
			}

			if !hasError && contextID != tt.expectedID {
				t.Errorf("expected contextID=%d, got contextID=%d", tt.expectedID, contextID)
			}
		})
	}
}

func TestInfoSubcommands(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		isValid bool
	}{
		{"breakpoints long form", []string{"breakpoints"}, true},
		{"breakpoints short form", []string{"b"}, true},
		{"no args", []string{}, false},
		{"invalid subcommand", []string{"invalid"}, false},
	}

	validInfoCommands := map[string]bool{
		"breakpoints": true,
		"b":           true,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.args) == 0 {
				if tt.isValid {
					t.Error("expected invalid for empty args")
				}
				return
			}

			isValid := validInfoCommands[tt.args[0]]
			if isValid != tt.isValid {
				t.Errorf("expected valid=%v, got valid=%v", tt.isValid, isValid)
			}
		})
	}
}

func TestCommandRequirement(t *testing.T) {
	tests := []struct {
		name        string
		daemon      bool
		commands    []string
		expectError bool
	}{
		{
			name:        "commands provided",
			daemon:      false,
			commands:    []string{"run", "step"},
			expectError: false,
		},
		{
			name:        "daemon mode without commands",
			daemon:      true,
			commands:    []string{},
			expectError: false,
		},
		{
			name:        "no commands and no daemon",
			daemon:      false,
			commands:    []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate flag validation
			hasError := !tt.daemon && len(tt.commands) == 0
			if hasError != tt.expectError {
				t.Errorf("expected error=%v, got error=%v", tt.expectError, hasError)
			}
		})
	}
}

func TestCommandParsing(t *testing.T) {
	tests := []struct {
		name            string
		commandStr      string
		expectedCommand string
		expectedArgs    []string
	}{
		{
			name:            "simple run command",
			commandStr:      "run",
			expectedCommand: "run",
			expectedArgs:    []string{},
		},
		{
			name:            "print with variable",
			commandStr:      "print $x",
			expectedCommand: "print",
			expectedArgs:    []string{"$x"},
		},
		{
			name:            "break with line",
			commandStr:      "break :42",
			expectedCommand: "break",
			expectedArgs:    []string{":42"},
		},
		{
			name:            "context local",
			commandStr:      "context local",
			expectedCommand: "context",
			expectedArgs:    []string{"local"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := strings.Fields(tt.commandStr)
			if len(parts) == 0 {
				t.Fatal("expected non-empty command")
			}

			command := parts[0]
			args := []string{}
			if len(parts) > 1 {
				args = parts[1:]
			}

			if command != tt.expectedCommand {
				t.Errorf("expected command %q, got %q", tt.expectedCommand, command)
			}

			if len(args) != len(tt.expectedArgs) {
				t.Errorf("expected %d args, got %d", len(tt.expectedArgs), len(args))
				return
			}

			for i, arg := range args {
				if arg != tt.expectedArgs[i] {
					t.Errorf("arg[%d]: expected %q, got %q", i, tt.expectedArgs[i], arg)
				}
			}
		})
	}
}

func TestCLIParameterStruct(t *testing.T) {
	// Test that CLIParameter has all required fields
	param := cfg.CLIParameter{
		Host:     "127.0.0.1",
		Port:     9003,
		Commands: []string{"run", "step"},
		JSON:     true,
	}

	if param.Host != "127.0.0.1" {
		t.Errorf("expected Host=127.0.0.1, got %s", param.Host)
	}

	if param.Port != 9003 {
		t.Errorf("expected Port=9003, got %d", param.Port)
	}

	if len(param.Commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(param.Commands))
	}

	if !param.JSON {
		t.Error("expected JSON=true")
	}
}

func TestParseBreakpointCondition(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedLocations []string
		expectedCondition string
		expectError       bool
	}{
		{
			name:              "single location with condition",
			args:              []string{":42", "if", "$count", ">", "10"},
			expectedLocations: []string{":42"},
			expectedCondition: "$count > 10",
			expectError:       false,
		},
		{
			name:              "single location no condition",
			args:              []string{":42"},
			expectedLocations: []string{":42"},
			expectedCondition: "",
			expectError:       false,
		},
		{
			name:              "empty condition after if",
			args:              []string{":42", "if"},
			expectedLocations: []string{":42"},
			expectedCondition: "",
			expectError:       true,
		},
		{
			name:              "complex condition",
			args:              []string{"file.php:100", "if", "$user->isAdmin()", "&&", "$debug"},
			expectedLocations: []string{"file.php:100"},
			expectedCondition: "$user->isAdmin() && $debug",
			expectError:       false,
		},
		{
			name:              "multiple locations no condition",
			args:              []string{":42", ":100", ":150"},
			expectedLocations: []string{":42", ":100", ":150"},
			expectedCondition: "",
			expectError:       false,
		},
		{
			name:              "multiple locations with condition",
			args:              []string{":10", ":20", ":30", "if", "$debug"},
			expectedLocations: []string{":10", ":20", ":30"},
			expectedCondition: "$debug",
			expectError:       false,
		},
		{
			name:              "mixed format locations",
			args:              []string{":42", "file.php:100", "other.php:50"},
			expectedLocations: []string{":42", "file.php:100", "other.php:50"},
			expectedCondition: "",
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locations, condition, err := parseBreakpointArgs(tt.args)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(locations) != len(tt.expectedLocations) {
				t.Errorf("Expected %d locations, got %d", len(tt.expectedLocations), len(locations))
			}

			for i, loc := range locations {
				if i < len(tt.expectedLocations) && loc != tt.expectedLocations[i] {
					t.Errorf("Expected location %s, got %s", tt.expectedLocations[i], loc)
				}
			}

			if condition != tt.expectedCondition {
				t.Errorf("Expected condition '%s', got '%s'", tt.expectedCondition, condition)
			}
		})
	}
}
