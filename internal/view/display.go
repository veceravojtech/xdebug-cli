package view

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"
)

// ShowInfoBreakpoints displays a formatted table of breakpoints.
func (v *View) ShowInfoBreakpoints(breakpoints []ProtocolBreakpoint) {
	if len(breakpoints) == 0 {
		v.PrintLn("No breakpoints set.")
		return
	}

	v.PrintLn("")
	v.PrintLn("Breakpoints:")
	v.PrintLn(strings.Repeat("-", 80))
	v.PrintLn(fmt.Sprintf("%-4s %-12s %-8s %-40s %s", "ID", "Type", "State", "Location", "Function"))
	v.PrintLn(strings.Repeat("-", 80))

	for _, bp := range breakpoints {
		location := ""
		if bp.GetFilename() != "" {
			if bp.GetLineNumber() > 0 {
				location = fmt.Sprintf("%s:%d", bp.GetFilename(), bp.GetLineNumber())
			} else {
				location = bp.GetFilename()
			}
		}

		function := bp.GetFunction()
		if function == "" {
			function = "-"
		}

		// Truncate long paths for readability
		if len(location) > 40 {
			location = "..." + location[len(location)-37:]
		}

		v.PrintLn(fmt.Sprintf("%-4s %-12s %-8s %-40s %s",
			bp.GetID(),
			bp.GetType(),
			bp.GetState(),
			location,
			function,
		))
	}
	v.PrintLn(strings.Repeat("-", 80))
	v.PrintLn("")
}

// PrintPropertyListWithDetails displays a list of properties (variables) with their values.
// scope is the context name (e.g., "Local", "Global", "Constants").
// properties is the list of variables to display.
func (v *View) PrintPropertyListWithDetails(scope string, properties []ProtocolProperty) {
	if len(properties) == 0 {
		v.PrintLn(fmt.Sprintf("No %s variables.", strings.ToLower(scope)))
		return
	}

	v.PrintLn("")
	v.PrintLn(fmt.Sprintf("%s Variables:", scope))
	v.PrintLn(strings.Repeat("-", 80))

	for _, prop := range properties {
		v.printProperty(prop, 0)
	}

	v.PrintLn(strings.Repeat("-", 80))
	v.PrintLn("")
}

// TryDecodeBase64 attempts to decode a base64-encoded value for string types
// Returns the decoded value if successful, otherwise returns the original value
func TryDecodeBase64(value string, propType string) string {
	// Only try to decode if the type is string
	if propType != "string" {
		return value
	}

	// Skip if the string is too short to be meaningful base64
	if len(value) < 4 {
		return value
	}

	// Try to decode as base64
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return value
	}

	// Check if decoded value is valid UTF-8
	if !utf8.Valid(decoded) {
		return value
	}

	// Return decoded value
	return string(decoded)
}

// printProperty recursively prints a property and its children.
// depth controls indentation for nested properties.
func (v *View) printProperty(prop ProtocolProperty, depth int) {
	indent := strings.Repeat("  ", depth)
	name := prop.GetName()
	propType := prop.GetType()
	value := prop.GetValue()

	// Format the property line
	var line string

	if prop.HasChildren() {
		// For complex types (arrays, objects), show type and child count
		childCount := prop.GetNumChildren()
		line = fmt.Sprintf("%s%s (%s) [%d children]", indent, name, propType, childCount)
		v.PrintLn(line)

		// Recursively print children
		children := prop.GetChildren()
		for _, child := range children {
			if childProp, ok := child.(ProtocolProperty); ok {
				v.printProperty(childProp, depth+1)
			}
		}
	} else {
		// For simple types, show type and value
		if value == "" {
			line = fmt.Sprintf("%s%s (%s)", indent, name, propType)
		} else {
			// Try to decode base64 for string types
			displayValue := TryDecodeBase64(value, propType)

			// Truncate long values
			maxValueLen := 60 - len(indent) - len(name) - len(propType)
			if len(displayValue) > maxValueLen && maxValueLen > 10 {
				displayValue = displayValue[:maxValueLen-3] + "..."
			}

			line = fmt.Sprintf("%s%s (%s) = %s", indent, name, propType, displayValue)
		}
		v.PrintLn(line)
	}
}

// PrintProperty is a convenience method to print a single property tree.
// Useful for 'print' command output.
func (v *View) PrintProperty(prop ProtocolProperty) {
	v.PrintLn("")
	v.printProperty(prop, 0)
	v.PrintLn("")
}

// PrintJSONPropertyWithDepth formats and prints a JSONProperty recursively with specified indentation depth
func (v *View) PrintJSONPropertyWithDepth(prop JSONProperty, depth int) {
	indent := strings.Repeat("  ", depth)
	name := prop.Name
	propType := prop.Type
	value := prop.Value

	// Format the property line
	var line string

	if prop.NumChildren > 0 && len(prop.Children) > 0 {
		// For complex types (arrays, objects), show type and child count
		line = fmt.Sprintf("%s%s (%s) [%d children]", indent, name, propType, prop.NumChildren)
		v.PrintLn(line)

		// Recursively print children
		for _, child := range prop.Children {
			v.PrintJSONPropertyWithDepth(child, depth+1)
		}
	} else {
		// For simple types, show type and value
		if value == "" {
			line = fmt.Sprintf("%s%s (%s)", indent, name, propType)
		} else {
			// Try to decode base64 for string types
			displayValue := TryDecodeBase64(value, propType)

			// Truncate long values
			maxValueLen := 60 - len(indent) - len(name) - len(propType)
			if len(displayValue) > maxValueLen && maxValueLen > 10 {
				displayValue = displayValue[:maxValueLen-3] + "..."
			}

			line = fmt.Sprintf("%s%s (%s) = %s", indent, name, propType, displayValue)
		}
		v.PrintLn(line)
	}
}

// PrintJSONProperty prints a JSONProperty in human-readable format
// This is used by the attach command to display properties returned from the daemon
func (v *View) PrintJSONProperty(prop JSONProperty) {
	v.PrintLn("")
	v.PrintJSONPropertyWithDepth(prop, 0)
	v.PrintLn("")
}

// ShowInfoStack displays a formatted stack trace.
func (v *View) ShowInfoStack(frames []ProtocolStack) {
	if len(frames) == 0 {
		v.PrintLn("No stack frames available.")
		return
	}

	v.PrintLn("")
	v.PrintLn("Stack Trace:")
	v.PrintLn(strings.Repeat("-", 80))

	for _, frame := range frames {
		location := ""
		if frame.GetFilename() != "" {
			location = fmt.Sprintf("%s:%d", frame.GetFilename(), frame.GetLineNumber())
		}

		where := frame.GetWhere()
		if where == "" {
			where = "{unknown}"
		}

		// Format: #0 functionName() at /path/to/file.php:42
		v.PrintLn(fmt.Sprintf("#%d %s at %s", frame.GetLevel(), where, location))
	}

	v.PrintLn(strings.Repeat("-", 80))
	v.PrintLn("")
}
