# dbgp Specification

## Purpose
TBD - created by archiving change implement-xdebug-cli. Update Purpose after archive.
## Requirements
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

### Requirement: Connection Message Framing
The dbgp package SHALL handle DBGp protocol message framing with size validation.

#### Scenario: ReadMessage parses DBGp format
- **WHEN** connection receives data in format "size\0xml\0"
- **AND** ReadMessage() is called
- **THEN** returns parsed ProtocolInit or ProtocolResponse
- **AND** returns error on malformed data

#### Scenario: SendMessage writes with null terminator
- **WHEN** SendMessage("step_into -i 1") is called
- **THEN** writes "step_into -i 1\0" to socket

#### Scenario: ReadMessage validates size field format
- **WHEN** size field contains non-digit characters
- **THEN** returns error with descriptive message showing invalid content
- **AND** does not attempt to allocate memory based on invalid size

#### Scenario: ReadMessage enforces maximum message size
- **WHEN** size field value exceeds MaxMessageSize (100MB)
- **THEN** returns error indicating message too large
- **AND** prevents memory exhaustion from corrupted size values

#### Scenario: ReadMessage rejects negative sizes
- **WHEN** size field parses to negative value
- **THEN** returns error indicating invalid size
- **AND** does not attempt buffer allocation

### Requirement: Protocol XML Parsing
The dbgp package SHALL parse DBGp protocol XML messages.

#### Scenario: Parse init message
- **WHEN** CreateProtocolFromXML receives init XML
- **THEN** returns *ProtocolInit with FileURI, Language, AppID fields

#### Scenario: Parse response message
- **WHEN** CreateProtocolFromXML receives response XML
- **THEN** returns *ProtocolResponse with Command, Status, Reason fields
- **AND** includes PropertyList, BreakpointList, etc. when present

### Requirement: Debug Client Operations
The dbgp package SHALL provide high-level debugging operations.

#### Scenario: Initialize session
- **WHEN** Client.Init() is called
- **THEN** reads init protocol from Xdebug
- **AND** sets Session.State to StateStarting
- **AND** discovers target files from init FileURI

#### Scenario: Run until breakpoint
- **WHEN** Client.Run() is called
- **THEN** sends "run -i {id}" command
- **AND** returns ProtocolResponse with break/stopping status

#### Scenario: Step into next statement
- **WHEN** Client.Step() is called
- **THEN** sends "step_into -i {id}" command
- **AND** returns new file/line position

#### Scenario: Step over next statement
- **WHEN** Client.Next() is called
- **THEN** sends "step_over -i {id}" command
- **AND** does not enter function calls

#### Scenario: Set line breakpoint
- **WHEN** Client.SetBreakpoint("/app.php", 42, "") is called
- **THEN** sends "breakpoint_set -i {id} -t line -f file:///app.php -n 42"
- **AND** returns nil on success

#### Scenario: Get variable value
- **WHEN** Client.GetProperty("$myVar") is called
- **THEN** sends "property_get -i {id} -n $myVar"
- **AND** returns ProtocolResponse with PropertyList

#### Scenario: Stop debugging session
- **WHEN** Client.Finish() is called
- **THEN** sends "stop -i {id}" command
- **AND** session cannot be resumed

### Requirement: Session State Management
The dbgp package SHALL track debugging session state.

#### Scenario: Transaction ID increments
- **WHEN** NextTransactionID() is called multiple times
- **THEN** returns incrementing IDs (2, 3, 4, ...)
- **AND** starts at 1 for new session

#### Scenario: Command history tracks commands
- **WHEN** AddCommand("step") is called
- **AND** GetLastCommand() is called
- **THEN** returns ("step", true)

### Requirement: Message Read Timeout
The dbgp Connection SHALL support timeout-based message reading to prevent indefinite hangs.

#### Scenario: ReadMessageWithTimeout returns within deadline
- **WHEN** ReadMessageWithTimeout(30s) is called
- **AND** valid DBGp message arrives within 30 seconds
- **THEN** returns parsed message
- **AND** clears the read deadline

#### Scenario: ReadMessageWithTimeout times out
- **WHEN** ReadMessageWithTimeout(5s) is called
- **AND** no data arrives within 5 seconds
- **THEN** returns timeout error
- **AND** error message indicates read timeout

#### Scenario: Default timeout is 30 seconds
- **WHEN** connection reads Xdebug response
- **THEN** applies DefaultMessageTimeout (30 seconds) if no explicit timeout specified
- **AND** prevents indefinite blocking on unresponsive connections

