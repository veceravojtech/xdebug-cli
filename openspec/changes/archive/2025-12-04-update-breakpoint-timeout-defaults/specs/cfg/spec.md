## MODIFIED Requirements

### Requirement: CLI Parameters
The cfg package SHALL provide a CLIParameter struct for command-line arguments.

#### Scenario: CLIParameter contains connection settings
- **WHEN** CLIParameter is initialized
- **THEN** it contains Host (string), Port (uint16), Trigger (string), and Version (string) fields
- **AND** all fields are exported (public)

#### Scenario: Default breakpoint timeout is 30 seconds
- **WHEN** daemon starts without explicit `--breakpoint-timeout` flag
- **THEN** uses default timeout of 30 seconds for breakpoint validation
- **AND** provides enough time for cold PHP opcache and framework bootstrap

## ADDED Requirements

### Requirement: Wait Forever Flag
The CLI SHALL provide a `--wait-forever` flag for indefinite breakpoint waiting.

#### Scenario: Wait forever disables timeout
- **WHEN** daemon starts with `--wait-forever` flag
- **THEN** breakpoint timeout is disabled (set to 0)
- **AND** daemon waits indefinitely for breakpoint to be hit

#### Scenario: Wait forever overrides default timeout
- **WHEN** daemon starts with `--wait-forever`
- **AND** no explicit `--breakpoint-timeout` is provided
- **THEN** timeout is disabled regardless of default value

#### Scenario: Explicit timeout overrides wait forever
- **WHEN** both `--wait-forever` and `--breakpoint-timeout 60` are provided
- **THEN** explicit timeout value takes precedence
- **AND** timeout is set to 60 seconds
