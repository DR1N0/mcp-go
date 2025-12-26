package sse

import "github.com/DR1N0/mcp-go/transport"

// ServerTransport is an SSE (Server-Sent Events) transport for MCP servers
type ServerTransport interface {
	transport.ServerTransport
}

// ClientTransport is an SSE transport for MCP clients
type ClientTransport interface {
	transport.ClientTransport
}
