package dbgp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// MaxMessageSize is the maximum allowed DBGp message size (100MB)
	// This prevents memory exhaustion from corrupted size values
	MaxMessageSize = 100 * 1024 * 1024

	// DefaultMessageTimeout is the default timeout for reading DBGp messages
	DefaultMessageTimeout = 30 * time.Second
)

var (
	// digitsOnlyRegex validates that the size field contains only digits
	digitsOnlyRegex = regexp.MustCompile(`^\d+$`)
)

// Connection wraps a network connection and handles DBGp message framing
type Connection struct {
	conn   net.Conn
	reader *bufio.Reader
}

// NewConnection creates a new DBGp connection wrapper
func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
}

// ReadMessage reads a DBGp message with the format: size\0xml\0
func (c *Connection) ReadMessage() (string, error) {
	return c.ReadMessageWithTimeout(DefaultMessageTimeout)
}

// ReadMessageWithTimeout reads a DBGp message with a timeout
// The timeout is enforced by setting a read deadline on the underlying connection
func (c *Connection) ReadMessageWithTimeout(timeout time.Duration) (string, error) {
	// Set read deadline
	if err := c.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return "", fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Ensure deadline is cleared after read (success or failure)
	defer func() {
		// Clear the deadline by setting it to zero value
		_ = c.conn.SetReadDeadline(time.Time{})
	}()

	// Read the size part (up to first null byte)
	sizeBytes, err := c.reader.ReadBytes(0)
	if err != nil {
		return "", fmt.Errorf("failed to read message size: %w", err)
	}

	// Remove the null terminator
	sizeStr := strings.TrimSuffix(string(sizeBytes), "\x00")

	// Validate size field format (digits only) before parsing
	if !digitsOnlyRegex.MatchString(sizeStr) {
		// Show first 50 bytes of invalid size field for debugging
		preview := sizeStr
		if len(preview) > 50 {
			preview = preview[:50] + "..."
		}
		return "", fmt.Errorf("invalid message size field (expected digits only): '%s'", preview)
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		// Should not happen after regex validation, but keep for safety
		preview := sizeStr
		if len(preview) > 50 {
			preview = preview[:50] + "..."
		}
		return "", fmt.Errorf("invalid message size '%s': %w", preview, err)
	}

	// Validate size bounds
	if size < 0 {
		return "", fmt.Errorf("invalid message size: negative value %d", size)
	}
	if size > MaxMessageSize {
		return "", fmt.Errorf("message size %d exceeds maximum allowed size of %d bytes (%.1f MB)",
			size, MaxMessageSize, float64(MaxMessageSize)/(1024*1024))
	}

	// Read the XML content (exactly 'size' bytes)
	xmlBytes := make([]byte, size)
	_, err = io.ReadFull(c.reader, xmlBytes)
	if err != nil {
		return "", fmt.Errorf("failed to read message content: %w", err)
	}

	// Read and discard the trailing null byte
	trailingByte := make([]byte, 1)
	_, err = c.reader.Read(trailingByte)
	if err != nil {
		return "", fmt.Errorf("failed to read message terminator: %w", err)
	}
	if trailingByte[0] != 0 {
		return "", fmt.Errorf("expected null terminator, got byte %d", trailingByte[0])
	}

	return string(xmlBytes), nil
}

// SendMessage sends a DBGp command message with null terminator
func (c *Connection) SendMessage(message string) error {
	// Disable Nagle's algorithm for immediate delivery (critical for debugging)
	if tcpConn, ok := c.conn.(*net.TCPConn); ok {
		_ = tcpConn.SetNoDelay(true)
	}

	// DBGp commands are sent with a null terminator
	data := []byte(message + "\x00")
	_, err := c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

// GetResponse reads a message and parses it as a protocol response
func (c *Connection) GetResponse() (*ProtocolResponse, error) {
	xmlData, err := c.ReadMessage()
	if err != nil {
		return nil, err
	}

	result, err := CreateProtocolFromXML(xmlData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	response, ok := result.(*ProtocolResponse)
	if !ok {
		return nil, fmt.Errorf("expected response, got %T", result)
	}

	return response, nil
}

// Close closes the underlying connection
func (c *Connection) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetRemoteAddr returns the remote address of the connection
func (c *Connection) GetRemoteAddr() string {
	if c.conn != nil {
		return c.conn.RemoteAddr().String()
	}
	return ""
}
