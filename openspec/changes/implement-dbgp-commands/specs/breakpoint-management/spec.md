# Spec Delta: Breakpoint Management

## MODIFIED Requirements

### Requirement: Breakpoint Commands
The CLI SHALL provide comprehensive breakpoint management including creation, deletion, enabling, and disabling.

*(Existing breakpoint_set scenarios from cli/spec.md retained)*

#### Scenario: List breakpoints shows state
- **WHEN** user runs `xdebug-cli attach --commands "info breakpoints"`
- **AND** multiple breakpoints exist with different states
- **THEN** displays each breakpoint with:
  - Breakpoint ID
  - Type (line, call, exception)
  - File and line (for line breakpoints)
  - Function name (for call breakpoints)
  - State (enabled/disabled)
  - Hit count (if available)

#### Scenario: Delete breakpoint removes from list
- **WHEN** user sets breakpoint with `break :42`
- **AND** receives breakpoint ID 123
- **AND** runs `delete 123`
- **AND** runs `info breakpoints`
- **THEN** breakpoint 123 no longer appears in list

#### Scenario: Disabled breakpoint doesn't trigger
- **WHEN** user sets breakpoint at line 100
- **AND** receives breakpoint ID 456
- **AND** runs `disable 456`
- **AND** execution reaches line 100
- **THEN** execution does not pause at line 100
- **AND** continues past the disabled breakpoint

#### Scenario: Re-enabled breakpoint triggers
- **WHEN** breakpoint 456 is disabled
- **AND** user runs `enable 456`
- **AND** execution reaches the breakpoint location
- **THEN** execution pauses at breakpoint
- **AND** debugger shows break status

#### Scenario: Workflow with multiple breakpoint operations
- **WHEN** user executes sequence:
  - `break :50` (receives ID 100)
  - `break :60` (receives ID 101)
  - `break :70` (receives ID 102)
  - `disable 101`
  - `delete 102`
  - `run`
- **THEN** execution pauses at line 50 (ID 100, enabled)
- **AND** continues past line 60 (ID 101, disabled)
- **AND** continues past line 70 (ID 102, deleted)
