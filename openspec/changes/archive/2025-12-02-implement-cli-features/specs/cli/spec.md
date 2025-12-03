## ADDED Requirements

### Requirement: Project Structure
The CLI SHALL be a buildable Go project following the standard template structure.

#### Scenario: Project builds successfully
- **WHEN** running `go build ./...` in the project root
- **THEN** the project compiles without errors

#### Scenario: Project structure follows conventions
- **WHEN** examining the project layout
- **THEN** it contains `cmd/source-cli/main.go` as entry point
- **AND** `internal/cli/` for command implementations
- **AND** `go.mod` with module definition

### Requirement: Preview Command
The CLI SHALL provide a `preview` command that displays an animated progress indicator.

#### Scenario: Preview with duration
- **WHEN** user runs `source-cli preview source 10s`
- **THEN** an animated loading indicator is displayed for 10 seconds
- **AND** the animation runs until the duration expires

#### Scenario: Preview without arguments
- **WHEN** user runs `source-cli preview`
- **THEN** help text is displayed explaining the command usage

### Requirement: Install Command
The CLI SHALL provide an `install` command that installs the binary to the user's local bin directory.

#### Scenario: Install to local bin
- **WHEN** user runs `source-cli install`
- **THEN** the CLI binary is built and copied to `~/.local/bin/source-cli`
- **AND** the binary includes a build timestamp

#### Scenario: Install creates directory if missing
- **WHEN** user runs `source-cli install` and `~/.local/bin` does not exist
- **THEN** the directory is created
- **AND** the binary is installed successfully

### Requirement: TDD Best Practices
The CLI SHALL follow test-driven development best practices for Go.

#### Scenario: Test files exist alongside source
- **WHEN** examining any `.go` source file in `internal/`
- **THEN** a corresponding `_test.go` file exists with unit tests

#### Scenario: Tests pass
- **WHEN** running `go test ./...`
- **THEN** all tests pass successfully
