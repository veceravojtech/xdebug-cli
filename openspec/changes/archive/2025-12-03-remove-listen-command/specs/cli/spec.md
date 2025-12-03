# cli Specification Deltas

## REMOVED Requirements

### Requirement: Listen Command
**Reason**: Removing dual-mode execution model in favor of daemon-only workflow. The listen command's one-shot execution mode is redundant with the more powerful daemon + attach pattern.

**Migration**: Users should replace `xdebug-cli listen --commands "..."` with the two-step daemon workflow:
1. Start daemon: `xdebug-cli daemon start [--commands "..."]`
2. Execute commands: `xdebug-cli attach --commands "..."`

For one-shot execution needs, users can chain commands: `xdebug-cli daemon start && xdebug-cli attach --commands "..." && xdebug-cli connection kill`

### Requirement: Commands Flag
**Reason**: This requirement was specific to the `listen` command's command-based execution mode, which is being removed.

**Migration**: The `--commands` flag is now used exclusively with `daemon start` (for initial commands) and `attach` (for subsequent commands).

### Requirement: Non-Interactive Mode Flag
**Reason**: This requirement described command-based execution for the listen command, which is being removed.

**Migration**: All debugging is now non-interactive by default (daemon mode). Users execute commands via `attach --commands`.

## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Daemon Start Command
The CLI SHALL provide a `daemon start` command as the primary entry point for all debugging sessions.

#### Scenario: Start daemon without commands
- **WHEN** user runs `xdebug-cli daemon start`
- **THEN** process forks to background
- **AND** parent process exits successfully
- **AND** daemon starts DBGp server and waits for connections
- **AND** displays message: "Daemon started on port 9003"

#### Scenario: Start daemon with initial commands
- **WHEN** user runs `xdebug-cli daemon start --commands "break /path/file.php:100"`
- **THEN** daemon starts in background
- **AND** waits for Xdebug connection
- **AND** executes commands when connection established
- **AND** keeps session alive after commands complete

#### Scenario: Start daemon with force flag
- **WHEN** user runs `xdebug-cli daemon start --force`
- **AND** daemon already running on same port
- **THEN** kills existing daemon on that port
- **AND** starts new daemon successfully
- **AND** displays message: "Killed daemon on port 9003 (PID 12345)" followed by "Daemon started on port 9003"

#### Scenario: Daemon already running without force
- **WHEN** user runs `xdebug-cli daemon start`
- **AND** daemon already running on port 9003
- **THEN** command exits with error code 1
- **AND** displays message: "Error: daemon already running on port 9003 (PID 12345). Use 'xdebug-cli connection kill' to terminate it first or use --force to replace it."
