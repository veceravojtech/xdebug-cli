## ADDED Requirements

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
