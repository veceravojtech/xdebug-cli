# Change: Implement CLI Features

## Why

The xdebug-cli project currently has no Go code - only OpenSpec structure. The CLAUDE.md defines the expected features that need to be implemented to create a functional CLI tool based on the gitlab-cli template.

## What Changes

- **NEW**: Scaffold Go project structure (cmd/, internal/cli/, go.mod)
- **NEW**: Implement `source-cli preview` command with animated progress indicator
- **NEW**: Implement `source-cli install` command to install binary to `~/.local/bin` with build timestamp
- **NEW**: Add install.sh script for distribution
- **NEW**: Set up TDD with test files alongside source

## Impact

- Affected specs: `cli` (new capability)
- Affected code: All files (greenfield implementation)
- Source reference: `/home/console/PhpstormProjects/CLI/gitlab-cli` (template to follow)
