# cli Spec Delta

## REMOVED Requirements

### Requirement: Connection Command
The CLI SHALL provide connection status commands with support for daemon sessions.

#### Scenario: Show daemon connection status
- **WHEN** user runs `xdebug-cli connection`
- **AND** daemon is running
- **THEN** displays daemon mode indicator
- **AND** displays daemon PID and socket path
- **AND** displays current session state

#### Scenario: Kill daemon connection
- **WHEN** user runs `xdebug-cli connection kill`
- **AND** daemon is running
- **THEN** sends kill signal to daemon
- **AND** daemon terminates debug session
- **AND** daemon exits and cleans up

## ADDED Requirements

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
