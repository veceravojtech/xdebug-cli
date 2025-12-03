## ADDED Requirements

### Requirement: Non-Interactive Mode Flag
The CLI SHALL provide a `--non-interactive` flag for the listen command that processes debugging commands from CLI arguments instead of interactive REPL.

#### Scenario: Start non-interactive session with commands
- **WHEN** user runs `xdebug-cli listen --non-interactive --commands "run" "step" "print myVar"`
- **THEN** server waits for connection without REPL
- **AND** executes commands sequentially after connection establishes
- **AND** exits when all commands complete

#### Scenario: Non-interactive mode with JSON output
- **WHEN** user runs `xdebug-cli listen --non-interactive --json --commands "context local"`
- **THEN** outputs JSON-formatted results
- **AND** includes structured data for variables, breakpoints, and state

#### Scenario: Non-interactive mode exits on error
- **WHEN** user runs `xdebug-cli listen --non-interactive --commands "invalid"`
- **THEN** displays error message
- **AND** exits with non-zero exit code

#### Scenario: Non-interactive mode suppresses prompts
- **WHEN** user runs `xdebug-cli listen --non-interactive --commands "run"`
- **THEN** does not display REPL prompt or application banner
- **AND** outputs only command results

### Requirement: Commands Flag
The CLI SHALL accept multiple debugging commands via `--commands` flag in non-interactive mode.

#### Scenario: Multiple commands executed in order
- **WHEN** user provides `--commands "break :42" "run" "print $x"`
- **THEN** sets breakpoint at line 42
- **AND** continues execution until breakpoint
- **AND** prints variable $x value

#### Scenario: Commands flag requires non-interactive mode
- **WHEN** user runs `xdebug-cli listen --commands "run"` without `--non-interactive`
- **THEN** displays error about incompatible flags
- **AND** exits with non-zero code

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

## MODIFIED Requirements

### Requirement: Listen Command
The CLI SHALL provide a listen command to start the DBGp server with optional non-interactive mode.

#### Scenario: Start listening server
- **WHEN** user runs `xdebug-cli listen -p 9003`
- **THEN** server binds to port 9003
- **AND** displays application banner (unless `--non-interactive` is set)
- **AND** waits for incoming Xdebug connections

#### Scenario: Handle incoming connection in interactive mode
- **WHEN** PHP script with Xdebug connects
- **AND** no `--non-interactive` flag is present
- **THEN** initializes debug session
- **AND** sets exception breakpoint by default
- **AND** starts interactive REPL

#### Scenario: Handle incoming connection in non-interactive mode
- **WHEN** PHP script with Xdebug connects
- **AND** `--non-interactive` flag is present with `--commands`
- **THEN** initializes debug session
- **AND** executes provided commands sequentially
- **AND** exits after last command or on error
