# cli Specification

## Purpose
TBD - created by archiving change implement-xdebug-cli. Update Purpose after archive.
## Requirements
### Requirement: Global CLI Flags
The CLI SHALL provide global flags for connection settings.

#### Scenario: Host flag default
- **WHEN** user runs `xdebug-cli listen --daemon` without specifying `--host`
- **THEN** server binds to `0.0.0.0` (all interfaces) by default

#### Scenario: Host flag explicit
- **WHEN** user runs with `--host 127.0.0.1` or `-l 127.0.0.1`
- **THEN** server binds to specified address (localhost only)

#### Scenario: Port flag
- **WHEN** user runs with `--port 9003` or `-p 9003`
- **THEN** server listens on specified port

### Requirement: Project Structure
The CLI SHALL be a buildable Go project following the standard template structure.

#### Scenario: Project builds successfully
- **WHEN** running `go build ./...` in the project root
- **THEN** the project compiles without errors

#### Scenario: Project structure follows conventions
- **WHEN** examining the project layout
- **THEN** it contains `cmd/source-cli/main.go` as entry point
- **AND** `internal/cli/` for command implementations
- **AND** `go.mod` with module definition

### Requirement: Preview Command
The CLI SHALL provide a `preview` command that displays an animated progress indicator.

#### Scenario: Preview with duration
- **WHEN** user runs `source-cli preview source 10s`
- **THEN** an animated loading indicator is displayed for 10 seconds
- **AND** the animation runs until the duration expires

#### Scenario: Preview without arguments
- **WHEN** user runs `source-cli preview`
- **THEN** help text is displayed explaining the command usage

### Requirement: Install Command
The CLI SHALL provide an `install` command that installs the binary to the user's local bin directory.

#### Scenario: Install to local bin
- **WHEN** user runs `source-cli install`
- **THEN** the CLI binary is built and copied to `~/.local/bin/source-cli`
- **AND** the binary includes a build timestamp

#### Scenario: Install creates directory if missing
- **WHEN** user runs `source-cli install` and `~/.local/bin` does not exist
- **THEN** the directory is created
- **AND** the binary is installed successfully

### Requirement: TDD Best Practices
The CLI SHALL follow test-driven development best practices for Go.

#### Scenario: Test files exist alongside source
- **WHEN** examining any `.go` source file in `internal/`
- **THEN** a corresponding `_test.go` file exists with unit tests

#### Scenario: Tests pass
- **WHEN** running `go test ./...`
- **THEN** all tests pass successfully

### Requirement: JSON Output Mode
The CLI SHALL provide `--json` global flag for machine-readable output.

#### Scenario: JSON output for variable inspection
- **WHEN** user runs with `--json` and executes "print $myArray"
- **THEN** outputs JSON with structure: `{"command": "print", "variable": "$myArray", "type": "array", "value": [...], "success": true}`

#### Scenario: JSON output for breakpoint status
- **WHEN** user runs with `--json` and executes "run"
- **THEN** outputs JSON with structure: `{"command": "run", "status": "break", "file": "app.php", "line": 42, "success": true}`

#### Scenario: JSON output for errors
- **WHEN** command fails with `--json` enabled
- **THEN** outputs JSON with structure: `{"command": "...", "success": false, "error": "error message"}`

### Requirement: Attach Command
The CLI SHALL provide an attach command to interact with daemon sessions.

#### Scenario: Execute commands on daemon
- **WHEN** user runs `xdebug-cli attach --commands "context local"`
- **AND** daemon is running with active session
- **THEN** executes commands against daemon session
- **AND** displays results
- **AND** exits while daemon continues

#### Scenario: Attach with no daemon
- **WHEN** user runs `xdebug-cli attach --commands "run"`
- **AND** no daemon is running
- **THEN** exits with error code 1
- **AND** displays error message: "Error: no daemon running on port 9003. Start with: xdebug-cli daemon start"

#### Scenario: Attach with JSON output
- **WHEN** user runs `xdebug-cli attach --json --commands "print \$x"`
- **THEN** requests JSON output from daemon
- **AND** displays JSON-formatted results

### Requirement: Daemon Subcommands
The CLI SHALL provide daemon subcommands for lifecycle and session management.

#### Scenario: Show daemon status
- **WHEN** user runs `xdebug-cli daemon status`
- **AND** daemon is running on the current port
- **THEN** displays "Connection Status: Daemon Mode"
- **AND** displays daemon PID, port, and socket path
- **AND** displays start timestamp
- **AND** shows help text for killing the daemon

#### Scenario: Show status when no daemon
- **WHEN** user runs `xdebug-cli daemon status`
- **AND** no daemon is running on the current port
- **THEN** displays "Connection Status: Not connected"
- **AND** shows help text for starting a daemon

#### Scenario: List all daemon sessions
- **WHEN** user runs `xdebug-cli daemon list`
- **THEN** displays table with columns: PID, Port, Started, Socket Path
- **AND** shows count of active sessions
- **AND** only includes sessions with running processes

#### Scenario: List daemon sessions in JSON
- **WHEN** user runs `xdebug-cli daemon list --json`
- **THEN** outputs JSON array of session objects
- **AND** each object contains: pid, port, socket_path, started_at

#### Scenario: List when no daemons
- **WHEN** user runs `xdebug-cli daemon list`
- **AND** no daemon sessions exist
- **THEN** displays "No active daemon sessions found."
- **AND** shows help text for starting a daemon

#### Scenario: Kill daemon on current port
- **WHEN** user runs `xdebug-cli daemon kill`
- **AND** daemon is running on current port
- **THEN** sends kill request via IPC socket
- **AND** displays success message
- **AND** daemon process terminates and cleans up

#### Scenario: Kill when no daemon
- **WHEN** user runs `xdebug-cli daemon kill`
- **AND** no daemon is running on current port
- **THEN** exits with code 1
- **AND** displays error "No active session to kill."
- **AND** shows help text for checking status

#### Scenario: Kill all daemon sessions with confirmation
- **WHEN** user runs `xdebug-cli daemon kill --all`
- **AND** multiple daemon sessions exist
- **THEN** prompts "Found N active session(s). Terminate all? (y/N):"
- **AND** waits for user input
- **AND** kills all sessions if user confirms with "y" or "yes"
- **AND** cancels operation if user enters anything else

#### Scenario: Kill all daemon sessions without confirmation
- **WHEN** user runs `xdebug-cli daemon kill --all --force`
- **AND** daemon sessions exist
- **THEN** kills all sessions without prompting
- **AND** displays progress for each session
- **AND** shows summary of successful/failed terminations

#### Scenario: Check if daemon is alive
- **WHEN** user runs `xdebug-cli daemon isAlive`
- **AND** daemon is running on current port
- **AND** process exists
- **THEN** prints "connected"
- **AND** exits with code 0

#### Scenario: Check when daemon not alive
- **WHEN** user runs `xdebug-cli daemon isAlive`
- **AND** no daemon is running on current port
- **THEN** prints "not connected"
- **AND** exits with code 1

#### Scenario: Kill detects stale processes
- **WHEN** user runs `xdebug-cli daemon kill`
- **AND** registry entry exists but process is dead
- **THEN** finds process using lsof (if available)
- **AND** verifies it's xdebug-cli by checking /proc/<pid>/comm
- **AND** kills the stale process
- **AND** displays "Stale process terminated successfully."

### Requirement: Daemon Start Command
The CLI SHALL provide a `daemon start` command as the primary entry point for all debugging sessions.

#### Scenario: Start daemon without commands
- **WHEN** user runs `xdebug-cli daemon start --curl "http://localhost/app.php"`
- **THEN** process forks to background
- **AND** parent process exits successfully
- **AND** daemon starts DBGp server and waits for connections
- **AND** displays message: "Daemon started on port 9003"

#### Scenario: Start daemon with initial commands
- **WHEN** user runs `xdebug-cli daemon start --curl "http://localhost/app.php" --commands "break /path/file.php:100"`
- **THEN** daemon starts in background
- **AND** waits for Xdebug connection
- **AND** executes commands when connection established
- **AND** keeps session alive after commands complete

#### Scenario: Start daemon with force flag
- **WHEN** user runs `xdebug-cli daemon start --curl "http://localhost/app.php" --force`
- **AND** daemon already running on same port
- **THEN** kills existing daemon on that port
- **AND** starts new daemon successfully
- **AND** displays message: "Killed daemon on port 9003 (PID 12345)" followed by "Daemon started on port 9003"

#### Scenario: Daemon already running without force
- **WHEN** user runs `xdebug-cli daemon start --curl "http://localhost/app.php"`
- **AND** daemon already running on port 9003
- **THEN** command exits with error code 1
- **AND** displays message: "Error: daemon already running on port 9003 (PID 12345). Use 'xdebug-cli connection kill' to terminate it first or use --force to replace it."

#### Scenario: Start daemon with external connection flag
- **WHEN** user runs `xdebug-cli daemon start --enable-external-connection`
- **THEN** process forks to background
- **AND** parent process exits successfully
- **AND** daemon starts DBGp server on specified port
- **AND** waits indefinitely for external Xdebug connection
- **AND** displays message: "Daemon started on port 9003 (waiting for external connection)"

#### Scenario: External connection with initial commands
- **WHEN** user runs `xdebug-cli daemon start --enable-external-connection --commands "break /path/file.php:100"`
- **THEN** daemon starts in background
- **AND** waits for external Xdebug connection
- **AND** executes commands when connection established
- **AND** keeps session alive after commands complete

#### Scenario: Start daemon without required flags
- **WHEN** user runs `xdebug-cli daemon start` without `--curl` or `--enable-external-connection`
- **THEN** command exits with error code 1
- **AND** displays error message explaining that either `--curl` or `--enable-external-connection` is required
- **AND** shows usage examples for both options

### Requirement: Command Aliases for Multiple Debugger Conventions
The CLI SHALL support command aliases from multiple debugger conventions (GDB, DBGp protocol, VS Code) to improve usability for users from different debugging backgrounds.

#### Scenario: GDB-style continue command
- **WHEN** user executes `xdebug-cli attach --commands "continue"`
- **THEN** execution continues to next breakpoint (same as `run`)
- **AND** returns current execution state

#### Scenario: Short continue alias
- **WHEN** user executes `xdebug-cli attach --commands "cont"` or `--commands "c"`
- **THEN** execution continues (same as `continue` and `run`)

#### Scenario: Step into with alternative names
- **WHEN** user executes `--commands "into"` or `--commands "step_into"`
- **THEN** steps into next function call (same as `step`)

#### Scenario: Step over with alternative name
- **WHEN** user executes `--commands "over"`
- **THEN** steps over next line without entering functions (same as `next`)

#### Scenario: Step out with alternative name
- **WHEN** user executes `--commands "step_out"`
- **THEN** steps out of current function (same as `out`)

#### Scenario: DBGp protocol breakpoint list
- **WHEN** user executes `--commands "breakpoint_list"`
- **THEN** displays all active breakpoints (same as `info breakpoints`)

#### Scenario: DBGp protocol breakpoint remove
- **WHEN** user executes `--commands "breakpoint_remove 1"`
- **THEN** removes breakpoint with ID 1 (same as `delete 1`)

#### Scenario: DBGp protocol property get
- **WHEN** user executes `--commands "property_get -n myVar"`
- **THEN** displays variable value (same as `print myVar`)
- **AND** supports both `$myVar` and `myVar` syntax

#### Scenario: Property get without flag error
- **WHEN** user executes `--commands "property_get myVar"` without `-n` flag
- **THEN** returns error: "Usage: property_get -n <variable>"

#### Scenario: GDB-style clear by line
- **WHEN** user executes `--commands "clear :42"`
- **AND** breakpoint exists at line 42 in current file
- **THEN** removes the breakpoint
- **AND** returns success with message "Removed 1 breakpoint(s) at :42"

#### Scenario: GDB-style clear by file and line
- **WHEN** user executes `--commands "clear app.php:100"`
- **AND** breakpoint exists at app.php:100
- **THEN** removes the breakpoint
- **AND** returns success message

#### Scenario: Clear with no breakpoint at location
- **WHEN** user executes `--commands "clear :50"`
- **AND** no breakpoint exists at line 50
- **THEN** returns error: "No breakpoint at location :50"

#### Scenario: Clear removes multiple breakpoints at same location
- **WHEN** multiple breakpoints exist at same file:line
- **AND** user executes `--commands "clear file.php:42"`
- **THEN** removes all breakpoints at that location
- **AND** returns count of removed breakpoints

#### Scenario: Aliases work with JSON output mode
- **WHEN** user executes `--json --commands "continue"`
- **THEN** returns JSON output with same structure as `run` command
- **AND** command field in JSON shows "continue"

#### Scenario: Help text shows aliases
- **WHEN** user executes `--commands "help"`
- **THEN** displays commands with their aliases
- **AND** shows "run, r, continue, c" on same line
- **AND** shows "step, s, into, step_into" on same line
- **AND** shows other command groups with their aliases

### Requirement: Breakpoint Path Validation with Fail-Fast
The CLI SHALL validate breakpoints with non-absolute paths by waiting for the first hit, and terminate the daemon if the breakpoint is not triggered within a timeout period.

#### Scenario: Absolute path breakpoint - normal operation
- **WHEN** user runs `daemon start --curl "..." --commands "break /var/www/app/File.php:100"`
- **THEN** daemon starts and sets breakpoint
- **AND** curl triggers request
- **AND** when breakpoint hits, daemon continues normally
- **AND** the full path is stored for future suggestions

#### Scenario: Non-absolute path breakpoint hits - success
- **WHEN** user runs `daemon start --curl "..." --commands "break File.php:100"`
- **AND** breakpoint path resolves correctly and is hit within timeout
- **THEN** warning is shown: "Warning: Breakpoint path 'File.php' is not absolute"
- **AND** daemon continues normally after breakpoint hit
- **AND** resolved full path is stored for future suggestions

#### Scenario: Non-absolute path breakpoint not hit - fail fast
- **WHEN** user runs `daemon start --curl "..." --commands "break File.php:100"`
- **AND** breakpoint is not hit within timeout (default 10s)
- **THEN** daemon is terminated
- **AND** error is displayed: "Error: Breakpoint at 'File.php:100' was not hit within 10s"
- **AND** if known path exists, shows: "Use full path: /var/www/app/File.php:100"
- **AND** exit code is non-zero

#### Scenario: Non-absolute path with known suggestion
- **WHEN** user previously used path `/var/www/app/controllers/PriceLoader.php`
- **AND** user runs `daemon start --curl "..." --commands "break PriceLoader.php:369"`
- **AND** breakpoint not hit within timeout
- **THEN** error includes suggestion: "Use full path: /var/www/app/controllers/PriceLoader.php:369"

#### Scenario: Non-absolute path with no known suggestion
- **WHEN** no previous path matches filename `NewFile.php`
- **AND** user runs `daemon start --curl "..." --commands "break NewFile.php:50"`
- **AND** breakpoint not hit within timeout
- **THEN** error shows: "Error: Breakpoint at 'NewFile.php:50' was not hit. Ensure you use an absolute path."

#### Scenario: Current file line syntax - no validation needed
- **WHEN** user runs with `--commands "break :100"`
- **THEN** breakpoint uses current file from session (absolute path)
- **AND** no timeout validation is applied

#### Scenario: Custom timeout flag
- **WHEN** user runs `daemon start --breakpoint-timeout 30s --curl "..." --commands "break File.php:100"`
- **THEN** daemon waits up to 30 seconds for breakpoint hit before failing

### Requirement: Timeout Exit Feedback
The CLI SHALL provide clear feedback when breakpoint validation times out.

#### Scenario: Warning message on timeout
- **WHEN** breakpoint validation timeout occurs
- **THEN** prints warning to stderr: "Warning: breakpoint not hit within Xs"
- **AND** includes list of pending breakpoints that were not hit
- **AND** suggests increasing timeout with `--breakpoint-timeout` or `--wait-forever`

#### Scenario: Distinct exit code for timeout
- **WHEN** breakpoint validation timeout occurs
- **THEN** exits with code 124 (Unix timeout convention)
- **AND** exit code is distinct from general errors (code 1)

#### Scenario: Log file for timeout events
- **WHEN** breakpoint validation timeout occurs
- **THEN** writes event to `/tmp/xdebug-cli-daemon-<port>.log`
- **AND** includes timestamp, timeout duration, and breakpoint details
- **AND** log file aids post-mortem debugging

#### Scenario: Normal exit when breakpoint hit
- **WHEN** breakpoint is hit within timeout period
- **THEN** no timeout warning is printed
- **AND** daemon continues running normally

