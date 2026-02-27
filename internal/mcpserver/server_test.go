package mcpserver

import "testing"

func TestNew(t *testing.T) {
	s := New("/fake/binary")
	if s == nil {
		t.Fatal("New() returned nil")
	}
	if s.binary != "/fake/binary" {
		t.Errorf("got binary %q, want %q", s.binary, "/fake/binary")
	}
}

func TestNew_ServerInitialized(t *testing.T) {
	s := New("/fake/binary")
	if s.server == nil {
		t.Fatal("server field should not be nil after New()")
	}
}
