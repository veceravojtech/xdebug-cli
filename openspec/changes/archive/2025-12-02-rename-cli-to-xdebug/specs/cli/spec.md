# CLI Capability - Rename Delta

## MODIFIED Requirements

### Requirement: CLI Project Structure
The project SHALL follow Go CLI conventions with proper module naming.

#### Scenario: Go module is properly named
- **WHEN** the project is initialized
- **THEN** `go.mod` declares module `github.com/console/xdebug-cli`
- **AND** it contains `cmd/xdebug-cli/main.go` as entry point
- **AND** CLI commands are in `internal/cli/` package

### Requirement: Preview Command
The CLI SHALL provide a preview command with animated progress indicator.

#### Scenario: Preview command shows animation
- **WHEN** user runs `xdebug-cli preview source 10s`
- **THEN** an animated progress indicator is displayed for 10 seconds
- **AND** the animation shows the source name and elapsed time

#### Scenario: Preview command requires arguments
- **WHEN** user runs `xdebug-cli preview`
- **THEN** an error message is displayed indicating source and duration are required

### Requirement: Install Command
The CLI SHALL provide an install command to install itself to `~/.local/bin`.

#### Scenario: Install command copies binary
- **WHEN** user runs `xdebug-cli install`
- **THEN** the CLI binary is built and copied to `~/.local/bin/xdebug-cli`
- **AND** the binary is made executable

#### Scenario: Install command creates directory if needed
- **WHEN** user runs `xdebug-cli install` and `~/.local/bin` does not exist
- **THEN** the directory is created before copying

### Requirement: Version Command
The CLI SHALL provide version information including build timestamp.

#### Scenario: Version flag shows build info
- **WHEN** user runs `xdebug-cli --version` or `xdebug-cli version`
- **THEN** the version number and build timestamp are displayed

### Requirement: Configuration File
The CLI SHALL support configuration via YAML file.

#### Scenario: Config file naming
- **WHEN** the CLI looks for configuration
- **THEN** it reads from `~/.xdebug-cli.yaml`
- **AND** an example file `.xdebug-cli.yaml.example` is provided in the repository
