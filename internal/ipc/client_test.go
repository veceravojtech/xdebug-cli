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

func TestClient_ConnectWithRetry_SuccessFirstAttempt(t *testing.T) {
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

	// Connect with retry should succeed on first attempt (no delay)
	client := NewClient(socketPath)
	start := time.Now()
	conn, err := client.ConnectWithRetry(3)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("ConnectWithRetry() error = %v", err)
	}
	defer conn.Close()

	// Should not have delayed (should be under 50ms for immediate success)
	if elapsed > 50*time.Millisecond {
		t.Errorf("ConnectWithRetry() took %v, expected immediate success", elapsed)
	}
}

func TestClient_ConnectWithRetry_SuccessSecondAttempt(t *testing.T) {
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("test-client-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	handler := func(req *CommandRequest) *CommandResponse {
		return NewSuccessResponse([]CommandResult{})
	}

	server := NewServer(socketPath, handler)

	// Start server in background after a delay (don't call Listen() yet)
	go func() {
		time.Sleep(150 * time.Millisecond) // Wait before starting server
		if err := server.Listen(); err != nil {
			t.Logf("Failed to start server: %v", err)
			return
		}
		_ = server.Serve()
	}()

	defer server.Shutdown()

	// Connect with retry should succeed on second attempt after backoff
	client := NewClient(socketPath)
	client.SetTimeout(1 * time.Second)
	start := time.Now()
	conn, err := client.ConnectWithRetry(3)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("ConnectWithRetry() error = %v", err)
	}
	defer conn.Close()

	// Should have waited at least 100ms (first backoff)
	if elapsed < 100*time.Millisecond {
		t.Errorf("ConnectWithRetry() took %v, expected at least 100ms for retry", elapsed)
	}

	// Should not have waited for all 3 attempts (100ms + 200ms = 300ms)
	if elapsed > 350*time.Millisecond {
		t.Errorf("ConnectWithRetry() took %v, expected success on second attempt", elapsed)
	}
}

func TestClient_ConnectWithRetry_AllAttemptsFail(t *testing.T) {
	socketPath := "/tmp/nonexistent-retry-socket.sock"

	client := NewClient(socketPath)
	client.SetTimeout(100 * time.Millisecond)
	start := time.Now()
	conn, err := client.ConnectWithRetry(3)
	elapsed := time.Since(start)

	if err == nil {
		conn.Close()
		t.Error("ConnectWithRetry() expected error for nonexistent socket")
	}

	// Verify error message mentions attempt count
	expectedMsg := "failed to connect after 3 attempts"
	if err != nil && !containsString(err.Error(), expectedMsg) {
		t.Errorf("ConnectWithRetry() error = %v, expected message containing %q", err, expectedMsg)
	}

	// Should have waited: 100ms + 200ms = 300ms between attempts
	// Total should be at least 300ms
	if elapsed < 300*time.Millisecond {
		t.Errorf("ConnectWithRetry() took %v, expected at least 300ms for 3 attempts", elapsed)
	}
}

func TestClient_ConnectWithRetry_CustomAttemptCount(t *testing.T) {
	socketPath := "/tmp/nonexistent-custom-retry-socket.sock"

	client := NewClient(socketPath)
	client.SetTimeout(100 * time.Millisecond)
	conn, err := client.ConnectWithRetry(5)

	if err == nil {
		conn.Close()
		t.Error("ConnectWithRetry() expected error for nonexistent socket")
	}

	// Verify error message mentions correct attempt count
	expectedMsg := "failed to connect after 5 attempts"
	if err != nil && !containsString(err.Error(), expectedMsg) {
		t.Errorf("ConnectWithRetry() error = %v, expected message containing %q", err, expectedMsg)
	}
}

func TestClient_ConnectWithRetry_MinAttempts(t *testing.T) {
	socketPath := "/tmp/nonexistent-min-retry-socket.sock"

	client := NewClient(socketPath)
	client.SetTimeout(100 * time.Millisecond)

	// Test with 0 attempts - should default to 1
	conn, err := client.ConnectWithRetry(0)

	if err == nil {
		conn.Close()
		t.Error("ConnectWithRetry() expected error for nonexistent socket")
	}

	// Should have made at least 1 attempt
	expectedMsg := "failed to connect after 1 attempt"
	if err != nil && !containsString(err.Error(), expectedMsg) {
		t.Errorf("ConnectWithRetry() with 0 attempts, error = %v, expected message containing %q", err, expectedMsg)
	}
}

func TestClient_SendCommandsWithRetry(t *testing.T) {
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

	// Send commands with retry
	client := NewClient(socketPath)
	resp, err := client.SendCommandsWithRetry([]string{"run", "step"}, false, 3)
	if err != nil {
		t.Fatalf("SendCommandsWithRetry() error = %v", err)
	}

	// Verify response
	if !resp.Success {
		t.Errorf("Response success = false, want true")
	}
	if len(resp.Results) != 2 {
		t.Fatalf("Response results length = %d, want 2", len(resp.Results))
	}
}

func TestClient_SendCommandsWithRetry_NoServer(t *testing.T) {
	socketPath := "/tmp/nonexistent-sendcommands-socket.sock"

	client := NewClient(socketPath)
	client.SetTimeout(100 * time.Millisecond)
	_, err := client.SendCommandsWithRetry([]string{"run"}, false, 3)

	if err == nil {
		t.Error("SendCommandsWithRetry() expected error for nonexistent socket")
	}

	// Verify error mentions attempt count
	expectedMsg := "failed to connect after 3 attempts"
	if err != nil && !containsString(err.Error(), expectedMsg) {
		t.Errorf("SendCommandsWithRetry() error = %v, expected message containing %q", err, expectedMsg)
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
