package sse

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/DR1N0/mcp-go/transport"
	"github.com/DR1N0/mcp-go/types"
)

// sseSession represents a single SSE client session
type sseSession struct {
	id             string
	messageChan    chan []byte
	requestChan    chan *types.BaseJSONRPCMessage
	ctx            context.Context
	cancel         context.CancelFunc
	messageHandler transport.MessageHandler
}

// sseServerTransport implements SSE transport for MCP servers
type sseServerTransport struct {
	sseEndpoint    string
	messageHandler transport.MessageHandler
	errorHandler   transport.ErrorHandler
	closeHandler   transport.CloseHandler
	server         *http.Server
	middleware     []transport.HTTPMiddleware
	mu             sync.RWMutex
	sessions       map[string]*sseSession
	ctx            context.Context
	cancel         context.CancelFunc
	closed         bool
}

// NewServerTransport creates a new SSE server transport
// sseEndpoint is the path for SSE streaming (e.g., "/mcp/sse")
// addr is the server address (e.g., ":8001")
func NewServerTransport(sseEndpoint string, addr string) ServerTransport {
	ctx, cancel := context.WithCancel(context.Background())
	return &sseServerTransport{
		sseEndpoint: sseEndpoint,
		sessions:    make(map[string]*sseSession),
		middleware:  make([]transport.HTTPMiddleware, 0),
		ctx:         ctx,
		cancel:      cancel,
		closed:      false,
		server: &http.Server{
			Addr: addr,
		},
	}
}

// WithMiddleware adds HTTP middleware to be chained before the MCP handler
// Middleware is chained in reverse order (last added = outermost wrapper)
func (t *sseServerTransport) WithMiddleware(middleware ...transport.HTTPMiddleware) ServerTransport {
	t.middleware = append(t.middleware, middleware...)
	return t
}

// generateSessionID creates a new random session ID
func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Start begins the HTTP server
func (t *sseServerTransport) Start(ctx context.Context) error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return fmt.Errorf("transport is closed")
	}
	t.mu.Unlock()

	// Set up HTTP handlers
	mux := http.NewServeMux()

	// SSE endpoint (GET only)
	mux.HandleFunc(t.sseEndpoint, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			t.handleSSE(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Message endpoint (POST only) - appends /message to sseEndpoint
	messageEndpoint := t.sseEndpoint
	if messageEndpoint[len(messageEndpoint)-1] != '/' {
		messageEndpoint += "/"
	}
	messageEndpoint += "message"

	mux.HandleFunc(messageEndpoint, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			t.handleMessage(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Health check endpoint
	healthEndpoint := t.sseEndpoint + "/health"
	mux.HandleFunc(healthEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Chain middleware in reverse order (last added = outermost)
	var handler http.Handler = mux
	for i := len(t.middleware) - 1; i >= 0; i-- {
		handler = t.middleware[i](handler)
	}

	t.server.Handler = handler

	// Start server in goroutine
	go func() {
		if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.mu.RLock()
			errorHandler := t.errorHandler
			t.mu.RUnlock()
			if errorHandler != nil {
				errorHandler(fmt.Errorf("HTTP server error: %w", err))
			}
		}
	}()

	return nil
}

// handleSSE handles SSE connections from clients
func (t *sseServerTransport) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create new session
	sessionID := generateSessionID()
	sessionCtx, sessionCancel := context.WithCancel(r.Context())

	session := &sseSession{
		id:             sessionID,
		messageChan:    make(chan []byte, 10),
		requestChan:    make(chan *types.BaseJSONRPCMessage, 10),
		ctx:            sessionCtx,
		cancel:         sessionCancel,
		messageHandler: nil, // Will be set when needed
	}

	// Register session
	t.mu.Lock()
	t.sessions[sessionID] = session
	session.messageHandler = t.messageHandler
	t.mu.Unlock()

	// Deregister session on disconnect
	defer func() {
		sessionCancel()
		t.mu.Lock()
		delete(t.sessions, sessionID)
		close(session.messageChan)
		close(session.requestChan)
		t.mu.Unlock()
	}()

	// Get flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send endpoint event per MCP SSE protocol
	messageEndpoint := t.sseEndpoint
	if messageEndpoint[len(messageEndpoint)-1] != '/' {
		messageEndpoint += "/"
	}
	messageEndpoint += "message?session_id=" + sessionID

	fmt.Fprintf(w, "event: endpoint\n")
	fmt.Fprintf(w, "data: %s\n\n", messageEndpoint)
	flusher.Flush()

	// Start message processor for this session
	go t.processSessionMessages(session)

	// Stream events to client
	for {
		select {
		case <-t.ctx.Done():
			return
		case <-session.ctx.Done():
			return
		case data := <-session.messageChan:
			// Write SSE event
			fmt.Fprintf(w, "event: message\n")
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// processSessionMessages processes incoming messages for a session
func (t *sseServerTransport) processSessionMessages(session *sseSession) {
	for {
		select {
		case <-session.ctx.Done():
			return
		case msg := <-session.requestChan:
			if session.messageHandler != nil {
				// Create context with session ID
				ctx := context.WithValue(session.ctx, "session_id", session.id)
				session.messageHandler(ctx, msg)
			}
		}
	}
}

// handleMessage handles incoming POST requests from clients
func (t *sseServerTransport) handleMessage(w http.ResponseWriter, r *http.Request) {
	// Get session ID from query parameter
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	// Find session
	t.mu.RLock()
	session, exists := t.sessions[sessionID]
	t.mu.RUnlock()

	if !exists {
		http.Error(w, "Could not find session", http.StatusNotFound)
		return
	}

	// Parse JSON-RPC message
	var msg types.BaseJSONRPCMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Send to session's request channel
	select {
	case session.requestChan <- &msg:
		w.WriteHeader(http.StatusAccepted)
	case <-session.ctx.Done():
		http.Error(w, "Session closed", http.StatusGone)
	}
}

// Send sends a message to the appropriate SSE client based on context
func (t *sseServerTransport) Send(ctx context.Context, msg *types.BaseJSONRPCMessage) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	// Marshal message
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Try to get session ID from context
	sessionID, ok := ctx.Value("session_id").(string)
	if ok && sessionID != "" {
		// Send to specific session
		if session, exists := t.sessions[sessionID]; exists {
			select {
			case session.messageChan <- data:
				return nil
			default:
				return fmt.Errorf("session buffer full")
			}
		}
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Fallback: broadcast to all sessions (for notifications)
	for _, session := range t.sessions {
		select {
		case session.messageChan <- data:
		default:
			// Session buffer full, skip
		}
	}

	return nil
}

// Close shuts down the server
func (t *sseServerTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	t.cancel()

	// Close all sessions
	for _, session := range t.sessions {
		session.cancel()
	}
	t.sessions = make(map[string]*sseSession)

	// Shutdown HTTP server
	if t.server != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5000)
		defer cancel()
		t.server.Shutdown(shutdownCtx)
	}

	if t.closeHandler != nil {
		t.closeHandler()
	}

	return nil
}

// SetMessageHandler sets the callback for incoming messages
func (t *sseServerTransport) SetMessageHandler(handler func(ctx context.Context, msg *types.BaseJSONRPCMessage)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.messageHandler = handler
}

// SetErrorHandler sets the callback for errors
func (t *sseServerTransport) SetErrorHandler(handler func(error)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.errorHandler = handler
}

// SetCloseHandler sets the callback for when the connection is closed
func (t *sseServerTransport) SetCloseHandler(handler func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closeHandler = handler
}
