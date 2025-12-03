# Spec Delta: DBGp Client Methods

## ADDED Requirements

### Requirement: Status Query Method
The dbgp.Client SHALL provide a Status() method to query debugger execution state.

#### Scenario: Query status returns current state
- **WHEN** Client.Status() is called
- **AND** debugger is at breakpoint
- **THEN** sends "status -i {id}" command
- **AND** returns ProtocolResponse with status="break"
- **AND** includes filename and line number in response

#### Scenario: Status in different execution states
- **WHEN** Client.Status() is called in various states
- **THEN** returns appropriate status value:
  - "starting" - session just initialized
  - "running" - code executing
  - "break" - paused at breakpoint
  - "stopping" - shutting down
  - "stopped" - session ended

### Requirement: Detach Method
The dbgp.Client SHALL provide a Detach() method to stop debugging without terminating script execution.

#### Scenario: Detach from active session
- **WHEN** Client.Detach() is called
- **AND** session is active
- **THEN** sends "detach -i {id}" command
- **AND** script continues execution without debugger
- **AND** returns success response

#### Scenario: Detach updates session state
- **WHEN** Client.Detach() is called
- **THEN** session state transitions appropriately
- **AND** subsequent commands fail with session ended error

### Requirement: Eval Method
The dbgp.Client SHALL provide an Eval() method to execute PHP expressions at runtime.

#### Scenario: Evaluate expression with result
- **WHEN** Client.Eval("$x + 10") is called
- **THEN** sends "eval -i {id} -- base64($x + 10)" command
- **AND** returns ProtocolResponse with PropertyList
- **AND** property contains evaluated result value and type

#### Scenario: Eval with syntax error
- **WHEN** Client.Eval("invalid syntax") is called
- **THEN** sends eval command
- **AND** returns error response from debugger
- **AND** error indicates syntax problem

#### Scenario: Eval encodes expression as base64
- **WHEN** Client.Eval() is called with any expression
- **THEN** expression is base64-encoded before transmission
- **AND** follows DBGp protocol requirement for expression data

### Requirement: Set Property Method
The dbgp.Client SHALL provide a SetProperty() method to modify variable values during debugging.

#### Scenario: Set integer property
- **WHEN** Client.SetProperty("$count", "42", "int") is called
- **THEN** sends "property_set -i {id} -n $count -t int -l 2 -- base64(42)"
- **AND** returns success response
- **AND** variable is updated in debugger

#### Scenario: Set string property
- **WHEN** Client.SetProperty("$name", "John", "string") is called
- **THEN** encodes value as base64
- **AND** calculates correct data length
- **AND** sends property_set with type="string"

#### Scenario: Set property validation
- **WHEN** Client.SetProperty() is called with empty name
- **THEN** returns validation error
- **AND** does not send command to debugger

### Requirement: Get Source Method
The dbgp.Client SHALL provide a GetSource() method to retrieve source code from debugger.

#### Scenario: Get full file source
- **WHEN** Client.GetSource("file:///app/index.php", 0, 0) is called
- **THEN** sends "source -i {id} -f file:///app/index.php"
- **AND** returns ProtocolResponse with base64-encoded source
- **AND** source is decoded and available

#### Scenario: Get source with line range
- **WHEN** Client.GetSource("file:///app/lib.php", 100, 120) is called
- **THEN** sends "source -i {id} -f file:///app/lib.php -b 100 -e 120"
- **AND** returns only specified line range

#### Scenario: Get current file source
- **WHEN** Client.GetSource("", 0, 0) is called
- **AND** execution is at specific file
- **THEN** sends "source -i {id}" without -f parameter
- **AND** debugger returns current file source

### Requirement: Update Breakpoint Method
The dbgp.Client SHALL provide an UpdateBreakpoint() method to modify breakpoint state.

#### Scenario: Disable breakpoint
- **WHEN** Client.UpdateBreakpoint("123", "disabled") is called
- **THEN** sends "breakpoint_update -i {id} -d 123 -s disabled"
- **AND** returns success response
- **AND** breakpoint remains but doesn't trigger

#### Scenario: Enable breakpoint
- **WHEN** Client.UpdateBreakpoint("123", "enabled") is called
- **THEN** sends "breakpoint_update -i {id} -d 123 -s enabled"
- **AND** breakpoint becomes active

#### Scenario: Update non-existent breakpoint
- **WHEN** Client.UpdateBreakpoint("999", "disabled") is called
- **THEN** sends command to debugger
- **AND** returns error response indicating breakpoint not found

## MODIFIED Requirements

### Requirement: Debug Client Operations
The dbgp package SHALL provide high-level debugging operations.

*(Existing scenarios retained, added scenarios below)*

#### Scenario: Remove breakpoint by ID
- **WHEN** Client.RemoveBreakpoint("123") is called
- **AND** breakpoint 123 exists
- **THEN** sends "breakpoint_remove -i {id} -d 123"
- **AND** returns success response
- **AND** breakpoint no longer triggers

*(Note: RemoveBreakpoint already exists in client.go:326, this scenario documents its expected behavior)*
