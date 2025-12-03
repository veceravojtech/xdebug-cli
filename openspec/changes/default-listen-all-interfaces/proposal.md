# Proposal: Default Listen All Interfaces

## Summary
Change the default listen host from `127.0.0.1` to `0.0.0.0` so the CLI listens on all network interfaces by default. This enables remote debugging scenarios (e.g., Docker, VMs, remote servers) without requiring explicit `--host` configuration.

## Motivation
The current default of `127.0.0.1` limits the CLI to accepting connections only from localhost. In common development scenarios like Docker containers or remote servers, Xdebug needs to connect from a different network interface. Requiring users to always specify `--host 0.0.0.0` adds friction to these workflows.

## Scope
- **In Scope**: Change default value for `--host` / `-l` flag from `127.0.0.1` to `0.0.0.0`
- **Out of Scope**: Removing the `--host` flag entirely, adding new security features

## Impact Analysis
- **Breaking Change**: No - users who explicitly set `--host` are unaffected
- **Behavioral Change**: Yes - CLI now listens on all interfaces by default
- **Security Consideration**: Users should be aware the CLI is network-accessible; this matches common development patterns

## Files Affected
- `internal/cli/root.go` - Change default value in flag definition
- `openspec/specs/cli/spec.md` - Update scenario examples
- `CLAUDE.md` - Update documentation to reflect new default
- `README.md` - Update documentation

## Alternatives Considered
1. **Keep 127.0.0.1 default**: Rejected - creates friction for common Docker/VM workflows
2. **Auto-detect Docker environment**: Rejected - over-engineered for this use case
3. **Remove --host flag entirely**: Rejected - users may still need to bind to specific interface

## Decision
Proceed with changing default to `0.0.0.0` as it aligns with typical Xdebug development patterns.
