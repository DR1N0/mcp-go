package sse

import (
	"github.com/DR1N0/mcp-go/transport"
	"github.com/DR1N0/mcp-go/types"
)

// ServerTransport is an SSE (Server-Sent Events) transport for MCP servers
type ServerTransport interface {
	transport.ServerTransport
	// WithMiddleware adds HTTP middleware to be chained before the MCP handler
	WithMiddleware(middleware ...types.HTTPMiddleware) ServerTransport
}

// ClientTransport is an SSE transport for MCP clients
type ClientTransport interface {
	transport.ClientTransport
}
