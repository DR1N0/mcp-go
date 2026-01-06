package streamable

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/DR1N0/mcp-go/transport"
	"github.com/DR1N0/mcp-go/types"
)

// httpClientTransport implements streamable HTTP transport for MCP clients
type httpClientTransport struct {
	url            string
	client         *http.Client
	messageHandler transport.MessageHandler
	errorHandler   transport.ErrorHandler
	closeHandler   transport.CloseHandler
	mu             sync.RWMutex
	timeout        time.Duration
	closed         bool
}

// NewClientTransport creates a new streamable HTTP client transport
// url is the full endpoint URL (e.g., "http://localhost:8000/mcp")
func NewClientTransport(url string, opts ...ClientTransportOption) ClientTransport {
	t := &httpClientTransport{
		url: url,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		timeout: 30 * time.Second,
		closed:  false,
	}

	// Apply options
	for _, opt := range opts {
		opt(t)
	}

	return t
}

// ClientTransportOption configures the client transport
type ClientTransportOption func(*httpClientTransport)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) ClientTransportOption {
	return func(t *httpClientTransport) {
		t.client = client
	}
}

// WithTimeout sets the request timeout
func WithTimeout(timeout time.Duration) ClientTransportOption {
	return func(t *httpClientTransport) {
		t.timeout = timeout
		if t.client != nil {
			t.client.Timeout = timeout
		}
	}
}

// Start initializes the client transport (no-op for HTTP client)
func (t *httpClientTransport) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	return nil
}

// Send sends a message to the server and waits for the response
func (t *httpClientTransport) Send(ctx context.Context, msg *types.BaseJSONRPCMessage) error {
	t.mu.RLock()
	if t.closed {
		t.mu.RUnlock()
		return fmt.Errorf("transport is closed")
	}
	messageHandler := t.messageHandler
	t.mu.RUnlock()

	// Marshal the message
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read response
	var response types.BaseJSONRPCMessage
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Deliver response to message handler if set
	if messageHandler != nil {
		messageHandler(ctx, &response)
	}

	return nil
}

// Close shuts down the client transport
func (t *httpClientTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true

	if t.closeHandler != nil {
		t.closeHandler()
	}

	return nil
}

// SetMessageHandler sets the callback for incoming messages
func (t *httpClientTransport) SetMessageHandler(handler func(ctx context.Context, msg *types.BaseJSONRPCMessage)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.messageHandler = handler
}

// SetErrorHandler sets the callback for errors
func (t *httpClientTransport) SetErrorHandler(handler func(error)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.errorHandler = handler
}

// SetCloseHandler sets the callback for when the connection is closed
func (t *httpClientTransport) SetCloseHandler(handler func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closeHandler = handler
}
