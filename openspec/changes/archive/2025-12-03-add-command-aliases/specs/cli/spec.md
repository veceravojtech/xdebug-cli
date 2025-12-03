## ADDED Requirements

### Requirement: Command Aliases for Multiple Debugger Conventions
The CLI SHALL support command aliases from multiple debugger conventions (GDB, DBGp protocol, VS Code) to improve usability for users from different debugging backgrounds.

#### Scenario: GDB-style continue command
- **WHEN** user executes `xdebug-cli attach --commands "continue"`
- **THEN** execution continues to next breakpoint (same as `run`)
- **AND** returns current execution state

#### Scenario: Short continue alias
- **WHEN** user executes `xdebug-cli attach --commands "cont"` or `--commands "c"`
- **THEN** execution continues (same as `continue` and `run`)

#### Scenario: Step into with alternative names
- **WHEN** user executes `--commands "into"` or `--commands "step_into"`
- **THEN** steps into next function call (same as `step`)

#### Scenario: Step over with alternative name
- **WHEN** user executes `--commands "over"`
- **THEN** steps over next line without entering functions (same as `next`)

#### Scenario: Step out with alternative name
- **WHEN** user executes `--commands "step_out"`
- **THEN** steps out of current function (same as `out`)

#### Scenario: DBGp protocol breakpoint list
- **WHEN** user executes `--commands "breakpoint_list"`
- **THEN** displays all active breakpoints (same as `info breakpoints`)

#### Scenario: DBGp protocol breakpoint remove
- **WHEN** user executes `--commands "breakpoint_remove 1"`
- **THEN** removes breakpoint with ID 1 (same as `delete 1`)

#### Scenario: DBGp protocol property get
- **WHEN** user executes `--commands "property_get -n myVar"`
- **THEN** displays variable value (same as `print myVar`)
- **AND** supports both `$myVar` and `myVar` syntax

#### Scenario: Property get without flag error
- **WHEN** user executes `--commands "property_get myVar"` without `-n` flag
- **THEN** returns error: "Usage: property_get -n <variable>"

#### Scenario: GDB-style clear by line
- **WHEN** user executes `--commands "clear :42"`
- **AND** breakpoint exists at line 42 in current file
- **THEN** removes the breakpoint
- **AND** returns success with message "Removed 1 breakpoint(s) at :42"

#### Scenario: GDB-style clear by file and line
- **WHEN** user executes `--commands "clear app.php:100"`
- **AND** breakpoint exists at app.php:100
- **THEN** removes the breakpoint
- **AND** returns success message

#### Scenario: Clear with no breakpoint at location
- **WHEN** user executes `--commands "clear :50"`
- **AND** no breakpoint exists at line 50
- **THEN** returns error: "No breakpoint at location :50"

#### Scenario: Clear removes multiple breakpoints at same location
- **WHEN** multiple breakpoints exist at same file:line
- **AND** user executes `--commands "clear file.php:42"`
- **THEN** removes all breakpoints at that location
- **AND** returns count of removed breakpoints

#### Scenario: Aliases work with JSON output mode
- **WHEN** user executes `--json --commands "continue"`
- **THEN** returns JSON output with same structure as `run` command
- **AND** command field in JSON shows "continue"

#### Scenario: Help text shows aliases
- **WHEN** user executes `--commands "help"`
- **THEN** displays commands with their aliases
- **AND** shows "run, r, continue, c" on same line
- **AND** shows "step, s, into, step_into" on same line
- **AND** shows other command groups with their aliases
