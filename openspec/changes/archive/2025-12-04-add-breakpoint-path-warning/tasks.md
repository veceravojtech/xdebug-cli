## 1. Implementation

- [x] 1.1 Add persistent breakpoint path storage in `internal/daemon/paths.go`
  - Create file `~/.xdebug-cli/breakpoint-paths.json`
  - Structure: `map[string]string` (filename -> full path)
  - Function: `SaveBreakpointPath(fullPath string) error`
  - Function: `LoadBreakpointPath(filename string) string`
  - Handle file creation, read/write errors gracefully

- [x] 1.2 Add path validation helper functions in `internal/daemon/paths.go`
  - Function: `IsAbsolutePath(path string) bool` - returns true if starts with `/`
  - Function: `extractFilename(path string) string` - returns filename from path
  - Function: `HasNonAbsoluteBreakpoint(commands []string) (bool, string)` - scan commands for non-absolute break

- [x] 1.3 Add breakpoint validation mode to daemon start
  - Add `--breakpoint-timeout` flag to `internal/cli/daemon.go` (default 10s)
  - In `internal/cli/daemon.go`, after setting breakpoint:
    - If non-absolute path detected, start timeout timer
    - After curl triggers and `run` executes, check if status is "break"
    - If status is "break" within timeout: save path, continue normally
    - If timeout expires: terminate daemon, return error with suggestion
  - Look up suggestion from persistent path history

- [x] 1.4 Show warning for non-absolute paths before forking
  - Warning displayed in parent process (visible to user)
  - Includes suggested full path from history if available
  - Mentions timeout duration for breakpoint hit

- [x] 1.5 Add breakpoint hit detection
  - After `run` command, check if status is "break"
  - If status is "break", save path to history for future suggestions
  - If status is "stopping"/"stopped", fail with error message

- [x] 1.6 Update error messages with path suggestions
  - Error format: "Breakpoint at 'File.php:100' was not hit. Use full path: /var/www/app/File.php:100"
  - If no suggestion available: "Breakpoint at 'File.php:100' was not hit. Ensure you use an absolute path (starting with /)."

## 2. Testing

- [x] 2.1 Add unit tests for path validation helpers
  - Test absolute paths: `/var/www/file.php` -> is absolute
  - Test filename-only: `PriceLoader.php` -> not absolute
  - Test relative paths: `app/models/User.php` -> not absolute

- [x] 2.2 Add unit tests for persistent path storage
  - Test saving and loading paths from JSON file
  - Test filename extraction and matching
  - Test handling of missing/corrupted file
  - Test persistence across instances

- [x] 2.3 Add integration test for breakpoint path warning
  - Non-absolute breakpoint shows warning message
  - Warning includes Xdebug absolute path requirement
  - Warning mentions timeout duration
  - Absolute paths do not show warning
  - :line format, break call, break exception do not show warning
  - --breakpoint-timeout=0 disables timeout message

- [x] 2.4 Verify all tests pass
  - `go test ./...` passes
  - Integration tests pass
