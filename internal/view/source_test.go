package view

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewSourceFileCache(t *testing.T) {
	cache := NewSourceFileCache()
	if cache == nil {
		t.Fatal("NewSourceFileCache() returned nil")
	}
	if cache.files == nil {
		t.Error("files map not initialized")
	}
}

func TestSourceFileCache_CacheFile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.php")
	content := "<?php\necho 'hello';\n$var = 42;\n"

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cache := NewSourceFileCache()

	// First cache should read the file
	err = cache.cacheFile(testFile)
	if err != nil {
		t.Errorf("cacheFile() error = %v", err)
	}

	// Check that file was cached
	if _, exists := cache.files[testFile]; !exists {
		t.Error("File was not cached")
	}

	// Second cache should use existing cache
	err = cache.cacheFile(testFile)
	if err != nil {
		t.Errorf("cacheFile() second call error = %v", err)
	}
}

func TestSourceFileCache_CacheFile_NotFound(t *testing.T) {
	cache := NewSourceFileCache()

	err := cache.cacheFile("/nonexistent/file.php")
	if err == nil {
		t.Error("cacheFile() expected error for nonexistent file")
	}
}

func TestSourceFileCache_GetLines(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.php")
	content := `<?php
function test() {
    echo 'line 3';
    $var = 42;
    return $var;
}
test();
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cache := NewSourceFileCache()

	tests := []struct {
		name       string
		begin      int
		length     int
		wantLines  int
		wantFirst  string
		wantErr    bool
	}{
		{
			name:      "get all lines",
			begin:     1,
			length:    0,
			wantLines: 8,
			wantFirst: "<?php",
			wantErr:   false,
		},
		{
			name:      "get specific range",
			begin:     2,
			length:    3,
			wantLines: 3,
			wantFirst: "function test() {",
			wantErr:   false,
		},
		{
			name:      "begin at 0 defaults to 1",
			begin:     0,
			length:    2,
			wantLines: 2,
			wantFirst: "<?php",
			wantErr:   false,
		},
		{
			name:      "begin beyond file",
			begin:     100,
			length:    5,
			wantLines: 0,
			wantErr:   false,
		},
		{
			name:      "length exceeds file",
			begin:     5,
			length:    100,
			wantLines: 4,
			wantFirst: "    return $var;",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines, err := cache.getLines(testFile, tt.begin, tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("getLines() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(lines) != tt.wantLines {
				t.Errorf("getLines() got %d lines, want %d", len(lines), tt.wantLines)
			}
			if len(lines) > 0 && tt.wantFirst != "" {
				if lines[0] != tt.wantFirst {
					t.Errorf("getLines() first line = %q, want %q", lines[0], tt.wantFirst)
				}
			}
		})
	}
}

func TestView_PrintSourceLn(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.php")
	content := `<?php
function test() {
    echo 'line 3';
    $var = 42;
    return $var;
}
test();
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var buf bytes.Buffer
	v := &View{
		stdout: &buf,
		stderr: &buf,
		source: NewSourceFileCache(),
	}

	// Test with absolute path
	v.PrintSourceLn(testFile, 4, 2)

	output := buf.String()

	// Check that output contains line numbers and content
	if !strings.Contains(output, "$var = 42") {
		t.Error("PrintSourceLn() output missing expected line content")
	}
	if !strings.Contains(output, ">") {
		t.Error("PrintSourceLn() output missing current line marker")
	}

	// Test with file:// URI
	buf.Reset()
	fileURI := "file://" + testFile
	v.PrintSourceLn(fileURI, 3, 1)

	output = buf.String()
	if !strings.Contains(output, "echo 'line 3'") {
		t.Error("PrintSourceLn() with file:// URI failed")
	}
}

func TestView_PrintSourceChangeLn(t *testing.T) {
	var buf bytes.Buffer
	v := &View{stdout: &buf}

	v.PrintSourceChangeLn("/path/to/file.php")

	output := buf.String()
	want := "=> Execution in: /path/to/file.php"

	if !strings.Contains(output, want) {
		t.Errorf("PrintSourceChangeLn() = %q, want to contain %q", output, want)
	}
}
