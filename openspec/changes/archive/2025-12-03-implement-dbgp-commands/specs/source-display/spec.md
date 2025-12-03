# Spec Delta: Source Code Display

## ADDED Requirements

### Requirement: Source Code Retrieval
The CLI SHALL provide source code display using the debugger's source command as an alternative to local file reading.

#### Scenario: Source command uses debugger data
- **WHEN** user runs `source app.php`
- **THEN** retrieves source via DBGp `source` command
- **AND** does not read from local filesystem
- **AND** shows source code as debugger sees it

#### Scenario: Remote debugging source access
- **WHEN** debugging remote PHP application
- **AND** source files not available locally
- **AND** user runs `source /var/www/app.php`
- **THEN** debugger provides source code over DBGp protocol
- **AND** displays correctly even though local file doesn't exist

#### Scenario: Source display with line numbers
- **WHEN** user runs `source lib.php`
- **THEN** displays source with line numbers in format:
  ```
  100 | <?php
  101 | function example() {
  102 |     return true;
  103 | }
  ```

#### Scenario: Source highlights current line
- **WHEN** execution paused at app.php:150
- **AND** user runs `source app.php:145-155`
- **THEN** displays line range 145-155
- **AND** highlights or marks line 150 as current execution point

#### Scenario: Source without arguments shows current file
- **WHEN** execution paused at helper.php:42
- **AND** user runs `source` with no arguments
- **THEN** displays source for helper.php
- **AND** defaults to showing region around line 42

#### Scenario: Source with only line range
- **WHEN** user runs `source :100-120`
- **THEN** displays lines 100-120 of current file

#### Scenario: Parse file path with line range
- **WHEN** user runs `source /var/www/lib/util.php:50-75`
- **THEN** extracts file: `/var/www/lib/util.php`
- **AND** extracts range: start=50, end=75
- **AND** requests specific range from debugger

#### Scenario: Parse relative file path
- **WHEN** user runs `source lib/helper.php`
- **AND** current execution is in /var/www/app.php
- **THEN** resolves relative path against project root
- **AND** converts to file:// URI for debugger

#### Scenario: Source handles base64 decoding
- **WHEN** debugger returns base64-encoded source
- **THEN** CLI decodes source correctly
- **AND** displays readable PHP code
- **AND** preserves whitespace and formatting

#### Scenario: Source handles Unicode and special characters
- **WHEN** source file contains Unicode characters or emojis
- **THEN** displays correctly in terminal
- **AND** preserves character encoding

#### Scenario: Compare source vs list commands
- **WHEN** user runs `list` (existing command)
- **THEN** displays local file contents from filesystem
- **WHEN** user runs `source` (new command)
- **THEN** displays source from debugger via DBGp protocol
- **AND** both show same content for local debugging
- **AND** `source` works for remote debugging where `list` fails

### Requirement: Source Command Integration
The source command SHALL work seamlessly in all execution modes.

#### Scenario: Source in listen mode
- **WHEN** user runs `xdebug-cli listen --commands "run" "source"`
- **THEN** executes both commands in sequence
- **AND** displays source after hitting breakpoint

#### Scenario: Source in daemon mode
- **WHEN** daemon is running with active session
- **AND** user runs `xdebug-cli attach --commands "source lib.php"`
- **THEN** retrieves and displays source
- **AND** session remains active for subsequent commands

#### Scenario: Source with JSON output mode
- **WHEN** user runs with `--json` flag
- **THEN** returns source in JSON format
- **AND** includes metadata: file URI, start line, end line
- **AND** source text included as string (preserving newlines)
