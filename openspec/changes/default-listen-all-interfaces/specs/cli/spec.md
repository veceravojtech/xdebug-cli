# cli Specification Delta

## MODIFIED Requirements

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
