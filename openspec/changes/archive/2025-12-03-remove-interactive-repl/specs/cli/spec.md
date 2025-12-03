## REMOVED Requirements

### Requirement: Interactive REPL
**Reason**: Interactive REPL mode is legacy functionality. Non-interactive mode (command-based execution) and daemon mode (persistent sessions with attach) cover all use cases more effectively. Removing REPL simplifies the codebase and focuses the tool on automation and programmatic debugging.

**Migration**:
- Users who ran `xdebug-cli listen` for interactive REPL should use `xdebug-cli listen --commands "cmd1" "cmd2"` for single-session debugging
- Users who need multi-step exploration should use daemon mode: `xdebug-cli listen --daemon` followed by `xdebug-cli attach --commands "..."`

## MODIFIED Requirements

### Requirement: Listen Command
The CLI SHALL provide a listen command to start the DBGp server with command-based execution or daemon mode.

#### Scenario: Start listening server with commands
- **WHEN** user runs `xdebug-cli listen --commands "run" "print \$var"`
- **THEN** server waits for Xdebug connection
- **AND** executes commands sequentially when connection established
- **AND** exits after commands complete or session ends

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

#### Scenario: Missing commands flag without daemon
- **WHEN** user runs `xdebug-cli listen` without `--commands` or `--daemon` flags
- **THEN** displays error message explaining `--commands` is required
- **AND** shows usage examples for command-based and daemon modes
- **AND** exits with non-zero code

### Requirement: Non-Interactive Mode Flag
The CLI SHALL execute debugging commands from the `--commands` flag for the listen command. Note: The `--non-interactive` flag has been removed; command-based execution is now the default and only mode for the listen command (outside of daemon mode).

#### Scenario: Execute commands from arguments
- **WHEN** user runs `xdebug-cli listen --commands "run" "step" "print myVar"`
- **THEN** server waits for connection
- **AND** executes commands sequentially after connection establishes
- **AND** exits when all commands complete

#### Scenario: Commands with JSON output
- **WHEN** user runs `xdebug-cli listen --json --commands "context local"`
- **THEN** outputs JSON-formatted results
- **AND** includes structured data for variables, breakpoints, and state

#### Scenario: Commands execution exits on error
- **WHEN** user runs `xdebug-cli listen --commands "invalid"`
- **THEN** displays error message
- **AND** exits with non-zero exit code

#### Scenario: Commands suppress prompts
- **WHEN** user runs `xdebug-cli listen --commands "run"`
- **THEN** does not display REPL prompt or interactive messages
- **AND** outputs only command results

### Requirement: Commands Flag
The CLI SHALL require the `--commands` flag for listen command unless using daemon mode.

#### Scenario: Multiple commands executed in order
- **WHEN** user provides `--commands "break :42" "run" "print $x"`
- **THEN** sets breakpoint at line 42
- **AND** continues execution until breakpoint
- **AND** prints variable $x value

#### Scenario: Commands required for listen
- **WHEN** user runs `xdebug-cli listen` without `--commands` and without `--daemon`
- **THEN** displays error about missing required flag
- **AND** exits with non-zero code

#### Scenario: Commands optional for daemon mode
- **WHEN** user runs `xdebug-cli listen --daemon`
- **THEN** starts daemon without requiring `--commands` flag
- **AND** waits for attach commands via separate invocations
