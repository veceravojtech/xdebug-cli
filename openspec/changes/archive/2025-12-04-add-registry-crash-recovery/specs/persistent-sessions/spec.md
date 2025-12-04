## MODIFIED Requirements

### Requirement: Session Registry
The CLI SHALL maintain a registry of active daemon sessions with crash recovery.

#### Scenario: Register new session
- **WHEN** daemon starts
- **THEN** creates registry file `~/.xdebug-cli/sessions.json` if missing
- **AND** adds entry with PID, port, socket path, start time

#### Scenario: Query active sessions
- **WHEN** user runs `xdebug-cli connection`
- **AND** daemon is running
- **THEN** reads registry to find session
- **AND** displays session information

#### Scenario: Cleanup stale entries
- **WHEN** daemon starts
- **OR** connection command queries sessions
- **THEN** validates each registry entry process exists
- **AND** validates process is xdebug-cli (not recycled PID)
- **AND** removes entries for non-existent processes
- **AND** removes entries for recycled PIDs running different executables
- **AND** removes stale PID and socket files

#### Scenario: Find session by port
- **WHEN** attach command needs to connect
- **THEN** reads registry to find session on specified port
- **AND** uses socket path from registry entry

#### Scenario: Validate process identity
- **WHEN** registry contains entry with PID 12345
- **AND** process 12345 exists but is not xdebug-cli
- **THEN** entry is considered stale
- **AND** entry is removed during cleanup
- **AND** associated socket and PID files are removed
