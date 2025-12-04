package daemon

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewSessionRegistry(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	registry, err := NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	// Registry directory should be created
	registryDir := filepath.Join(tempDir, ".xdebug-cli")
	if _, err := os.Stat(registryDir); os.IsNotExist(err) {
		t.Errorf("Registry directory was not created: %s", registryDir)
	}

	// Registry file path should be correct
	expectedPath := filepath.Join(registryDir, "sessions.json")
	if registry.Path() != expectedPath {
		t.Errorf("Registry path = %s, want %s", registry.Path(), expectedPath)
	}
}

func TestSessionRegistry_Add(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	registry, err := NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	session := SessionInfo{
		PID:        12345,
		Port:       9003,
		SocketPath: "/tmp/test.sock",
		StartedAt:  time.Now(),
	}

	// Add should succeed
	if err := registry.Add(session); err != nil {
		t.Errorf("Add() error = %v", err)
	}

	// Session should be in registry
	retrieved, err := registry.Get(9003)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if retrieved.PID != 12345 {
		t.Errorf("Retrieved PID = %d, want 12345", retrieved.PID)
	}
	if retrieved.Port != 9003 {
		t.Errorf("Retrieved Port = %d, want 9003", retrieved.Port)
	}
}

func TestSessionRegistry_Add_Duplicate(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	registry, err := NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	session := SessionInfo{
		PID:        12345,
		Port:       9003,
		SocketPath: "/tmp/test.sock",
		StartedAt:  time.Now(),
	}

	// First add succeeds
	if err := registry.Add(session); err != nil {
		t.Fatalf("First Add() error = %v", err)
	}

	// Second add with same port should fail
	session.PID = 67890
	if err := registry.Add(session); err == nil {
		t.Error("Second Add() expected error for duplicate port")
	}
}

func TestSessionRegistry_Remove(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	registry, err := NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	session := SessionInfo{
		PID:        12345,
		Port:       9003,
		SocketPath: "/tmp/test.sock",
		StartedAt:  time.Now(),
	}

	registry.Add(session)

	// Remove should succeed
	if err := registry.Remove(9003); err != nil {
		t.Errorf("Remove() error = %v", err)
	}

	// Session should no longer exist
	_, err = registry.Get(9003)
	if err == nil {
		t.Error("Get() expected error after Remove()")
	}
}

func TestSessionRegistry_Remove_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	registry, err := NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	// Remove non-existent session should fail
	if err := registry.Remove(9999); err == nil {
		t.Error("Remove() expected error for non-existent session")
	}
}

func TestSessionRegistry_Get(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	registry, err := NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	session := SessionInfo{
		PID:        12345,
		Port:       9003,
		SocketPath: "/tmp/test.sock",
		StartedAt:  time.Now(),
	}

	registry.Add(session)

	// Get should return the session
	retrieved, err := registry.Get(9003)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if retrieved.PID != session.PID {
		t.Errorf("PID = %d, want %d", retrieved.PID, session.PID)
	}
	if retrieved.Port != session.Port {
		t.Errorf("Port = %d, want %d", retrieved.Port, session.Port)
	}
	if retrieved.SocketPath != session.SocketPath {
		t.Errorf("SocketPath = %s, want %s", retrieved.SocketPath, session.SocketPath)
	}
}

func TestSessionRegistry_Get_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	registry, err := NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	_, err = registry.Get(9999)
	if err == nil {
		t.Error("Get() expected error for non-existent session")
	}
}

func TestSessionRegistry_List(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	registry, err := NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	// Initially empty
	list := registry.List()
	if len(list) != 0 {
		t.Errorf("List() length = %d, want 0", len(list))
	}

	// Add sessions
	session1 := SessionInfo{PID: 111, Port: 9003, SocketPath: "/tmp/1.sock", StartedAt: time.Now()}
	session2 := SessionInfo{PID: 222, Port: 9004, SocketPath: "/tmp/2.sock", StartedAt: time.Now()}

	registry.Add(session1)
	registry.Add(session2)

	// List should return both
	list = registry.List()
	if len(list) != 2 {
		t.Errorf("List() length = %d, want 2", len(list))
	}
}

func TestSessionRegistry_Persistence(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	// Create registry and add session
	registry1, err := NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	session := SessionInfo{
		PID:        os.Getpid(), // Use current PID so it won't be cleaned up
		Port:       9003,
		SocketPath: "/tmp/test.sock",
		StartedAt:  time.Now(),
	}

	if err := registry1.Add(session); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Create new registry instance (should load from file)
	registry2, err := NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() second instance error = %v", err)
	}

	// Session should still exist
	retrieved, err := registry2.Get(9003)
	if err != nil {
		t.Fatalf("Get() error = %v after reload", err)
	}
	if retrieved.PID != session.PID {
		t.Errorf("Retrieved PID = %d, want %d", retrieved.PID, session.PID)
	}
}

func TestSessionRegistry_CleanupStale(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	registry, err := NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	// Add session with current PID (exists)
	existingSession := SessionInfo{
		PID:        os.Getpid(),
		Port:       9003,
		SocketPath: "/tmp/existing.sock",
		StartedAt:  time.Now(),
	}
	registry.Add(existingSession)

	// Add session with non-existent PID
	staleSession := SessionInfo{
		PID:        999999, // Unlikely to exist
		Port:       9004,
		SocketPath: "/tmp/stale.sock",
		StartedAt:  time.Now(),
	}
	registry.Add(staleSession)

	// Create new registry (triggers cleanup)
	registry2, err := NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	// Existing session should still be there
	if _, err := registry2.Get(9003); err != nil {
		t.Errorf("Existing session was cleaned up: %v", err)
	}

	// Stale session should be removed
	if _, err := registry2.Get(9004); err == nil {
		t.Error("Stale session was not cleaned up")
	}
}

func TestProcessExists(t *testing.T) {
	// Current process should exist
	if !processExists(os.Getpid()) {
		t.Error("processExists() = false for current process")
	}

	// Non-existent PID
	if processExists(999999) {
		t.Error("processExists() = true for non-existent PID")
	}
}

func TestValidateProcess(t *testing.T) {
	// Current process should be valid (it's xdebug-cli running tests)
	if !validateProcess(os.Getpid()) {
		t.Error("validateProcess() = false for current xdebug-cli process")
	}

	// Non-existent PID should be invalid
	if validateProcess(999999) {
		t.Error("validateProcess() = true for non-existent PID")
	}
}

func TestValidateProcess_RecycledPID(t *testing.T) {
	// Test that validateProcess correctly identifies non-xdebug-cli processes
	// We'll use PID 1 (init/systemd) as a known non-xdebug-cli process
	if validateProcess(1) {
		t.Error("validateProcess() = true for init process (PID 1), expected false")
	}
}

func TestCleanupStaleEntries_NonExistentProcess(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	registry, err := NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	// Add session with current PID (exists and is xdebug-cli)
	validSession := SessionInfo{
		PID:        os.Getpid(),
		Port:       9003,
		SocketPath: "/tmp/valid.sock",
		StartedAt:  time.Now(),
	}
	registry.Add(validSession)

	// Add session with non-existent PID
	staleSession := SessionInfo{
		PID:        999999, // Very unlikely to exist
		Port:       9004,
		SocketPath: "/tmp/stale.sock",
		StartedAt:  time.Now(),
	}
	registry.Add(staleSession)

	// Call CleanupStaleEntries
	if err := registry.CleanupStaleEntries(); err != nil {
		t.Fatalf("CleanupStaleEntries() error = %v", err)
	}

	// Valid session should still exist
	if _, err := registry.Get(9003); err != nil {
		t.Errorf("Valid session was removed: %v", err)
	}

	// Stale session should be removed
	if _, err := registry.Get(9004); err == nil {
		t.Error("Stale session was not removed")
	}
}

func TestCleanupStaleEntries_RecycledPID(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	registry, err := NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	// Add session with current PID (exists and is xdebug-cli)
	validSession := SessionInfo{
		PID:        os.Getpid(),
		Port:       9003,
		SocketPath: "/tmp/valid.sock",
		StartedAt:  time.Now(),
	}
	registry.Add(validSession)

	// Add session with PID 1 (init/systemd - exists but is NOT xdebug-cli)
	// This simulates a recycled PID
	recycledSession := SessionInfo{
		PID:        1,
		Port:       9004,
		SocketPath: "/tmp/recycled.sock",
		StartedAt:  time.Now(),
	}
	registry.Add(recycledSession)

	// Call CleanupStaleEntries
	if err := registry.CleanupStaleEntries(); err != nil {
		t.Fatalf("CleanupStaleEntries() error = %v", err)
	}

	// Valid session should still exist
	if _, err := registry.Get(9003); err != nil {
		t.Errorf("Valid session was removed: %v", err)
	}

	// Recycled PID session should be removed
	if _, err := registry.Get(9004); err == nil {
		t.Error("Recycled PID session was not removed")
	}
}

func TestCleanupStaleEntries_PreservesValidEntries(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	registry, err := NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	// Add multiple valid sessions (all with current PID)
	session1 := SessionInfo{
		PID:        os.Getpid(),
		Port:       9003,
		SocketPath: "/tmp/session1.sock",
		StartedAt:  time.Now(),
	}
	session2 := SessionInfo{
		PID:        os.Getpid(),
		Port:       9004,
		SocketPath: "/tmp/session2.sock",
		StartedAt:  time.Now(),
	}
	registry.Add(session1)
	registry.Add(session2)

	// Call CleanupStaleEntries
	if err := registry.CleanupStaleEntries(); err != nil {
		t.Fatalf("CleanupStaleEntries() error = %v", err)
	}

	// Both sessions should still exist
	if _, err := registry.Get(9003); err != nil {
		t.Errorf("Session 1 was removed: %v", err)
	}
	if _, err := registry.Get(9004); err != nil {
		t.Errorf("Session 2 was removed: %v", err)
	}

	// List should have 2 entries
	sessions := registry.List()
	if len(sessions) != 2 {
		t.Errorf("List() length = %d, want 2", len(sessions))
	}
}

func TestCleanupStaleEntries_OrphanedSocketFiles(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	registry, err := NewSessionRegistry()
	if err != nil {
		t.Fatalf("NewSessionRegistry() error = %v", err)
	}

	// Create a socket file
	socketPath := filepath.Join(tempDir, "orphaned.sock")
	if err := os.WriteFile(socketPath, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create socket file: %v", err)
	}

	// Add session with non-existent PID and the socket path
	staleSession := SessionInfo{
		PID:        999999,
		Port:       9004,
		SocketPath: socketPath,
		StartedAt:  time.Now(),
	}
	registry.Add(staleSession)

	// Verify socket file exists before cleanup
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Fatal("Socket file should exist before cleanup")
	}

	// Call CleanupStaleEntries
	if err := registry.CleanupStaleEntries(); err != nil {
		t.Fatalf("CleanupStaleEntries() error = %v", err)
	}

	// Socket file should be removed
	if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
		t.Error("Orphaned socket file was not removed")
	}

	// Stale session should be removed from registry
	if _, err := registry.Get(9004); err == nil {
		t.Error("Stale session was not removed from registry")
	}
}
