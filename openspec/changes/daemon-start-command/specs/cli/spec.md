# cli Specification Delta

## MODIFIED Requirements

### Requirement: Listen Command
The CLI SHALL provide a listen command to start the DBGp server with command-based execution only (daemon mode removed).

#### Scenario: Start listening server with commands
- **WHEN** user runs `xdebug-cli listen --commands "run" "print \$var"`
- **THEN** server waits for Xdebug connection
- **AND** executes commands sequentially when connection established
- **AND** exits after commands complete or session ends

#### Scenario: Missing commands flag
- **WHEN** user runs `xdebug-cli listen` without `--commands` flag
- **THEN** displays error message explaining `--commands` is required
- **AND** shows usage examples for command-based execution
- **AND** suggests `daemon start` for persistent sessions
- **AND** exits with non-zero code

#### REMOVED Scenario: Start listening server with daemon flag
*(Scenario removed - daemon mode now uses `daemon start` command)*

#### REMOVED Scenario: Daemon mode with commands
*(Scenario removed - daemon mode now uses `daemon start` command)*

### Requirement: Commands Flag
The CLI SHALL require the `--commands` flag for listen command (daemon mode no longer supported on listen).

#### Scenario: Multiple commands executed in order
- **WHEN** user provides `--commands "break :42" "run" "print $x"`
- **THEN** sets breakpoint at line 42
- **AND** continues execution until breakpoint
- **AND** prints variable $x value

#### Scenario: Commands required for listen
- **WHEN** user runs `xdebug-cli listen` without `--commands`
- **THEN** displays error about missing required flag
- **AND** exits with non-zero code

#### REMOVED Scenario: Commands optional for daemon mode
*(Scenario removed - daemon mode moved to separate command)*

## ADDED Requirements

### Requirement: Daemon Command
The CLI SHALL provide a daemon command with subcommands for managing persistent debug sessions.

#### Scenario: Daemon parent command without subcommand
- **WHEN** user runs `xdebug-cli daemon`
- **THEN** displays help text listing available subcommands
- **AND** shows usage examples for `daemon start`

#### Scenario: Daemon command with invalid subcommand
- **WHEN** user runs `xdebug-cli daemon invalid`
- **THEN** displays error about unknown subcommand
- **AND** lists valid subcommands

### Requirement: Daemon Start Subcommand
The CLI SHALL provide a `daemon start` subcommand to start persistent background debug sessions.

#### Scenario: Start daemon with default settings
- **WHEN** user runs `xdebug-cli daemon start`
- **THEN** kills any existing daemon on port 9003 automatically
- **AND** forks process to background
- **AND** parent exits immediately
- **AND** daemon binds to `0.0.0.0:9003`
- **AND** daemon waits for Xdebug connections

#### Scenario: Start daemon with custom port
- **WHEN** user runs `xdebug-cli daemon start -p 9004`
- **THEN** kills any existing daemon on port 9004 automatically
- **AND** daemon binds to `0.0.0.0:9004`

#### Scenario: Start daemon with initial commands
- **WHEN** user runs `xdebug-cli daemon start --commands "break /path/file.php:100"`
- **THEN** daemon starts in background
- **AND** waits for Xdebug connection
- **AND** executes commands when connection established
- **AND** keeps session alive after command execution

#### Scenario: Start daemon with JSON output
- **WHEN** user runs `xdebug-cli daemon start --json --commands "break :42"`
- **THEN** daemon starts in background
- **AND** command results are formatted as JSON when connection established

#### Scenario: Daemon auto-force behavior
- **WHEN** user runs `xdebug-cli daemon start` on port with existing daemon
- **THEN** automatically kills existing daemon on that port
- **AND** cleans up stale PID and socket files
- **AND** starts new daemon successfully
- **AND** displays message about killed daemon

#### Scenario: Daemon inherits global flags
- **WHEN** user runs `xdebug-cli daemon start --port 9005` or `-p 9005`
- **THEN** daemon uses port 9005
- **WHEN** user runs `xdebug-cli daemon start --json`
- **THEN** daemon output uses JSON format
