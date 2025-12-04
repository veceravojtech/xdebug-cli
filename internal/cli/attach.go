package cli

import (
	"fmt"
	"os"

	"github.com/console/xdebug-cli/internal/daemon"
	"github.com/console/xdebug-cli/internal/ipc"
	"github.com/console/xdebug-cli/internal/view"
	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach to a running daemon session",
	Long: `Attach to an existing daemon session and execute commands.
The daemon must be running for the specified port (default 9003).

This command allows you to interact with a persistent debug session started
with 'xdebug-cli daemon start'. You can issue multiple commands across
separate invocations without terminating the debug connection.

Examples:
  # Inspect state (attaches to existing session)
  xdebug-cli attach --commands "context local" "print \$myVar"

  # Continue execution
  xdebug-cli attach --commands "run"

  # Get JSON output for automation
  xdebug-cli attach --json --commands "context local"

  # Set breakpoint and step through
  xdebug-cli attach --commands "break :100"
  xdebug-cli attach --commands "run"
  xdebug-cli attach --commands "step" "step"`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runAttachCmd(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	attachCmd.Flags().StringArrayVar(&CLIArgs.Commands, "commands", []string{}, "Commands to execute")
	attachCmd.Flags().IntVar(&CLIArgs.RetryAttempts, "retry", ipc.DefaultRetryAttempts, "Number of connection retry attempts (with exponential backoff)")
	rootCmd.AddCommand(attachCmd)
}

// runAttachCmd connects to a running daemon and executes commands
func runAttachCmd() error {
	v := view.NewView()

	// Validate that commands are provided
	if len(CLIArgs.Commands) == 0 {
		return fmt.Errorf("--commands flag is required for attach command")
	}

	// Create session registry
	registry, err := daemon.NewSessionRegistry()
	if err != nil {
		return fmt.Errorf("failed to create session registry: %w", err)
	}

	// Find session for the specified port
	sessionInfo, err := registry.Get(CLIArgs.Port)
	if err != nil {
		return fmt.Errorf("no daemon running on port %d. Start with:\n  xdebug-cli daemon start", CLIArgs.Port)
	}

	// Create IPC client
	client := ipc.NewClient(sessionInfo.SocketPath)

	// Send commands to daemon with retry logic
	response, err := client.SendCommandsWithRetry(CLIArgs.Commands, CLIArgs.JSON, CLIArgs.RetryAttempts)
	if err != nil {
		return fmt.Errorf("failed to connect to daemon socket: %s\nThe daemon may have crashed or ended.", sessionInfo.SocketPath)
	}

	// Handle response
	if !response.Success {
		if CLIArgs.JSON {
			// Output error in JSON format
			v.OutputJSON("attach", false, response.Error, nil)
		} else {
			return fmt.Errorf("command execution failed: %s", response.Error)
		}
		os.Exit(1)
	}

	// Display results
	if CLIArgs.JSON {
		// If only one command, output single object for easier parsing
		// If multiple commands, output array
		if len(response.Results) == 1 {
			result := response.Results[0]
			v.OutputJSON(result.Command, result.Success, result.Error, result.Result)
		} else {
			// Output all results as JSON array
			v.OutputJSONArray("attach", response.Results)
		}
	} else {
		// Display each result in human-readable format
		for _, result := range response.Results {
			if !result.Success {
				v.PrintErrorLn(fmt.Sprintf("Command '%s' failed: %s", result.Command, result.Error))
				os.Exit(1)
			}
			// Display result based on command type
			displayCommandResult(v, result)
		}
	}

	return nil
}

// displayCommandResult displays the result of a command in human-readable format
func displayCommandResult(v *view.View, result ipc.CommandResult) {
	if result.Result == nil {
		return
	}

	switch result.Command {
	case "print", "p", "property_get":
		// result.Result is a JSONProperty (or map representation of it)
		if propMap, ok := result.Result.(map[string]interface{}); ok {
			prop := mapToJSONProperty(propMap)
			v.PrintJSONProperty(prop)
		}

	case "context", "c":
		// result.Result is a map with "scope" and "variables"
		if ctxMap, ok := result.Result.(map[string]interface{}); ok {
			if scope, ok := ctxMap["scope"].(string); ok {
				v.PrintLn(fmt.Sprintf("\n%s Variables:", scope))
				v.PrintLn("----------------------------------------")
			}
			if vars, ok := ctxMap["variables"].([]interface{}); ok {
				for _, varItem := range vars {
					if varMap, ok := varItem.(map[string]interface{}); ok {
						prop := mapToJSONProperty(varMap)
						v.PrintJSONPropertyWithDepth(prop, 0)
					}
				}
				v.PrintLn("")
			}
		}

	case "run", "r", "continue", "cont", "step", "s", "into", "step_into", "next", "n", "over", "out", "o", "step_out":
		// result.Result is a map with status, filename, line
		if stateMap, ok := result.Result.(map[string]interface{}); ok {
			status := stateMap["status"].(string)
			filename := stateMap["filename"].(string)
			line := int(stateMap["line"].(float64))

			switch status {
			case "break":
				v.PrintLn(fmt.Sprintf("Breakpoint hit at %s:%d", filename, line))
			case "stopping", "stopped", "running":
				v.PrintLn("Execution finished.")
			default:
				v.PrintLn(fmt.Sprintf("Status: %s at %s:%d", status, filename, line))
			}
		}

	case "break", "b":
		// result.Result is a map with id, location, and optionally condition
		if bpMap, ok := result.Result.(map[string]interface{}); ok {
			id := bpMap["id"].(string)
			location := bpMap["location"].(string)
			if condition, ok := bpMap["condition"].(string); ok && condition != "" {
				v.PrintLn(fmt.Sprintf("Breakpoint set at %s with condition '%s' (ID: %s)", location, condition, id))
			} else {
				v.PrintLn(fmt.Sprintf("Breakpoint set at %s (ID: %s)", location, id))
			}
		}

	case "info", "i":
		// result.Result is a map with "type" and either "breakpoints" or "frames"
		if infoMap, ok := result.Result.(map[string]interface{}); ok {
			infoType := infoMap["type"].(string)

			if infoType == "breakpoints" {
				if bps, ok := infoMap["breakpoints"].([]interface{}); ok {
					v.PrintLn("\nBreakpoints:")
					v.PrintLn("----------------------------------------")
					for _, bpItem := range bps {
						if bpMap, ok := bpItem.(map[string]interface{}); ok {
							id := bpMap["id"].(string)
							bpType := bpMap["type"].(string)
							state := bpMap["state"].(string)
							location := ""
							if filename, ok := bpMap["filename"].(string); ok && filename != "" {
								if line, ok := bpMap["line"].(float64); ok && line > 0 {
									location = fmt.Sprintf("%s:%d", filename, int(line))
								}
							}
							v.PrintLn(fmt.Sprintf("  [%s] %s (%s) %s", id, bpType, state, location))
						}
					}
					v.PrintLn("")
				}
			} else if infoType == "stack" {
				if frames, ok := infoMap["frames"].([]interface{}); ok {
					v.PrintLn("\nStack Trace:")
					v.PrintLn("----------------------------------------")
					for _, frameItem := range frames {
						if frameMap, ok := frameItem.(map[string]interface{}); ok {
							level := int(frameMap["level"].(float64))
							where := frameMap["where"].(string)
							filename := frameMap["filename"].(string)
							line := int(frameMap["line"].(float64))
							v.PrintLn(fmt.Sprintf("#%d %s at %s:%d", level, where, filename, line))
						}
					}
					v.PrintLn("")
				}
			}
		}

	case "list", "l":
		// result.Result is a map with file and line
		if listMap, ok := result.Result.(map[string]interface{}); ok {
			filename := listMap["file"].(string)
			line := int(listMap["line"].(float64))
			v.PrintSourceLn(filename, line, 5)
		}

	case "finish", "f":
		// result.Result is a map with message
		if finishMap, ok := result.Result.(map[string]interface{}); ok {
			if msg, ok := finishMap["message"].(string); ok {
				v.PrintLn(msg)
			}
		}

	case "help", "h", "?":
		// result.Result is a map with help text
		if helpMap, ok := result.Result.(map[string]interface{}); ok {
			if helpText, ok := helpMap["help"].(string); ok {
				v.PrintLn(helpText)
			}
		}

	case "clear":
		// result.Result is a map with location, removed_ids, and count
		if clearMap, ok := result.Result.(map[string]interface{}); ok {
			location := clearMap["location"].(string)
			count := int(clearMap["count"].(float64))
			if count == 1 {
				v.PrintLn(fmt.Sprintf("Cleared breakpoint at %s", location))
			} else {
				v.PrintLn(fmt.Sprintf("Cleared %d breakpoints at %s", count, location))
			}
		}

	case "delete", "del", "breakpoint_remove":
		// result.Result is a map with breakpoint_id
		if delMap, ok := result.Result.(map[string]interface{}); ok {
			bpID := delMap["breakpoint_id"].(string)
			v.PrintLn(fmt.Sprintf("Deleted breakpoint %s", bpID))
		}
	}
}

// mapToJSONProperty converts a map[string]interface{} to a view.JSONProperty
func mapToJSONProperty(m map[string]interface{}) view.JSONProperty {
	prop := view.JSONProperty{}

	if name, ok := m["name"].(string); ok {
		prop.Name = name
	}
	if fullname, ok := m["fullname"].(string); ok {
		prop.FullName = fullname
	}
	if propType, ok := m["type"].(string); ok {
		prop.Type = propType
	}
	if value, ok := m["value"].(string); ok {
		prop.Value = value
	}
	if numChildren, ok := m["num_children"].(float64); ok {
		prop.NumChildren = int(numChildren)
	}

	// Handle children recursively
	if children, ok := m["children"].([]interface{}); ok {
		prop.Children = make([]view.JSONProperty, 0, len(children))
		for _, child := range children {
			if childMap, ok := child.(map[string]interface{}); ok {
				prop.Children = append(prop.Children, mapToJSONProperty(childMap))
			}
		}
	}

	return prop
}
