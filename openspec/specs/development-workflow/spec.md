# development-workflow Specification

## Purpose
TBD - created by archiving change add-post-apply-install-step. Update Purpose after archive.
## Requirements
### Requirement: Post-Apply Installation
After applying an OpenSpec proposal, the CLI binary SHALL be rebuilt and installed to ensure the development environment has the latest version.

#### Scenario: Binary is updated after apply
- **WHEN** an OpenSpec proposal is applied via `/openspec:apply`
- **THEN** the `./install.sh` script is executed
- **AND** `xdebug-cli version` confirms the binary is updated

