package view

import (
	"bytes"
	"strings"
	"testing"
)

func TestView_ShowHelpMessage(t *testing.T) {
	var buf bytes.Buffer
	v := &View{stdout: &buf}

	v.ShowHelpMessage()

	output := buf.String()

	expectedCommands := []string{
		"run, r",
		"step, s",
		"next, n",
		"out, o",
		"break, b",
		"print, p",
		"context, c",
		"list, l",
		"info, i",
		"finish, f",
		"help, h, ?",
	}

	for _, cmd := range expectedCommands {
		if !strings.Contains(output, cmd) {
			t.Errorf("ShowHelpMessage() missing command: %s", cmd)
		}
	}
}

func TestView_ShowInfoHelpMessage(t *testing.T) {
	var buf bytes.Buffer
	v := &View{stdout: &buf}

	v.ShowInfoHelpMessage()

	output := buf.String()

	expectedContent := []string{
		"info",
		"breakpoints",
		"info b",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("ShowInfoHelpMessage() missing content: %s", content)
		}
	}
}

func TestView_ShowStepHelpMessage(t *testing.T) {
	var buf bytes.Buffer
	v := &View{stdout: &buf}

	v.ShowStepHelpMessage()

	output := buf.String()

	expectedContent := []string{
		"step, s",
		"next, n",
		"out, o",
		"Step into",
		"Step over",
		"Step out",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("ShowStepHelpMessage() missing content: %s", content)
		}
	}
}

func TestView_ShowBreakpointHelpMessage(t *testing.T) {
	var buf bytes.Buffer
	v := &View{stdout: &buf}

	v.ShowBreakpointHelpMessage()

	output := buf.String()

	expectedContent := []string{
		"break",
		"<line>",
		"<file>:<line>",
		"call",
		"exception",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("ShowBreakpointHelpMessage() missing content: %s", content)
		}
	}
}

func TestView_ShowPrintHelpMessage(t *testing.T) {
	var buf bytes.Buffer
	v := &View{stdout: &buf}

	v.ShowPrintHelpMessage()

	output := buf.String()

	expectedContent := []string{
		"print",
		"<variable>",
		"$myVar",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("ShowPrintHelpMessage() missing content: %s", content)
		}
	}
}

func TestView_ShowContextHelpMessage(t *testing.T) {
	var buf bytes.Buffer
	v := &View{stdout: &buf}

	v.ShowContextHelpMessage()

	output := buf.String()

	expectedContent := []string{
		"context",
		"local",
		"global",
		"constant",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("ShowContextHelpMessage() missing content: %s", content)
		}
	}
}

func TestView_ShowCommandHelp(t *testing.T) {
	tests := []struct {
		name            string
		command         string
		expectedContent string
	}{
		{
			name:            "help for info",
			command:         "info",
			expectedContent: "info breakpoints",
		},
		{
			name:            "help for step",
			command:         "step",
			expectedContent: "Step into",
		},
		{
			name:            "help for break",
			command:         "break",
			expectedContent: "Set breakpoint",
		},
		{
			name:            "help for print",
			command:         "print",
			expectedContent: "Print variable",
		},
		{
			name:            "help for context",
			command:         "context",
			expectedContent: "context [type]",
		},
		{
			name:            "help for unknown command",
			command:         "unknown",
			expectedContent: "No help available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			v := &View{stdout: &buf}

			v.ShowCommandHelp(tt.command)

			output := buf.String()
			if !strings.Contains(output, tt.expectedContent) {
				t.Errorf("ShowCommandHelp(%q) missing expected content: %s", tt.command, tt.expectedContent)
			}
		})
	}
}
