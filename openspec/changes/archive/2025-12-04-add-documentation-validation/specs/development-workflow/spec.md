## ADDED Requirements
### Requirement: Documentation Validation
After applying or archiving an OpenSpec proposal, both CLAUDE.md and README.md SHALL be reviewed and updated to reflect any user-facing changes.

#### Scenario: New command added
- **WHEN** a proposal adds a new CLI command or subcommand
- **THEN** CLAUDE.md "Available Commands" section is updated with the new command
- **AND** README.md "Usage" and "Debugging Commands" sections are updated
- **AND** usage examples are added if applicable

#### Scenario: Command behavior changed
- **WHEN** a proposal modifies existing command behavior, flags, or output
- **THEN** CLAUDE.md and README.md examples are verified to still be accurate
- **AND** any changed flag names or defaults are updated in both files
- **AND** error message examples are updated if changed

#### Scenario: Command removed or deprecated
- **WHEN** a proposal removes or deprecates a command
- **THEN** CLAUDE.md and README.md are updated to remove references to the command
- **AND** any workflows using the command are updated or removed

#### Scenario: New workflow or pattern added
- **WHEN** a proposal introduces a new debugging workflow or usage pattern
- **THEN** a corresponding workflow example is added to CLAUDE.md
- **AND** README.md Quick Start or Usage section is updated if relevant
- **AND** shell escaping examples are updated if new special characters are involved

#### Scenario: Spec-to-documentation mapping
- **WHEN** reviewing documentation for accuracy
- **THEN** CLAUDE.md sections correspond to these specs:
  - "Available Commands" -> cli spec (Requirement: Daemon Subcommands, Attach Command, etc.)
  - "Debugging Commands" -> dbgp spec (command parsing and execution)
  - "Daemon Workflow" -> persistent-sessions spec (Requirement: Daemon Mode)
  - "Managing Daemon Sessions" -> persistent-sessions spec (Requirement: Session Registry)
  - "Error Messages" -> cli spec (error scenarios)
  - "Troubleshooting" -> cli spec (Requirement: Timeout Exit Feedback)
- **AND** README.md sections correspond to these specs:
  - "Usage" -> cli spec (all commands)
  - "Debugging Commands" -> dbgp spec (command table)
  - "Exit Codes" -> cli spec (error handling)
