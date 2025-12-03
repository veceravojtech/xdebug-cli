# cfg Specification

## Purpose
TBD - created by archiving change implement-xdebug-cli. Update Purpose after archive.
## Requirements
### Requirement: Version Constant
The cfg package SHALL provide a Version variable for build-time version injection.

#### Scenario: Version is overridable at build time
- **WHEN** building with `-ldflags "-X github.com/console/xdebug-cli/internal/cfg.Version=1.0.0"`
- **THEN** Version variable contains "1.0.0"
- **AND** default value is "develop"

### Requirement: CLI Parameters
The cfg package SHALL provide a CLIParameter struct for command-line arguments.

#### Scenario: CLIParameter contains connection settings
- **WHEN** CLIParameter is initialized
- **THEN** it contains Host (string), Port (uint16), Trigger (string), and Version (string) fields
- **AND** all fields are exported (public)

