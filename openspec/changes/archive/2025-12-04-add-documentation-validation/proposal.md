# Change: Add Documentation Validation Workflow

## Why
CLAUDE.md serves as the primary user-facing documentation for xdebug-cli, containing command syntax, examples, and workflows. When specs are added or modified via OpenSpec proposals, the documentation can drift out of sync, leading to outdated examples, missing commands, or incorrect usage patterns.

## What Changes
- Add a manual validation checklist requirement to the development-workflow spec
- Require CLAUDE.md review during `/openspec:apply` and archiving stages
- Document which CLAUDE.md sections correspond to which specs

## Impact
- Affected specs: development-workflow
- Affected code: None (process change only)
- No breaking changes
