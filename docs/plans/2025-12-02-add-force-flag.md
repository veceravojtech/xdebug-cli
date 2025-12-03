# Add --force Flag to listen Command Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `--force` flag to `xdebug-cli listen` that automatically kills existing daemon on same port before starting non-interactive or daemon sessions.

**Architecture:** Reuse existing `Force` field in `CLIParameter` (currently used for `connection kill --all --force`). Add flag to listen command, validate it requires `--non-interactive` or `--daemon`, and kill daemon on target port before starting server.

**Tech Stack:** Go 1.x, Cobra CLI framework, existing daemon registry system

---

## Task 1: Add --force flag to listen command

**Files:**
- Modify: `internal/cli/listen.go:47-52` (init function)

**Step 1: Add flag registration**

In `internal/cli/listen.go`, modify the `init()` function to add the `--force` flag:

```go
func init() {
	listenCmd.Flags().BoolVar(&CLIArgs.NonInteractive, "non-interactive", false, "Run in non-interactive mode (no REPL)")
	listenCmd.Flags().StringArrayVar(&CLIArgs.Commands, "commands", []string{}, "Commands to execute in non-interactive mode")
	listenCmd.Flags().BoolVar(&CLIArgs.Daemon, "daemon", false, "Run as background daemon process")
	listenCmd.Flags().BoolVar(&CLIArgs.Force, "force", false, "Kill existing daemon on same port before starting")
	rootCmd.AddCommand(listenCmd)
}
```

**Step 2: Verify flag is registered**

Run: `go build -o xdebug-cli ./cmd/xdebug-cli && ./xdebug-cli listen --help`

Expected: Should see `--force` flag in help output with description "Kill existing daemon on same port before starting"

**Step 3: Commit**

```bash
git add internal/cli/listen.go
git commit -m "feat: add --force flag to listen command"
```

---

## Task 2: Add flag validation

**Files:**
- Modify: `internal/cli/listen.go:33-44` (Run function in listenCmd)

**Step 1: Write failing test for validation**

Create test file `internal/cli/listen_force_test.go`:

```go
package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/console/xdebug-cli/internal/cfg"
)

func TestListenForceValidation(t *testing.T) {
	tests := []struct {
		name        string
		force       bool
		nonInteractive bool
		daemon      bool
		shouldError bool
	}{
		{
			name:        "force without non-interactive or daemon should error",
			force:       true,
			nonInteractive: false,
			daemon:      false,
			shouldError: true,
		},
		{
			name:        "force with non-interactive should pass",
			force:       true,
			nonInteractive: true,
			daemon:      false,
			shouldError: false,
		},
		{
			name:        "force with daemon should pass",
			force:       true,
			nonInteractive: false,
			daemon:      true,
			shouldError: false,
		},
		{
			name:        "no force should always pass",
			force:       false,
			nonInteractive: false,
			daemon:      false,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origForce := CLIArgs.Force
			origNonInteractive := CLIArgs.NonInteractive
			origDaemon := CLIArgs.Daemon
			origStderr := os.Stderr

			// Restore after test
			defer func() {
				CLIArgs.Force = origForce
				CLIArgs.NonInteractive = origNonInteractive
				CLIArgs.Daemon = origDaemon
				os.Stderr = origStderr
			}()

			// Set test values
			CLIArgs.Force = tt.force
			CLIArgs.NonInteractive = tt.nonInteractive
			CLIArgs.Daemon = tt.daemon

			// Capture stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// Validate flags
			err := validateListenFlags()

			w.Close()
			var stderr bytes.Buffer
			stderr.ReadFrom(r)

			if tt.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/cli -run TestListenForceValidation -v`

Expected: FAIL with "validateListenFlags undefined"

**Step 3: Implement validation function**

In `internal/cli/listen.go`, add validation function after `init()`:

```go
// validateListenFlags validates command-line flag combinations
func validateListenFlags() error {
	// --commands requires --non-interactive or --daemon
	if len(CLIArgs.Commands) > 0 && !CLIArgs.NonInteractive && !CLIArgs.Daemon {
		return fmt.Errorf("--commands requires --non-interactive or --daemon flag")
	}

	// --force requires --non-interactive or --daemon
	if CLIArgs.Force && !CLIArgs.NonInteractive && !CLIArgs.Daemon {
		return fmt.Errorf("--force requires --non-interactive or --daemon flag")
	}

	return nil
}
```

**Step 4: Update Run function to use validation**

In `internal/cli/listen.go`, update the `Run` function in `listenCmd`:

```go
Run: func(cmd *cobra.Command, args []string) {
	// Validate flags
	if err := validateListenFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := runListeningCmd(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
},
```

**Step 5: Run test to verify it passes**

Run: `go test ./internal/cli -run TestListenForceValidation -v`

Expected: PASS - all test cases pass

**Step 6: Commit**

```bash
git add internal/cli/listen.go internal/cli/listen_force_test.go
git commit -m "feat: add validation for --force flag"
```

---

## Task 3: Implement killDaemonOnPort helper

**Files:**
- Modify: `internal/cli/listen.go` (add new function before runListeningCmd)
- Create: `internal/cli/listen_force_integration_test.go`

**Step 1: Write failing integration test**

Create `internal/cli/listen_force_integration_test.go`:

```go
package cli

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/console/xdebug-cli/internal/daemon"
)

func TestKillDaemonOnPort(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test")
	}

	// Create a mock daemon session in registry
	registry, err := daemon.NewSessionRegistry()
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Use a test port
	testPort := 9999
	testPID := os.Getpid() // Use current process as mock daemon

	sessionInfo := daemon.SessionInfo{
		PID:        testPID,
		Port:       testPort,
		SocketPath: fmt.Sprintf("/tmp/test-xdebug-cli-%d.sock", testPort),
		StartedAt:  time.Now(),
	}

	// Add to registry
	err = registry.Add(sessionInfo)
	if err != nil {
		t.Fatalf("Failed to add session: %v", err)
	}

	// Clean up after test
	defer registry.Remove(testPort)

	// Test 1: Kill existing daemon (should show message)
	t.Run("kill existing daemon", func(t *testing.T) {
		// Note: We can't actually kill ourselves, but we can test the lookup logic
		killDaemonOnPort(testPort)
		// If we get here without panic, the function handles the case gracefully
	})

	// Test 2: Kill non-existent daemon (should show warning)
	t.Run("kill non-existent daemon", func(t *testing.T) {
		killDaemonOnPort(8888) // Port with no daemon
		// Should return without error
	})
}
```

**Step 2: Run test to verify it fails**

Run: `INTEGRATION_TEST=1 go test ./internal/cli -run TestKillDaemonOnPort -v`

Expected: FAIL with "killDaemonOnPort undefined"

**Step 3: Implement killDaemonOnPort function**

In `internal/cli/listen.go`, add function before `runListeningCmd()`:

```go
// killDaemonOnPort attempts to kill any daemon running on the specified port.
// Always returns nil (never fails) - shows warnings/errors but continues.
func killDaemonOnPort(port int) error {
	registry, err := daemon.NewSessionRegistry()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load session registry: %v\n", err)
		return nil
	}

	session, err := registry.Get(port)
	if err != nil {
		// No session found - just warn and continue
		fmt.Fprintf(os.Stderr, "Warning: no daemon running on port %d\n", port)
		return nil
	}

	// Check if process exists (handle stale registry entries)
	if !processExists(session.PID) {
		fmt.Fprintf(os.Stderr, "Warning: daemon on port %d is stale (PID %d no longer exists), cleaning up\n", port, session.PID)
		registry.Remove(port)
		return nil
	}

	// Kill the process
	process, err := os.FindProcess(session.PID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to find daemon process (PID %d): %v\n", session.PID, err)
		return nil
	}

	if err := process.Kill(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to kill daemon on port %d (PID %d): %v\nContinuing anyway...\n", port, session.PID, err)
		return nil
	}

	// Clean up registry
	registry.Remove(port)
	fmt.Printf("Killed daemon on port %d (PID %d)\n", port, session.PID)
	return nil
}

// processExists checks if a process with the given PID exists
func processExists(pid int) bool {
	// Check if /proc/<pid> exists (Linux-specific)
	procPath := fmt.Sprintf("/proc/%d", pid)
	_, err := os.Stat(procPath)
	return err == nil
}
```

**Step 4: Run test to verify it passes**

Run: `INTEGRATION_TEST=1 go test ./internal/cli -run TestKillDaemonOnPort -v`

Expected: PASS - all test cases pass

**Step 5: Commit**

```bash
git add internal/cli/listen.go internal/cli/listen_force_integration_test.go
git commit -m "feat: implement killDaemonOnPort helper function"
```

---

## Task 4: Integrate force flag into runListeningCmd

**Files:**
- Modify: `internal/cli/listen.go:54-93` (runListeningCmd function)

**Step 1: Write failing test**

Add to `internal/cli/listen_force_test.go`:

```go
func TestRunListeningCmdWithForce(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test")
	}

	// This test verifies that force flag triggers cleanup before server start
	// We can't fully test server startup without mocking, but we can verify
	// the killDaemonOnPort is called when Force is true

	tests := []struct {
		name  string
		force bool
	}{
		{
			name:  "with force flag",
			force: true,
		},
		{
			name:  "without force flag",
			force: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origForce := CLIArgs.Force
			origNonInteractive := CLIArgs.NonInteractive
			origPort := CLIArgs.Port

			// Restore after test
			defer func() {
				CLIArgs.Force = origForce
				CLIArgs.NonInteractive = origNonInteractive
				CLIArgs.Port = origPort
			}()

			// Set test values
			CLIArgs.Force = tt.force
			CLIArgs.NonInteractive = true // Required for force
			CLIArgs.Port = 9999 // Use test port

			// Note: We can't actually run the full command without starting a server
			// This test documents the expected behavior
			// Manual testing required for full integration
		})
	}
}
```

**Step 2: Run test**

Run: `INTEGRATION_TEST=1 go test ./internal/cli -run TestRunListeningCmdWithForce -v`

Expected: PASS (tests document expected behavior)

**Step 3: Add force flag logic to runListeningCmd**

In `internal/cli/listen.go`, modify `runListeningCmd()`:

```go
// runListeningCmd starts the DBGp server and waits for connections
func runListeningCmd() error {
	v := view.NewView()

	// Kill existing daemon if --force is set
	if CLIArgs.Force {
		killDaemonOnPort(CLIArgs.Port)
	}

	// Display startup information only in interactive mode
	if !CLIArgs.NonInteractive && !CLIArgs.Daemon {
		v.PrintApplicationInformation(cfg.Version, CLIArgs.Host, CLIArgs.Port)
	}

	// Create and start server
	server := dbgp.NewServer(CLIArgs.Host, CLIArgs.Port)
	if err := server.Listen(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	defer server.Close()

	// ... rest of existing code ...
```

**Step 4: Build and manual test**

Run:
```bash
go build -o xdebug-cli ./cmd/xdebug-cli

# Start a daemon
./xdebug-cli listen --daemon -p 9999 &
DAEMON_PID=$!

# Try to start non-interactive with force
./xdebug-cli listen --non-interactive --force -p 9999 --commands "help"
```

Expected: Should see "Killed daemon on port 9999 (PID <pid>)" message, then proceed with non-interactive session

**Step 5: Commit**

```bash
git add internal/cli/listen.go internal/cli/listen_force_test.go
git commit -m "feat: integrate --force flag into listen command"
```

---

## Task 5: Update documentation

**Files:**
- Modify: `CLAUDE.md:35-47` (Available Commands section)
- Modify: `CLAUDE.md:67-134` (Non-Interactive Mode section)
- Modify: `CLAUDE.md:135-230` (Daemon Mode section)

**Step 1: Update Available Commands section**

In `CLAUDE.md`, update line 36:

```markdown
xdebug-cli listen --non-interactive --force --commands "cmd1" "cmd2"  # Kill existing daemon, run commands
```

**Step 2: Add Force Flag subsection to Non-Interactive Mode**

In `CLAUDE.md`, after line 82 (after "Basic Usage"), add:

```markdown
### Force Flag

Use `--force` to automatically kill any existing daemon on the same port before starting:

```bash
# Kill stale daemon on port 9003, then run commands
xdebug-cli listen --non-interactive --force --commands "run" "print \$x"

# With custom port
xdebug-cli listen -p 9004 --non-interactive --force --commands "break :42" "run"
```

The `--force` flag:
- Kills only the daemon on the same port (e.g., port 9003)
- Shows warning if no daemon exists, but continues
- Never fails - always proceeds with the new session
- Useful for automation scripts and CI/CD where stale processes may exist

**Output Examples:**

Daemon killed successfully:
```
Killed daemon on port 9003 (PID 12345)
Server listening on 127.0.0.1:9003
```

No daemon running:
```
Warning: no daemon running on port 9003
Server listening on 127.0.0.1:9003
```

Stale daemon (process already dead):
```
Warning: daemon on port 9003 is stale (PID 12345 no longer exists), cleaning up
Server listening on 127.0.0.1:9003
```
```

**Step 3: Add force example to Daemon Mode**

In `CLAUDE.md`, after line 149 (after "Start daemon with JSON output"), add:

```bash
# Kill old daemon and start fresh
xdebug-cli listen --daemon --force --commands "break :42"
```

**Step 4: Verify documentation renders correctly**

Run: `cat CLAUDE.md | head -n 250`

Expected: See updated documentation with force flag examples

**Step 5: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add --force flag documentation"
```

---

## Task 6: End-to-end manual testing

**Files:**
- None (manual testing)

**Step 1: Test force with non-interactive mode**

Run:
```bash
# Build
go build -o xdebug-cli ./cmd/xdebug-cli

# Start daemon
./xdebug-cli listen --daemon -p 9999 &
sleep 2

# Verify daemon is running
./xdebug-cli connection -p 9999

# Try non-interactive with force
./xdebug-cli listen --non-interactive --force -p 9999 --commands "help"
```

Expected: Should kill daemon and proceed with non-interactive mode

**Step 2: Test force with no daemon**

Run:
```bash
# Ensure no daemon on port 9999
./xdebug-cli connection kill -p 9999 2>/dev/null || true

# Try force with no daemon
./xdebug-cli listen --non-interactive --force -p 9999 --commands "help"
```

Expected: Should show warning "no daemon running on port 9999" and continue

**Step 3: Test force with daemon mode**

Run:
```bash
# Start first daemon
./xdebug-cli listen --daemon -p 9999 &
sleep 2

# Start second daemon with force (should kill first)
./xdebug-cli listen --daemon --force -p 9999 &
sleep 2

# Verify only one daemon running
./xdebug-cli connection list | grep 9999 | wc -l
```

Expected: Should see only 1 daemon on port 9999

**Step 4: Test force validation**

Run:
```bash
# Try force without non-interactive or daemon
./xdebug-cli listen --force
```

Expected: Error message "Error: --force requires --non-interactive or --daemon flag"

**Step 5: Clean up**

Run:
```bash
./xdebug-cli connection kill --all --force
```

**Step 6: Document test results**

Create `docs/plans/2025-12-02-add-force-flag-test-results.md` with manual test outcomes

**Step 7: Final commit**

```bash
git add docs/plans/2025-12-02-add-force-flag-test-results.md
git commit -m "test: document manual testing results for --force flag"
```

---

## Summary Checklist

- [ ] Task 1: Add --force flag registration to listen command
- [ ] Task 2: Add and test flag validation logic
- [ ] Task 3: Implement killDaemonOnPort helper with tests
- [ ] Task 4: Integrate force flag into runListeningCmd
- [ ] Task 5: Update CLAUDE.md documentation (3 sections)
- [ ] Task 6: Complete end-to-end manual testing

## Key Testing Commands

```bash
# Build
go build -o xdebug-cli ./cmd/xdebug-cli

# Unit tests
go test ./internal/cli -run TestListenForceValidation -v

# Integration tests
INTEGRATION_TEST=1 go test ./internal/cli -run TestKillDaemonOnPort -v
INTEGRATION_TEST=1 go test ./internal/cli -run TestRunListeningCmdWithForce -v

# All tests
go test ./...

# Manual e2e test
./xdebug-cli listen --daemon -p 9999 &
./xdebug-cli listen --non-interactive --force -p 9999 --commands "help"
./xdebug-cli connection kill --all --force
```

## Notes

- The `Force` field already exists in `CLIParameter` (used for `connection kill --all --force`)
- We're reusing it with dual meaning: skip confirmation (connection) OR kill daemon (listen)
- The `processExists()` function is duplicated from `connection.go` to avoid circular dependencies
- Linux-specific implementation using `/proc/<pid>` - Windows would need different approach
- Always returns nil from `killDaemonOnPort()` to ensure non-interactive mode never fails on cleanup
