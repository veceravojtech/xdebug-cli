package cfg

// Version is the current version of the xdebug-cli application
const Version = "1.0.1"

// CLIParameter contains the command-line parameters for the Xdebug CLI
type CLIParameter struct {
	// Host is the address to listen on for Xdebug connections
	Host string

	// Port is the port number to listen on for Xdebug connections
	Port int

	// Trigger is the IDE key/trigger value for Xdebug sessions
	Trigger string

	// Commands is the list of commands to execute
	Commands []string

	// JSON enables JSON output format
	JSON bool

	// KillAll enables killing all daemon sessions
	KillAll bool

	// Force skips confirmation prompts
	Force bool

	// Curl is the curl arguments for triggering Xdebug connections
	Curl string

	// BreakpointTimeout is the timeout in seconds for breakpoint validation (0 = disabled)
	BreakpointTimeout int

	// WaitForever disables breakpoint timeout (sets BreakpointTimeout to 0)
	WaitForever bool

	// EnableExternalConnection allows daemon to start without --curl, waiting for external Xdebug trigger
	EnableExternalConnection bool

	// RetryAttempts is the number of connection retry attempts for attach command
	RetryAttempts int
}
