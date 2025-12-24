package transport

import (
	"github.com/DR1N0/mcp-go/types"
)

// ServerTransport is a transport implementation for MCP servers
type ServerTransport interface {
	types.Transport
	// Additional server-specific methods can be added here if needed
}

// ClientTransport is a transport implementation for MCP clients
type ClientTransport interface {
	types.Transport
	// Additional client-specific methods can be added here if needed
}
