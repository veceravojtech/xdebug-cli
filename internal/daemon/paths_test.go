package daemon

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsAbsolutePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"absolute path", "/var/www/file.php", true},
		{"absolute path with spaces", "/var/www/my file.php", true},
		{"filename only", "PriceLoader.php", false},
		{"relative path", "app/models/User.php", false},
		{"relative with dot", "./file.php", false},
		{"relative with parent", "../file.php", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAbsolutePath(tt.path)
			if result != tt.expected {
				t.Errorf("IsAbsolutePath(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestExtractFilename(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"absolute path", "/var/www/file.php", "file.php"},
		{"filename only", "PriceLoader.php", "PriceLoader.php"},
		{"relative path", "app/models/User.php", "User.php"},
		{"file:// URI", "file:///var/www/file.php", "file.php"},
		{"empty string", "", "."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFilename(tt.path)
			if result != tt.expected {
				t.Errorf("extractFilename(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestHasNonAbsoluteBreakpoint(t *testing.T) {
	tests := []struct {
		name           string
		commands       []string
		expectNonAbs   bool
		expectPath     string
	}{
		{
			name:         "absolute path",
			commands:     []string{"break /var/www/file.php:100"},
			expectNonAbs: false,
			expectPath:   "",
		},
		{
			name:         "filename only",
			commands:     []string{"break PriceLoader.php:100"},
			expectNonAbs: true,
			expectPath:   "PriceLoader.php:100",
		},
		{
			name:         "relative path",
			commands:     []string{"break app/models/User.php:50"},
			expectNonAbs: true,
			expectPath:   "app/models/User.php:50",
		},
		{
			name:         "line number only",
			commands:     []string{"break :42"},
			expectNonAbs: false,
			expectPath:   "",
		},
		{
			name:         "break call",
			commands:     []string{"break call myFunction"},
			expectNonAbs: false,
			expectPath:   "",
		},
		{
			name:         "break exception",
			commands:     []string{"break exception"},
			expectNonAbs: false,
			expectPath:   "",
		},
		{
			name:         "short form b",
			commands:     []string{"b file.php:100"},
			expectNonAbs: true,
			expectPath:   "file.php:100",
		},
		{
			name:         "multiple commands with non-absolute",
			commands:     []string{"run", "break file.php:100", "context local"},
			expectNonAbs: true,
			expectPath:   "file.php:100",
		},
		{
			name:         "multiple commands all absolute",
			commands:     []string{"run", "break /var/www/file.php:100", "context local"},
			expectNonAbs: false,
			expectPath:   "",
		},
		{
			name:         "no break commands",
			commands:     []string{"run", "context local"},
			expectNonAbs: false,
			expectPath:   "",
		},
		{
			name:         "empty commands",
			commands:     []string{},
			expectNonAbs: false,
			expectPath:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasNonAbs, path := HasNonAbsoluteBreakpoint(tt.commands)
			if hasNonAbs != tt.expectNonAbs {
				t.Errorf("HasNonAbsoluteBreakpoint(%v) hasNonAbs = %v, want %v", tt.commands, hasNonAbs, tt.expectNonAbs)
			}
			if path != tt.expectPath {
				t.Errorf("HasNonAbsoluteBreakpoint(%v) path = %q, want %q", tt.commands, path, tt.expectPath)
			}
		})
	}
}

func TestBreakpointPathStore(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()

	// Override home directory for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("save and load path", func(t *testing.T) {
		store, err := NewBreakpointPathStore()
		if err != nil {
			t.Fatalf("NewBreakpointPathStore() error = %v", err)
		}

		// Save a path
		fullPath := "/var/www/app/PriceLoader.php"
		err = store.SaveBreakpointPath(fullPath)
		if err != nil {
			t.Fatalf("SaveBreakpointPath() error = %v", err)
		}

		// Load the path using filename
		loaded := store.LoadBreakpointPath("PriceLoader.php")
		if loaded != fullPath {
			t.Errorf("LoadBreakpointPath() = %q, want %q", loaded, fullPath)
		}
	})

	t.Run("load non-existent path", func(t *testing.T) {
		store, err := NewBreakpointPathStore()
		if err != nil {
			t.Fatalf("NewBreakpointPathStore() error = %v", err)
		}

		loaded := store.LoadBreakpointPath("NonExistent.php")
		if loaded != "" {
			t.Errorf("LoadBreakpointPath() = %q, want empty string", loaded)
		}
	})

	t.Run("persistence across instances", func(t *testing.T) {
		// Create first store and save path
		store1, err := NewBreakpointPathStore()
		if err != nil {
			t.Fatalf("NewBreakpointPathStore() error = %v", err)
		}

		fullPath := "/var/www/app/Controller.php"
		err = store1.SaveBreakpointPath(fullPath)
		if err != nil {
			t.Fatalf("SaveBreakpointPath() error = %v", err)
		}

		// Create second store instance and verify path persisted
		store2, err := NewBreakpointPathStore()
		if err != nil {
			t.Fatalf("NewBreakpointPathStore() error = %v", err)
		}

		loaded := store2.LoadBreakpointPath("Controller.php")
		if loaded != fullPath {
			t.Errorf("LoadBreakpointPath() = %q, want %q", loaded, fullPath)
		}
	})

	t.Run("update existing path", func(t *testing.T) {
		store, err := NewBreakpointPathStore()
		if err != nil {
			t.Fatalf("NewBreakpointPathStore() error = %v", err)
		}

		// Save original path
		originalPath := "/var/www/old/File.php"
		err = store.SaveBreakpointPath(originalPath)
		if err != nil {
			t.Fatalf("SaveBreakpointPath() error = %v", err)
		}

		// Update with new path
		newPath := "/var/www/new/File.php"
		err = store.SaveBreakpointPath(newPath)
		if err != nil {
			t.Fatalf("SaveBreakpointPath() error = %v", err)
		}

		// Verify new path is returned
		loaded := store.LoadBreakpointPath("File.php")
		if loaded != newPath {
			t.Errorf("LoadBreakpointPath() = %q, want %q", loaded, newPath)
		}
	})

	t.Run("config dir creation", func(t *testing.T) {
		// Verify config directory was created
		configPath := filepath.Join(tempDir, ".xdebug-cli")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Errorf("Config directory not created at %s", configPath)
		}
	})

	t.Run("strips file:// prefix from loaded path", func(t *testing.T) {
		store, err := NewBreakpointPathStore()
		if err != nil {
			t.Fatalf("NewBreakpointPathStore() error = %v", err)
		}

		// Save a path with file:// prefix (as Xdebug reports it)
		xdebugPath := "file:///var/www/app/Service.php"
		err = store.SaveBreakpointPath(xdebugPath)
		if err != nil {
			t.Fatalf("SaveBreakpointPath() error = %v", err)
		}

		// Load should return clean absolute path without file:// prefix
		loaded := store.LoadBreakpointPath("Service.php")
		expected := "/var/www/app/Service.php"
		if loaded != expected {
			t.Errorf("LoadBreakpointPath() = %q, want %q (should strip file:// prefix)", loaded, expected)
		}
	})
}
