package transport

import (
	"context"
	"net/http"

	"github.com/DR1N0/mcp-go/types"
)

// ServerTransport is a transport implementation for MCP servers
type ServerTransport interface {
	Transport
	// Additional server-specific methods can be added here if needed
}

// ClientTransport is a transport implementation for MCP clients
type ClientTransport interface {
	Transport
	// Additional client-specific methods can be added here if needed
}

// Transport defines the interface for MCP transports
type Transport interface {
	// Start initializes the transport
	Start(ctx context.Context) error

	// Send sends a message through the transport
	Send(ctx context.Context, msg *types.BaseJSONRPCMessage) error

	// Close shuts down the transport
	Close() error

	// SetMessageHandler sets the callback for incoming messages
	SetMessageHandler(handler func(ctx context.Context, msg *types.BaseJSONRPCMessage))

	// SetErrorHandler sets the callback for errors
	SetErrorHandler(handler func(error))

	// SetCloseHandler sets the callback for when the connection is closed
	SetCloseHandler(handler func())
}

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, msg *types.BaseJSONRPCMessage)

// ErrorHandler handles errors
type ErrorHandler func(error)

// CloseHandler handles connection closure
type CloseHandler func()

// HTTPMiddleware is a function that wraps an http.Handler to add functionality
// Middleware functions are chained in reverse order (following Chi router pattern)
// This allows for flexible composition of cross-cutting concerns like authentication,
// logging, CORS, rate limiting, and telemetry.
type HTTPMiddleware func(http.Handler) http.Handler
