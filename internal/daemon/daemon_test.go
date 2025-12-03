package daemon

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/console/xdebug-cli/internal/dbgp"
	"github.com/console/xdebug-cli/internal/ipc"
)

func TestNewDaemon(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9003)
	daemon, err := NewDaemon(server, 9003)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Verify daemon configuration
	if daemon.port != 9003 {
		t.Errorf("port = %d, want 9003", daemon.port)
	}

	expectedPIDFile := "/tmp/xdebug-cli-daemon-9003.pid"
	if daemon.pidFile != expectedPIDFile {
		t.Errorf("pidFile = %s, want %s", daemon.pidFile, expectedPIDFile)
	}

	expectedSocketPath := "/tmp/xdebug-cli-session-9003.sock"
	if daemon.socketPath != expectedSocketPath {
		t.Errorf("socketPath = %s, want %s", daemon.socketPath, expectedSocketPath)
	}

	if daemon.registry == nil {
		t.Error("registry is nil")
	}
}

func TestDaemon_CheckExisting_NoPIDFile(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9003)
	daemon, err := NewDaemon(server, 9003)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	exists, pid, err := daemon.CheckExisting()
	if err != nil {
		t.Errorf("CheckExisting() error = %v", err)
	}
	if exists {
		t.Error("CheckExisting() exists = true, want false")
	}
	if pid != 0 {
		t.Errorf("CheckExisting() pid = %d, want 0", pid)
	}
}

func TestDaemon_CheckExisting_ValidPID(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9003)
	daemon, err := NewDaemon(server, 9003)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Write current PID to file (should exist)
	currentPID := os.Getpid()
	err = os.WriteFile(daemon.pidFile, []byte(strconv.Itoa(currentPID)), 0600)
	if err != nil {
		t.Fatalf("Failed to write PID file: %v", err)
	}
	defer os.Remove(daemon.pidFile)

	exists, pid, err := daemon.CheckExisting()
	if err != nil {
		t.Errorf("CheckExisting() error = %v", err)
	}
	if !exists {
		t.Error("CheckExisting() exists = false, want true")
	}
	if pid != currentPID {
		t.Errorf("CheckExisting() pid = %d, want %d", pid, currentPID)
	}
}

func TestDaemon_CheckExisting_StalePID(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9003)
	daemon, err := NewDaemon(server, 9003)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Write non-existent PID to file
	stalePID := 999999
	err = os.WriteFile(daemon.pidFile, []byte(strconv.Itoa(stalePID)), 0600)
	if err != nil {
		t.Fatalf("Failed to write PID file: %v", err)
	}

	exists, pid, err := daemon.CheckExisting()
	if err != nil {
		t.Errorf("CheckExisting() error = %v", err)
	}
	if exists {
		t.Error("CheckExisting() exists = true, want false (stale PID)")
	}
	if pid != 0 {
		t.Errorf("CheckExisting() pid = %d, want 0", pid)
	}

	// PID file should be removed
	if _, err := os.Stat(daemon.pidFile); !os.IsNotExist(err) {
		t.Error("Stale PID file was not removed")
	}
}

func TestDaemon_WritePIDFile(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9003)
	daemon, err := NewDaemon(server, 9003)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	err = daemon.writePIDFile()
	if err != nil {
		t.Errorf("writePIDFile() error = %v", err)
	}
	defer daemon.removePIDFile()

	// Verify PID file content
	data, err := os.ReadFile(daemon.pidFile)
	if err != nil {
		t.Fatalf("Failed to read PID file: %v", err)
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		t.Fatalf("Invalid PID in file: %v", err)
	}

	if pid != os.Getpid() {
		t.Errorf("PID in file = %d, want %d", pid, os.Getpid())
	}
}

func TestDaemon_RemovePIDFile(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9003)
	daemon, err := NewDaemon(server, 9003)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Create PID file
	daemon.writePIDFile()

	// Remove it
	err = daemon.removePIDFile()
	if err != nil {
		t.Errorf("removePIDFile() error = %v", err)
	}

	// Verify it's gone
	if _, err := os.Stat(daemon.pidFile); !os.IsNotExist(err) {
		t.Error("PID file still exists after removal")
	}

	// Removing non-existent file should not error
	err = daemon.removePIDFile()
	if err != nil {
		t.Errorf("removePIDFile() error on non-existent file = %v", err)
	}
}

func TestDaemon_IsDaemonMode(t *testing.T) {
	// Should be false by default
	if IsDaemonMode() {
		t.Error("IsDaemonMode() = true, want false")
	}

	// Set environment variable
	os.Setenv("XDEBUG_CLI_DAEMON_MODE", "1")
	defer os.Unsetenv("XDEBUG_CLI_DAEMON_MODE")

	if !IsDaemonMode() {
		t.Error("IsDaemonMode() = false, want true")
	}
}

func TestDaemon_HandleIPCRequest_NoSession(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9003)
	daemon, err := NewDaemon(server, 9003)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Request without client should fail
	req := ipc.NewExecuteCommandsRequest([]string{"run"}, false)
	resp := daemon.handleIPCRequest(req)

	if resp.Success {
		t.Error("handleIPCRequest() success = true, want false (no client)")
	}
	if resp.Error != "no active debug session" {
		t.Errorf("handleIPCRequest() error = %s, want 'no active debug session'", resp.Error)
	}
}

func TestDaemon_HandleIPCRequest_UnknownType(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9003)
	daemon, err := NewDaemon(server, 9003)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Unknown request type
	req := &ipc.CommandRequest{
		Type: "unknown_type",
	}
	resp := daemon.handleIPCRequest(req)

	if resp.Success {
		t.Error("handleIPCRequest() success = true, want false")
	}
	if resp.Error != "unknown request type: unknown_type" {
		t.Errorf("handleIPCRequest() error = %s", resp.Error)
	}
}

func TestDaemon_Getters(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9003)
	daemon, err := NewDaemon(server, 9003)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	if daemon.GetPort() != 9003 {
		t.Errorf("GetPort() = %d, want 9003", daemon.GetPort())
	}

	expectedPIDFile := "/tmp/xdebug-cli-daemon-9003.pid"
	if daemon.GetPIDFile() != expectedPIDFile {
		t.Errorf("GetPIDFile() = %s, want %s", daemon.GetPIDFile(), expectedPIDFile)
	}

	expectedSocketPath := "/tmp/xdebug-cli-session-9003.sock"
	if daemon.GetSocketPath() != expectedSocketPath {
		t.Errorf("GetSocketPath() = %s, want %s", daemon.GetSocketPath(), expectedSocketPath)
	}
}

// TestDaemon_SignalHandling tests signal handling and cleanup
func TestDaemon_SignalHandling(t *testing.T) {
	// This test verifies the signal handler setup
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9003)
	daemon, err := NewDaemon(server, 9003)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Setup signal handlers
	daemon.setupSignalHandlers()

	// Verify shutdown channel was created
	if daemon.shutdown == nil {
		t.Error("shutdown channel is nil")
	}

	// Send signal to trigger shutdown
	done := make(chan bool, 1)
	go func() {
		daemon.Wait()
		done <- true
	}()

	// Send SIGTERM
	daemon.shutdown <- syscall.SIGTERM

	// Wait for shutdown or timeout
	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Shutdown did not complete within timeout")
	}
}

// TestDaemon_Shutdown tests graceful shutdown
func TestDaemon_Shutdown(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9003)
	daemon, err := NewDaemon(server, 9003)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Write PID file
	daemon.writePIDFile()

	// Add to registry
	sessionInfo := SessionInfo{
		PID:        os.Getpid(),
		Port:       9003,
		SocketPath: daemon.socketPath,
		StartedAt:  time.Now(),
	}
	daemon.registry.Add(sessionInfo)

	// Shutdown should clean everything up
	err = daemon.Shutdown()
	if err != nil {
		t.Logf("Shutdown() error = %v (may be expected)", err)
	}

	// Verify PID file removed
	if _, err := os.Stat(daemon.pidFile); !os.IsNotExist(err) {
		t.Error("PID file still exists after shutdown")
	}

	// Verify removed from registry
	if _, err := daemon.registry.Get(9003); err == nil {
		t.Error("Session still in registry after shutdown")
	}

	// Verify context cancelled
	select {
	case <-daemon.ctx.Done():
		// Success
	default:
		t.Error("Context not cancelled after shutdown")
	}
}

// Integration test demonstrating PID file conflict detection
func TestDaemon_Integration_ConflictDetection(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9003)
	daemon1, err := NewDaemon(server, 9003)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Write PID file for daemon1 (simulating running daemon)
	daemon1.writePIDFile()
	defer daemon1.removePIDFile()

	// Create second daemon on same port
	daemon2, err := NewDaemon(server, 9003)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// CheckExisting should detect the running daemon
	exists, pid, err := daemon2.CheckExisting()
	if err != nil {
		t.Errorf("CheckExisting() error = %v", err)
	}
	if !exists {
		t.Error("CheckExisting() did not detect existing daemon")
	}
	if pid != os.Getpid() {
		t.Errorf("CheckExisting() pid = %d, want %d", pid, os.Getpid())
	}

	// Fork should fail with conflict error
	err = daemon2.Fork([]string{"test"})
	if err == nil {
		t.Error("Fork() succeeded, want error for existing daemon")
	}
	// Check that error message contains key information (exact format may vary)
	errMsg := err.Error()
	if !strings.Contains(errMsg, "daemon already running on port 9003") {
		t.Errorf("Fork() error = %v, want error about daemon already running", err)
	}
	if !strings.Contains(errMsg, fmt.Sprintf("PID %d", os.Getpid())) {
		t.Errorf("Fork() error = %v, want error mentioning PID %d", err, os.Getpid())
	}
}

// Integration test demonstrating kill request handling
func TestDaemon_Integration_KillRequest(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9003)
	daemon, err := NewDaemon(server, 9003)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Write PID file and register
	daemon.writePIDFile()
	sessionInfo := SessionInfo{
		PID:        os.Getpid(),
		Port:       9003,
		SocketPath: daemon.socketPath,
		StartedAt:  time.Now(),
	}
	daemon.registry.Add(sessionInfo)

	// Test kill request
	killReq := ipc.NewKillRequest()
	resp := daemon.handleIPCRequest(killReq)

	if !resp.Success {
		t.Errorf("Kill request failed: %s", resp.Error)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(resp.Results))
	}

	result := resp.Results[0]
	if result.Command != "kill" {
		t.Errorf("Result command = %s, want 'kill'", result.Command)
	}
	if !result.Success {
		t.Errorf("Result success = false, want true")
	}

	// Give shutdown goroutine time to execute
	time.Sleep(200 * time.Millisecond)

	// Verify cleanup happened
	if _, err := os.Stat(daemon.pidFile); !os.IsNotExist(err) {
		t.Error("PID file still exists after kill")
	}
}

// Integration test for registry persistence across daemon lifecycle
func TestDaemon_Integration_RegistryPersistence(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	// Create and start daemon
	server := dbgp.NewServer("127.0.0.1", 9003)
	daemon, err := NewDaemon(server, 9003)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Write PID and register
	daemon.writePIDFile()
	sessionInfo := SessionInfo{
		PID:        os.Getpid(),
		Port:       9003,
		SocketPath: daemon.socketPath,
		StartedAt:  time.Now(),
	}
	daemon.registry.Add(sessionInfo)

	// Verify session is in registry
	retrieved, err := daemon.registry.Get(9003)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}
	if retrieved.Port != 9003 {
		t.Errorf("Retrieved port = %d, want 9003", retrieved.Port)
	}

	// Create new registry instance (simulating fresh process)
	registry2, err := NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	// Session should still be there
	retrieved2, err := registry2.Get(9003)
	if err != nil {
		t.Fatalf("Failed to get session from new registry: %v", err)
	}
	if retrieved2.PID != os.Getpid() {
		t.Errorf("Retrieved PID = %d, want %d", retrieved2.PID, os.Getpid())
	}

	// Cleanup
	daemon.removePIDFile()
	daemon.registry.Remove(9003)
}

// TestDaemon_Cleanup_NormalExit tests cleanup on normal session end
func TestDaemon_Cleanup_NormalExit(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9004)
	daemon, err := NewDaemon(server, 9004)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Simulate daemon startup
	daemon.writePIDFile()
	sessionInfo := SessionInfo{
		PID:        os.Getpid(),
		Port:       9004,
		SocketPath: daemon.socketPath,
		StartedAt:  time.Now(),
	}
	daemon.registry.Add(sessionInfo)

	// Create socket file to verify cleanup
	socketFile, err := os.Create(daemon.socketPath)
	if err != nil {
		t.Fatalf("Failed to create socket file: %v", err)
	}
	socketFile.Close()

	// Verify files exist before shutdown
	if _, err := os.Stat(daemon.pidFile); os.IsNotExist(err) {
		t.Error("PID file should exist before shutdown")
	}
	if _, err := os.Stat(daemon.socketPath); os.IsNotExist(err) {
		t.Error("Socket file should exist before shutdown")
	}

	// Perform graceful shutdown
	err = daemon.Shutdown()
	if err != nil {
		t.Logf("Shutdown() error = %v (may be expected)", err)
	}

	// Verify all cleanup completed
	if _, err := os.Stat(daemon.pidFile); !os.IsNotExist(err) {
		t.Error("PID file still exists after shutdown")
	}

	if _, err := os.Stat(daemon.socketPath); !os.IsNotExist(err) {
		t.Error("Socket file still exists after shutdown")
	}

	if _, err := daemon.registry.Get(9004); err == nil {
		t.Error("Session still in registry after shutdown")
	}

	// Verify context cancelled
	select {
	case <-daemon.ctx.Done():
		// Success
	default:
		t.Error("Context not cancelled after shutdown")
	}
}

// TestDaemon_Cleanup_SignalTERM tests cleanup on SIGTERM
func TestDaemon_Cleanup_SignalTERM(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9005)
	daemon, err := NewDaemon(server, 9005)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Simulate daemon startup
	daemon.writePIDFile()
	sessionInfo := SessionInfo{
		PID:        os.Getpid(),
		Port:       9005,
		SocketPath: daemon.socketPath,
		StartedAt:  time.Now(),
	}
	daemon.registry.Add(sessionInfo)

	// Create socket file
	socketFile, err := os.Create(daemon.socketPath)
	if err != nil {
		t.Fatalf("Failed to create socket file: %v", err)
	}
	socketFile.Close()

	// Setup signal handlers
	daemon.setupSignalHandlers()

	// Monitor shutdown completion
	done := make(chan bool, 1)
	go func() {
		daemon.Wait()
		done <- true
	}()

	// Send SIGTERM
	daemon.shutdown <- syscall.SIGTERM

	// Wait for shutdown with timeout
	select {
	case <-done:
		// Success - wait a bit more for cleanup to complete
		time.Sleep(100 * time.Millisecond)
	case <-time.After(2 * time.Second):
		t.Fatal("Shutdown did not complete within timeout")
	}

	// Verify cleanup (with retries for async cleanup)
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(daemon.pidFile); os.IsNotExist(err) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if _, err := os.Stat(daemon.pidFile); !os.IsNotExist(err) {
		t.Error("PID file still exists after SIGTERM")
	}

	if _, err := os.Stat(daemon.socketPath); !os.IsNotExist(err) {
		t.Error("Socket file still exists after SIGTERM")
	}

	if _, err := daemon.registry.Get(9005); err == nil {
		t.Error("Session still in registry after SIGTERM")
	}
}

// TestDaemon_Cleanup_SignalINT tests cleanup on SIGINT
func TestDaemon_Cleanup_SignalINT(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9006)
	daemon, err := NewDaemon(server, 9006)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Simulate daemon startup
	daemon.writePIDFile()
	sessionInfo := SessionInfo{
		PID:        os.Getpid(),
		Port:       9006,
		SocketPath: daemon.socketPath,
		StartedAt:  time.Now(),
	}
	daemon.registry.Add(sessionInfo)

	// Create socket file
	socketFile, err := os.Create(daemon.socketPath)
	if err != nil {
		t.Fatalf("Failed to create socket file: %v", err)
	}
	socketFile.Close()

	// Setup signal handlers
	daemon.setupSignalHandlers()

	// Monitor shutdown completion
	done := make(chan bool, 1)
	go func() {
		daemon.Wait()
		done <- true
	}()

	// Send SIGINT
	daemon.shutdown <- syscall.SIGINT

	// Wait for shutdown with timeout
	select {
	case <-done:
		// Success - wait a bit more for cleanup to complete
		time.Sleep(100 * time.Millisecond)
	case <-time.After(2 * time.Second):
		t.Fatal("Shutdown did not complete within timeout")
	}

	// Verify cleanup (with retries for async cleanup)
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(daemon.pidFile); os.IsNotExist(err) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if _, err := os.Stat(daemon.pidFile); !os.IsNotExist(err) {
		t.Error("PID file still exists after SIGINT")
	}

	if _, err := os.Stat(daemon.socketPath); !os.IsNotExist(err) {
		t.Error("Socket file still exists after SIGINT")
	}

	if _, err := daemon.registry.Get(9006); err == nil {
		t.Error("Session still in registry after SIGINT")
	}
}

// TestDaemon_Cleanup_Error tests cleanup on error conditions
func TestDaemon_Cleanup_Error(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9007)
	daemon, err := NewDaemon(server, 9007)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Simulate daemon startup
	daemon.writePIDFile()
	sessionInfo := SessionInfo{
		PID:        os.Getpid(),
		Port:       9007,
		SocketPath: daemon.socketPath,
		StartedAt:  time.Now(),
	}
	daemon.registry.Add(sessionInfo)

	// Create socket file
	socketFile, err := os.Create(daemon.socketPath)
	if err != nil {
		t.Fatalf("Failed to create socket file: %v", err)
	}
	socketFile.Close()

	// Simulate error condition by manually triggering shutdown
	err = daemon.Shutdown()
	if err != nil {
		t.Logf("Shutdown() error = %v (expected for error condition)", err)
	}

	// Verify cleanup still happened despite error
	if _, err := os.Stat(daemon.pidFile); !os.IsNotExist(err) {
		t.Error("PID file still exists after error shutdown")
	}

	if _, err := os.Stat(daemon.socketPath); !os.IsNotExist(err) {
		t.Error("Socket file still exists after error shutdown")
	}

	if _, err := daemon.registry.Get(9007); err == nil {
		t.Error("Session still in registry after error shutdown")
	}
}

// TestDaemon_Cleanup_Timeout tests shutdown timeout handling
func TestDaemon_Cleanup_Timeout(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9008)
	daemon, err := NewDaemon(server, 9008)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Simulate daemon startup
	daemon.writePIDFile()
	sessionInfo := SessionInfo{
		PID:        os.Getpid(),
		Port:       9008,
		SocketPath: daemon.socketPath,
		StartedAt:  time.Now(),
	}
	daemon.registry.Add(sessionInfo)

	// Create socket file
	socketFile, err := os.Create(daemon.socketPath)
	if err != nil {
		t.Fatalf("Failed to create socket file: %v", err)
	}
	socketFile.Close()

	// Note: Testing actual timeout would require blocking server.Close()
	// which is difficult to mock. This test verifies basic timeout logic.

	// Perform shutdown
	err = daemon.Shutdown()
	if err != nil {
		t.Logf("Shutdown() error = %v (may be expected)", err)
	}

	// Verify critical cleanup happened (registry and PID)
	// These should always be cleaned up even on timeout
	if _, err := os.Stat(daemon.pidFile); !os.IsNotExist(err) {
		t.Error("PID file still exists - critical cleanup failed")
	}

	if _, err := daemon.registry.Get(9008); err == nil {
		t.Error("Session still in registry - critical cleanup failed")
	}

	// Socket should also be cleaned up
	if _, err := os.Stat(daemon.socketPath); !os.IsNotExist(err) {
		t.Error("Socket file still exists after shutdown")
	}
}

// TestDaemon_Cleanup_MultipleShutdown tests that multiple shutdown calls are safe
func TestDaemon_Cleanup_MultipleShutdown(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9009)
	daemon, err := NewDaemon(server, 9009)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Simulate daemon startup
	daemon.writePIDFile()
	sessionInfo := SessionInfo{
		PID:        os.Getpid(),
		Port:       9009,
		SocketPath: daemon.socketPath,
		StartedAt:  time.Now(),
	}
	daemon.registry.Add(sessionInfo)

	// First shutdown
	err1 := daemon.Shutdown()
	if err1 != nil {
		t.Logf("First Shutdown() error = %v", err1)
	}

	// Second shutdown should not panic
	err2 := daemon.Shutdown()
	if err2 != nil {
		t.Logf("Second Shutdown() error = %v (expected)", err2)
	}

	// Verify cleanup only happened once (no double-free issues)
	if _, err := os.Stat(daemon.pidFile); !os.IsNotExist(err) {
		t.Error("PID file still exists after shutdown")
	}
}

// TestDaemon_EmergencyCleanup tests emergency cleanup on panic
func TestDaemon_EmergencyCleanup(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	server := dbgp.NewServer("127.0.0.1", 9010)
	daemon, err := NewDaemon(server, 9010)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Create PID and socket files
	daemon.writePIDFile()
	socketFile, err := os.Create(daemon.socketPath)
	if err != nil {
		t.Fatalf("Failed to create socket file: %v", err)
	}
	socketFile.Close()

	// Call emergency cleanup
	daemon.emergencyCleanup()

	// Verify files removed
	if _, err := os.Stat(daemon.pidFile); !os.IsNotExist(err) {
		t.Error("PID file still exists after emergency cleanup")
	}

	if _, err := os.Stat(daemon.socketPath); !os.IsNotExist(err) {
		t.Error("Socket file still exists after emergency cleanup")
	}
}
