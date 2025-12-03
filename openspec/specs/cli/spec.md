# cli Specification

## Purpose
TBD - created by archiving change implement-xdebug-cli. Update Purpose after archive.
## Requirements
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

### Requirement: Global CLI Flags
The CLI SHALL provide global flags for connection settings.

#### Scenario: Host flag default
- **WHEN** user runs `xdebug-cli listen --daemon` without specifying `--host`
- **THEN** server binds to `0.0.0.0` (all interfaces) by default

#### Scenario: Host flag explicit
- **WHEN** user runs with `--host 127.0.0.1` or `-l 127.0.0.1`
- **THEN** server binds to specified address (localhost only)

#### Scenario: Port flag
- **WHEN** user runs with `--port 9003` or `-p 9003`
- **THEN** server listens on specified port

### Requirement: Project Structure
The CLI SHALL be a buildable Go project following the standard template structure.

#### Scenario: Project builds successfully
- **WHEN** running `go build ./...` in the project root
- **THEN** the project compiles without errors

#### Scenario: Project structure follows conventions
- **WHEN** examining the project layout
- **THEN** it contains `cmd/source-cli/main.go` as entry point
- **AND** `internal/cli/` for command implementations
- **AND** `go.mod` with module definition

### Requirement: Preview Command
The CLI SHALL provide a `preview` command that displays an animated progress indicator.

#### Scenario: Preview with duration
- **WHEN** user runs `source-cli preview source 10s`
- **THEN** an animated loading indicator is displayed for 10 seconds
- **AND** the animation runs until the duration expires

#### Scenario: Preview without arguments
- **WHEN** user runs `source-cli preview`
- **THEN** help text is displayed explaining the command usage

### Requirement: Install Command
The CLI SHALL provide an `install` command that installs the binary to the user's local bin directory.

#### Scenario: Install to local bin
- **WHEN** user runs `source-cli install`
- **THEN** the CLI binary is built and copied to `~/.local/bin/source-cli`
- **AND** the binary includes a build timestamp

#### Scenario: Install creates directory if missing
- **WHEN** user runs `source-cli install` and `~/.local/bin` does not exist
- **THEN** the directory is created
- **AND** the binary is installed successfully

### Requirement: TDD Best Practices
The CLI SHALL follow test-driven development best practices for Go.

#### Scenario: Test files exist alongside source
- **WHEN** examining any `.go` source file in `internal/`
- **THEN** a corresponding `_test.go` file exists with unit tests

#### Scenario: Tests pass
- **WHEN** running `go test ./...`
- **THEN** all tests pass successfully

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

### Requirement: JSON Output Mode
The CLI SHALL provide `--json` global flag for machine-readable output.

#### Scenario: JSON output for variable inspection
- **WHEN** user runs with `--json` and executes "print $myArray"
- **THEN** outputs JSON with structure: `{"command": "print", "variable": "$myArray", "type": "array", "value": [...], "success": true}`

#### Scenario: JSON output for breakpoint status
- **WHEN** user runs with `--json` and executes "run"
- **THEN** outputs JSON with structure: `{"command": "run", "status": "break", "file": "app.php", "line": 42, "success": true}`

#### Scenario: JSON output for errors
- **WHEN** command fails with `--json` enabled
- **THEN** outputs JSON with structure: `{"command": "...", "success": false, "error": "error message"}`

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

