## ADDED Requirements

### Requirement: Breakpoint Path History
The CLI SHALL persist successfully used breakpoint paths across daemon sessions to provide suggestions when non-absolute paths fail.

#### Scenario: Store path on successful breakpoint hit
- **WHEN** breakpoint at `/var/www/app/File.php:100` is hit successfully
- **THEN** path is stored in persistent history file `~/.xdebug-cli/breakpoint-paths.json`
- **AND** mapping is: filename `File.php` -> full path `/var/www/app/File.php`

#### Scenario: Lookup path by filename
- **WHEN** path lookup is requested for filename `File.php`
- **AND** history has stored path for `File.php`
- **THEN** returns the stored full path `/var/www/app/File.php`

#### Scenario: No stored path for filename
- **WHEN** path lookup is requested for filename `Unknown.php`
- **AND** no stored path exists for that filename
- **THEN** returns empty string

#### Scenario: Path history persists across sessions
- **WHEN** daemon session ends
- **AND** new daemon session starts later
- **THEN** previous breakpoint paths are still available for suggestions

#### Scenario: Multiple paths for same filename
- **WHEN** different full paths exist for same filename (e.g., `/app1/File.php` and `/app2/File.php`)
- **THEN** most recently used path is stored and suggested
