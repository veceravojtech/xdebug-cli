package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install xdebug-cli to ~/.local/bin",
	Long:  `Install xdebug-cli binary to ~/.local/bin directory and make it executable.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runInstall(); err != nil {
			fmt.Fprintf(os.Stderr, "Installation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully installed xdebug-cli to %s\n", filepath.Join(getInstallDir(), "xdebug-cli"))
		fmt.Println("Make sure ~/.local/bin is in your PATH")
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

// getInstallDir returns the installation directory path (~/.local/bin)
func getInstallDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// This should rarely happen, but fall back to a sensible default
		return filepath.Join("/tmp", ".local", "bin")
	}
	return filepath.Join(homeDir, ".local", "bin")
}

// ensureDir creates the directory if it doesn't exist
func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// copyBinary copies a file from src to dst and makes it executable
func copyBinary(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy content
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Make executable (0755 = rwxr-xr-x)
	err = os.Chmod(dst, 0755)
	if err != nil {
		return fmt.Errorf("failed to make file executable: %w", err)
	}

	return nil
}

// runInstall performs the installation process
func runInstall() error {
	// Get the current executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks to get the real path
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	installDir := getInstallDir()
	return runInstallWithPaths(exePath, installDir)
}

// runInstallWithPaths performs installation with specified paths (useful for testing)
func runInstallWithPaths(exePath, installDir string) error {
	// Ensure the installation directory exists
	if err := ensureDir(installDir); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Define destination path
	dstPath := filepath.Join(installDir, "xdebug-cli")

	// Copy the binary
	if err := copyBinary(exePath, dstPath); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	return nil
}
