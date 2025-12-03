package daemon

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// BreakpointPathStore manages persistent storage of successful breakpoint paths.
// It stores a mapping from filename to full path for use in suggestions.
type BreakpointPathStore struct {
	mu       sync.RWMutex
	filePath string
	paths    map[string]string // filename -> full path
}

// configDir returns the xdebug-cli config directory, creating it if needed
func configDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(homeDir, ".xdebug-cli")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	return dir, nil
}

// NewBreakpointPathStore creates a new breakpoint path store
func NewBreakpointPathStore() (*BreakpointPathStore, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}

	store := &BreakpointPathStore{
		filePath: filepath.Join(dir, "breakpoint-paths.json"),
		paths:    make(map[string]string),
	}

	// Load existing paths (ignore errors - file may not exist)
	store.load()

	return store, nil
}

// load reads the paths from the JSON file
func (s *BreakpointPathStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's OK
		}
		return err
	}

	return json.Unmarshal(data, &s.paths)
}

// save writes the paths to the JSON file
func (s *BreakpointPathStore) save() error {
	data, err := json.MarshalIndent(s.paths, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

// SaveBreakpointPath stores a successful breakpoint path for future suggestions.
// The fullPath should be an absolute path that was successfully hit.
func (s *BreakpointPathStore) SaveBreakpointPath(fullPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Extract filename from path
	filename := extractFilename(fullPath)
	if filename == "" {
		return nil // Nothing to save
	}

	// Store the mapping
	s.paths[filename] = fullPath

	return s.save()
}

// LoadBreakpointPath looks up a suggested full path for a filename.
// Returns empty string if no suggestion is available.
// The returned path is a clean absolute path (file:// prefix stripped).
func (s *BreakpointPathStore) LoadBreakpointPath(filename string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Extract just the filename if a path was provided
	filename = extractFilename(filename)

	path := s.paths[filename]
	// Strip file:// prefix if present (Xdebug stores paths with this prefix)
	path = strings.TrimPrefix(path, "file://")
	return path
}

// IsAbsolutePath returns true if the path is absolute (starts with /)
func IsAbsolutePath(path string) bool {
	return strings.HasPrefix(path, "/")
}

// extractFilename returns the filename component from a path
func extractFilename(path string) string {
	// Handle file:// URIs
	path = strings.TrimPrefix(path, "file://")

	// Get the base name
	return filepath.Base(path)
}

// HasNonAbsoluteBreakpoint scans commands for break commands with non-absolute paths.
// Returns (hasNonAbsolute, breakpointPath) where breakpointPath is the first
// non-absolute path found, or empty if all are absolute.
func HasNonAbsoluteBreakpoint(commands []string) (bool, string) {
	for _, cmd := range commands {
		// Check if this is a break command
		if !strings.HasPrefix(cmd, "break ") && !strings.HasPrefix(cmd, "b ") {
			continue
		}

		// Parse the command to extract the breakpoint target
		parts := strings.Fields(cmd)
		if len(parts) < 2 {
			continue
		}

		// Skip "break call" and "break exception" which don't have file paths
		if parts[1] == "call" || parts[1] == "exception" {
			continue
		}

		// Get the location (could be :line, file:line, or just line)
		location := parts[1]

		// Skip :line format (uses current file) - this is OK
		if strings.HasPrefix(location, ":") {
			continue
		}

		// Check for file:line format
		if strings.Contains(location, ":") {
			filePart := strings.Split(location, ":")[0]
			if !IsAbsolutePath(filePart) {
				return true, location
			}
		}
	}

	return false, ""
}
