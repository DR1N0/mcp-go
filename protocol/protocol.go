package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/DR1N0/mcp-go/types"
)

// jsonRpcProtocol implements the Protocol interface
type jsonRpcProtocol struct {
	transport            types.Transport
	requestHandlers      map[string]RequestHandler
	notificationHandlers map[string]NotificationHandler
	pendingRequests      map[interface{}]chan interface{}
	requestID            atomic.Int64
	mu                   sync.RWMutex
}

// NewProtocol creates a new JSON-RPC 2.0 protocol handler
func NewProtocol() Protocol {
	return &jsonRpcProtocol{
		requestHandlers:      make(map[string]RequestHandler),
		notificationHandlers: make(map[string]NotificationHandler),
		pendingRequests:      make(map[interface{}]chan interface{}),
	}
}

// Connect attaches the protocol to a transport
func (p *jsonRpcProtocol) Connect(transport types.Transport) error {
	p.transport = transport

	// Set up the message handler to route incoming messages
	transport.SetMessageHandler(p.handleMessage)

	// Start the transport
	return transport.Start(context.Background())
}

// handleMessage routes incoming JSON-RPC messages
func (p *jsonRpcProtocol) handleMessage(ctx context.Context, msg *types.BaseJSONRPCMessage) {
	// Check if this is a response to a pending request
	if msg.Result != nil || msg.Error != nil {
		p.handleResponse(msg)
		return
	}

	// Check if this is a request (has method and id)
	if msg.Method != "" && msg.ID != nil {
		p.handleRequest(ctx, msg)
		return
	}

	// Otherwise, it's a notification (has method but no id)
	if msg.Method != "" {
		p.handleNotification(msg)
		return
	}
}

// handleRequest processes an incoming request
func (p *jsonRpcProtocol) handleRequest(ctx context.Context, msg *types.BaseJSONRPCMessage) {
	p.mu.RLock()
	handler, ok := p.requestHandlers[msg.Method]
	p.mu.RUnlock()

	var response types.BaseJSONRPCMessage
	response.JSONRPC = "2.0"
	response.ID = msg.ID

	if !ok {
		// Method not found
		response.Error = &types.RPCError{
			Code:    -32601,
			Message: "Method not found",
			Data:    fmt.Sprintf("No handler registered for method: %s", msg.Method),
		}
	} else {
		// Parse params
		var params interface{}
		if msg.Params != nil {
			if err := json.Unmarshal(msg.Params, &params); err != nil {
				response.Error = &types.RPCError{
					Code:    -32602,
					Message: "Invalid params",
					Data:    err.Error(),
				}
			}
		}

		if response.Error == nil {
			// Call the handler
			result, err := handler(ctx, params)
			if err != nil {
				response.Error = &types.RPCError{
					Code:    -32603,
					Message: "Internal error",
					Data:    err.Error(),
				}
			} else {
				// Serialize the result
				resultBytes, err := json.Marshal(result)
				if err != nil {
					response.Error = &types.RPCError{
						Code:    -32603,
						Message: "Failed to serialize result",
						Data:    err.Error(),
					}
				} else {
					response.Result = resultBytes
				}
			}
		}
	}

	// Send the response
	if err := p.transport.Send(ctx, &response); err != nil {
		// Log error but don't panic - transport will handle it
		fmt.Printf("Error sending response: %v\n", err)
	}
}

// handleResponse processes an incoming response
func (p *jsonRpcProtocol) handleResponse(msg *types.BaseJSONRPCMessage) {
	p.mu.RLock()
	responseChan, ok := p.pendingRequests[msg.ID]
	p.mu.RUnlock()

	if !ok {
		fmt.Printf("Received response for unknown request ID: %v\n", msg.ID)
		return
	}

	// Send the response to the waiting goroutine
	select {
	case responseChan <- msg:
	default:
		// Channel might be closed or full
	}
}

// handleNotification processes an incoming notification
func (p *jsonRpcProtocol) handleNotification(msg *types.BaseJSONRPCMessage) {
	p.mu.RLock()
	handler, ok := p.notificationHandlers[msg.Method]
	p.mu.RUnlock()

	if !ok {
		// Silently ignore unknown notifications (as per JSON-RPC spec)
		return
	}

	// Parse params
	var params interface{}
	if msg.Params != nil {
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			fmt.Printf("Error parsing notification params: %v\n", err)
			return
		}
	}

	// Call the handler (ignore errors for notifications)
	_ = handler(params)
}

// Request sends a request and waits for a response
func (p *jsonRpcProtocol) Request(ctx context.Context, method string, params interface{}) (interface{}, error) {
	// Generate a unique request ID
	id := p.requestID.Add(1)

	// Serialize params
	var paramsBytes json.RawMessage
	if params != nil {
		var err error
		paramsBytes, err = json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
	}

	// Create the request message
	msg := &types.BaseJSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  paramsBytes,
	}

	// Create a response channel
	responseChan := make(chan interface{}, 1)

	p.mu.Lock()
	p.pendingRequests[id] = responseChan
	p.mu.Unlock()

	// Ensure cleanup
	defer func() {
		p.mu.Lock()
		delete(p.pendingRequests, id)
		p.mu.Unlock()
		close(responseChan)
	}()

	// Send the request
	if err := p.transport.Send(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for response
	select {
	case response := <-responseChan:
		responseMsg, ok := response.(*types.BaseJSONRPCMessage)
		if !ok {
			return nil, fmt.Errorf("invalid response type")
		}

		// Check for errors
		if responseMsg.Error != nil {
			return nil, fmt.Errorf("RPC error %d: %s", responseMsg.Error.Code, responseMsg.Error.Message)
		}

		return responseMsg.Result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Notification sends a notification (no response expected)
func (p *jsonRpcProtocol) Notification(method string, params interface{}) error {
	// Serialize params
	var paramsBytes json.RawMessage
	if params != nil {
		var err error
		paramsBytes, err = json.Marshal(params)
		if err != nil {
			return fmt.Errorf("failed to marshal params: %w", err)
		}
	}

	// Create the notification message (no ID)
	msg := &types.BaseJSONRPCMessage{
		JSONRPC: "2.0",
		Method:  method,
		Params:  paramsBytes,
	}

	// Send the notification
	return p.transport.Send(context.Background(), msg)
}

// SetRequestHandler registers a handler for incoming requests
func (p *jsonRpcProtocol) SetRequestHandler(method string, handler RequestHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.requestHandlers[method] = handler
}

// SetNotificationHandler registers a handler for incoming notifications
func (p *jsonRpcProtocol) SetNotificationHandler(method string, handler NotificationHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.notificationHandlers[method] = handler
}

// Close shuts down the protocol
func (p *jsonRpcProtocol) Close() error {
	if p.transport != nil {
		return p.transport.Close()
	}
	return nil
}
