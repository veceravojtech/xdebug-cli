## ADDED Requirements

### Requirement: Timeout Exit Feedback
The CLI SHALL provide clear feedback when breakpoint validation times out.

#### Scenario: Warning message on timeout
- **WHEN** breakpoint validation timeout occurs
- **THEN** prints warning to stderr: "Warning: breakpoint not hit within Xs"
- **AND** includes list of pending breakpoints that were not hit
- **AND** suggests increasing timeout with `--breakpoint-timeout` or `--wait-forever`

#### Scenario: Distinct exit code for timeout
- **WHEN** breakpoint validation timeout occurs
- **THEN** exits with code 124 (Unix timeout convention)
- **AND** exit code is distinct from general errors (code 1)

#### Scenario: Log file for timeout events
- **WHEN** breakpoint validation timeout occurs
- **THEN** writes event to `/tmp/xdebug-cli-daemon-<port>.log`
- **AND** includes timestamp, timeout duration, and breakpoint details
- **AND** log file aids post-mortem debugging

#### Scenario: Normal exit when breakpoint hit
- **WHEN** breakpoint is hit within timeout period
- **THEN** no timeout warning is printed
- **AND** daemon continues running normally
