## ADDED Requirements

### Requirement: Breakpoint Path Validation with Fail-Fast
The CLI SHALL validate breakpoints with non-absolute paths by waiting for the first hit, and terminate the daemon if the breakpoint is not triggered within a timeout period.

#### Scenario: Absolute path breakpoint - normal operation
- **WHEN** user runs `daemon start --curl "..." --commands "break /var/www/app/File.php:100"`
- **THEN** daemon starts and sets breakpoint
- **AND** curl triggers request
- **AND** when breakpoint hits, daemon continues normally
- **AND** the full path is stored for future suggestions

#### Scenario: Non-absolute path breakpoint hits - success
- **WHEN** user runs `daemon start --curl "..." --commands "break File.php:100"`
- **AND** breakpoint path resolves correctly and is hit within timeout
- **THEN** warning is shown: "Warning: Breakpoint path 'File.php' is not absolute"
- **AND** daemon continues normally after breakpoint hit
- **AND** resolved full path is stored for future suggestions

#### Scenario: Non-absolute path breakpoint not hit - fail fast
- **WHEN** user runs `daemon start --curl "..." --commands "break File.php:100"`
- **AND** breakpoint is not hit within timeout (default 10s)
- **THEN** daemon is terminated
- **AND** error is displayed: "Error: Breakpoint at 'File.php:100' was not hit within 10s"
- **AND** if known path exists, shows: "Use full path: /var/www/app/File.php:100"
- **AND** exit code is non-zero

#### Scenario: Non-absolute path with known suggestion
- **WHEN** user previously used path `/var/www/app/controllers/PriceLoader.php`
- **AND** user runs `daemon start --curl "..." --commands "break PriceLoader.php:369"`
- **AND** breakpoint not hit within timeout
- **THEN** error includes suggestion: "Use full path: /var/www/app/controllers/PriceLoader.php:369"

#### Scenario: Non-absolute path with no known suggestion
- **WHEN** no previous path matches filename `NewFile.php`
- **AND** user runs `daemon start --curl "..." --commands "break NewFile.php:50"`
- **AND** breakpoint not hit within timeout
- **THEN** error shows: "Error: Breakpoint at 'NewFile.php:50' was not hit. Ensure you use an absolute path."

#### Scenario: Current file line syntax - no validation needed
- **WHEN** user runs with `--commands "break :100"`
- **THEN** breakpoint uses current file from session (absolute path)
- **AND** no timeout validation is applied

#### Scenario: Custom timeout flag
- **WHEN** user runs `daemon start --breakpoint-timeout 30s --curl "..." --commands "break File.php:100"`
- **THEN** daemon waits up to 30 seconds for breakpoint hit before failing
