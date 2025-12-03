package view

import (
	"fmt"
	"io"
	"os"
)

// View handles terminal output operations for the debugger CLI.
type View struct {
	stdout io.Writer
	stderr io.Writer
	source *SourceFileCache
}

// NewView creates a new View instance with source cache.
func NewView() *View {
	return &View{
		stdout: os.Stdout,
		stderr: os.Stderr,
		source: NewSourceFileCache(),
	}
}

// Print outputs a string without a newline.
func (v *View) Print(s string) {
	fmt.Fprint(v.stdout, s)
}

// PrintLn outputs a string with a newline.
func (v *View) PrintLn(s string) {
	fmt.Fprintln(v.stdout, s)
}

// PrintErrorLn outputs an error message to stderr with a newline.
func (v *View) PrintErrorLn(s string) {
	fmt.Fprintln(v.stderr, s)
}
