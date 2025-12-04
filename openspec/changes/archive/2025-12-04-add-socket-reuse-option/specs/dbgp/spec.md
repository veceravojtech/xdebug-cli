## MODIFIED Requirements

### Requirement: TCP Server
The dbgp package SHALL provide a TCP server for accepting Xdebug connections.

#### Scenario: Server binds to address and port
- **WHEN** NewServer("127.0.0.1", 9003) is called
- **AND** Listen() is called
- **THEN** server binds to 127.0.0.1:9003
- **AND** returns nil error on success

#### Scenario: Server accepts connections
- **WHEN** server is listening
- **AND** Accept(handler) is called
- **THEN** handler is invoked for each incoming connection
- **AND** each connection runs in its own goroutine

#### Scenario: Server uses SO_REUSEADDR socket option
- **WHEN** Listen() is called
- **THEN** the TCP socket is created with SO_REUSEADDR option enabled
- **AND** the server can bind immediately after a previous server closes
- **AND** TIME_WAIT state does not prevent rebinding
