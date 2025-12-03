# CLI Commands Capability

## ADDED Requirements

### Requirement: Listen Command
The CLI SHALL provide a listen command to start the DBGp server.

#### Scenario: Start listening server
- **WHEN** user runs `xdebug-cli listen -p 9003`
- **THEN** server binds to port 9003
- **AND** displays application banner
- **AND** waits for incoming Xdebug connections

#### Scenario: Handle incoming connection
- **WHEN** PHP script with Xdebug connects
- **THEN** initializes debug session
- **AND** sets exception breakpoint by default
- **AND** starts interactive REPL

### Requirement: Interactive REPL
The CLI SHALL provide an interactive debugging REPL in listen mode.

#### Scenario: Execute run command
- **WHEN** user types "run" or "r"
- **THEN** continues execution until breakpoint
- **AND** displays breakpoint location and source

#### Scenario: Execute step command
- **WHEN** user types "step" or "s"
- **THEN** steps into next statement
- **AND** displays new location and source

#### Scenario: Execute next command
- **WHEN** user types "next" or "n"
- **THEN** steps over next statement
- **AND** does not enter function calls

#### Scenario: Set breakpoint
- **WHEN** user types "break :42"
- **THEN** sets breakpoint at line 42 in current file
- **AND** supports "file.php:42" and "call funcName" syntax

#### Scenario: Print variable
- **WHEN** user types "print $myVar"
- **THEN** displays variable value with type information

#### Scenario: Show context
- **WHEN** user types "context local"
- **THEN** displays all local variables

#### Scenario: List source
- **WHEN** user types "list" or "l"
- **THEN** displays source code around current line

#### Scenario: Show breakpoint info
- **WHEN** user types "info breakpoints"
- **THEN** displays table of all breakpoints

#### Scenario: Repeat last command
- **WHEN** user presses Enter with empty input
- **THEN** repeats previous command if one exists

#### Scenario: Quit debugger
- **WHEN** user types "quit" or "q"
- **AND** confirms with "y"
- **THEN** closes debug session

### Requirement: Connection Command
The CLI SHALL provide connection status commands.

#### Scenario: Show connection status
- **WHEN** user runs `xdebug-cli connection`
- **THEN** displays "connected" or "not connected"

#### Scenario: Check if alive
- **WHEN** user runs `xdebug-cli connection isAlive`
- **THEN** exits with code 0 if connected
- **AND** exits with code 1 if not connected

#### Scenario: Kill connection
- **WHEN** user runs `xdebug-cli connection kill`
- **THEN** terminates active session if one exists
- **AND** displays confirmation message

### Requirement: Global CLI Flags
The CLI SHALL provide global flags for connection settings.

#### Scenario: Host flag
- **WHEN** user runs with `--host 0.0.0.0` or `-l 0.0.0.0`
- **THEN** server binds to specified address

#### Scenario: Port flag
- **WHEN** user runs with `--port 9003` or `-p 9003`
- **THEN** server listens on specified port
