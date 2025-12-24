package streamable

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/DR1N0/mcp-go/types"
)

// pendingRequest holds a pending request waiting for a response
type pendingRequest struct {
	responseChan chan *types.BaseJSONRPCMessage
	ctx          context.Context
}

// httpServerTransport implements streamable HTTP transport for MCP servers
type httpServerTransport struct {
	endpoint        string
	addr            string
	server          *http.Server
	messageHandler  types.MessageHandler
	errorHandler    types.ErrorHandler
	closeHandler    types.CloseHandler
	pendingRequests map[interface{}]*pendingRequest
	mu              sync.RWMutex
	timeout         time.Duration
}

// NewServerTransport creates a new streamable HTTP server transport
// endpoint is the HTTP path (e.g., "/mcp")
// addr is the address to listen on (e.g., ":8000")
func NewServerTransport(endpoint, addr string) ServerTransport {
	return &httpServerTransport{
		endpoint:        endpoint,
		addr:            addr,
		pendingRequests: make(map[interface{}]*pendingRequest),
		timeout:         30 * time.Second, // Default 30 second timeout
	}
}

// Start initializes the HTTP server and begins listening
func (t *httpServerTransport) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc(t.endpoint, t.ServeHTTP)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	t.server = &http.Server{
		Addr:    t.addr,
		Handler: mux,
	}

	log.Printf("Streamable HTTP server starting on http://localhost%s%s", t.addr, t.endpoint)

	// Run server in a goroutine
	go func() {
		if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.mu.RLock()
			errorHandler := t.errorHandler
			t.mu.RUnlock()

			if errorHandler != nil {
				errorHandler(fmt.Errorf("server error: %w", err))
			}
		}
	}()

	return nil
}

// ServeHTTP handles incoming HTTP requests
func (t *httpServerTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode the JSON-RPC request
	var req types.BaseJSONRPCMessage
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Printf("Received request: method=%s, id=%v (type: %T)", req.Method, req.ID, req.ID)

	// Get the message handler
	t.mu.RLock()
	handler := t.messageHandler
	t.mu.RUnlock()

	if handler == nil {
		http.Error(w, "No message handler registered", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	// Handle notifications (no ID) - don't wait for response
	if req.ID == nil {
		log.Printf("Handling notification: method=%s", req.Method)
		go func() {
			handler(ctx, &req)
		}()
		// Notifications get immediate 200 OK
		w.WriteHeader(http.StatusOK)
		return
	}

	// For requests (have ID), register pending request and wait for response
	responseChan := make(chan *types.BaseJSONRPCMessage, 1)

	pending := &pendingRequest{
		responseChan: responseChan,
		ctx:          ctx,
	}

	// Register the pending request
	t.mu.Lock()
	t.pendingRequests[req.ID] = pending
	log.Printf("Registered pending request with id=%v (type: %T), total pending: %d", req.ID, req.ID, len(t.pendingRequests))
	t.mu.Unlock()

	// Ensure cleanup
	defer func() {
		t.mu.Lock()
		delete(t.pendingRequests, req.ID)
		t.mu.Unlock()
		close(responseChan)
	}()

	// Call the message handler in a goroutine
	go func() {
		handler(ctx, &req)
	}()

	// Wait for response with timeout
	timeout := time.After(t.timeout)
	select {
	case response := <-responseChan:
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	case <-timeout:
		log.Printf("Request timeout for id=%v", req.ID)
		http.Error(w, "Request timeout", http.StatusRequestTimeout)
	case <-ctx.Done():
		log.Printf("Request cancelled for id=%v", req.ID)
		http.Error(w, "Request cancelled", http.StatusRequestTimeout)
	}
}

// Send sends a message (response) back to the client
func (t *httpServerTransport) Send(ctx context.Context, msg *types.BaseJSONRPCMessage) error {
	if msg == nil {
		return fmt.Errorf("cannot send nil message")
	}

	log.Printf("Send called: id=%v (type: %T), method=%s", msg.ID, msg.ID, msg.Method)

	// Look up the pending request by ID
	t.mu.RLock()
	pending, ok := t.pendingRequests[msg.ID]
	log.Printf("Looking up pending request for id=%v (type: %T), found=%v, total pending=%d", msg.ID, msg.ID, ok, len(t.pendingRequests))
	t.mu.RUnlock()

	if !ok {
		// If there's no pending request, this might be a notification
		log.Printf("No pending request found for id=%v (type: %T) - might be a notification", msg.ID, msg.ID)
		return nil
	}

	log.Printf("Sending response to channel for id=%v", msg.ID)
	// Send the response to the waiting HTTP handler
	select {
	case pending.responseChan <- msg:
		log.Printf("Successfully sent response for id=%v", msg.ID)
		return nil
	case <-pending.ctx.Done():
		return fmt.Errorf("request context cancelled")
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout sending response")
	}
}

// Close shuts down the HTTP server
func (t *httpServerTransport) Close() error {
	t.mu.RLock()
	closeHandler := t.closeHandler
	t.mu.RUnlock()

	if t.server != nil {
		if err := t.server.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
	}

	if closeHandler != nil {
		closeHandler()
	}

	return nil
}

// SetMessageHandler sets the callback for handling incoming messages
func (t *httpServerTransport) SetMessageHandler(handler func(ctx context.Context, msg *types.BaseJSONRPCMessage)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.messageHandler = handler
}

// SetErrorHandler sets the callback for handling errors
func (t *httpServerTransport) SetErrorHandler(handler func(error)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.errorHandler = handler
}

// SetCloseHandler sets the callback for when the connection is closed
func (t *httpServerTransport) SetCloseHandler(handler func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closeHandler = handler
}
