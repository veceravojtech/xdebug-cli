# View Package

The `view` package provides the terminal UI layer for the Xdebug CLI debugger. It handles all user interaction including input, output formatting, and display of debugging information.

## Components

### view.go - Core View
The main View struct provides basic terminal I/O operations:
- `NewView()` - Creates a new view instance with stdin reader and source cache
- `Print(s)`, `PrintLn(s)` - Output text to stdout
- `PrintErrorLn(s)` - Output errors to stderr
- `PrintInputPrefix()` - Display the "(xdbg) " command prompt
- `GetInputLine()` - Read and parse user input
- `PrintApplicationInformation(version, host, port)` - Display startup banner

### source.go - Source Code Display
Handles caching and displaying PHP source files:
- `SourceFileCache` - Caches file contents for efficient repeated access
- `PrintSourceLn(fileURI, line, length)` - Display source code with line numbers, highlighting the current line
- `PrintSourceChangeLn(filename)` - Notify when execution moves to a new file

Features:
- Automatic file:// URI to path conversion
- Configurable context lines (before/after current line)
- Current line marked with ">" indicator
- Line number formatting

### help.go - Help Messages
Comprehensive help system for all debugger commands:
- `ShowHelpMessage()` - Main help menu with all commands
- `ShowInfoHelpMessage()` - Help for info command
- `ShowStepHelpMessage()` - Help for step/next commands
- `ShowBreakpointHelpMessage()` - Help for breakpoint command
- `ShowPrintHelpMessage()` - Help for print command
- `ShowContextHelpMessage()` - Help for context command
- `ShowCommandHelp(command)` - Route to specific command help

### display.go - Formatted Display
Rich formatting for debugging information:
- `ShowInfoBreakpoints([]ProtocolBreakpoint)` - Formatted table of breakpoints
- `PrintPropertyListWithDetails(scope, []ProtocolProperty)` - Display variables with nested structure
- `PrintProperty(ProtocolProperty)` - Display single property tree

Features:
- Automatic path truncation for long file names
- Recursive property display for arrays/objects
- Value truncation for long strings
- Proper indentation for nested structures

### types.go - Protocol Types
Placeholder interfaces for DBGp protocol structures:
- `ProtocolBreakpoint` - Interface for breakpoint information
- `ProtocolProperty` - Interface for variable/property information

**Note:** These are temporary placeholder interfaces. They will be replaced with concrete implementations from `internal/dbgp/protocol.go` when Phase 2 (DBGp Protocol Layer) is completed.

## Usage Example

```go
import "github.com/console/xdebug-cli/internal/view"

// Create view instance
v := view.NewView()

// Display application banner
v.PrintApplicationInformation("1.0.0", "localhost", 9003)

// REPL loop
for {
    v.PrintInputPrefix()
    input, err := v.GetInputLine()
    if err != nil {
        v.PrintErrorLn("Error reading input")
        break
    }

    if input == "help" {
        v.ShowHelpMessage()
    } else if input == "quit" {
        break
    }
    // Handle other commands...
}

// Display source code around line 42
v.PrintSourceLn("file:///path/to/script.php", 42, 5)

// Display breakpoints
breakpoints := []view.ProtocolBreakpoint{...}
v.ShowInfoBreakpoints(breakpoints)

// Display variables
properties := []view.ProtocolProperty{...}
v.PrintPropertyListWithDetails("Local", properties)
```

## Testing

All components have comprehensive unit tests:
- `view_test.go` - Tests for core I/O operations
- `source_test.go` - Tests for source file caching and display
- `help_test.go` - Tests for help message content
- `display_test.go` - Tests for formatting breakpoints and properties

Run tests with:
```bash
go test ./internal/view/...
```

## Integration with Other Packages

The view package is designed to integrate with:
- `internal/dbgp` - Receives ProtocolBreakpoint and ProtocolProperty types
- `internal/cli` - Called by CLI commands (especially listen command REPL)
- `internal/cfg` - Receives version information for display

## Future Enhancements

Potential improvements for future phases:
- Color output support (ANSI codes)
- Configurable prompt
- Watch expression display
- Call stack formatting
- Conditional breakpoint display
- Pagination for long output
