# persistent-sessions Spec Delta

## MODIFIED Requirements

### Requirement: Enhanced Connection Commands
The CLI SHALL provide daemon subcommands for session management and lifecycle operations.

#### Scenario: Kill daemon session
- **WHEN** user runs `xdebug-cli daemon kill`
- **AND** daemon is running on current port (--port flag)
- **THEN** looks up session in registry by port
- **AND** sends kill request via IPC socket
- **AND** daemon receives request and terminates gracefully
- **AND** registry entry is removed
- **AND** displays "Daemon terminated successfully."

#### Scenario: Check daemon status
- **WHEN** user runs `xdebug-cli daemon status`
- **AND** daemon is running
- **THEN** displays daemon mode indicator
- **AND** displays PID, port, socket path, and start time

#### Scenario: Show daemon details
- **WHEN** user runs `xdebug-cli daemon status`
- **AND** daemon exists on current port
- **THEN** displays connection status header
- **AND** shows PID from registry
- **AND** shows port from registry
- **AND** shows socket path from registry
- **AND** shows formatted start timestamp
- **AND** includes usage hint for killing the daemon

#### Scenario: List all daemon sessions
- **WHEN** user runs `xdebug-cli daemon list`
- **THEN** reads all entries from session registry
- **AND** filters out sessions with dead processes
- **AND** displays table with PID, Port, Started, Socket Path
- **AND** shows total count of active sessions

#### Scenario: List daemon sessions in JSON
- **WHEN** user runs `xdebug-cli daemon list --json`
- **THEN** outputs JSON array of active sessions
- **AND** includes pid, port, socket_path, started_at for each

#### Scenario: Check if daemon is alive
- **WHEN** user runs `xdebug-cli daemon isAlive`
- **THEN** checks registry for session on current port
- **AND** verifies process exists via /proc/<pid>
- **AND** exits with code 0 and prints "connected" if alive
- **AND** exits with code 1 and prints "not connected" if not alive

#### Scenario: Kill all daemon sessions
- **WHEN** user runs `xdebug-cli daemon kill --all`
- **THEN** lists all active sessions from registry
- **AND** prompts for confirmation unless --force flag used
- **AND** sends kill request to each session via IPC
- **AND** displays progress and summary
- **AND** exits with code 1 if any kills fail
