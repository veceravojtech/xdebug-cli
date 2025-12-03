# persistent-sessions Specification

## Purpose
TBD - created by archiving change enable-persistent-debug-sessions. Update Purpose after archive.
## Requirements
### Requirement: Daemon Mode
The CLI SHALL support running the debug server as a background daemon that persists after the CLI exits.

#### Scenario: Start daemon without commands
- **WHEN** user runs `xdebug-cli listen --daemon`
- **THEN** process forks to background
- **AND** parent process exits successfully
- **AND** daemon writes PID to `/tmp/xdebug-cli-daemon-<port>.pid`
- **AND** daemon starts DBGp server and waits for connections

#### Scenario: Start daemon with initial commands
- **WHEN** user runs `xdebug-cli listen --daemon --commands "break :42" "run"`
- **THEN** daemon starts in background
- **AND** waits for Xdebug connection
- **AND** executes commands when connection established
- **AND** keeps session alive after commands complete

#### Scenario: Daemon already running
- **WHEN** user runs `xdebug-cli listen --daemon -p 9003`
- **AND** daemon already running on port 9003
- **THEN** command exits with error
- **AND** displays existing daemon PID and socket path

#### Scenario: Daemon cleanup on exit
- **WHEN** debug session ends or is killed
- **THEN** daemon removes PID file
- **AND** daemon removes Unix socket
- **AND** daemon exits gracefully

### Requirement: IPC Communication
The CLI SHALL provide Unix socket-based IPC for communicating with daemon sessions.

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
The CLI SHALL maintain a registry of active daemon sessions.

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
- **AND** removes entries for non-existent processes
- **AND** removes stale PID and socket files

#### Scenario: Find session by port
- **WHEN** attach command needs to connect
- **THEN** reads registry to find session on specified port
- **AND** uses socket path from registry entry

### Requirement: Enhanced Connection Commands
The CLI connection commands SHALL support persistent daemon sessions.

#### Scenario: Kill daemon session
- **WHEN** user runs `xdebug-cli connection kill`
- **AND** daemon is running
- **THEN** connects to Unix socket
- **AND** sends kill request
- **AND** daemon sends DBGp stop command
- **AND** daemon closes connection and exits
- **AND** removes PID file and registry entry

#### Scenario: Check daemon status
- **WHEN** user runs `xdebug-cli connection isAlive`
- **AND** daemon is running
- **THEN** checks PID file exists and process is alive
- **AND** exits with code 0

#### Scenario: Show daemon details
- **WHEN** user runs `xdebug-cli connection`
- **AND** daemon is running
- **THEN** displays daemon PID, port, socket path
- **AND** displays session state and current location
- **AND** indicates session is running in daemon mode

