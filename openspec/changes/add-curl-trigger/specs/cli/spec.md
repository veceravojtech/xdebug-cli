# cli Specification Delta

## MODIFIED Requirements

### Requirement: Daemon Start Command
The CLI SHALL require a `--curl` flag on `daemon start` to trigger Xdebug connections.

#### Scenario: Start daemon with curl trigger
- **WHEN** user runs `xdebug-cli daemon start --curl "http://localhost/app.php"`
- **THEN** daemon starts DBGp server on configured port
- **AND** executes `curl http://localhost/app.php -b "XDEBUG_TRIGGER=1"` in background
- **AND** waits for Xdebug connection
- **AND** parent process exits after fork

#### Scenario: Start daemon with curl and commands
- **WHEN** user runs `xdebug-cli daemon start --curl "http://localhost/app.php" --commands "break :42"`
- **THEN** daemon starts and executes curl
- **AND** waits for Xdebug connection
- **AND** executes commands when connection established
- **AND** auto-appends "run" if breakpoint set without run command

#### Scenario: Start daemon with complex curl
- **WHEN** user runs `xdebug-cli daemon start --curl "http://localhost/api -X POST -d 'data' -H 'Content-Type: application/json'"`
- **THEN** daemon executes curl with all provided arguments
- **AND** appends `-b "XDEBUG_TRIGGER=1"` to curl command
- **AND** existing curl cookies/headers are preserved

#### Scenario: Missing curl flag shows error
- **WHEN** user runs `xdebug-cli daemon start` without `--curl` flag
- **THEN** command exits with code 1
- **AND** displays error: "Error: --curl flag is required"
- **AND** shows usage examples with --curl flag
- **AND** explains that XDEBUG_TRIGGER is added automatically

#### Scenario: Curl failure terminates daemon
- **WHEN** user runs `xdebug-cli daemon start --curl "http://invalid-host/app.php"`
- **AND** curl command fails (connection refused, DNS error, etc.)
- **THEN** daemon process terminates
- **AND** displays error with curl exit code and message
- **AND** cleans up resources (socket, registry)

#### Scenario: Curl not found
- **WHEN** user runs `xdebug-cli daemon start --curl "http://localhost/app.php"`
- **AND** curl binary is not in PATH
- **THEN** command exits with code 1
- **AND** displays error: "Error: curl not found in PATH"

## REMOVED Requirements

### Requirement: Daemon Start Command (previous scenarios)

#### Scenario: Start daemon without commands
- REMOVED: `--curl` is now required, daemon cannot start without it

#### Scenario: Daemon already running without force
- REMOVED: Auto-kill on same port is always enabled, no --force flag needed
