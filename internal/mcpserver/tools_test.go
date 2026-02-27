package mcpserver

import (
	"context"
	"reflect"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestBuildDaemonStartArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    DaemonStartInput
		expected []string
	}{
		{
			name:     "curl only",
			input:    DaemonStartInput{Curl: "http://localhost/app.php"},
			expected: []string{"daemon", "start", "--curl", "http://localhost/app.php"},
		},
		{
			name:     "external connection only",
			input:    DaemonStartInput{EnableExternalConnection: true},
			expected: []string{"daemon", "start", "--enable-external-connection"},
		},
		{
			name:     "with port",
			input:    DaemonStartInput{Curl: "http://localhost/app.php", Port: 9004},
			expected: []string{"daemon", "start", "--curl", "http://localhost/app.php", "-p", "9004"},
		},
		{
			name:     "with commands",
			input:    DaemonStartInput{Curl: "http://localhost/app.php", Commands: []string{"break :42", "break :100"}},
			expected: []string{"daemon", "start", "--curl", "http://localhost/app.php", "--commands", "break :42", "--commands", "break :100"},
		},
		{
			name:     "with breakpoint timeout",
			input:    DaemonStartInput{Curl: "http://localhost/app.php", BreakpointTimeout: 60},
			expected: []string{"daemon", "start", "--curl", "http://localhost/app.php", "--breakpoint-timeout", "60"},
		},
		{
			name:     "with wait forever",
			input:    DaemonStartInput{Curl: "http://localhost/app.php", WaitForever: true},
			expected: []string{"daemon", "start", "--curl", "http://localhost/app.php", "--wait-forever"},
		},
		{
			name: "full combination",
			input: DaemonStartInput{
				Curl:              "http://localhost/app.php",
				Port:              9004,
				Commands:          []string{"break :42"},
				BreakpointTimeout: 60,
				WaitForever:       true,
			},
			expected: []string{"daemon", "start", "--curl", "http://localhost/app.php", "-p", "9004", "--commands", "break :42", "--breakpoint-timeout", "60", "--wait-forever"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDaemonStartArgs(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBuildDaemonKillArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    DaemonKillInput
		expected []string
	}{
		{
			name:     "default",
			input:    DaemonKillInput{},
			expected: []string{"daemon", "kill"},
		},
		{
			name:     "with port",
			input:    DaemonKillInput{Port: 9004},
			expected: []string{"daemon", "kill", "-p", "9004"},
		},
		{
			name:     "with all",
			input:    DaemonKillInput{All: true},
			expected: []string{"daemon", "kill", "--all", "--force"},
		},
		{
			name:     "with port and all",
			input:    DaemonKillInput{Port: 9004, All: true},
			expected: []string{"daemon", "kill", "-p", "9004", "--all", "--force"},
		},
		{
			name:     "force without all",
			input:    DaemonKillInput{Force: true},
			expected: []string{"daemon", "kill", "--force"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDaemonKillArgs(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBuildDaemonStatusArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    DaemonStatusInput
		expected []string
	}{
		{
			name:     "default",
			input:    DaemonStatusInput{},
			expected: []string{"daemon", "status"},
		},
		{
			name:     "with port",
			input:    DaemonStatusInput{Port: 9004},
			expected: []string{"daemon", "status", "-p", "9004"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDaemonStatusArgs(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBuildDaemonListArgs(t *testing.T) {
	got := buildDaemonListArgs()
	expected := []string{"daemon", "list", "--json"}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestBuildDaemonIsAliveArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    DaemonIsAliveInput
		expected []string
	}{
		{
			name:     "default",
			input:    DaemonIsAliveInput{},
			expected: []string{"daemon", "isAlive"},
		},
		{
			name:     "with port",
			input:    DaemonIsAliveInput{Port: 9004},
			expected: []string{"daemon", "isAlive", "-p", "9004"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDaemonIsAliveArgs(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBuildExecuteArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    ExecuteInput
		expected []string
	}{
		{
			name:     "single command",
			input:    ExecuteInput{Commands: []string{"run"}},
			expected: []string{"attach", "--json", "--commands", "run"},
		},
		{
			name:     "multiple commands",
			input:    ExecuteInput{Commands: []string{"step", "print $x"}},
			expected: []string{"attach", "--json", "--commands", "step", "--commands", "print $x"},
		},
		{
			name:     "with port",
			input:    ExecuteInput{Commands: []string{"run"}, Port: 9004},
			expected: []string{"attach", "--json", "-p", "9004", "--commands", "run"},
		},
		{
			name:     "commands and port",
			input:    ExecuteInput{Commands: []string{"break :42", "run"}, Port: 9004},
			expected: []string{"attach", "--json", "-p", "9004", "--commands", "break :42", "--commands", "run"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildExecuteArgs(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRun_Success(t *testing.T) {
	output, err := run(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "hello" {
		t.Errorf("got %q, want %q", output, "hello")
	}
}

func TestRun_Failure(t *testing.T) {
	_, err := run(context.Background(), "false")
	if err == nil {
		t.Fatal("expected error from 'false' command")
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := run(ctx, "sleep", "10")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestTextResult(t *testing.T) {
	result, err := textResult("ok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected IsError to be false")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *mcp.TextContent, got %T", result.Content[0])
	}
	if tc.Text != "ok" {
		t.Errorf("got text %q, want %q", tc.Text, "ok")
	}
}

func TestErrorResult(t *testing.T) {
	result, err := errorResult("fail")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError to be true")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *mcp.TextContent, got %T", result.Content[0])
	}
	if tc.Text != "fail" {
		t.Errorf("got text %q, want %q", tc.Text, "fail")
	}
}
