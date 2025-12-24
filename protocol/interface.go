package protocol

import (
	"context"

	"github.com/DR1N0/mcp-go/types"
)

// Protocol handles JSON-RPC 2.0 message routing and processing
type Protocol interface {
	// Connect attaches the protocol to a transport
	Connect(transport types.Transport) error

	// Request sends a request and waits for a response
	Request(ctx context.Context, method string, params interface{}) (interface{}, error)

	// Notification sends a notification (no response expected)
	Notification(method string, params interface{}) error

	// SetRequestHandler registers a handler for incoming requests
	SetRequestHandler(method string, handler RequestHandler)

	// SetNotificationHandler registers a handler for incoming notifications
	SetNotificationHandler(method string, handler NotificationHandler)

	// Close shuts down the protocol
	Close() error
}

// RequestHandler handles incoming JSON-RPC requests
type RequestHandler func(ctx context.Context, params interface{}) (interface{}, error)

// NotificationHandler handles incoming JSON-RPC notifications
type NotificationHandler func(params interface{}) error
