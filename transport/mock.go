package transport

import (
	"context"
	"sync"

	"github.com/DR1N0/mcp-go/types"
)

// MockTransport is a mock implementation of the Transport interface for testing
type MockTransport struct {
	mu             sync.RWMutex
	sentMessages   []*types.BaseJSONRPCMessage
	messageHandler func(ctx context.Context, msg *types.BaseJSONRPCMessage)
	errorHandler   func(error)
	closeHandler   func()
	started        bool
	closed         bool
	sendError      error // Set this to simulate send errors
	startError     error // Set this to simulate start errors
}

// NewMock creates a new mock transport
func NewMock() *MockTransport {
	return &MockTransport{
		sentMessages: make([]*types.BaseJSONRPCMessage, 0),
	}
}

// Start simulates starting the transport
func (m *MockTransport) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.startError != nil {
		return m.startError
	}

	m.started = true
	return nil
}

// Send records the message and optionally returns an error
func (m *MockTransport) Send(ctx context.Context, msg *types.BaseJSONRPCMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.sendError != nil {
		return m.sendError
	}

	m.sentMessages = append(m.sentMessages, msg)
	return nil
}

// Close simulates closing the transport
func (m *MockTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	if m.closeHandler != nil {
		m.closeHandler()
	}
	return nil
}

// SetMessageHandler sets the message handler
func (m *MockTransport) SetMessageHandler(handler func(ctx context.Context, msg *types.BaseJSONRPCMessage)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messageHandler = handler
}

// SetErrorHandler sets the error handler
func (m *MockTransport) SetErrorHandler(handler func(error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorHandler = handler
}

// SetCloseHandler sets the close handler
func (m *MockTransport) SetCloseHandler(handler func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeHandler = handler
}

// Helper methods for testing

// SimulateReceive simulates receiving a message from the transport
func (m *MockTransport) SimulateReceive(ctx context.Context, msg *types.BaseJSONRPCMessage) {
	m.mu.RLock()
	handler := m.messageHandler
	m.mu.RUnlock()

	if handler != nil {
		handler(ctx, msg)
	}
}

// GetSentMessages returns all messages that were sent
func (m *MockTransport) GetSentMessages() []*types.BaseJSONRPCMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	messages := make([]*types.BaseJSONRPCMessage, len(m.sentMessages))
	copy(messages, m.sentMessages)
	return messages
}

// ClearSentMessages clears the sent messages buffer
func (m *MockTransport) ClearSentMessages() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentMessages = make([]*types.BaseJSONRPCMessage, 0)
}

// IsStarted returns whether the transport was started
func (m *MockTransport) IsStarted() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.started
}

// IsClosed returns whether the transport was closed
func (m *MockTransport) IsClosed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.closed
}

// SetSendError sets an error to return on Send
func (m *MockTransport) SetSendError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendError = err
}

// SetStartError sets an error to return on Start
func (m *MockTransport) SetStartError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startError = err
}
