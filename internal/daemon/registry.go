package daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SessionInfo represents information about an active daemon session
type SessionInfo struct {
	PID        int       `json:"pid"`
	Port       int       `json:"port"`
	SocketPath string    `json:"socket_path"`
	StartedAt  time.Time `json:"started_at"`
}

// SessionRegistry manages the registry of active daemon sessions
type SessionRegistry struct {
	path     string
	mu       sync.Mutex
	sessions []SessionInfo
}

// NewSessionRegistry creates a new session registry
func NewSessionRegistry() (*SessionRegistry, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	registryDir := filepath.Join(homeDir, ".xdebug-cli")
	registryPath := filepath.Join(registryDir, "sessions.json")

	// Create registry directory if it doesn't exist
	if err := os.MkdirAll(registryDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create registry directory: %w", err)
	}

	registry := &SessionRegistry{
		path:     registryPath,
		sessions: []SessionInfo{},
	}

	// Load existing sessions
	if err := registry.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load registry: %w", err)
	}

	// Clean up stale entries
	if err := registry.cleanupStale(); err != nil {
		return nil, fmt.Errorf("failed to cleanup stale entries: %w", err)
	}

	return registry, nil
}

// Add adds a new session to the registry
func (r *SessionRegistry) Add(session SessionInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate port
	for _, s := range r.sessions {
		if s.Port == session.Port {
			return fmt.Errorf("session already exists on port %d (PID %d)", session.Port, s.PID)
		}
	}

	r.sessions = append(r.sessions, session)
	return r.save()
}

// Remove removes a session from the registry by port
func (r *SessionRegistry) Remove(port int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	newSessions := []SessionInfo{}
	found := false
	for _, s := range r.sessions {
		if s.Port != port {
			newSessions = append(newSessions, s)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("session not found for port %d", port)
	}

	r.sessions = newSessions
	return r.save()
}

// Get retrieves a session by port
func (r *SessionRegistry) Get(port int) (*SessionInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, s := range r.sessions {
		if s.Port == port {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("no session found for port %d", port)
}

// List returns all active sessions
func (r *SessionRegistry) List() []SessionInfo {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Return copy to avoid external modification
	result := make([]SessionInfo, len(r.sessions))
	copy(result, r.sessions)
	return result
}

// load loads sessions from the registry file
func (r *SessionRegistry) load() error {
	data, err := os.ReadFile(r.path)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		r.sessions = []SessionInfo{}
		return nil
	}

	if err := json.Unmarshal(data, &r.sessions); err != nil {
		return fmt.Errorf("failed to unmarshal registry: %w", err)
	}

	return nil
}

// save writes sessions to the registry file
func (r *SessionRegistry) save() error {
	data, err := json.MarshalIndent(r.sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	if err := os.WriteFile(r.path, data, 0600); err != nil {
		return fmt.Errorf("failed to write registry: %w", err)
	}

	return nil
}

// cleanupStale removes sessions whose processes no longer exist or have been recycled
func (r *SessionRegistry) cleanupStale() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	activeSessions := []SessionInfo{}
	for _, s := range r.sessions {
		if validateProcess(s.PID) {
			activeSessions = append(activeSessions, s)
		} else {
			// Clean up orphaned socket file
			if s.SocketPath != "" {
				os.Remove(s.SocketPath)
			}
		}
	}

	if len(activeSessions) != len(r.sessions) {
		r.sessions = activeSessions
		return r.save()
	}

	return nil
}

// CleanupStaleEntries is the public method to clean up stale entries
// This is called explicitly by daemon start command
func (r *SessionRegistry) CleanupStaleEntries() error {
	return r.cleanupStale()
}

// validateProcess checks if a process with the given PID exists and is xdebug-cli
func validateProcess(pid int) bool {
	// Check if /proc/<pid> exists (Linux-specific)
	procPath := fmt.Sprintf("/proc/%d", pid)
	if _, err := os.Stat(procPath); err != nil {
		return false
	}

	// Verify process is xdebug-cli by checking /proc/<pid>/comm
	commPath := fmt.Sprintf("/proc/%d/comm", pid)
	commData, err := os.ReadFile(commPath)
	if err != nil {
		return false
	}

	// comm contains the executable name, usually with a trailing newline
	// For xdebug-cli, it will be "xdebug-cli\n"
	// For test binaries, it will be "daemon.test\n", "cli.test\n", etc.
	comm := string(commData)

	// Check if it starts with "xdebug-cli" (the actual binary)
	if len(comm) >= 10 && comm[0:10] == "xdebug-cli" {
		return true
	}
	// Check for test binaries - any .test suffix
	if len(comm) >= 6 && comm[len(comm)-6:len(comm)-1] == ".test" {
		return true
	}

	return false
}

// processExists checks if a process with the given PID exists
// NOTE: This is less strict than validateProcess - it only checks existence
func processExists(pid int) bool {
	// Check if /proc/<pid> exists (Linux-specific)
	procPath := fmt.Sprintf("/proc/%d", pid)
	_, err := os.Stat(procPath)
	return err == nil
}

// Path returns the registry file path
func (r *SessionRegistry) Path() string {
	return r.path
}
