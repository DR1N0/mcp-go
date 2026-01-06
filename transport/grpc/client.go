package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/DR1N0/mcp-go/transport"
	pb "github.com/DR1N0/mcp-go/transport/grpc/protogen"
	"github.com/DR1N0/mcp-go/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
)

// grpcClientTransport implements ClientTransport for gRPC
type grpcClientTransport struct {
	mu             sync.RWMutex
	address        string
	dialOpts       []grpc.DialOption
	conn           *grpc.ClientConn
	client         pb.JSONRPCServiceClient
	stream         pb.JSONRPCService_TransportClient
	messageHandler transport.MessageHandler
	errorHandler   transport.ErrorHandler
	closeHandler   transport.CloseHandler
	ctx            context.Context
	cancel         context.CancelFunc
	closed         bool
	sendChan       chan *pb.JSONRPCMessage
}

// NewClientTransport creates a new gRPC client transport
func NewClientTransport(address string, opts ...ClientOption) transport.ClientTransport {
	ctx, cancel := context.WithCancel(context.Background())
	c := &grpcClientTransport{
		address:  address,
		dialOpts: []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
		ctx:      ctx,
		cancel:   cancel,
		closed:   false,
		sendChan: make(chan *pb.JSONRPCMessage, 10),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithGRPCDialOptions adds gRPC dial options
func (c *grpcClientTransport) WithGRPCDialOptions(opts ...grpc.DialOption) transport.ClientTransport {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dialOpts = append(c.dialOpts, opts...)
	return c
}

// Start initializes the client connection
func (c *grpcClientTransport) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return fmt.Errorf("transport is closed")
	}
	c.mu.Unlock()

	// Create connection with timeout
	connCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(c.address, c.dialOpts...)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.client = pb.NewJSONRPCServiceClient(conn)
	c.mu.Unlock()

	// Establish stream
	stream, err := c.client.Transport(c.ctx)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create stream: %w", err)
	}

	c.mu.Lock()
	c.stream = stream
	c.mu.Unlock()

	// Start send goroutine
	go c.sendLoop()

	// Start receive goroutine
	go c.receiveLoop()

	// Wait for connection to be ready
	select {
	case <-connCtx.Done():
		return fmt.Errorf("connection timeout")
	case <-time.After(100 * time.Millisecond):
		// Give it a moment to establish
	}

	return nil
}

// sendLoop handles sending messages to the server
func (c *grpcClientTransport) sendLoop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.sendChan:
			c.mu.RLock()
			stream := c.stream
			c.mu.RUnlock()

			if stream == nil {
				continue
			}

			if err := stream.Send(msg); err != nil {
				c.mu.RLock()
				errorHandler := c.errorHandler
				c.mu.RUnlock()
				if errorHandler != nil {
					errorHandler(fmt.Errorf("send error: %w", err))
				}
				return
			}
		}
	}
}

// receiveLoop handles receiving messages from the server
func (c *grpcClientTransport) receiveLoop() {
	for {
		c.mu.RLock()
		stream := c.stream
		c.mu.RUnlock()

		if stream == nil {
			return
		}

		msg, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			c.mu.RLock()
			errorHandler := c.errorHandler
			c.mu.RUnlock()
			if errorHandler != nil {
				errorHandler(fmt.Errorf("receive error: %w", err))
			}
			return
		}

		// Convert protobuf message to BaseJSONRPCMessage
		baseMsg, err := protoToBase(msg)
		if err != nil {
			c.mu.RLock()
			errorHandler := c.errorHandler
			c.mu.RUnlock()
			if errorHandler != nil {
				errorHandler(fmt.Errorf("failed to convert message: %w", err))
			}
			continue
		}

		// Handle message
		c.mu.RLock()
		handler := c.messageHandler
		c.mu.RUnlock()

		if handler != nil {
			handler(c.ctx, baseMsg)
		}
	}
}

// Send sends a message through the transport
func (c *grpcClientTransport) Send(ctx context.Context, msg *types.BaseJSONRPCMessage) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("transport is closed")
	}

	// Convert to protobuf message
	protoMsg, err := baseToProto(msg)
	if err != nil {
		return fmt.Errorf("failed to convert message: %w", err)
	}

	// Send message
	select {
	case c.sendChan <- protoMsg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("send channel full")
	}
}

// Close shuts down the client
func (c *grpcClientTransport) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	c.cancel()

	if c.stream != nil {
		c.stream.CloseSend()
	}

	if c.conn != nil {
		c.conn.Close()
	}

	close(c.sendChan)

	if c.closeHandler != nil {
		c.closeHandler()
	}

	return nil
}

// SetMessageHandler sets the callback for incoming messages
func (c *grpcClientTransport) SetMessageHandler(handler func(ctx context.Context, msg *types.BaseJSONRPCMessage)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messageHandler = handler
}

// SetErrorHandler sets the callback for errors
func (c *grpcClientTransport) SetErrorHandler(handler func(error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errorHandler = handler
}

// SetCloseHandler sets the callback for when the connection is closed
func (c *grpcClientTransport) SetCloseHandler(handler func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closeHandler = handler
}

// Helper functions (shared with server.go but repeated here for clarity)

func protoToBaseClient(msg *pb.JSONRPCMessage) (*types.BaseJSONRPCMessage, error) {
	base := &types.BaseJSONRPCMessage{
		JSONRPC: msg.Jsonrpc,
		Method:  msg.Method,
	}

	// Handle ID
	switch id := msg.Id.(type) {
	case *pb.JSONRPCMessage_IdString:
		base.ID = id.IdString
	case *pb.JSONRPCMessage_IdNumber:
		base.ID = id.IdNumber
	}

	// Handle params
	if msg.Params != nil {
		paramsStruct, err := structpb.NewStruct(msg.Params.AsMap())
		if err != nil {
			return nil, fmt.Errorf("failed to create params struct: %w", err)
		}
		params, err := paramsStruct.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		base.Params = params
	}

	// Handle result
	if msg.Result != nil {
		resultStruct, err := structpb.NewStruct(msg.Result.AsMap())
		if err != nil {
			return nil, fmt.Errorf("failed to create result struct: %w", err)
		}
		result, err := resultStruct.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		base.Result = result
	}

	// Handle error
	if msg.Error != nil {
		base.Error = &types.RPCError{
			Code:    int(msg.Error.Code),
			Message: msg.Error.Message,
		}
		if msg.Error.Data != nil {
			base.Error.Data = msg.Error.Data.AsMap()
		}
	}

	return base, nil
}

func baseToProtoClient(msg *types.BaseJSONRPCMessage) (*pb.JSONRPCMessage, error) {
	proto := &pb.JSONRPCMessage{
		Jsonrpc: msg.JSONRPC,
		Method:  msg.Method,
	}

	// Handle ID
	switch id := msg.ID.(type) {
	case string:
		proto.Id = &pb.JSONRPCMessage_IdString{IdString: id}
	case int:
		proto.Id = &pb.JSONRPCMessage_IdNumber{IdNumber: int64(id)}
	case int64:
		proto.Id = &pb.JSONRPCMessage_IdNumber{IdNumber: id}
	case float64:
		proto.Id = &pb.JSONRPCMessage_IdNumber{IdNumber: int64(id)}
	}

	// Handle params
	if len(msg.Params) > 0 {
		var paramsMap map[string]interface{}
		if err := json.Unmarshal(msg.Params, &paramsMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal params: %w", err)
		}
		params, err := structpb.NewStruct(paramsMap)
		if err != nil {
			return nil, fmt.Errorf("failed to create params struct: %w", err)
		}
		proto.Params = params
	}

	// Handle result
	if len(msg.Result) > 0 {
		var resultMap map[string]interface{}
		if err := json.Unmarshal(msg.Result, &resultMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
		result, err := structpb.NewStruct(resultMap)
		if err != nil {
			return nil, fmt.Errorf("failed to create result struct: %w", err)
		}
		proto.Result = result
	}

	// Handle error
	if msg.Error != nil {
		proto.Error = &pb.JSONRPCError{
			Code:    int32(msg.Error.Code),
			Message: msg.Error.Message,
		}
		if msg.Error.Data != nil {
			data, err := structpb.NewStruct(msg.Error.Data.(map[string]interface{}))
			if err != nil {
				return nil, fmt.Errorf("failed to create error data struct: %w", err)
			}
			proto.Error.Data = data
		}
	}

	return proto, nil
}
