# Project Context

## Purpose
OpenSpec template for building Go CLI applications. This template provides a standardized structure and conventions for creating command-line tools that interact with external services and APIs.

When a new CLI is needed, this template is used with OpenSpec to scaffold the project with proper structure, dependencies, and conventions already in place.

## Tech Stack
- **Language**: Go 1.25+
- **CLI Framework**: [spf13/cobra](https://github.com/spf13/cobra) - Command-line interface
- **Configuration**: [spf13/viper](https://github.com/spf13/viper) - Configuration management
- **Config Format**: YAML (`.{cli-name}.yaml`)

## Project Conventions

### Project Structure
```
cmd/{cli-name}/main.go     # Entry point - minimal, delegates to internal/cli
internal/
  cli/                     # CLI commands
    root.go                # Root command setup with persistent flags
    {feature}.go           # Feature-specific subcommands
  {service}/               # Business logic for external service
    client.go              # HTTP client with auth and request handling
    types.go               # Data structures and API response types
    {entity}.go            # Entity-specific operations
    {entity}_test.go       # Unit tests
docs/
  plans/                   # Implementation plans
go.mod
go.sum
.{cli-name}.yaml.example   # Example configuration
.gitignore
install.sh                 # Installation script
CLAUDE.md                  # AI assistant instructions
```

### Code Style
- Follow Go best practices and idioms
- **No spaghetti code** - well-structured, readable code only
- **Max file length**: 500 lines - split by logical meaning if longer
- Use meaningful names for packages, types, and functions
- Wrap errors with context: `fmt.Errorf("doing X: %w", err)`
- Close resources properly with `defer`

### Architecture Patterns
- **Separation of concerns**: CLI layer (`internal/cli`) vs business logic (`internal/{service}`)
- **Client pattern**: Dedicated HTTP client struct with auth and helper methods
- **Minimal main.go**: Entry point only calls `cli.Execute()`
- **Persistent flags**: Common flags (config, verbose) defined on root command
- **JSON output**: Support JSON output with `--json` flag for scripting

### Testing Strategy
- Unit tests in `_test.go` files alongside source
- Use table-driven tests for multiple scenarios
- Mock HTTP responses for API client tests
- Run tests: `go test ./...`

### Git Workflow
- **Branching**: Feature branches from main (format: `{issue-number}-{description}`)
- **Conventional commits**:
  - `feat:` - New features
  - `fix:` - Bug fixes
  - `docs:` - Documentation
  - `refactor:` - Code refactoring
  - `test:` - Test additions/changes

## Domain Context
- CLIs interact with REST APIs using JSON
- Configuration stored in home directory as hidden YAML file
- API keys and URLs are required configuration
- Common operations: list/filter, CRUD, reports, automation

## Important Constraints
- Must build as standalone binary (no runtime dependencies)
- Configuration must never be committed (use .example files)
- API keys stored in config, never hardcoded
- Support both human-readable and JSON output formats

## External Dependencies
- Each CLI targets a specific external service API
- Document API endpoints and authentication in CLAUDE.md
- Link to official API documentation when available

## Development Workflow
- **Before development**: `/superpowers:brainstorm` to refine the feature design
- **Planning**: `/superpowers:write-plan` to create detailed implementation plan
- **Execution**: `/superpowers:execute-plan` to implement the plan
