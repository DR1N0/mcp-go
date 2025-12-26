package stdio

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/DR1N0/mcp-go/types"
)

// stdioServerTransport implements stdio transport for MCP servers
type stdioServerTransport struct {
	reader         *bufio.Reader
	writer         *bufio.Writer
	messageHandler types.MessageHandler
	errorHandler   types.ErrorHandler
	closeHandler   types.CloseHandler
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	closed         bool
}

// NewServerTransport creates a new stdio server transport
// Reads from stdin, writes to stdout
func NewServerTransport() ServerTransport {
	ctx, cancel := context.WithCancel(context.Background())
	return &stdioServerTransport{
		reader: bufio.NewReader(os.Stdin),
		writer: bufio.NewWriter(os.Stdout),
		ctx:    ctx,
		cancel: cancel,
		closed: false,
	}
}

// Start begins reading messages from stdin
func (t *stdioServerTransport) Start(ctx context.Context) error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return fmt.Errorf("transport is closed")
	}
	t.mu.Unlock()

	// Start reading loop in goroutine
	go t.readLoop()

	return nil
}

// readLoop continuously reads JSON-RPC messages from stdin
func (t *stdioServerTransport) readLoop() {
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
			// Read line from stdin
			line, err := t.reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					// Stdin closed, shut down gracefully
					t.Close()
					return
				}
				t.mu.RLock()
				errorHandler := t.errorHandler
				t.mu.RUnlock()
				if errorHandler != nil {
					errorHandler(fmt.Errorf("failed to read from stdin: %w", err))
				}
				continue
			}

			// Parse JSON-RPC message
			var msg types.BaseJSONRPCMessage
			if err := json.Unmarshal(line, &msg); err != nil {
				t.mu.RLock()
				errorHandler := t.errorHandler
				t.mu.RUnlock()
				if errorHandler != nil {
					errorHandler(fmt.Errorf("failed to parse JSON: %w", err))
				}
				continue
			}

			// Deliver message to handler
			t.mu.RLock()
			messageHandler := t.messageHandler
			t.mu.RUnlock()

			if messageHandler != nil {
				messageHandler(t.ctx, &msg)
			}
		}
	}
}

// Send writes a JSON-RPC message to stdout
func (t *stdioServerTransport) Send(ctx context.Context, msg *types.BaseJSONRPCMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	// Marshal message to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write to stdout with newline
	if _, err := t.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}
	if err := t.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	// Flush to ensure message is sent immediately
	if err := t.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush stdout: %w", err)
	}

	return nil
}

// Close shuts down the transport
func (t *stdioServerTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	t.cancel()

	// Flush any remaining output
	if t.writer != nil {
		t.writer.Flush()
	}

	if t.closeHandler != nil {
		t.closeHandler()
	}

	return nil
}

// SetMessageHandler sets the callback for incoming messages
func (t *stdioServerTransport) SetMessageHandler(handler func(ctx context.Context, msg *types.BaseJSONRPCMessage)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.messageHandler = handler
}

// SetErrorHandler sets the callback for errors
func (t *stdioServerTransport) SetErrorHandler(handler func(error)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.errorHandler = handler
}

// SetCloseHandler sets the callback for when the connection is closed
func (t *stdioServerTransport) SetCloseHandler(handler func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closeHandler = handler
}
