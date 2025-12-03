# cli Specification Delta

## MODIFIED Requirements

### Requirement: Listen Command
The CLI SHALL provide a listen command to start the DBGp server with optional daemon mode.

#### Scenario: Start listening server with daemon flag
- **WHEN** user runs `xdebug-cli listen -p 9003 --daemon`
- **THEN** forks process to background
- **AND** parent exits immediately
- **AND** daemon continues running and waiting for connections

#### Scenario: Daemon mode with commands
- **WHEN** user runs `xdebug-cli listen --daemon --commands "break :42"`
- **THEN** daemon starts in background
- **AND** executes commands when connection established
- **AND** keeps session alive after command execution

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
- **AND** displays helpful error message

#### Scenario: Attach with JSON output
- **WHEN** user runs `xdebug-cli attach --json --commands "print \$x"`
- **THEN** requests JSON output from daemon
- **AND** displays JSON-formatted results
