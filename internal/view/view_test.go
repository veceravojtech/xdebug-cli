package view

import (
	"bytes"
	"testing"
)

func TestNewView(t *testing.T) {
	v := NewView()
	if v == nil {
		t.Fatal("NewView() returned nil")
	}
	if v.source == nil {
		t.Error("source cache not initialized")
	}
}

func TestPrint(t *testing.T) {
	var buf bytes.Buffer
	v := &View{stdout: &buf}

	v.Print("test message")

	got := buf.String()
	want := "test message"
	if got != want {
		t.Errorf("Print() = %q, want %q", got, want)
	}
}

func TestPrintLn(t *testing.T) {
	var buf bytes.Buffer
	v := &View{stdout: &buf}

	v.PrintLn("test message")

	got := buf.String()
	want := "test message\n"
	if got != want {
		t.Errorf("PrintLn() = %q, want %q", got, want)
	}
}

func TestPrintErrorLn(t *testing.T) {
	var buf bytes.Buffer
	v := &View{stderr: &buf}

	v.PrintErrorLn("error message")

	got := buf.String()
	want := "error message\n"
	if got != want {
		t.Errorf("PrintErrorLn() = %q, want %q", got, want)
	}
}

