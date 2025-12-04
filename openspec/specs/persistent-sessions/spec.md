# persistent-sessions Specification

## Purpose
TBD - created by archiving change enable-persistent-debug-sessions. Update Purpose after archive.
## Requirements
### Requirement: Daemon Mode
The CLI SHALL run all debug sessions as background daemons. Daemon mode is now the primary and only execution model.

#### Scenario: Start daemon without commands
- **WHEN** user runs `xdebug-cli daemon start`
- **THEN** process forks to background
- **AND** parent process exits successfully
- **AND** daemon writes session entry to registry with PID, port, socket path, and start time
- **AND** daemon starts DBGp server and waits for connections
- **AND** displays message: "Daemon started on port 9003"

#### Scenario: Start daemon with initial commands
- **WHEN** user runs `xdebug-cli daemon start --commands "break :42" "run"`
- **THEN** daemon starts in background
- **AND** waits for Xdebug connection
- **AND** executes commands when connection established
- **AND** keeps session alive after commands complete

#### Scenario: Daemon already running without force flag
- **WHEN** user runs `xdebug-cli daemon start -p 9003`
- **AND** daemon already running on port 9003
- **THEN** command exits with error code 1
- **AND** displays existing daemon PID and socket path
- **AND** suggests using `--force` to replace or `connection kill` to terminate

#### Scenario: Daemon already running with force flag
- **WHEN** user runs `xdebug-cli daemon start -p 9003 --force`
- **AND** daemon already running on port 9003
- **THEN** kills existing daemon on port 9003
- **AND** starts new daemon successfully
- **AND** displays message: "Killed daemon on port 9003 (PID 12345)" followed by "Daemon started on port 9003"

#### Scenario: Daemon cleanup on exit
- **WHEN** debug session ends or is killed
- **THEN** daemon removes session entry from registry
- **AND** daemon removes Unix socket
- **AND** daemon exits gracefully

### Requirement: IPC Communication
The CLI SHALL provide Unix socket-based IPC for communicating with daemon sessions with automatic retry.

#### Scenario: Create IPC socket
- **WHEN** daemon starts
- **THEN** creates Unix socket at `/tmp/xdebug-cli-session-<pid>.sock`
- **AND** socket has permissions 0600 (owner only)
- **AND** socket path is written to session registry

#### Scenario: Accept IPC connections
- **WHEN** client connects to Unix socket
- **THEN** daemon accepts connection
- **AND** reads JSON command request
- **AND** executes commands via existing dispatcher
- **AND** returns JSON response with results

#### Scenario: Handle concurrent connections
- **WHEN** multiple clients connect simultaneously
- **THEN** daemon processes requests sequentially
- **AND** maintains thread-safe session access
- **AND** returns results to correct client

#### Scenario: IPC protocol structure
- **WHEN** client sends command request
- **THEN** request contains command list and output format flag
- **AND** response contains success status and result array
- **AND** each result includes command name, success flag, and data

#### Scenario: Client retries on transient connection failure
- **WHEN** initial connection attempt fails
- **THEN** client retries with exponential backoff (100ms, 200ms, 400ms)
- **AND** makes up to 3 attempts by default
- **AND** returns error only after all attempts exhausted

#### Scenario: Client connects immediately when daemon ready
- **WHEN** daemon socket is accepting connections
- **AND** client attempts to connect
- **THEN** connection succeeds on first attempt
- **AND** no unnecessary delay is introduced

### Requirement: Attach Command
The CLI SHALL provide an attach command to execute commands against running daemon sessions.

#### Scenario: Attach to running session
- **WHEN** user runs `xdebug-cli attach --commands "context local"`
- **AND** daemon is running
- **THEN** connects to daemon Unix socket
- **AND** sends command batch
- **AND** displays results
- **AND** exits while daemon continues running

#### Scenario: Attach with JSON output
- **WHEN** user runs `xdebug-cli attach --json --commands "print \$x"`
- **THEN** sends JSON output flag in IPC request
- **AND** displays JSON-formatted results

#### Scenario: Attach when no session
- **WHEN** user runs `xdebug-cli attach --commands "run"`
- **AND** no daemon is running
- **THEN** exits with error
- **AND** displays message to start daemon first

#### Scenario: Attach command fails
- **WHEN** user runs `xdebug-cli attach --commands "invalid"`
- **THEN** displays error for that command
- **AND** exits with non-zero status

### Requirement: Session Registry
The CLI SHALL maintain a registry of active daemon sessions with crash recovery.

#### Scenario: Register new session
- **WHEN** daemon starts
- **THEN** creates registry file `~/.xdebug-cli/sessions.json` if missing
- **AND** adds entry with PID, port, socket path, start time

#### Scenario: Query active sessions
- **WHEN** user runs `xdebug-cli connection`
- **AND** daemon is running
- **THEN** reads registry to find session
- **AND** displays session information

#### Scenario: Cleanup stale entries
- **WHEN** daemon starts
- **OR** connection command queries sessions
- **THEN** validates each registry entry process exists
- **AND** validates process is xdebug-cli (not recycled PID)
- **AND** removes entries for non-existent processes
- **AND** removes entries for recycled PIDs running different executables
- **AND** removes stale PID and socket files

#### Scenario: Find session by port
- **WHEN** attach command needs to connect
- **THEN** reads registry to find session on specified port
- **AND** uses socket path from registry entry

#### Scenario: Validate process identity
- **WHEN** registry contains entry with PID 12345
- **AND** process 12345 exists but is not xdebug-cli
- **THEN** entry is considered stale
- **AND** entry is removed during cleanup
- **AND** associated socket and PID files are removed

### Requirement: Enhanced Connection Commands
The CLI connection commands SHALL support daemon session management.

#### Scenario: Kill daemon session
- **WHEN** user runs `xdebug-cli connection kill`
- **AND** daemon is running
- **THEN** connects to Unix socket
- **AND** sends kill request
- **AND** daemon sends DBGp stop command
- **AND** daemon closes connection and exits
- **AND** removes session from registry

#### Scenario: Check daemon status
- **WHEN** user runs `xdebug-cli connection isAlive`
- **AND** daemon is running
- **THEN** checks session registry and process is alive
- **AND** exits with code 0

#### Scenario: Show daemon details
- **WHEN** user runs `xdebug-cli connection`
- **AND** daemon is running
- **THEN** displays daemon PID, port, socket path
- **AND** displays session state and current location
- **AND** indicates session is running in daemon mode

#### Scenario: List all daemon sessions
- **WHEN** user runs `xdebug-cli connection list`
- **THEN** reads session registry
- **AND** displays all active sessions with PID, port, socket path, and start time
- **AND** validates process existence and cleans up stale entries
- **AND** supports `--json` flag for machine-readable output

#### Scenario: Kill all daemon sessions
- **WHEN** user runs `xdebug-cli connection kill --all`
- **THEN** prompts user for confirmation
- **AND** iterates through all active sessions and terminates them
- **AND** displays count of sessions terminated
- **AND** supports `--force` flag to skip confirmation prompt

### Requirement: Breakpoint Path History
The CLI SHALL persist successfully used breakpoint paths across daemon sessions to provide suggestions when non-absolute paths fail.

#### Scenario: Store path on successful breakpoint hit
- **WHEN** breakpoint at `/var/www/app/File.php:100` is hit successfully
- **THEN** path is stored in persistent history file `~/.xdebug-cli/breakpoint-paths.json`
- **AND** mapping is: filename `File.php` -> full path `/var/www/app/File.php`

#### Scenario: Lookup path by filename
- **WHEN** path lookup is requested for filename `File.php`
- **AND** history has stored path for `File.php`
- **THEN** returns the stored full path `/var/www/app/File.php`

#### Scenario: No stored path for filename
- **WHEN** path lookup is requested for filename `Unknown.php`
- **AND** no stored path exists for that filename
- **THEN** returns empty string

#### Scenario: Path history persists across sessions
- **WHEN** daemon session ends
- **AND** new daemon session starts later
- **THEN** previous breakpoint paths are still available for suggestions

#### Scenario: Multiple paths for same filename
- **WHEN** different full paths exist for same filename (e.g., `/app1/File.php` and `/app2/File.php`)
- **THEN** most recently used path is stored and suggested

