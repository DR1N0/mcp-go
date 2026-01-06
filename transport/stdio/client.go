package stdio

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/DR1N0/mcp-go/transport"
	"github.com/DR1N0/mcp-go/types"
)

// stdioClientTransport implements stdio transport for MCP clients
type stdioClientTransport struct {
	command        string
	args           []string
	cmd            *exec.Cmd
	stdin          io.WriteCloser
	stdout         io.ReadCloser
	stderr         io.ReadCloser
	reader         *bufio.Reader
	writer         *bufio.Writer
	messageHandler transport.MessageHandler
	errorHandler   transport.ErrorHandler
	closeHandler   transport.CloseHandler
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	closed         bool
}

// ClientTransportOption configures the client transport
type ClientTransportOption func(*stdioClientTransport)

// WithStderr redirects stderr to the error handler
func WithStderr(redirect bool) ClientTransportOption {
	return func(t *stdioClientTransport) {
		// stderr handling is set up in Start()
	}
}

// NewClientTransport creates a new stdio client transport
// command is the server executable path
// args are command-line arguments for the server
func NewClientTransport(command string, args []string, opts ...ClientTransportOption) ClientTransport {
	ctx, cancel := context.WithCancel(context.Background())
	t := &stdioClientTransport{
		command: command,
		args:    args,
		ctx:     ctx,
		cancel:  cancel,
		closed:  false,
	}

	// Apply options
	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Start spawns the server process and begins communication
func (t *stdioClientTransport) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	// Create command
	t.cmd = exec.CommandContext(t.ctx, t.command, t.args...)

	// Set up pipes
	var err error
	t.stdin, err = t.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	t.stdout, err = t.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	t.stderr, err = t.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Create buffered reader/writer
	t.reader = bufio.NewReader(t.stdout)
	t.writer = bufio.NewWriter(t.stdin)

	// Start the server process
	if err := t.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server process: %w", err)
	}

	// Start reading loops
	go t.readLoop()
	go t.readStderrLoop()

	return nil
}

// readLoop continuously reads JSON-RPC messages from stdout
func (t *stdioClientTransport) readLoop() {
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
			// Read line from stdout
			line, err := t.reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					// Server closed stdout, shut down gracefully
					t.Close()
					return
				}
				t.mu.RLock()
				errorHandler := t.errorHandler
				t.mu.RUnlock()
				if errorHandler != nil {
					errorHandler(fmt.Errorf("failed to read from stdout: %w", err))
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

// readStderrLoop reads and logs stderr from the server process
func (t *stdioClientTransport) readStderrLoop() {
	scanner := bufio.NewScanner(t.stderr)
	for scanner.Scan() {
		select {
		case <-t.ctx.Done():
			return
		default:
			line := scanner.Text()
			t.mu.RLock()
			errorHandler := t.errorHandler
			t.mu.RUnlock()
			if errorHandler != nil {
				errorHandler(fmt.Errorf("server stderr: %s", line))
			}
		}
	}
}

// Send writes a JSON-RPC message to the server's stdin
func (t *stdioClientTransport) Send(ctx context.Context, msg *types.BaseJSONRPCMessage) error {
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

	// Write to stdin with newline
	if _, err := t.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write to stdin: %w", err)
	}
	if err := t.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	// Flush to ensure message is sent immediately
	if err := t.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush stdin: %w", err)
	}

	return nil
}

// Close shuts down the transport and terminates the server process
func (t *stdioClientTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	t.cancel()

	// Close stdin to signal server to shut down
	if t.stdin != nil {
		t.stdin.Close()
	}

	// Wait for process to exit
	if t.cmd != nil && t.cmd.Process != nil {
		t.cmd.Wait()
	}

	if t.closeHandler != nil {
		t.closeHandler()
	}

	return nil
}

// SetMessageHandler sets the callback for incoming messages
func (t *stdioClientTransport) SetMessageHandler(handler func(ctx context.Context, msg *types.BaseJSONRPCMessage)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.messageHandler = handler
}

// SetErrorHandler sets the callback for errors
func (t *stdioClientTransport) SetErrorHandler(handler func(error)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.errorHandler = handler
}

// SetCloseHandler sets the callback for when the connection is closed
func (t *stdioClientTransport) SetCloseHandler(handler func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closeHandler = handler
}
