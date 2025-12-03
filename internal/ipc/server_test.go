package ipc

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServer_Listen(t *testing.T) {
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("test-server-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	handler := func(req *CommandRequest) *CommandResponse {
		return NewSuccessResponse([]CommandResult{})
	}

	server := NewServer(socketPath, handler)

	// Listen should succeed
	if err := server.Listen(); err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer server.Shutdown()

	// Socket file should exist
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Errorf("Socket file does not exist at %s", socketPath)
	}

	// Socket should have 0600 permissions
	info, err := os.Stat(socketPath)
	if err != nil {
		t.Fatalf("Failed to stat socket: %v", err)
	}
	mode := info.Mode() & os.ModePerm
	if mode != 0600 {
		t.Errorf("Socket permissions = %o, want 0600", mode)
	}
}

func TestServer_Listen_DoubleCall(t *testing.T) {
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("test-server-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	handler := func(req *CommandRequest) *CommandResponse {
		return NewSuccessResponse([]CommandResult{})
	}

	server := NewServer(socketPath, handler)
	defer server.Shutdown()

	if err := server.Listen(); err != nil {
		t.Fatalf("First Listen() error = %v", err)
	}

	// Second Listen should fail
	if err := server.Listen(); err == nil {
		t.Error("Second Listen() expected error, got nil")
	}
}

func TestServer_HandleRequest(t *testing.T) {
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("test-server-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	// Handler that echoes back command count
	handler := func(req *CommandRequest) *CommandResponse {
		if req.Type == "execute_commands" {
			results := make([]CommandResult, len(req.Commands))
			for i, cmd := range req.Commands {
				results[i] = CommandResult{
					Command: cmd,
					Success: true,
					Result:  map[string]interface{}{"executed": true},
				}
			}
			return NewSuccessResponse(results)
		}
		return NewErrorResponse("unknown request type")
	}

	server := NewServer(socketPath, handler)
	if err := server.Listen(); err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer server.Shutdown()

	// Start server in background
	go func() {
		_ = server.Serve()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Connect as client
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Send request
	req := NewExecuteCommandsRequest([]string{"run", "step"}, false)
	reqData, _ := req.ToJSON()
	reqData = append(reqData, '\n')
	if _, err := conn.Write(reqData); err != nil {
		t.Fatalf("Failed to write request: %v", err)
	}

	// Read response
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Parse response
	var resp CommandResponse
	if err := resp.FromJSON(buf[:n]); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify response
	if !resp.Success {
		t.Errorf("Response success = false, want true")
	}
	if len(resp.Results) != 2 {
		t.Errorf("Response results length = %d, want 2", len(resp.Results))
	}
	if resp.Results[0].Command != "run" {
		t.Errorf("Results[0].Command = %s, want run", resp.Results[0].Command)
	}
	if resp.Results[1].Command != "step" {
		t.Errorf("Results[1].Command = %s, want step", resp.Results[1].Command)
	}
}

func TestServer_InvalidRequest(t *testing.T) {
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("test-server-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	handler := func(req *CommandRequest) *CommandResponse {
		return NewSuccessResponse([]CommandResult{})
	}

	server := NewServer(socketPath, handler)
	if err := server.Listen(); err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer server.Shutdown()

	go func() {
		_ = server.Serve()
	}()

	time.Sleep(100 * time.Millisecond)

	// Connect and send invalid JSON
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("invalid json\n")); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	// Read error response
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var resp CommandResponse
	if err := resp.FromJSON(buf[:n]); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Success {
		t.Error("Expected error response for invalid JSON")
	}
}

func TestServer_Shutdown(t *testing.T) {
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("test-server-%d.sock", time.Now().UnixNano()))

	handler := func(req *CommandRequest) *CommandResponse {
		return NewSuccessResponse([]CommandResult{})
	}

	server := NewServer(socketPath, handler)
	if err := server.Listen(); err != nil {
		t.Fatalf("Listen() error = %v", err)
	}

	go func() {
		_ = server.Serve()
	}()

	time.Sleep(100 * time.Millisecond)

	// Shutdown should succeed
	if err := server.Shutdown(); err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}

	// Socket file should be removed
	if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
		t.Errorf("Socket file still exists after shutdown")
	}
}

func TestServer_IsRunning(t *testing.T) {
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("test-server-%d.sock", time.Now().UnixNano()))
	defer os.Remove(socketPath)

	handler := func(req *CommandRequest) *CommandResponse {
		return NewSuccessResponse([]CommandResult{})
	}

	server := NewServer(socketPath, handler)

	// Not running initially
	if server.IsRunning() {
		t.Error("IsRunning() = true before Listen(), want false")
	}

	// Running after Listen
	if err := server.Listen(); err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer server.Shutdown()

	if !server.IsRunning() {
		t.Error("IsRunning() = false after Listen(), want true")
	}
}

func TestServer_SocketPath(t *testing.T) {
	socketPath := "/tmp/test.sock"
	handler := func(req *CommandRequest) *CommandResponse {
		return NewSuccessResponse([]CommandResult{})
	}

	server := NewServer(socketPath, handler)

	if server.SocketPath() != socketPath {
		t.Errorf("SocketPath() = %s, want %s", server.SocketPath(), socketPath)
	}
}
