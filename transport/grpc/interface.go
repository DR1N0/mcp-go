package grpc

import (
	"github.com/DR1N0/mcp-go/transport"
	"google.golang.org/grpc"
)

// ServerTransport is a gRPC-specific server transport
type ServerTransport interface {
	transport.ServerTransport
	// WithPort sets the port for the gRPC server
	WithPort(port int) ServerTransport
	// WithGRPCOptions adds gRPC server options
	WithGRPCOptions(opts ...grpc.ServerOption) ServerTransport
	// WithInterceptor adds a unary interceptor (similar to HTTP middleware)
	WithInterceptor(interceptor grpc.UnaryServerInterceptor) ServerTransport
	// WithStreamInterceptor adds a stream interceptor
	WithStreamInterceptor(interceptor grpc.StreamServerInterceptor) ServerTransport
}

// ClientTransport is a gRPC-specific client transport
type ClientTransport interface {
	transport.ClientTransport
	// WithGRPCDialOptions adds gRPC dial options
	WithGRPCDialOptions(opts ...grpc.DialOption) ClientTransport
}

// ServerOption configures a gRPC server transport
type ServerOption func(*grpcServerTransport)

// ClientOption configures a gRPC client transport
type ClientOption func(*grpcClientTransport)

// WithServerPort sets the server port
func WithServerPort(port int) ServerOption {
	return func(s *grpcServerTransport) {
		s.port = port
	}
}

// WithServerGRPCOptions adds gRPC server options
func WithServerGRPCOptions(opts ...grpc.ServerOption) ServerOption {
	return func(s *grpcServerTransport) {
		s.grpcOpts = append(s.grpcOpts, opts...)
	}
}

// WithClientGRPCDialOptions adds gRPC dial options
func WithClientGRPCDialOptions(opts ...grpc.DialOption) ClientOption {
	return func(c *grpcClientTransport) {
		c.dialOpts = append(c.dialOpts, opts...)
	}
}
