package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetInstallDir(t *testing.T) {
	installDir := getInstallDir()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	expectedPath := filepath.Join(homeDir, ".local", "bin")

	if installDir != expectedPath {
		t.Errorf("getInstallDir() = %s; want %s", installDir, expectedPath)
	}

	// Verify it's an absolute path
	if !filepath.IsAbs(installDir) {
		t.Errorf("getInstallDir() returned relative path: %s", installDir)
	}
}

func TestEnsureDir(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test", "nested", "dir")

	// Verify directory doesn't exist yet
	if _, err := os.Stat(testPath); !os.IsNotExist(err) {
		t.Fatalf("Test directory should not exist yet: %s", testPath)
	}

	// Test ensureDir creates the directory
	err := ensureDir(testPath)
	if err != nil {
		t.Fatalf("ensureDir() failed: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(testPath)
	if err != nil {
		t.Fatalf("Directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Errorf("Path exists but is not a directory: %s", testPath)
	}

	// Test that calling ensureDir on existing directory doesn't fail
	err = ensureDir(testPath)
	if err != nil {
		t.Errorf("ensureDir() should not fail on existing directory: %v", err)
	}
}

func TestCopyBinary(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a source file with test content
	srcPath := filepath.Join(tmpDir, "source")
	testContent := []byte("test binary content")
	err := os.WriteFile(srcPath, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Test copying to destination
	dstPath := filepath.Join(tmpDir, "destination")
	err = copyBinary(srcPath, dstPath)
	if err != nil {
		t.Fatalf("copyBinary() failed: %v", err)
	}

	// Verify destination file exists
	info, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("Destination file was not created: %v", err)
	}

	// Verify file is executable (has execute permission)
	mode := info.Mode()
	if mode&0111 == 0 {
		t.Errorf("Destination file is not executable: %v", mode)
	}

	// Verify content was copied correctly
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != string(testContent) {
		t.Errorf("File content mismatch: got %s, want %s", dstContent, testContent)
	}

	// Test copying to non-existent directory fails appropriately
	badDstPath := filepath.Join(tmpDir, "nonexistent", "dir", "file")
	err = copyBinary(srcPath, badDstPath)
	if err == nil {
		t.Error("copyBinary() should fail when destination directory doesn't exist")
	}

	// Test copying non-existent source file fails
	err = copyBinary(filepath.Join(tmpDir, "nonexistent"), dstPath)
	if err == nil {
		t.Error("copyBinary() should fail when source file doesn't exist")
	}
}

func TestRunInstall(t *testing.T) {
	// Create a temporary directory to act as install directory
	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, ".local", "bin")

	// Create a temporary executable to install
	tmpExe := filepath.Join(tmpDir, "test-exe")
	testContent := []byte("#!/bin/bash\necho test")
	err := os.WriteFile(tmpExe, testContent, 0755)
	if err != nil {
		t.Fatalf("Failed to create test executable: %v", err)
	}

	// Test the install process
	err = runInstallWithPaths(tmpExe, installDir)
	if err != nil {
		t.Fatalf("runInstallWithPaths() failed: %v", err)
	}

	// Verify the binary was installed
	installedPath := filepath.Join(installDir, "xdebug-cli")
	info, err := os.Stat(installedPath)
	if err != nil {
		t.Fatalf("Installed binary not found: %v", err)
	}

	// Verify it's executable
	if info.Mode()&0111 == 0 {
		t.Errorf("Installed binary is not executable: %v", info.Mode())
	}

	// Verify content
	installedContent, err := os.ReadFile(installedPath)
	if err != nil {
		t.Fatalf("Failed to read installed binary: %v", err)
	}

	if string(installedContent) != string(testContent) {
		t.Errorf("Installed binary content mismatch")
	}
}
