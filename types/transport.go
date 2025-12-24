package types

import "context"

// Transport defines the interface for MCP transports
type Transport interface {
	// Start initializes the transport
	Start(ctx context.Context) error

	// Send sends a message through the transport
	Send(ctx context.Context, msg *BaseJSONRPCMessage) error

	// Close shuts down the transport
	Close() error

	// SetMessageHandler sets the callback for incoming messages
	SetMessageHandler(handler func(ctx context.Context, msg *BaseJSONRPCMessage))

	// SetErrorHandler sets the callback for errors
	SetErrorHandler(handler func(error))

	// SetCloseHandler sets the callback for when the connection is closed
	SetCloseHandler(handler func())
}

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, msg *BaseJSONRPCMessage)

// ErrorHandler handles errors
type ErrorHandler func(error)

// CloseHandler handles connection closure
type CloseHandler func()
