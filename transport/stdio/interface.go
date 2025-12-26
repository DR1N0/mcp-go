package stdio

import "github.com/DR1N0/mcp-go/transport"

// ServerTransport is a stdio transport for MCP servers
type ServerTransport interface {
	transport.ServerTransport
}

// ClientTransport is a stdio transport for MCP clients
type ClientTransport interface {
	transport.ClientTransport
}
