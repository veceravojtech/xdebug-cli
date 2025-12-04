## MODIFIED Requirements

### Requirement: Daemon Start Command
The CLI SHALL provide a `daemon start` command as the primary entry point for all debugging sessions.

#### Scenario: Start daemon without commands
- **WHEN** user runs `xdebug-cli daemon start --curl "http://localhost/app.php"`
- **THEN** process forks to background
- **AND** parent process exits successfully
- **AND** daemon starts DBGp server and waits for connections
- **AND** displays message: "Daemon started on port 9003"

#### Scenario: Start daemon with initial commands
- **WHEN** user runs `xdebug-cli daemon start --curl "http://localhost/app.php" --commands "break /path/file.php:100"`
- **THEN** daemon starts in background
- **AND** waits for Xdebug connection
- **AND** executes commands when connection established
- **AND** keeps session alive after commands complete

#### Scenario: Start daemon with force flag
- **WHEN** user runs `xdebug-cli daemon start --curl "http://localhost/app.php" --force`
- **AND** daemon already running on same port
- **THEN** kills existing daemon on that port
- **AND** starts new daemon successfully
- **AND** displays message: "Killed daemon on port 9003 (PID 12345)" followed by "Daemon started on port 9003"

#### Scenario: Daemon already running without force
- **WHEN** user runs `xdebug-cli daemon start --curl "http://localhost/app.php"`
- **AND** daemon already running on port 9003
- **THEN** command exits with error code 1
- **AND** displays message: "Error: daemon already running on port 9003 (PID 12345). Use 'xdebug-cli connection kill' to terminate it first or use --force to replace it."

#### Scenario: Start daemon with external connection flag
- **WHEN** user runs `xdebug-cli daemon start --enable-external-connection`
- **THEN** process forks to background
- **AND** parent process exits successfully
- **AND** daemon starts DBGp server on specified port
- **AND** waits indefinitely for external Xdebug connection
- **AND** displays message: "Daemon started on port 9003 (waiting for external connection)"

#### Scenario: External connection with initial commands
- **WHEN** user runs `xdebug-cli daemon start --enable-external-connection --commands "break /path/file.php:100"`
- **THEN** daemon starts in background
- **AND** waits for external Xdebug connection
- **AND** executes commands when connection established
- **AND** keeps session alive after commands complete

#### Scenario: Start daemon without required flags
- **WHEN** user runs `xdebug-cli daemon start` without `--curl` or `--enable-external-connection`
- **THEN** command exits with error code 1
- **AND** displays error message explaining that either `--curl` or `--enable-external-connection` is required
- **AND** shows usage examples for both options
