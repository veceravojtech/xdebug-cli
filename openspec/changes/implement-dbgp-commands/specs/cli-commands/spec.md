# Spec Delta: CLI Commands

## ADDED Requirements

### Requirement: Status Command
The CLI SHALL provide a `status` command to query debugger execution state.

#### Scenario: Query current execution status
- **WHEN** user runs `xdebug-cli attach --commands "status"`
- **AND** debugger is at a breakpoint
- **THEN** displays current status (starting, running, break, stopping, stopped)
- **AND** displays current file and line number if available
- **AND** displays reason (ok, error, aborted, exception)

#### Scenario: Status with alias
- **WHEN** user runs `xdebug-cli attach --commands "st"`
- **THEN** executes status command
- **AND** displays same output as `status`

#### Scenario: Status command JSON output
- **WHEN** user runs `xdebug-cli attach --json --commands "status"`
- **THEN** outputs JSON with structure: `{"command":"status","success":true,"result":{"status":"break","reason":"ok","filename":"/path/file.php","lineno":42}}`

### Requirement: Detach Command
The CLI SHALL provide a `detach` command to stop debugging without terminating the script.

#### Scenario: Detach from debug session
- **WHEN** user runs `xdebug-cli attach --commands "detach"`
- **AND** debugger is active
- **THEN** sends detach command to debugger
- **AND** script continues execution without debugging
- **AND** session ends gracefully

#### Scenario: Detach with alias
- **WHEN** user runs `xdebug-cli attach --commands "d"`
- **THEN** executes detach command

#### Scenario: Detach command JSON output
- **WHEN** user runs `xdebug-cli attach --json --commands "detach"`
- **THEN** outputs JSON with structure: `{"command":"detach","success":true}`

### Requirement: Eval Command
The CLI SHALL provide an `eval` command to execute arbitrary PHP expressions during debugging.

#### Scenario: Evaluate simple expression
- **WHEN** user runs `xdebug-cli attach --commands "eval \$x + 10"`
- **AND** variable $x exists with value 32
- **THEN** evaluates expression
- **AND** displays result: "42"
- **AND** displays result type: "int"

#### Scenario: Evaluate complex expression
- **WHEN** user runs `xdebug-cli attach --commands "eval \$user->getName()"`
- **THEN** calls method and displays return value

#### Scenario: Eval with syntax error
- **WHEN** user runs `xdebug-cli attach --commands "eval invalid syntax"`
- **THEN** displays error message about syntax error
- **AND** exits with error code

#### Scenario: Eval with alias
- **WHEN** user runs `xdebug-cli attach --commands "e \$count"`
- **THEN** executes eval command

#### Scenario: Eval command JSON output
- **WHEN** user runs `xdebug-cli attach --json --commands "eval \$x"`
- **THEN** outputs JSON with structure: `{"command":"eval","success":true,"result":{"expression":"$x","type":"int","value":"42"}}`

### Requirement: Stack Command
The CLI SHALL provide a `stack` command as a standalone alternative to `info stack`.

#### Scenario: Display call stack
- **WHEN** user runs `xdebug-cli attach --commands "stack"`
- **AND** execution is paused
- **THEN** displays full call stack
- **AND** shows function names, file paths, and line numbers
- **AND** indicates current stack frame

#### Scenario: Stack command JSON output
- **WHEN** user runs `xdebug-cli attach --json --commands "stack"`
- **THEN** outputs JSON array of stack frames
- **AND** each frame contains: function, file, line, depth

### Requirement: Set Command
The CLI SHALL provide a `set` command to modify variable values during debugging.

#### Scenario: Set integer variable
- **WHEN** user runs `xdebug-cli attach --commands "set \$count = 42"`
- **AND** variable $count exists
- **THEN** updates variable to value 42
- **AND** displays confirmation

#### Scenario: Set string variable
- **WHEN** user runs `xdebug-cli attach --commands "set \$name = \"John\""`
- **THEN** updates variable to string "John"

#### Scenario: Set variable that doesn't exist
- **WHEN** user runs `xdebug-cli attach --commands "set \$nonexistent = 1"`
- **THEN** displays error message
- **AND** exits with error code

#### Scenario: Set with invalid syntax
- **WHEN** user runs `xdebug-cli attach --commands "set invalid"`
- **THEN** displays usage help
- **AND** shows example: `set $variable = value`

#### Scenario: Set command JSON output
- **WHEN** user runs `xdebug-cli attach --json --commands "set \$x = 100"`
- **THEN** outputs JSON with structure: `{"command":"set","success":true,"result":{"variable":"$x","value":"100","type":"int"}}`

### Requirement: Source Command
The CLI SHALL provide a `source` command to display source code from the debugger.

#### Scenario: Display current file source
- **WHEN** user runs `xdebug-cli attach --commands "source"`
- **AND** execution is paused at app.php:100
- **THEN** displays source code for app.php
- **AND** shows line numbers

#### Scenario: Display specific file source
- **WHEN** user runs `xdebug-cli attach --commands "source lib/helper.php"`
- **THEN** displays source code for lib/helper.php

#### Scenario: Display source with line range
- **WHEN** user runs `xdebug-cli attach --commands "source app.php:100-120"`
- **THEN** displays lines 100-120 of app.php
- **AND** shows line numbers

#### Scenario: Display current file with line range
- **WHEN** user runs `xdebug-cli attach --commands "source :50-60"`
- **THEN** displays lines 50-60 of current file

#### Scenario: Source with alias
- **WHEN** user runs `xdebug-cli attach --commands "src app.php"`
- **THEN** executes source command

#### Scenario: Source command JSON output
- **WHEN** user runs `xdebug-cli attach --json --commands "source app.php:10-20"`
- **THEN** outputs JSON with structure: `{"command":"source","success":true,"result":{"file":"file:///path/app.php","start_line":10,"end_line":20,"source":"<?php..."}}`

### Requirement: Delete Breakpoint Command
The CLI SHALL provide a `delete` command to remove breakpoints.

#### Scenario: Delete breakpoint by ID
- **WHEN** user runs `xdebug-cli attach --commands "delete 123"`
- **AND** breakpoint 123 exists
- **THEN** removes breakpoint
- **AND** displays confirmation

#### Scenario: Delete with alias
- **WHEN** user runs `xdebug-cli attach --commands "del 456"`
- **THEN** executes delete command

#### Scenario: Delete non-existent breakpoint
- **WHEN** user runs `xdebug-cli attach --commands "delete 999"`
- **THEN** displays error message
- **AND** indicates breakpoint not found

#### Scenario: Delete without ID
- **WHEN** user runs `xdebug-cli attach --commands "delete"`
- **THEN** displays usage help
- **AND** shows example: `delete <breakpoint_id>`

#### Scenario: Delete command JSON output
- **WHEN** user runs `xdebug-cli attach --json --commands "delete 123"`
- **THEN** outputs JSON with structure: `{"command":"delete","success":true,"result":{"breakpoint_id":"123"}}`

### Requirement: Disable Breakpoint Command
The CLI SHALL provide a `disable` command to temporarily disable breakpoints without removing them.

#### Scenario: Disable breakpoint by ID
- **WHEN** user runs `xdebug-cli attach --commands "disable 123"`
- **AND** breakpoint 123 exists and is enabled
- **THEN** disables breakpoint
- **AND** displays confirmation
- **AND** breakpoint remains in breakpoint list but inactive

#### Scenario: Disable already disabled breakpoint
- **WHEN** user runs `xdebug-cli attach --commands "disable 123"`
- **AND** breakpoint 123 is already disabled
- **THEN** displays message indicating already disabled

#### Scenario: Disable without ID
- **WHEN** user runs `xdebug-cli attach --commands "disable"`
- **THEN** displays usage help

#### Scenario: Disable command JSON output
- **WHEN** user runs `xdebug-cli attach --json --commands "disable 123"`
- **THEN** outputs JSON with structure: `{"command":"disable","success":true,"result":{"breakpoint_id":"123","state":"disabled"}}`

### Requirement: Enable Breakpoint Command
The CLI SHALL provide an `enable` command to re-enable disabled breakpoints.

#### Scenario: Enable disabled breakpoint
- **WHEN** user runs `xdebug-cli attach --commands "enable 123"`
- **AND** breakpoint 123 exists and is disabled
- **THEN** enables breakpoint
- **AND** displays confirmation

#### Scenario: Enable already enabled breakpoint
- **WHEN** user runs `xdebug-cli attach --commands "enable 123"`
- **AND** breakpoint 123 is already enabled
- **THEN** displays message indicating already enabled

#### Scenario: Enable without ID
- **WHEN** user runs `xdebug-cli attach --commands "enable"`
- **THEN** displays usage help

#### Scenario: Enable command JSON output
- **WHEN** user runs `xdebug-cli attach --json --commands "enable 123"`
- **THEN** outputs JSON with structure: `{"command":"enable","success":true,"result":{"breakpoint_id":"123","state":"enabled"}}`
