package view

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// SourceFileCache caches source file contents for display.
type SourceFileCache struct {
	files map[string][]string
}

// NewSourceFileCache creates a new source file cache.
func NewSourceFileCache() *SourceFileCache {
	return &SourceFileCache{
		files: make(map[string][]string),
	}
}

// cacheFile reads a file from disk and caches its lines.
// Returns error if file cannot be read.
func (s *SourceFileCache) cacheFile(path string) error {
	if _, exists := s.files[path]; exists {
		return nil // Already cached
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}

	lines := strings.Split(string(content), "\n")
	s.files[path] = lines
	return nil
}

// getLines retrieves a range of lines from a cached file.
// Returns the lines and any error encountered.
// If begin is 0 or negative, starts from line 1.
// If length is 0 or negative, returns all lines from begin to end.
func (s *SourceFileCache) getLines(path string, begin, length int) ([]string, error) {
	if err := s.cacheFile(path); err != nil {
		return nil, err
	}

	lines := s.files[path]
	totalLines := len(lines)

	// Adjust begin to 1-indexed, ensure it's valid
	if begin < 1 {
		begin = 1
	}
	if begin > totalLines {
		return []string{}, nil
	}

	// Convert to 0-indexed
	startIdx := begin - 1
	endIdx := totalLines

	// If length specified, calculate end
	if length > 0 {
		endIdx = startIdx + length
		if endIdx > totalLines {
			endIdx = totalLines
		}
	}

	return lines[startIdx:endIdx], nil
}

// PrintSourceLn displays source code around the specified line.
// fileURI is expected to be a file:// URI or absolute path.
// line is the current line (1-indexed).
// length is the number of lines to display (before and after current line).
func (v *View) PrintSourceLn(fileURI string, line, length int) {
	// Convert file:// URI to path
	path := fileURI
	if strings.HasPrefix(fileURI, "file://") {
		parsedURL, err := url.Parse(fileURI)
		if err != nil {
			v.PrintErrorLn(fmt.Sprintf("Error parsing file URI: %v", err))
			return
		}
		path = parsedURL.Path
	}

	// Calculate range: show 'length' lines before and after current line
	contextLines := length
	if contextLines < 3 {
		contextLines = 5 // Default context
	}

	begin := line - contextLines
	totalLength := contextLines*2 + 1

	lines, err := v.source.getLines(path, begin, totalLength)
	if err != nil {
		// Source file not accessible locally - continue silently
		return
	}

	// Calculate starting line number
	startLine := line - contextLines
	if startLine < 1 {
		startLine = 1
	}

	// Print lines with line numbers
	v.PrintLn("")
	for i, content := range lines {
		lineNum := startLine + i
		marker := " "
		if lineNum == line {
			marker = ">" // Mark current line
		}
		v.PrintLn(fmt.Sprintf("%s %4d | %s", marker, lineNum, content))
	}
	v.PrintLn("")
}

// PrintSourceChangeLn displays a notification that execution moved to a new file.
func (v *View) PrintSourceChangeLn(filename string) {
	v.PrintLn(fmt.Sprintf("=> Execution in: %s", filename))
}
