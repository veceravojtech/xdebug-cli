package mcpserver

import (
	"context"
	"fmt"
	"os"

	"github.com/console/xdebug-cli/internal/cfg"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the MCP server and shells out to the xdebug-cli binary.
type Server struct {
	server *mcp.Server
	binary string
}

// New creates a new MCP server that delegates to the given xdebug-cli binary.
func New(binary string) *Server {
	// Prevent nullable array types in JSON schema (e.g. ["null","array"] → "array").
	os.Setenv("JSONSCHEMAGODEBUG", "typeschemasnull=1")

	s := &Server{
		server: mcp.NewServer(&mcp.Implementation{
			Name:    "xdebug-cli",
			Version: cfg.Version,
		}, nil),
		binary: binary,
	}
	s.registerTools()
	return s
}

// Run starts the MCP server on stdio and blocks until the client disconnects.
func (s *Server) Run(ctx context.Context) error {
	if err := s.server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		return fmt.Errorf("mcp server: %w", err)
	}
	return nil
}

func (s *Server) registerTools() {
	s.registerDaemonStart()
	s.registerDaemonKill()
	s.registerDaemonStatus()
	s.registerDaemonList()
	s.registerDaemonIsAlive()
	s.registerExecute()
}
