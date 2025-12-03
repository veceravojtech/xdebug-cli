package view

import "fmt"

// ShowHelpMessage displays the main help message with all available commands.
func (v *View) ShowHelpMessage() {
	help := `
Available Commands:
  run, r          Continue execution (aliases: continue, cont)
  step, s         Step into (aliases: into, step_into)
  next, n         Step over (alias: over)
  out, o          Step out (alias: step_out)
  finish, f       Stop the debugging session
  break, b        Set a breakpoint (see 'help break' for details)
  clear           Delete breakpoint by location (GDB-style)
  print, p        Print variable value (see 'help print' for details)
  property_get    Print variable (DBGp-style: property_get -n $var)
  context, c      Show variables in current context (see 'help context' for details)
  list, l         Show source code around current line
  info, i         Show debugging information (see 'help info' for details)
  breakpoint_list List breakpoints (DBGp-style)
  status, st      Show current execution status
  detach, d       Detach from debug session
  eval, e         Evaluate PHP expression (see 'help eval' for details)
  set             Set variable value (see 'help set' for details)
  source          Display source code (see 'help source' for details)
  stack           Show call stack trace
  delete, del     Delete breakpoint by ID (alias: breakpoint_remove)
  disable         Disable breakpoint by ID (see 'help disable' for details)
  enable          Enable breakpoint by ID (see 'help enable' for details)
  help, h, ?      Show this help message or help for specific command

For detailed help on a specific command, use: xdebug-cli attach --commands "help <command>"
Examples: xdebug-cli attach --commands "help break"
`
	v.PrintLn(help)
}

// ShowInfoHelpMessage displays help for the info command.
func (v *View) ShowInfoHelpMessage() {
	help := `
info - Show debugging information

Usage:
  info breakpoints    Show all breakpoints
  info b              Short form for 'info breakpoints'

Examples:
  xdebug-cli listen --commands "info breakpoints"
  xdebug-cli listen --commands "info b"
`
	v.PrintLn(help)
}

// ShowStepHelpMessage displays help for step, next, and out commands.
func (v *View) ShowStepHelpMessage() {
	help := `
step/next/out - Control execution flow

Commands:
  step, s    Step into next statement (aliases: into, step_into)
  next, n    Step over next statement (alias: over)
  out, o     Step out of current function (alias: step_out)

Usage:
  step       Execute one statement, entering any functions
  next       Execute one statement, treating function calls as single steps
  out        Execute until current function returns

The difference:
  - 'step' enters function definitions so you can debug inside them
  - 'next' executes the entire function and stops at the next line
  - 'out' finishes the current function and stops after it returns

Examples:
  (at line 10) step     # Enter function if line 10 has a call
  (at line 10) next     # Execute line 10, don't enter functions
  (inside func) out     # Return from current function
`
	v.PrintLn(help)
}

// ShowBreakpointHelpMessage displays help for the breakpoint command.
func (v *View) ShowBreakpointHelpMessage() {
	help := `
break - Set breakpoints

Usage:
  break <line>              Set breakpoint at line in current file
  break :<line>             Set breakpoint at line in current file (explicit form)
  break <file>:<line>       Set breakpoint at line in specific file
  break call <function>     Set breakpoint on function call
  break exception           Set breakpoint on exceptions

Arguments:
  <line>         Line number (1-indexed)
  <file>         File path (absolute or relative)
  <function>     Function name

Examples:
  xdebug-cli listen --commands "break 42"                    # Break at line 42 in current file
  xdebug-cli listen --commands "break :42"                   # Same as above
  xdebug-cli listen --commands "break /path/to/file.php:15"  # Break at line 15 in specific file
  xdebug-cli listen --commands "break call myFunction"       # Break when myFunction is called
  xdebug-cli listen --commands "break exception"             # Break on any exception

Note: Breakpoints are set before execution starts or while paused.
`
	v.PrintLn(help)
}

// ShowPrintHelpMessage displays help for the print command.
func (v *View) ShowPrintHelpMessage() {
	help := `
print - Print variable values

Usage:
  print <variable>     Print the value of a variable

Arguments:
  <variable>    Variable name, can include $ prefix for PHP variables

Examples:
  xdebug-cli listen --commands "print \$myVar"
  xdebug-cli listen --commands "print myVar"         # $ is optional
  xdebug-cli listen --commands "print \$obj->prop"
  xdebug-cli listen --commands "p \$arr['key']"      # 'p' is short form

The print command displays:
  - Variable type
  - Variable value
  - For arrays/objects: nested structure with all properties

Note: Variable must be in current scope (local, global, or constant context).
`
	v.PrintLn(help)
}

// ShowContextHelpMessage displays help for the context command.
func (v *View) ShowContextHelpMessage() {
	help := `
context - Show variables in current execution context

Usage:
  context [type]     Show all variables of specified type

Arguments:
  type    Optional context type:
          - local      Local variables (default)
          - global     Global variables
          - constant   Constants

Examples:
  xdebug-cli listen --commands "context"           # Show local variables
  xdebug-cli listen --commands "context local"     # Show local variables (explicit)
  xdebug-cli listen --commands "context global"    # Show global variables
  xdebug-cli listen --commands "c constant"        # Show constants ('c' is short form)

The context command displays:
  - All variables in the specified scope
  - Variable names and types
  - Variable values
  - Nested structure for arrays and objects

Note: The context shows the state at the current breakpoint or step position.
`
	v.PrintLn(help)
}

// ShowStatusHelpMessage displays help for the status command.
func (v *View) ShowStatusHelpMessage() {
	help := `
status - Show current execution status

Usage:
  status          Show execution status

The status command displays:
  - Current file and line number
  - Execution status (running, paused, stopped)
  - Current function or method
  - Connection state

Examples:
  xdebug-cli listen --commands "status"
  xdebug-cli attach --commands "st"       # 'st' is short form
`
	v.PrintLn(help)
}

// ShowDetachHelpMessage displays help for the detach command.
func (v *View) ShowDetachHelpMessage() {
	help := `
detach - Detach from debug session

Usage:
  detach          Disconnect from the debug session

The detach command:
  - Closes the debugger connection
  - Allows the script to continue execution
  - Does not terminate the daemon (in daemon mode)
  - Useful for ending interactive debugging without stopping execution

Examples:
  xdebug-cli listen --commands "detach"
  xdebug-cli attach --commands "d"        # 'd' is short form
`
	v.PrintLn(help)
}

// ShowEvalHelpMessage displays help for the eval command.
func (v *View) ShowEvalHelpMessage() {
	help := `
eval - Evaluate PHP expression

Usage:
  eval <expression>       Evaluate a PHP expression in current context

Arguments:
  <expression>    Any valid PHP expression

Examples:
  xdebug-cli listen --commands "eval \$x + 1"
  xdebug-cli listen --commands "e 'strlen(\$name)'"
  xdebug-cli listen --commands "eval \$obj->method()"
  xdebug-cli listen --commands "eval count(\$arr)"

The eval command:
  - Executes PHP code in the current execution context
  - Returns the result of the expression
  - Has access to local and global variables
  - Useful for testing logic or complex operations

Note: Expression evaluation depends on current execution context.
`
	v.PrintLn(help)
}

// ShowSetHelpMessage displays help for the set command.
func (v *View) ShowSetHelpMessage() {
	help := `
set - Set variable value

Usage:
  set <variable> = <value>     Set a variable to a new value

Arguments:
  <variable>    Variable name (may include $ prefix)
  <value>       New value (string, number, expression)

Examples:
  xdebug-cli listen --commands "set \$x = 42"
  xdebug-cli listen --commands "set \$name = 'John'"
  xdebug-cli listen --commands "set \$arr['key'] = 'value'"
  xdebug-cli listen --commands "set \$flag = true"

The set command:
  - Modifies variable values in the current context
  - Useful for testing error conditions
  - Can change object properties
  - Changes persist during the debugging session

Note: Variable must be in current scope.
`
	v.PrintLn(help)
}

// ShowSourceHelpMessage displays help for the source command.
func (v *View) ShowSourceHelpMessage() {
	help := `
source - Display source code

Usage:
  source              Show source around current line
  source <file>       Show source from specified file
  source <file>:<line>     Show source around specific line
  source <file>:<start>-<end>     Show source range

Arguments:
  <file>         File path (absolute or relative)
  <line>         Line number
  <start>        Starting line number
  <end>          Ending line number

Examples:
  xdebug-cli listen --commands "source"                 # Around current line
  xdebug-cli listen --commands "source app.php"         # File content
  xdebug-cli listen --commands "source app.php:42"      # Around line 42
  xdebug-cli listen --commands "source app.php:40-50"   # Lines 40-50

The source command displays:
  - Source code with line numbers
  - Current execution line highlighted
  - Context for understanding code flow
`
	v.PrintLn(help)
}

// ShowStackHelpMessage displays help for the stack command.
func (v *View) ShowStackHelpMessage() {
	help := `
stack - Show call stack trace

Usage:
  stack           Show complete call stack

The stack command displays:
  - Complete function call stack
  - File and line number for each frame
  - Function/method names
  - Parameters passed to each function
  - Helps trace execution flow

Examples:
  xdebug-cli listen --commands "stack"
  xdebug-cli attach --commands "stack"

The stack shows:
  - Current function at top (Frame 0)
  - Calling function below
  - Original entry point at bottom
`
	v.PrintLn(help)
}

// ShowDeleteHelpMessage displays help for the delete command.
func (v *View) ShowDeleteHelpMessage() {
	help := `
delete - Delete breakpoint

Usage:
  delete <id>         Delete breakpoint by ID

Arguments:
  <id>            Breakpoint ID (shown in 'info breakpoints')

Examples:
  xdebug-cli listen --commands "info breakpoints"    # See breakpoint IDs
  xdebug-cli listen --commands "delete 1"            # Delete breakpoint 1
  xdebug-cli listen --commands "del 2"               # 'del' is short form

The delete command:
  - Permanently removes a breakpoint
  - Breakpoint will not trigger again
  - Use 'disable' to temporarily disable a breakpoint
  - Use 'info breakpoints' to list current breakpoints
`
	v.PrintLn(help)
}

// ShowDisableHelpMessage displays help for the disable command.
func (v *View) ShowDisableHelpMessage() {
	help := `
disable - Disable breakpoint

Usage:
  disable <id>        Disable breakpoint by ID (keeps it for later re-enabling)

Arguments:
  <id>            Breakpoint ID (shown in 'info breakpoints')

Examples:
  xdebug-cli listen --commands "info breakpoints"     # See breakpoint IDs
  xdebug-cli listen --commands "disable 1"            # Disable breakpoint 1
  xdebug-cli listen --commands "enable 1"             # Re-enable breakpoint 1

The disable command:
  - Temporarily stops a breakpoint from triggering
  - Breakpoint remains in the list
  - Use 'enable' to re-activate it
  - Use 'delete' to permanently remove it
`
	v.PrintLn(help)
}

// ShowEnableHelpMessage displays help for the enable command.
func (v *View) ShowEnableHelpMessage() {
	help := `
enable - Enable breakpoint

Usage:
  enable <id>         Enable a previously disabled breakpoint

Arguments:
  <id>            Breakpoint ID (shown in 'info breakpoints')

Examples:
  xdebug-cli attach --commands "info breakpoints"     # See breakpoint IDs
  xdebug-cli attach --commands "disable 1"            # Disable breakpoint 1
  xdebug-cli attach --commands "enable 1"             # Re-enable breakpoint 1

The enable command:
  - Re-activates a disabled breakpoint
  - Breakpoint will trigger again
  - Only works on previously disabled breakpoints
  - Use 'disable' to temporarily stop a breakpoint
`
	v.PrintLn(help)
}

// ShowClearHelpMessage displays help for the clear command.
func (v *View) ShowClearHelpMessage() {
	help := `
clear - Delete breakpoint by location (GDB-style)

Usage:
  clear :<line>           Delete breakpoint at line in current file
  clear <file>:<line>     Delete breakpoint at specific location

Arguments:
  <line>         Line number
  <file>         File path

Examples:
  xdebug-cli attach --commands "clear :42"            # Clear breakpoint at line 42
  xdebug-cli attach --commands "clear app.php:100"    # Clear breakpoint at file:line

The clear command:
  - Removes breakpoints by location instead of ID
  - Removes all breakpoints at the specified location
  - Use 'delete <id>' to remove by breakpoint ID
  - Use 'info breakpoints' to list current breakpoints
`
	v.PrintLn(help)
}

// ShowRunHelpMessage displays help for the run command.
func (v *View) ShowRunHelpMessage() {
	help := `
run - Continue execution

Usage:
  run                 Continue execution to next breakpoint

Aliases:
  r                   Short form
  continue            GDB-style alias
  cont                GDB-style alias (abbreviated)

Examples:
  xdebug-cli attach --commands "run"
  xdebug-cli attach --commands "continue"
  xdebug-cli attach --commands "r"

The run command:
  - Resumes script execution
  - Stops at the next breakpoint
  - Stops when script completes
  - Use 'step' or 'next' for line-by-line execution
`
	v.PrintLn(help)
}

// ShowCommandHelp displays help for a specific command.
// This is a convenience method that routes to the appropriate help function.
func (v *View) ShowCommandHelp(command string) {
	switch command {
	case "info", "i", "breakpoint_list":
		v.ShowInfoHelpMessage()
	case "step", "s", "next", "n", "out", "o", "into", "step_into", "over", "step_out":
		v.ShowStepHelpMessage()
	case "break", "b", "breakpoint":
		v.ShowBreakpointHelpMessage()
	case "clear":
		v.ShowClearHelpMessage()
	case "print", "p", "property_get":
		v.ShowPrintHelpMessage()
	case "context", "c":
		v.ShowContextHelpMessage()
	case "run", "r", "continue", "cont":
		v.ShowRunHelpMessage()
	case "status", "st":
		v.ShowStatusHelpMessage()
	case "detach", "d":
		v.ShowDetachHelpMessage()
	case "eval", "e":
		v.ShowEvalHelpMessage()
	case "set":
		v.ShowSetHelpMessage()
	case "source":
		v.ShowSourceHelpMessage()
	case "stack":
		v.ShowStackHelpMessage()
	case "delete", "del", "breakpoint_remove":
		v.ShowDeleteHelpMessage()
	case "disable":
		v.ShowDisableHelpMessage()
	case "enable":
		v.ShowEnableHelpMessage()
	default:
		v.PrintLn(fmt.Sprintf("No help available for command: %s", command))
		v.PrintLn("Type 'help' to see all available commands.")
	}
}
