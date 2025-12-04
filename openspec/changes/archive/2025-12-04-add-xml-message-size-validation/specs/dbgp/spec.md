## MODIFIED Requirements

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
