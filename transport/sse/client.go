package sse

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"

	"github.com/DR1N0/mcp-go/types"
)

// sseClientTransport implements SSE transport for MCP clients
type sseClientTransport struct {
	sseURL         string
	messageURL     string // Will be set from endpoint event
	client         *http.Client
	messageHandler types.MessageHandler
	errorHandler   types.ErrorHandler
	closeHandler   types.CloseHandler
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	closed         bool
	timeout        time.Duration
	endpointReady  chan struct{} // Signals when endpoint URL is received
}

// ClientTransportOption configures the client transport
type ClientTransportOption func(*sseClientTransport)

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientTransportOption {
	return func(t *sseClientTransport) {
		t.timeout = timeout
	}
}

// NewClientTransport creates a new SSE client transport
// sseURL is the SSE endpoint URL (e.g., "http://localhost:8001/mcp/sse")
func NewClientTransport(sseURL string, opts ...ClientTransportOption) ClientTransport {
	ctx, cancel := context.WithCancel(context.Background())

	t := &sseClientTransport{
		sseURL:        sseURL,
		ctx:           ctx,
		cancel:        cancel,
		closed:        false,
		timeout:       30 * time.Second,
		endpointReady: make(chan struct{}),
	}

	// Apply options
	for _, opt := range opts {
		opt(t)
	}

	// Create cookie jar for session management
	jar, _ := cookiejar.New(nil)
	t.client = &http.Client{
		Timeout: t.timeout,
		Jar:     jar,
	}

	return t
}

// Start connects to the SSE endpoint and begins receiving events
func (t *sseClientTransport) Start(ctx context.Context) error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return fmt.Errorf("transport is closed")
	}
	t.mu.Unlock()

	// Start SSE connection in goroutine
	go t.connectSSE()

	return nil
}

// connectSSE establishes SSE connection and listens for events
func (t *sseClientTransport) connectSSE() {
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
			if err := t.listenToSSE(); err != nil {
				t.mu.RLock()
				errorHandler := t.errorHandler
				closed := t.closed
				t.mu.RUnlock()

				if closed {
					return
				}

				if errorHandler != nil {
					errorHandler(fmt.Errorf("SSE connection error: %w", err))
				}

				// Wait before reconnecting
				select {
				case <-t.ctx.Done():
					return
				case <-time.After(1 * time.Second):
					continue
				}
			}
		}
	}
}

// listenToSSE connects to SSE endpoint and processes events
func (t *sseClientTransport) listenToSSE() error {
	req, err := http.NewRequestWithContext(t.ctx, http.MethodGet, t.sseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read SSE events
	reader := bufio.NewReader(resp.Body)
	var dataBuffer bytes.Buffer
	var eventType string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return fmt.Errorf("SSE connection closed")
			}
			return fmt.Errorf("failed to read SSE event: %w", err)
		}

		line = strings.TrimRight(line, "\n\r")

		// Empty line signals end of event
		if line == "" {
			if dataBuffer.Len() > 0 {
				// Process event based on type
				if eventType == "endpoint" {
					t.handleEndpointEvent(dataBuffer.Bytes())
				} else if eventType == "message" {
					if err := t.processEvent(dataBuffer.Bytes()); err != nil {
						t.mu.RLock()
						errorHandler := t.errorHandler
						t.mu.RUnlock()
						if errorHandler != nil {
							errorHandler(fmt.Errorf("failed to process event: %w", err))
						}
					}
				}
				dataBuffer.Reset()
				eventType = ""
			}
			continue
		}

		// Parse SSE fields
		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)
			dataBuffer.WriteString(data)
		}
	}
}

// handleEndpointEvent processes the endpoint event from the server
func (t *sseClientTransport) handleEndpointEvent(data []byte) {
	endpoint := strings.TrimSpace(string(data))

	// Convert relative path to absolute URL
	messageURL := endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		// Parse base URL from sseURL
		if idx := strings.Index(t.sseURL, "/mcp/"); idx != -1 {
			baseURL := t.sseURL[:idx]
			messageURL = baseURL + endpoint
		}
	}

	t.mu.Lock()
	t.messageURL = messageURL
	t.mu.Unlock()

	// Signal that endpoint is ready
	select {
	case t.endpointReady <- struct{}{}:
	default:
		// Channel already signaled
	}
}

// processEvent processes an SSE event containing JSON-RPC message
func (t *sseClientTransport) processEvent(data []byte) error {
	var msg types.BaseJSONRPCMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	t.mu.RLock()
	messageHandler := t.messageHandler
	t.mu.RUnlock()

	if messageHandler != nil {
		messageHandler(t.ctx, &msg)
	}

	return nil
}

// Send sends a JSON-RPC message to the server via POST
func (t *sseClientTransport) Send(ctx context.Context, msg *types.BaseJSONRPCMessage) error {
	// Wait for messageURL to be set (poll with timeout)
	var messageURL string
	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		t.mu.RLock()
		messageURL = t.messageURL
		closed := t.closed
		t.mu.RUnlock()

		if closed {
			return fmt.Errorf("transport is closed")
		}

		if messageURL != "" {
			break
		}

		// Check context
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			// Continue polling
		}
	}

	if messageURL == "" {
		return fmt.Errorf("timeout waiting for endpoint URL from server")
	}

	// Marshal message
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send POST request to message endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, messageURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Close shuts down the transport
func (t *sseClientTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	t.cancel()

	if t.closeHandler != nil {
		t.closeHandler()
	}

	return nil
}

// SetMessageHandler sets the callback for incoming messages
func (t *sseClientTransport) SetMessageHandler(handler func(ctx context.Context, msg *types.BaseJSONRPCMessage)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.messageHandler = handler
}

// SetErrorHandler sets the callback for errors
func (t *sseClientTransport) SetErrorHandler(handler func(error)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.errorHandler = handler
}

// SetCloseHandler sets the callback for when the connection is closed
func (t *sseClientTransport) SetCloseHandler(handler func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closeHandler = handler
}
