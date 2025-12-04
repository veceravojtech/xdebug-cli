## MODIFIED Requirements

### Requirement: IPC Communication
The CLI SHALL provide Unix socket-based IPC for communicating with daemon sessions with automatic retry.

#### Scenario: Create IPC socket
- **WHEN** daemon starts
- **THEN** creates Unix socket at `/tmp/xdebug-cli-session-<pid>.sock`
- **AND** socket has permissions 0600 (owner only)
- **AND** socket path is written to session registry

#### Scenario: Accept IPC connections
- **WHEN** client connects to Unix socket
- **THEN** daemon accepts connection
- **AND** reads JSON command request
- **AND** executes commands via existing dispatcher
- **AND** returns JSON response with results

#### Scenario: Handle concurrent connections
- **WHEN** multiple clients connect simultaneously
- **THEN** daemon processes requests sequentially
- **AND** maintains thread-safe session access
- **AND** returns results to correct client

#### Scenario: IPC protocol structure
- **WHEN** client sends command request
- **THEN** request contains command list and output format flag
- **AND** response contains success status and result array
- **AND** each result includes command name, success flag, and data

#### Scenario: Client retries on transient connection failure
- **WHEN** initial connection attempt fails
- **THEN** client retries with exponential backoff (100ms, 200ms, 400ms)
- **AND** makes up to 3 attempts by default
- **AND** returns error only after all attempts exhausted

#### Scenario: Client connects immediately when daemon ready
- **WHEN** daemon socket is accepting connections
- **AND** client attempts to connect
- **THEN** connection succeeds on first attempt
- **AND** no unnecessary delay is introduced
