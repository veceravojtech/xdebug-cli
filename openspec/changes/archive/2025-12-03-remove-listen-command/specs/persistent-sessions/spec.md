# persistent-sessions Specification Deltas

## MODIFIED Requirements

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
