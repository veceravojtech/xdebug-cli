package cli

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/console/xdebug-cli/internal/daemon"
	"github.com/console/xdebug-cli/internal/dbgp"
	"github.com/console/xdebug-cli/internal/ipc"
)

// mockConn implements a mock network connection for testing
type mockConn struct {
	closed bool
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	return 0, fmt.Errorf("mock read not implemented")
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (m *mockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0}
}

func (m *mockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0}
}

func (m *mockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// createMockClient creates a mock DBGp client for testing
func createMockClient() *dbgp.Client {
	// Create a mock network connection
	mockNetConn := &mockConn{}
	// Create a DBGp connection wrapper
	mockConnection := dbgp.NewConnection(mockNetConn)
	// Use NewClient to create a properly initialized client
	return dbgp.NewClient(mockConnection)
}

// TestDaemonIntegration_BasicLifecycle tests the complete daemon lifecycle
func TestDaemonIntegration_BasicLifecycle(t *testing.T) {
	// Setup test environment
	tmpDir := setupTestEnv(t)
	port := 9100

	// Create daemon
	server := dbgp.NewServer("127.0.0.1", port)
	d, err := daemon.NewDaemon(server, port)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Verify daemon can start
	mockClient := createMockClient()
	err = d.Start(mockClient)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Give daemon time to initialize
	time.Sleep(100 * time.Millisecond)

	// Verify PID file exists
	if _, err := os.Stat(d.GetPIDFile()); os.IsNotExist(err) {
		t.Error("PID file should exist after daemon start")
	}

	// Verify socket exists
	if _, err := os.Stat(d.GetSocketPath()); os.IsNotExist(err) {
		t.Error("Socket file should exist after daemon start")
	}

	// Verify registry entry
	registry, err := daemon.NewSessionRegistry()
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	session, err := registry.Get(port)
	if err != nil {
		t.Errorf("Session not in registry: %v", err)
	}
	if session != nil && session.Port != port {
		t.Errorf("Session port = %d, want %d", session.Port, port)
	}

	// Verify daemon responds to IPC
	client := ipc.NewClient(d.GetSocketPath())
	err = client.Ping()
	if err != nil {
		t.Errorf("Daemon not responding to IPC: %v", err)
	}

	// Shutdown daemon
	err = d.Shutdown()
	if err != nil {
		t.Logf("Shutdown() error = %v (may be expected)", err)
	}

	// Wait a bit for cleanup to complete
	time.Sleep(200 * time.Millisecond)

	// Verify cleanup
	if _, err := os.Stat(d.GetPIDFile()); !os.IsNotExist(err) {
		t.Error("PID file still exists after shutdown")
	}

	if _, err := os.Stat(d.GetSocketPath()); !os.IsNotExist(err) {
		t.Error("Socket file still exists after shutdown")
	}

	// Cleanup
	cleanupTestEnv(tmpDir)
}

// TestDaemonIntegration_ConcurrentAttach tests concurrent attach commands
func TestDaemonIntegration_ConcurrentAttach(t *testing.T) {
	tmpDir := setupTestEnv(t)
	port := 9101

	// Create and start daemon
	server := dbgp.NewServer("127.0.0.1", port)
	d, err := daemon.NewDaemon(server, port)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	mockClient := createMockClient()
	err = d.Start(mockClient)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer d.Shutdown()

	// Give daemon time to start
	time.Sleep(100 * time.Millisecond)

	// Create IPC client
	client := ipc.NewClient(d.GetSocketPath())

	// Test concurrent requests
	concurrentRequests := 5
	var wg sync.WaitGroup
	errors := make(chan error, concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Send command request
			// Note: This will fail because we don't have a real DBGp connection
			// but we're testing that the IPC mechanism handles concurrency
			resp, err := client.SendCommands([]string{"list"}, false)
			if err != nil {
				errors <- fmt.Errorf("request %d failed: %w", id, err)
				return
			}

			// Expect failure due to no active session, but no crash
			if resp.Success {
				// If it succeeds, that's actually unexpected without a real connection
				// but not necessarily an error in the IPC layer
				t.Logf("Request %d unexpectedly succeeded (may be OK for mock)", id)
			}
		}(i)
	}

	// Wait for all requests to complete
	wg.Wait()
	close(errors)

	// Check for errors
	var errorCount int
	for err := range errors {
		t.Logf("Concurrent request error: %v", err)
		errorCount++
	}

	// Some errors are expected (no active session), but server should still be running
	err = client.Ping()
	if err != nil {
		t.Error("Daemon stopped responding after concurrent requests")
	}

	cleanupTestEnv(tmpDir)
}

// TestDaemonIntegration_StaleFileCleanup tests cleanup of stale PID files
func TestDaemonIntegration_StaleFileCleanup(t *testing.T) {
	tmpDir := setupTestEnv(t)
	port := 9102

	// Create stale PID file with non-existent PID
	stalePIDFile := fmt.Sprintf("/tmp/xdebug-cli-daemon-%d.pid", port)
	stalePID := 999999
	err := os.WriteFile(stalePIDFile, []byte(fmt.Sprintf("%d", stalePID)), 0600)
	if err != nil {
		t.Fatalf("Failed to create stale PID file: %v", err)
	}

	// Create daemon
	server := dbgp.NewServer("127.0.0.1", port)
	d, err := daemon.NewDaemon(server, port)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// CheckExisting should detect and clean up stale file
	exists, pid, err := d.CheckExisting()
	if err != nil {
		t.Errorf("CheckExisting() error = %v", err)
	}

	if exists {
		t.Error("CheckExisting() should return false for stale PID")
	}

	if pid != 0 {
		t.Errorf("CheckExisting() pid = %d, want 0 for stale PID", pid)
	}

	// Verify stale PID file was removed
	if _, err := os.Stat(stalePIDFile); !os.IsNotExist(err) {
		t.Error("Stale PID file should be removed")
	}

	// Now start the daemon successfully
	mockClient := createMockClient()
	err = d.Start(mockClient)
	if err != nil {
		t.Fatalf("Start() after cleanup error = %v", err)
	}
	defer d.Shutdown()

	// Verify new PID file created
	if _, err := os.Stat(d.GetPIDFile()); os.IsNotExist(err) {
		t.Error("New PID file should exist after start")
	}

	cleanupTestEnv(tmpDir)
}

// TestDaemonIntegration_SessionEndBehavior tests daemon exit on session end
func TestDaemonIntegration_SessionEndBehavior(t *testing.T) {
	tmpDir := setupTestEnv(t)
	port := 9103

	// Create and start daemon
	server := dbgp.NewServer("127.0.0.1", port)
	d, err := daemon.NewDaemon(server, port)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	mockClient := createMockClient()
	err = d.Start(mockClient)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Give daemon time to start
	time.Sleep(100 * time.Millisecond)

	// Verify daemon is running
	client := ipc.NewClient(d.GetSocketPath())
	err = client.Ping()
	if err != nil {
		t.Errorf("Daemon should be running: %v", err)
	}

	// Send kill command
	resp, err := client.Kill()
	if err != nil {
		t.Errorf("Kill command failed: %v", err)
	}

	if resp != nil && !resp.Success {
		t.Errorf("Kill response indicates failure: %s", resp.Error)
	}

	// Wait for daemon to shutdown
	time.Sleep(300 * time.Millisecond)

	// Verify daemon is no longer responding
	err = client.Ping()
	if err == nil {
		t.Error("Daemon should not be responding after kill")
	}

	// Verify cleanup happened
	if _, err := os.Stat(d.GetPIDFile()); !os.IsNotExist(err) {
		t.Error("PID file should be removed after kill")
	}

	if _, err := os.Stat(d.GetSocketPath()); !os.IsNotExist(err) {
		t.Error("Socket file should be removed after kill")
	}

	// Verify registry cleaned up
	registry, err := daemon.NewSessionRegistry()
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	_, err = registry.Get(port)
	if err == nil {
		t.Error("Session should be removed from registry after kill")
	}

	cleanupTestEnv(tmpDir)
}

// TestDaemonIntegration_ConflictDetection tests daemon conflict detection
func TestDaemonIntegration_ConflictDetection(t *testing.T) {
	tmpDir := setupTestEnv(t)
	port := 9104

	// Start first daemon
	server1 := dbgp.NewServer("127.0.0.1", port)
	d1, err := daemon.NewDaemon(server1, port)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	mockClient := createMockClient()
	err = d1.Start(mockClient)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer d1.Shutdown()

	// Give daemon time to start
	time.Sleep(100 * time.Millisecond)

	// Try to create second daemon on same port
	server2 := dbgp.NewServer("127.0.0.1", port)
	d2, err := daemon.NewDaemon(server2, port)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// CheckExisting should detect the running daemon
	exists, pid, err := d2.CheckExisting()
	if err != nil {
		t.Errorf("CheckExisting() error = %v", err)
	}

	if !exists {
		t.Error("CheckExisting() should detect existing daemon")
	}

	if pid != os.Getpid() {
		t.Errorf("CheckExisting() pid = %d, want %d", pid, os.Getpid())
	}

	// Fork should fail
	err = d2.Fork([]string{"test"})
	if err == nil {
		t.Error("Fork() should fail when daemon already running")
	}

	// Error message should be informative
	if err != nil {
		errMsg := err.Error()
		if len(errMsg) < 10 {
			t.Errorf("Error message too short: %q", errMsg)
		}
	}

	cleanupTestEnv(tmpDir)
}

// TestDaemonIntegration_RegistryPersistence tests registry persistence
func TestDaemonIntegration_RegistryPersistence(t *testing.T) {
	tmpDir := setupTestEnv(t)
	port := 9105

	// Create and start daemon
	server := dbgp.NewServer("127.0.0.1", port)
	d, err := daemon.NewDaemon(server, port)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	mockClient := createMockClient()
	err = d.Start(mockClient)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Give daemon time to start
	time.Sleep(100 * time.Millisecond)

	// Create new registry instance (simulating fresh process)
	registry2, err := daemon.NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	// Session should be accessible from new registry
	session, err := registry2.Get(port)
	if err != nil {
		t.Errorf("Session not accessible from new registry: %v", err)
	}

	if session != nil && session.Port != port {
		t.Errorf("Session port = %d, want %d", session.Port, port)
	}

	if session != nil && session.PID != os.Getpid() {
		t.Errorf("Session PID = %d, want %d", session.PID, os.Getpid())
	}

	// Verify socket path is correct
	if session != nil && session.SocketPath != d.GetSocketPath() {
		t.Errorf("Session socket path = %q, want %q", session.SocketPath, d.GetSocketPath())
	}

	// Cleanup
	d.Shutdown()
	time.Sleep(200 * time.Millisecond)

	// Create third registry instance
	registry3, err := daemon.NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	// Session should be removed after shutdown
	_, err = registry3.Get(port)
	if err == nil {
		t.Error("Session should be removed from registry after shutdown")
	}

	cleanupTestEnv(tmpDir)
}

// TestDaemonIntegration_MultiplePortsDaemon tests multiple daemons on different ports
func TestDaemonIntegration_MultiplePortsDaemon(t *testing.T) {
	tmpDir := setupTestEnv(t)
	port1 := 9106
	port2 := 9107

	// Start first daemon
	server1 := dbgp.NewServer("127.0.0.1", port1)
	d1, err := daemon.NewDaemon(server1, port1)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	mockClient1 := createMockClient()
	err = d1.Start(mockClient1)
	if err != nil {
		t.Fatalf("Start() d1 error = %v", err)
	}
	defer d1.Shutdown()

	// Start second daemon on different port
	server2 := dbgp.NewServer("127.0.0.1", port2)
	d2, err := daemon.NewDaemon(server2, port2)
	if err != nil {
		t.Fatalf("NewDaemon() d2 error = %v", err)
	}

	mockClient2 := createMockClient()
	err = d2.Start(mockClient2)
	if err != nil {
		t.Fatalf("Start() d2 error = %v", err)
	}
	defer d2.Shutdown()

	// Give daemons time to start
	time.Sleep(100 * time.Millisecond)

	// Verify both daemons are running
	client1 := ipc.NewClient(d1.GetSocketPath())
	err = client1.Ping()
	if err != nil {
		t.Errorf("Daemon 1 not responding: %v", err)
	}

	client2 := ipc.NewClient(d2.GetSocketPath())
	err = client2.Ping()
	if err != nil {
		t.Errorf("Daemon 2 not responding: %v", err)
	}

	// Verify registry has both sessions
	registry, err := daemon.NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	sessions := registry.List()
	if len(sessions) < 2 {
		t.Errorf("Registry should have at least 2 sessions, got %d", len(sessions))
	}

	// Verify both sessions are accessible
	session1, err := registry.Get(port1)
	if err != nil {
		t.Errorf("Session 1 not in registry: %v", err)
	}
	if session1 != nil && session1.Port != port1 {
		t.Errorf("Session 1 port = %d, want %d", session1.Port, port1)
	}

	session2, err := registry.Get(port2)
	if err != nil {
		t.Errorf("Session 2 not in registry: %v", err)
	}
	if session2 != nil && session2.Port != port2 {
		t.Errorf("Session 2 port = %d, want %d", session2.Port, port2)
	}

	cleanupTestEnv(tmpDir)
}

// TestDaemonIntegration_IPCTimeout tests IPC timeout handling
func TestDaemonIntegration_IPCTimeout(t *testing.T) {
	tmpDir := setupTestEnv(t)
	port := 9108

	// Create daemon but don't start it
	server := dbgp.NewServer("127.0.0.1", port)
	d, err := daemon.NewDaemon(server, port)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	// Try to connect to non-existent daemon
	client := ipc.NewClient(d.GetSocketPath())
	client.SetTimeout(1 * time.Second)

	startTime := time.Now()
	err = client.Ping()
	duration := time.Since(startTime)

	if err == nil {
		t.Error("Ping() should fail for non-existent daemon")
	}

	// Should timeout within reasonable time
	if duration > 2*time.Second {
		t.Errorf("Timeout took too long: %v", duration)
	}

	cleanupTestEnv(tmpDir)
}

// TestDaemonIntegration_CleanupOnError tests cleanup when daemon encounters error
func TestDaemonIntegration_CleanupOnError(t *testing.T) {
	tmpDir := setupTestEnv(t)
	port := 9109

	// Create daemon
	server := dbgp.NewServer("127.0.0.1", port)
	d, err := daemon.NewDaemon(server, port)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	mockClient := createMockClient()
	err = d.Start(mockClient)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Give daemon time to start
	time.Sleep(100 * time.Millisecond)

	// Verify files exist
	if _, err := os.Stat(d.GetPIDFile()); os.IsNotExist(err) {
		t.Error("PID file should exist")
	}

	// Simulate error by calling Shutdown
	err = d.Shutdown()
	if err != nil {
		t.Logf("Shutdown() error = %v (may be expected)", err)
	}

	// Wait for cleanup
	time.Sleep(200 * time.Millisecond)

	// Verify cleanup despite error
	if _, err := os.Stat(d.GetPIDFile()); !os.IsNotExist(err) {
		t.Error("PID file should be cleaned up even on error")
	}

	if _, err := os.Stat(d.GetSocketPath()); !os.IsNotExist(err) {
		t.Error("Socket file should be cleaned up even on error")
	}

	// Verify registry cleanup
	registry, err := daemon.NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	_, err = registry.Get(port)
	if err == nil {
		t.Error("Session should be removed from registry")
	}

	cleanupTestEnv(tmpDir)
}

// TestDaemonIntegration_SocketPermissions tests Unix socket permissions
func TestDaemonIntegration_SocketPermissions(t *testing.T) {
	tmpDir := setupTestEnv(t)
	port := 9110

	// Create and start daemon
	server := dbgp.NewServer("127.0.0.1", port)
	d, err := daemon.NewDaemon(server, port)
	if err != nil {
		t.Fatalf("NewDaemon() error = %v", err)
	}

	mockClient := createMockClient()
	err = d.Start(mockClient)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer d.Shutdown()

	// Give daemon time to start
	time.Sleep(100 * time.Millisecond)

	// Check socket permissions
	info, err := os.Stat(d.GetSocketPath())
	if err != nil {
		t.Fatalf("Failed to stat socket: %v", err)
	}

	mode := info.Mode()
	// Socket should be 0600 (owner read/write only)
	expectedMode := os.FileMode(0600)
	if mode.Perm() != expectedMode {
		t.Errorf("Socket permissions = %o, want %o", mode.Perm(), expectedMode)
	}

	cleanupTestEnv(tmpDir)
}

// Helper functions

func setupTestEnv(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "xdebug-cli-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Set HOME to temp dir for registry
	os.Setenv("HOME", tmpDir)

	// Create .xdebug-cli directory
	registryDir := filepath.Join(tmpDir, ".xdebug-cli")
	if err := os.MkdirAll(registryDir, 0700); err != nil {
		t.Fatalf("Failed to create registry dir: %v", err)
	}

	return tmpDir
}

func cleanupTestEnv(tmpDir string) {
	// Restore HOME
	homeDir, err := os.UserHomeDir()
	if err == nil {
		os.Setenv("HOME", homeDir)
	}

	// Remove temp directory
	os.RemoveAll(tmpDir)
}
