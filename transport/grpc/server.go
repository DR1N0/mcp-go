package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/DR1N0/mcp-go/transport"
	pb "github.com/DR1N0/mcp-go/transport/grpc/protogen"
	"github.com/DR1N0/mcp-go/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/structpb"
)

// grpcServerTransport implements ServerTransport for gRPC
type grpcServerTransport struct {
	pb.UnimplementedJSONRPCServiceServer
	mu                 sync.RWMutex
	port               int
	grpcOpts           []grpc.ServerOption
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	grpcServer         *grpc.Server
	listener           net.Listener
	messageHandler     transport.MessageHandler
	errorHandler       transport.ErrorHandler
	closeHandler       transport.CloseHandler
	ctx                context.Context
	cancel             context.CancelFunc
	closed             bool
}

// NewServerTransport creates a new gRPC server transport
func NewServerTransport(opts ...ServerOption) ServerTransport {
	ctx, cancel := context.WithCancel(context.Background())
	s := &grpcServerTransport{
		port:               50051,
		grpcOpts:           make([]grpc.ServerOption, 0),
		unaryInterceptors:  make([]grpc.UnaryServerInterceptor, 0),
		streamInterceptors: make([]grpc.StreamServerInterceptor, 0),
		ctx:                ctx,
		cancel:             cancel,
		closed:             false,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// WithPort sets the port for the gRPC server
func (s *grpcServerTransport) WithPort(port int) ServerTransport {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.port = port
	return s
}

// WithGRPCOptions adds gRPC server options
func (s *grpcServerTransport) WithGRPCOptions(opts ...grpc.ServerOption) ServerTransport {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.grpcOpts = append(s.grpcOpts, opts...)
	return s
}

// WithInterceptor adds a unary interceptor
func (s *grpcServerTransport) WithInterceptor(interceptor grpc.UnaryServerInterceptor) ServerTransport {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.unaryInterceptors = append(s.unaryInterceptors, interceptor)
	return s
}

// WithStreamInterceptor adds a stream interceptor
func (s *grpcServerTransport) WithStreamInterceptor(interceptor grpc.StreamServerInterceptor) ServerTransport {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.streamInterceptors = append(s.streamInterceptors, interceptor)
	return s
}

// Start initializes and starts the gRPC server
func (s *grpcServerTransport) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return fmt.Errorf("transport is closed")
	}

	// Create listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.listener = lis

	// Build server options with interceptors
	opts := make([]grpc.ServerOption, 0, len(s.grpcOpts)+2)
	opts = append(opts, s.grpcOpts...)

	// Chain unary interceptors if any
	if len(s.unaryInterceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(s.unaryInterceptors...))
	}

	// Chain stream interceptors if any
	if len(s.streamInterceptors) > 0 {
		opts = append(opts, grpc.ChainStreamInterceptor(s.streamInterceptors...))
	}

	// Create gRPC server
	s.grpcServer = grpc.NewServer(opts...)
	pb.RegisterJSONRPCServiceServer(s.grpcServer, s)

	// Enable reflection for debugging
	reflection.Register(s.grpcServer)

	s.mu.Unlock()

	// Start server in goroutine
	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			s.mu.RLock()
			errorHandler := s.errorHandler
			s.mu.RUnlock()
			if errorHandler != nil {
				errorHandler(fmt.Errorf("gRPC server error: %w", err))
			}
		}
	}()

	return nil
}

// Transport implements the bidirectional streaming RPC
func (s *grpcServerTransport) Transport(stream pb.JSONRPCService_TransportServer) error {
	ctx := stream.Context()

	// Channel for sending messages to client
	sendChan := make(chan *pb.JSONRPCMessage, 10)
	errChan := make(chan error, 1)

	// Store stream context for sending messages
	streamCtx := context.WithValue(ctx, "grpc_stream", stream)
	streamCtx = context.WithValue(streamCtx, "send_chan", sendChan)

	// Goroutine to send messages to client
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-sendChan:
				if err := stream.Send(msg); err != nil {
					errChan <- err
					return
				}
			}
		}
	}()

	// Receive messages from client
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errChan:
			return err
		default:
			msg, err := stream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}

			// Convert protobuf message to BaseJSONRPCMessage
			baseMsg, err := protoToBase(msg)
			if err != nil {
				s.mu.RLock()
				errorHandler := s.errorHandler
				s.mu.RUnlock()
				if errorHandler != nil {
					errorHandler(fmt.Errorf("failed to convert message: %w", err))
				}
				continue
			}

			// Handle message
			s.mu.RLock()
			handler := s.messageHandler
			s.mu.RUnlock()

			if handler != nil {
				handler(streamCtx, baseMsg)
			}
		}
	}
}

// Send sends a message through the transport
func (s *grpcServerTransport) Send(ctx context.Context, msg *types.BaseJSONRPCMessage) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return fmt.Errorf("transport is closed")
	}

	// Convert to protobuf message
	protoMsg, err := baseToProto(msg)
	if err != nil {
		return fmt.Errorf("failed to convert message: %w", err)
	}

	// Get send channel from context
	sendChan, ok := ctx.Value("send_chan").(chan *pb.JSONRPCMessage)
	if !ok || sendChan == nil {
		return fmt.Errorf("send channel not found in context")
	}

	// Send message
	select {
	case sendChan <- protoMsg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("send channel full")
	}
}

// Close shuts down the server
func (s *grpcServerTransport) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true
	s.cancel()

	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	if s.closeHandler != nil {
		s.closeHandler()
	}

	return nil
}

// SetMessageHandler sets the callback for incoming messages
func (s *grpcServerTransport) SetMessageHandler(handler func(ctx context.Context, msg *types.BaseJSONRPCMessage)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messageHandler = handler
}

// SetErrorHandler sets the callback for errors
func (s *grpcServerTransport) SetErrorHandler(handler func(error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errorHandler = handler
}

// SetCloseHandler sets the callback for when the connection is closed
func (s *grpcServerTransport) SetCloseHandler(handler func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closeHandler = handler
}

// Helper functions for message conversion

func protoToBase(msg *pb.JSONRPCMessage) (*types.BaseJSONRPCMessage, error) {
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

func baseToProto(msg *types.BaseJSONRPCMessage) (*pb.JSONRPCMessage, error) {
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
