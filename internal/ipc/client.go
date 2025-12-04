package ipc

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

// DefaultRetryAttempts is the default number of connection retry attempts
const DefaultRetryAttempts = 3

// Client represents an IPC client that connects to a Unix domain socket
type Client struct {
	socketPath string
	timeout    time.Duration
}

// NewClient creates a new IPC client
func NewClient(socketPath string) *Client {
	return &Client{
		socketPath: socketPath,
		timeout:    5 * time.Second, // Default 5 second timeout
	}
}

// SetTimeout sets the connection timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// Connect establishes a connection to the IPC server
func (c *Client) Connect() (net.Conn, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to socket %s: %w", c.socketPath, err)
	}
	return conn, nil
}

// ConnectWithRetry attempts to connect to the IPC server with exponential backoff
// maxAttempts specifies the number of connection attempts (must be >= 1)
// Backoff delays: 100ms, 200ms, 400ms, 800ms, etc. (100ms * 2^attempt)
func (c *Client) ConnectWithRetry(maxAttempts int) (net.Conn, error) {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		conn, err := c.Connect()
		if err == nil {
			return conn, nil
		}
		lastErr = err

		// Don't sleep after the last attempt
		if attempt < maxAttempts-1 {
			// Exponential backoff: 100ms * (2^attempt)
			backoff := time.Duration(100*(1<<uint(attempt))) * time.Millisecond
			time.Sleep(backoff)
		}
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxAttempts, lastErr)
}

// SendCommands sends a batch of commands to the daemon and returns the response
func (c *Client) SendCommands(commands []string, jsonOutput bool) (*CommandResponse, error) {
	// Connect to server
	conn, err := c.Connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Set read/write deadlines
	deadline := time.Now().Add(c.timeout)
	if err := conn.SetDeadline(deadline); err != nil {
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	// Create request
	req := NewExecuteCommandsRequest(commands, jsonOutput)

	// Serialize and send request
	reqData, err := req.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}

	// Write request with newline delimiter
	if _, err := conn.Write(append(reqData, '\n')); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	reader := bufio.NewReader(conn)
	respData, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var resp CommandResponse
	if err := resp.FromJSON(respData); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

// SendCommandsWithRetry sends commands with connection retry logic
func (c *Client) SendCommandsWithRetry(commands []string, jsonOutput bool, maxAttempts int) (*CommandResponse, error) {
	// Connect to server with retry
	conn, err := c.ConnectWithRetry(maxAttempts)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Set read/write deadlines
	deadline := time.Now().Add(c.timeout)
	if err := conn.SetDeadline(deadline); err != nil {
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	// Create request
	req := NewExecuteCommandsRequest(commands, jsonOutput)

	// Serialize and send request
	reqData, err := req.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}

	// Write request with newline delimiter
	if _, err := conn.Write(append(reqData, '\n')); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	reader := bufio.NewReader(conn)
	respData, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var resp CommandResponse
	if err := resp.FromJSON(respData); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

// Kill sends a kill request to the daemon
func (c *Client) Kill() (*CommandResponse, error) {
	// Connect to server
	conn, err := c.Connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Set deadlines
	deadline := time.Now().Add(c.timeout)
	if err := conn.SetDeadline(deadline); err != nil {
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	// Create kill request
	req := NewKillRequest()

	// Serialize and send
	reqData, err := req.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}

	if _, err := conn.Write(append(reqData, '\n')); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	reader := bufio.NewReader(conn)
	respData, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var resp CommandResponse
	if err := resp.FromJSON(respData); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

// Ping checks if the daemon is responsive
func (c *Client) Ping() error {
	conn, err := c.Connect()
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
