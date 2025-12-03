# view Specification

## Purpose
TBD - created by archiving change implement-xdebug-cli. Update Purpose after archive.
## Requirements
### Requirement: Terminal Output
The view package SHALL provide console output methods for displaying command results.

#### Scenario: Print with and without newline
- **WHEN** Print("text") or PrintLn("text") is called
- **THEN** writes to stdout with or without trailing newline

#### Scenario: Print application banner
- **WHEN** PrintApplicationInformation("1.0.0", "127.0.0.1", 9003) is called
- **THEN** displays version, listening address, and help links

#### Scenario: Print errors
- **WHEN** PrintErrorLn("error message") is called
- **THEN** writes to stderr with newline

### Requirement: Source Code Display
The view package SHALL display source code with line numbers.

#### Scenario: Display source lines
- **WHEN** PrintSourceLn("file:///app.php", 10, 5) is called
- **THEN** displays 5 lines starting at line 10
- **AND** each line is prefixed with its line number

#### Scenario: Cache file contents
- **WHEN** same file is accessed multiple times
- **THEN** file is read from cache, not disk

#### Scenario: Handle file errors
- **WHEN** file does not exist or is not accessible
- **THEN** prints error message

### Requirement: Help Messages
The view package SHALL provide formatted help text for CLI-level help.

#### Scenario: Show command-specific help
- **WHEN** ShowBreakpointHelpMessage() is called
- **THEN** displays breakpoint command syntax and examples

#### Scenario: Show main command help
- **WHEN** help-related view methods are called
- **THEN** displays usage information for debugging commands

### Requirement: Property Display
The view package SHALL format debug properties for display.

#### Scenario: Display variable tree
- **WHEN** PrintPropertyListWithDetails("local", properties) is called
- **THEN** displays scope name and property tree
- **AND** nested properties are indented
- **AND** base64 values are decoded

#### Scenario: Truncate long values
- **WHEN** property value exceeds 40 characters
- **THEN** truncates to 36 characters plus "..."

### Requirement: Breakpoint Display
The view package SHALL display breakpoint information.

#### Scenario: Display breakpoint table
- **WHEN** ShowInfoBreakpoints(breakpoints) is called
- **THEN** displays table with Num, Type, Enabled, What columns
- **AND** file paths are shortened to basename

