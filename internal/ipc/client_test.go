package ipc

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestClient_Connect(t *testing.T) {
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("test-client-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	// Start a server
	handler := func(req *CommandRequest) *CommandResponse {
		return NewSuccessResponse([]CommandResult{})
	}

	server := NewServer(socketPath, handler)
	if err := server.Listen(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Shutdown()

	go func() {
		_ = server.Serve()
	}()

	time.Sleep(100 * time.Millisecond)

	// Client should connect successfully
	client := NewClient(socketPath)
	conn, err := client.Connect()
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer conn.Close()
}

func TestClient_Connect_NoServer(t *testing.T) {
	socketPath := "/tmp/nonexistent-socket.sock"

	client := NewClient(socketPath)
	conn, err := client.Connect()
	if err == nil {
		conn.Close()
		t.Error("Connect() expected error for nonexistent socket")
	}
}

func TestClient_SendCommands(t *testing.T) {
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("test-client-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	// Start server with echo handler
	handler := func(req *CommandRequest) *CommandResponse {
		results := make([]CommandResult, len(req.Commands))
		for i, cmd := range req.Commands {
			results[i] = CommandResult{
				Command: cmd,
				Success: true,
				Result:  map[string]interface{}{"echo": cmd},
			}
		}
		return NewSuccessResponse(results)
	}

	server := NewServer(socketPath, handler)
	if err := server.Listen(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Shutdown()

	go func() {
		_ = server.Serve()
	}()

	time.Sleep(100 * time.Millisecond)

	// Send commands
	client := NewClient(socketPath)
	resp, err := client.SendCommands([]string{"run", "step", "next"}, false)
	if err != nil {
		t.Fatalf("SendCommands() error = %v", err)
	}

	// Verify response
	if !resp.Success {
		t.Errorf("Response success = false, want true")
	}
	if len(resp.Results) != 3 {
		t.Fatalf("Response results length = %d, want 3", len(resp.Results))
	}

	expectedCommands := []string{"run", "step", "next"}
	for i, expected := range expectedCommands {
		if resp.Results[i].Command != expected {
			t.Errorf("Results[%d].Command = %s, want %s", i, resp.Results[i].Command, expected)
		}
		if !resp.Results[i].Success {
			t.Errorf("Results[%d].Success = false, want true", i)
		}
	}
}

func TestClient_SendCommands_JSONOutput(t *testing.T) {
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("test-client-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	// Server handler that checks JSONOutput flag
	handler := func(req *CommandRequest) *CommandResponse {
		if !req.JSONOutput {
			return NewErrorResponse("expected json_output=true")
		}
		return NewSuccessResponse([]CommandResult{
			{Command: "run", Success: true, Result: map[string]interface{}{"status": "ok"}},
		})
	}

	server := NewServer(socketPath, handler)
	if err := server.Listen(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Shutdown()

	go func() {
		_ = server.Serve()
	}()

	time.Sleep(100 * time.Millisecond)

	// Send with JSON output
	client := NewClient(socketPath)
	resp, err := client.SendCommands([]string{"run"}, true)
	if err != nil {
		t.Fatalf("SendCommands() error = %v", err)
	}

	if !resp.Success {
		t.Errorf("Response success = false, error: %s", resp.Error)
	}
}

func TestClient_Kill(t *testing.T) {
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("test-client-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	// Server handler that handles kill request
	handler := func(req *CommandRequest) *CommandResponse {
		if req.Type == "kill" {
			return NewSuccessResponse([]CommandResult{
				{Command: "kill", Success: true, Result: nil},
			})
		}
		return NewErrorResponse("unknown request type")
	}

	server := NewServer(socketPath, handler)
	if err := server.Listen(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Shutdown()

	go func() {
		_ = server.Serve()
	}()

	time.Sleep(100 * time.Millisecond)

	// Send kill request
	client := NewClient(socketPath)
	resp, err := client.Kill()
	if err != nil {
		t.Fatalf("Kill() error = %v", err)
	}

	if !resp.Success {
		t.Errorf("Kill response success = false, want true")
	}
}

func TestClient_Ping(t *testing.T) {
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("test-client-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	handler := func(req *CommandRequest) *CommandResponse {
		return NewSuccessResponse([]CommandResult{})
	}

	server := NewServer(socketPath, handler)
	if err := server.Listen(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Shutdown()

	go func() {
		_ = server.Serve()
	}()

	time.Sleep(100 * time.Millisecond)

	// Ping should succeed
	client := NewClient(socketPath)
	if err := client.Ping(); err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

func TestClient_Ping_NoServer(t *testing.T) {
	socketPath := "/tmp/nonexistent-socket.sock"

	client := NewClient(socketPath)
	if err := client.Ping(); err == nil {
		t.Error("Ping() expected error for nonexistent socket")
	}
}

func TestClient_SetTimeout(t *testing.T) {
	client := NewClient("/tmp/test.sock")

	// Default timeout
	if client.timeout != 5*time.Second {
		t.Errorf("Default timeout = %v, want 5s", client.timeout)
	}

	// Set custom timeout
	client.SetTimeout(10 * time.Second)
	if client.timeout != 10*time.Second {
		t.Errorf("Timeout after SetTimeout = %v, want 10s", client.timeout)
	}
}

func TestClient_Timeout(t *testing.T) {
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("test-client-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	// Server that delays response
	handler := func(req *CommandRequest) *CommandResponse {
		time.Sleep(2 * time.Second)
		return NewSuccessResponse([]CommandResult{})
	}

	server := NewServer(socketPath, handler)
	if err := server.Listen(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Shutdown()

	go func() {
		_ = server.Serve()
	}()

	time.Sleep(100 * time.Millisecond)

	// Client with short timeout
	client := NewClient(socketPath)
	client.SetTimeout(500 * time.Millisecond)

	_, err := client.SendCommands([]string{"run"}, false)
	if err == nil {
		t.Error("SendCommands() expected timeout error")
	}
}
