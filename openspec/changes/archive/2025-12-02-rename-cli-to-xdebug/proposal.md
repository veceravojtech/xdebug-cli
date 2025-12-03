# Change: Rename CLI from source-cli to xdebug-cli

## Why
The CLI is currently named `source-cli` which is a generic placeholder name. The project should be named `xdebug-cli` to reflect its actual purpose and match the project directory name.

## What Changes
- Rename Go module from `github.com/console/source-cli` to `github.com/console/xdebug-cli`
- Rename `cmd/source-cli/` directory to `cmd/xdebug-cli/`
- Rename binary output from `source-cli` to `xdebug-cli`
- Rename config file from `.source-cli.yaml` to `.xdebug-cli.yaml`
- Update all references in documentation (README.md, CLAUDE.md)
- Update all internal package imports
- Update install script references
- Update .gitignore entries

## Impact
- Affected specs: cli (command names and binary references)
- Affected code: All files that reference `source-cli`
- Affected files:
  - `go.mod` - module name
  - `cmd/source-cli/main.go` - directory rename + import path
  - `internal/cli/root.go` - Use field and descriptions
  - `internal/cli/preview.go` - import path + examples
  - `internal/cli/install.go` - descriptions and output paths
  - `internal/cli/install_test.go` - expected paths
  - `install.sh` - binary name and build paths
  - `.gitignore` - binary and config file patterns
  - `.source-cli.yaml.example` - rename to `.xdebug-cli.yaml.example`
  - `README.md` - all documentation
  - `CLAUDE.md` - project documentation
  - `openspec/changes/implement-cli-features/*` - historical references (archive)
