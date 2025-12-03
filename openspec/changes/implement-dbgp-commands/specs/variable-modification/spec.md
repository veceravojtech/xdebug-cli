# Spec Delta: Variable Modification

## ADDED Requirements

### Requirement: Variable Value Modification
The CLI SHALL allow modifying variable values during debugging sessions to test different execution paths.

#### Scenario: Modify integer to change control flow
- **WHEN** user pauses at conditional statement `if ($count > 10)`
- **AND** $count currently equals 5
- **AND** runs `set $count = 15`
- **AND** continues execution with `next`
- **THEN** execution follows the true branch
- **AND** demonstrates changed behavior from modified value

#### Scenario: Modify string for testing
- **WHEN** variable $username = "admin"
- **AND** user runs `set $username = "guest"`
- **AND** prints value with `print $username`
- **THEN** displays "guest"

#### Scenario: Type detection for integers
- **WHEN** user runs `set $x = 42`
- **THEN** CLI detects type as "int"
- **AND** sends property_set with -t int

#### Scenario: Type detection for strings
- **WHEN** user runs `set $name = "John"`
- **THEN** CLI detects type as "string"
- **AND** sends property_set with -t string

#### Scenario: Type detection for booleans
- **WHEN** user runs `set $flag = true`
- **THEN** CLI detects type as "bool"
- **AND** converts to appropriate boolean representation

#### Scenario: Type detection for floats
- **WHEN** user runs `set $price = 19.99`
- **THEN** CLI detects type as "float"
- **AND** sends property_set with -t float

#### Scenario: Parse assignment with spaces
- **WHEN** user runs variations:
  - `set $x=42` (no spaces)
  - `set $x = 42` (spaces around =)
  - `set $x= 42` (space after =)
  - `set $x =42` (space before =)
- **THEN** all parse successfully
- **AND** all set $x to 42

#### Scenario: Validate variable name format
- **WHEN** user runs `set count = 10` (missing $)
- **THEN** displays error about invalid variable name
- **AND** suggests correct format: `set $count = 10`

#### Scenario: Handle assignment without equals sign
- **WHEN** user runs `set $x 42`
- **THEN** displays usage error
- **AND** shows example: `set $variable = value`

#### Scenario: Modify array elements (future consideration)
- **WHEN** user runs `set $array[0] = "new"`
- **THEN** (implementation dependent on DBGp support)
- **NOTE**: May require complex property name syntax

#### Scenario: Modify object properties (future consideration)
- **WHEN** user runs `set $user->name = "Bob"`
- **THEN** (implementation dependent on DBGp support)
- **NOTE**: May require complex property name syntax
