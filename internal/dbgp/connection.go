package dbgp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
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
	// Read the size part (up to first null byte)
	sizeBytes, err := c.reader.ReadBytes(0)
	if err != nil {
		return "", fmt.Errorf("failed to read message size: %w", err)
	}

	// Remove the null terminator
	sizeStr := strings.TrimSuffix(string(sizeBytes), "\x00")
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return "", fmt.Errorf("invalid message size '%s': %w", sizeStr, err)
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
