# persistent-sessions Specification Delta

## MODIFIED Requirements

### Requirement: Daemon Mode
The CLI SHALL support running the debug server as a background daemon using the `daemon start` command.

#### Scenario: Start daemon without commands
- **WHEN** user runs `xdebug-cli daemon start`
- **THEN** automatically kills any existing daemon on port 9003
- **AND** process forks to background
- **AND** parent process exits successfully
- **AND** daemon writes PID to `/tmp/xdebug-cli-daemon-<port>.pid`
- **AND** daemon starts DBGp server on `0.0.0.0:9003`
- **AND** waits for connections

#### Scenario: Start daemon with initial commands
- **WHEN** user runs `xdebug-cli daemon start --commands "break :42" "run"`
- **THEN** daemon starts in background
- **AND** waits for Xdebug connection
- **AND** executes commands when connection established
- **AND** keeps session alive after commands complete

#### Scenario: Daemon auto-cleanup
- **WHEN** user runs `xdebug-cli daemon start` and existing daemon on same port
- **THEN** validates existing daemon PID
- **AND** kills existing process if alive
- **AND** cleans up stale PID if process dead
- **AND** displays appropriate message (killed or stale cleanup)
- **AND** starts new daemon successfully

#### REMOVED Scenario: Daemon already running
*(Scenario removed - daemon start now auto-kills with --force behavior)*

#### Scenario: Daemon cleanup on exit
- **WHEN** debug session ends or is killed
- **THEN** daemon removes PID file
- **AND** daemon removes Unix socket
- **AND** daemon exits gracefully

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
- **AND** indicates session was started with `daemon start`

## ADDED Requirements

### Requirement: Daemon Start Command Behavior
The `daemon start` command SHALL automatically handle stale daemon cleanup.

#### Scenario: Force behavior always enabled
- **WHEN** user runs `xdebug-cli daemon start`
- **THEN** checks for existing daemon on target port
- **AND** if found, validates PID exists
- **AND** kills process if alive
- **AND** cleans up stale registry entry if process dead
- **AND** continues with starting new daemon
- **AND** never fails due to existing daemon

#### Scenario: Force cleanup messages
- **WHEN** `daemon start` kills existing daemon
- **THEN** displays: "Killed daemon on port <port> (PID <pid>)"
- **WHEN** `daemon start` finds stale daemon
- **THEN** displays: "Warning: daemon on port <port> is stale (PID <pid> no longer exists), cleaning up"
- **WHEN** no existing daemon found
- **THEN** displays: "Server listening on 0.0.0.0:<port>" (no warning)

#### Scenario: Port defaults and overrides
- **WHEN** user runs `xdebug-cli daemon start` without `-p` flag
- **THEN** daemon uses port 9003 (global default)
- **WHEN** user runs `xdebug-cli daemon start -p 9005`
- **THEN** daemon uses port 9005
- **AND** force-kills any daemon on port 9005 only
