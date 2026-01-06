# Transport Selection Guide

This guide helps you choose the right MCP transport for your use case and shows how to use each one effectively.

## Table of Contents

- [Quick Decision Matrix](#quick-decision-matrix)
- [Transport Comparison](#transport-comparison)
- [Detailed Transport Guides](#detailed-transport-guides)
- [Migration Between Transports](#migration-between-transports)

## Quick Decision Matrix

| Your Use Case | Recommended Transport | Why |
|--------------|----------------------|-----|
| CLI tool / command-line app | **Stdio** | Simplest, lowest overhead |
| Claude Desktop integration | **Stdio** | Native support |
| Web application | **SSE** | Real-time updates, persistent connection |
| Microservice / REST API | **Streamable HTTP** or **gRPC** | Stateless, scales horizontally |
| Real-time collaboration | **SSE** or **gRPC** | Push notifications / bidirectional streaming |
| Serverless / Lambda | **Streamable HTTP** | No persistent state required |
| Desktop app with subprocess | **Stdio** | Process lifecycle management |
| Mobile app backend | **Streamable HTTP** or **gRPC** | Works with standard HTTP clients |
| Remote MCP server deployment | **gRPC** | Built for remote RPC, efficient binary protocol |
| Multi-language client support | **gRPC** | Native code generation for many languages |
| High-performance requirements | **gRPC** | Binary protocol with HTTP/2 |
| Enterprise/production deployment | **gRPC** | Battle-tested infrastructure, monitoring |

## Transport Comparison

### Feature Matrix

| Feature | Stdio | SSE | Streamable HTTP |
|---------|-------|-----|----------------|
| **Connection Type** | Process pipes | Persistent HTTP | Request/Response |
| **State Management** | Process-bound | Session-based | Stateless |
| **Latency** | Very Low | Low-Medium | Medium |
| **Scalability** | Single process | Moderate | High |
| **Python Compatible** | ✅ Yes | ✅ Yes | ✅ Yes |
| **Web Browser** | ❌ No | ✅ Yes | ✅ Yes |
| **Server Push** | ❌ No | ✅ Yes | ❌ No |
| **Load Balancing** | ❌ No | ⚠️ Sticky sessions | ✅ Yes |
| **Firewall Friendly** | ✅ Yes | ⚠️ HTTP only | ✅ Yes |

### Performance Characteristics

```
Latency (lower is better):
Stdio:      ████ 1-2ms (process IPC)
SSE:        ████████ 5-10ms (HTTP + persistent conn)
Streamable: ████████████ 10-20ms (HTTP request/response)

Throughput (higher is better):
Stdio:      ████████████ Very High (no network)
SSE:        ████████ Moderate (connection limit)
Streamable: ████████████ High (stateless, horizontal scale)

Resource Usage (lower is better):
Stdio:      ████ Process memory only
SSE:        ████████ Connection + session state
Streamable: ████ Request memory only
```

## Detailed Transport Guides

### Stdio Transport

**Best for**: CLI tools, Claude Desktop, subprocess-based tools

#### How It Works

```
┌─────────┐                    ┌─────────┐
│  Client │                    │  Server │
│ Process │                    │ Process │
└────┬────┘                    └────┬────┘
     │                              │
     │──── stdin (JSON-RPC) ───────→│
     │                              │
     │←─── stdout (JSON-RPC) ───────│
     │                              │
```

#### Server Setup

```go
package main

import (
    "log"
    mcpgo "github.com/DR1N0/mcp-go"
    "github.com/DR1N0/mcp-go/transport/stdio"
)

func main() {
    // Create stdio transport (reads from stdin, writes to stdout)
    transport := stdio.NewServerTransport()
    
    // Create MCP server
    server := mcpgo.NewServer(
        transport,
        mcpgo.WithName("my-tool"),
        mcpgo.WithVersion("1.0.0"),
    )
    
    // Register tools
    server.RegisterTool("echo", "Echoes input", func(args struct{
        Message string `json:"message" jsonschema:"required"`
    }) (*mcpgo.ToolResponse, error) {
        return mcpgo.NewToolResponse(
            mcpgo.NewTextContent("Echo: " + args.Message),
        ), nil
    })
    
    // Start server (blocks)
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}
```

#### Client Setup (Go)

```go
// Client spawns server process
transport := stdio.NewClientTransport(
    "./my-server",  // path to server binary
    []string{},     // command-line args
)

client := mcpgo.NewClient(transport)
defer client.Close()

// Use client
result, err := client.CallTool(ctx, "echo", map[string]interface{}{
    "message": "Hello!",
})
```

#### Client Setup (Python)

```python
from pydantic_ai.mcp import MCPServerStdio

server = MCPServerStdio(
    command="./my-server",
    args=[],
)

# Use with pydantic_ai
from pydantic_ai import Agent
agent = Agent(model, toolsets=[server])
```

#### Claude Desktop Integration

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "my-tool": {
      "command": "/path/to/my-server",
      "args": []
    }
  }
}
```

### SSE Transport

**Best for**: Web applications, real-time features, browser clients

#### How It Works

```
┌─────────┐                              ┌─────────┐
│  Client │                              │  Server │
└────┬────┘                              └────┬────┘
     │                                        │
     │──── GET /mcp/sse (establish) ─────────→│
     │                                        │
     │←──── event: endpoint ──────────────────│
     │      data: /message?session_id=abc     │
     │                                        │
     │──── POST /message?session_id=abc ─────→│
     │      {method: tools/call}              │
     │                                        │
     │←─── event: message ────────────────────│
     │      data: {result: ...}               │
```

#### Server Setup

```go
transport := sse.NewServerTransport(":8001", "/mcp/sse")

server := mcpgo.NewServer(
    transport,
    mcpgo.WithName("sse-server"),
)

// Register tools, prompts, resources...

server.Start()  // Starts HTTP server
```

#### Client Setup (Go)

```go
transport := sse.NewClientTransport("http://localhost:8001/mcp/sse")
client := mcpgo.NewClient(transport)

// Initialize and use
info, err := client.Initialize(ctx)
```

#### Client Setup (Python)

```python
from pydantic_ai.mcp import MCPServerSSE

server = MCPServerSSE("http://localhost:8001/mcp/sse")

# Use with pydantic_ai
agent = Agent(model, toolsets=[server])
```

#### Production Considerations

1. **HTTPS Required**: Always use HTTPS in production
2. **Session Management**: Server maintains session state
3. **Connection Limits**: Monitor open SSE connections
4. **Timeouts**: Configure appropriate read/write timeouts

```go
transport := sse.NewServerTransport(":8001", "/mcp/sse",
    sse.WithReadTimeout(30*time.Second),
    sse.WithWriteTimeout(30*time.Second),
)
```

### Streamable HTTP Transport

**Best for**: Microservices, REST APIs, serverless, stateless deployments

#### How It Works

```
┌─────────┐                              ┌─────────┐
│  Client │                              │  Server │
└────┬────┘                              └────┬────┘
     │                                        │
     │─── POST /mcp ─────────────────────────→│
     │     {method: initialize}               │
     │                                        │
     │←─── 200 OK ────────────────────────────│
     │     {result: {capabilities: ...}}      │
     │                                        │
     │─── POST /mcp ─────────────────────────→│
     │     {method: tools/call}               │
     │                                        │
     │←─── 200 OK ────────────────────────────│
     │     {result: {content: [...]}}         │
```

#### Server Setup

```go
transport := streamable.NewServerTransport(":8000", "/mcp")

server := mcpgo.NewServer(
    transport,
    mcpgo.WithName("api-server"),
)

// Register tools, prompts, resources...

server.Start()
```

#### Client Setup (Go)

```go
transport := streamable.NewClientTransport("http://localhost:8000/mcp")
client := mcpgo.NewClient(transport)

result, err := client.CallTool(ctx, "myTool", args)
```

#### Client Setup (Python)

```python
from pydantic_ai.mcp import MCPServerStreamableHTTP

server = MCPServerStreamableHTTP("http://localhost:8000/mcp")

agent = Agent(model, toolsets=[server])
```

#### Scaling Considerations

1. **Horizontal Scaling**: Add more server instances behind a load balancer
2. **No Session State**: Each request is independent
3. **Connection Pooling**: Reuse HTTP connections

```go
// Client with connection pooling
httpClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 100,
    },
    Timeout: 30 * time.Second,
}

transport := streamable.NewClientTransport(
    "http://api.example.com/mcp",
    streamable.WithHTTPClient(httpClient),
)
```

### gRPC Transport

**Best for**: Remote deployment, high-performance microservices, multi-language support

#### How It Works

```
┌─────────┐         Bidirectional Stream          ┌─────────┐
│  Client │◄─────────────────────────────────────►│  Server │
│         │          gRPC over HTTP/2             │         │
│         │         (Binary Protocol)             │         │
└─────────┘         Port: 50051 (default)         └─────────┘
```

#### Server Setup

```go
import (
    mcpgo "github.com/DR1N0/mcp-go"
    "github.com/DR1N0/mcp-go/transport/grpc"
)

// Create server with gRPC transport (default port 50051)
transport := grpc.NewServerTransport()

server := mcpgo.NewServer(
    transport,
    mcpgo.WithName("grpc-server"),
    mcpgo.WithVersion("1.0.0"),
)

// Register tools, prompts, resources...

server.Serve()  // Starts gRPC server
```

#### Custom Configuration

```go
import (
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
)

// With custom port
transport := grpc.NewServerTransport(
    grpc.WithServerPort(8080),
)

// With TLS
creds, _ := credentials.NewServerTLSFromFile("cert.pem", "key.pem")
transport := grpc.NewServerTransport(
    grpc.WithServerGRPCOptions(grpc.Creds(creds)),
)

// With interceptors (similar to HTTP middleware)
transport := grpc.NewServerTransport().
    WithInterceptor(authInterceptor).
    WithStreamInterceptor(loggingInterceptor)
```

#### Client Setup (Go)

```go
import (
    mcpgo "github.com/DR1N0/mcp-go"
    "github.com/DR1N0/mcp-go/transport/grpc"
)

// Create client transport
transport := grpc.NewClientTransport("localhost:50051")

// Create MCP client
client := mcpgo.NewClient(transport)

// Initialize and use
ctx := context.Background()
initResp, err := client.Initialize(ctx)
```

#### Secure Client Connection

```go
import (
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
)

// With TLS
creds, _ := credentials.NewClientTLSFromFile("ca.pem", "")
transport := grpc.NewClientTransport(
    "myserver.com:50051",
    grpc.WithClientGRPCDialOptions(grpc.WithTransportCredentials(creds)),
)
```

#### Production Deployment

**Docker Example**:
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest
COPY --from=builder /app/server /server
EXPOSE 50051
CMD ["/server"]
```

**Kubernetes Example**:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: mcp-grpc-server
spec:
  ports:
  - port: 50051
    protocol: TCP
  selector:
    app: mcp-grpc-server
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mcp-grpc-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: mcp-grpc-server
  template:
    metadata:
      labels:
        app: mcp-grpc-server
    spec:
      containers:
      - name: server
        image: your-registry/mcp-grpc-server:latest
        ports:
        - containerPort: 50051
```

#### Advantages

1. **Remote Deployment**: Host servers anywhere, not limited to local processes
2. **Binary Protocol**: Efficient HTTP/2-based binary protocol
3. **Load Balancing**: Native support for horizontal scaling
4. **Multi-Language**: Easy to create clients in other languages
5. **Interceptors**: Built-in support for auth, logging, tracing
6. **Streaming**: Native bidirectional streaming support

## HTTP Middleware

**Available on:** Streamable HTTP, SSE

HTTP middleware allows you to add cross-cutting concerns to your MCP servers without modifying your tool/prompt/resource logic. Middleware wraps the HTTP handler chain, enabling features like authentication, logging, CORS, rate limiting, and telemetry.

### Overview

Middleware functions follow the standard Go HTTP pattern:

```go
type HTTPMiddleware func(http.Handler) http.Handler
```

Middleware is chained in **reverse order** (following the Chi router pattern), so the last middleware added becomes the outermost wrapper.

### Basic Usage

```go
import (
    mcpgo "github.com/DR1N0/mcp-go"
    "github.com/DR1N0/mcp-go/transport/streamable"
    "github.com/DR1N0/mcp-go/types"
)

// Create middleware functions
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("Authorization") == "" {
            w.WriteHeader(http.StatusUnauthorized)
            w.Write([]byte("Authorization required"))
            return
        }
        next.ServeHTTP(w, r)
    })
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("[%s] %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

// Apply to transport
transport := streamable.NewServerTransport("/mcp", ":8080").
    WithMiddleware(authMiddleware).
    WithMiddleware(loggingMiddleware)

server := mcpgo.NewServer(transport)
```

### Execution Order

```
Request Flow (with 3 middleware):
1. CORS middleware      (outermost - applied first)
2. Logging middleware
3. Auth middleware      (innermost - applied last)
4. MCP Handler
```

### Common Patterns

#### Authentication (Bearer Token)

```go
func bearerAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if len(token) < 7 || token[:7] != "Bearer " {
            w.WriteHeader(http.StatusUnauthorized)
            w.Write([]byte("Invalid auth token"))
            return
        }
        
        // Validate token (example)
        if !isValidToken(token[7:]) {
            w.WriteHeader(http.StatusUnauthorized)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

#### CORS Headers

```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

#### Rate Limiting

```go
import "golang.org/x/time/rate"

func rateLimitMiddleware(next http.Handler) http.Handler {
    limiter := rate.NewLimiter(10, 100) // 10 req/sec, burst 100
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            w.WriteHeader(http.StatusTooManyRequests)
            w.Write([]byte("Rate limit exceeded"))
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

#### Request ID Tracking

```go
import "github.com/google/uuid"

func requestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := uuid.New().String()
        w.Header().Set("X-Request-ID", requestID)
        
        ctx := context.WithValue(r.Context(), "request_id", requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

#### OpenTelemetry Tracing

```go
import "go.opentelemetry.io/otel"

func telemetryMiddleware(next http.Handler) http.Handler {
    tracer := otel.Tracer("mcp-server")
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx, span := tracer.Start(r.Context(), "mcp.request")
        defer span.End()
        
        span.SetAttributes(
            attribute.String("http.method", r.Method),
            attribute.String("http.path", r.URL.Path),
        )
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Using with SSE Transport

Middleware works identically with SSE:

```go
transport := sse.NewServerTransport("/mcp/sse", ":8001").
    WithMiddleware(authMiddleware).
    WithMiddleware(loggingMiddleware).
    WithMiddleware(corsMiddleware)
```

### Complete Example

See `examples/middleware/` for a complete working example with:
- Bearer token authentication
- Request logging
- CORS headers
- Server and client implementations

```bash
# Run the middleware example
make server-middleware  # Terminal 1
make client-middleware  # Terminal 2
```

### Production Best Practices

1. **Order Matters**: Place auth middleware after logging for better debugging
2. **Error Handling**: Always handle errors gracefully in middleware
3. **Performance**: Avoid blocking operations in middleware
4. **Security**: Validate all inputs, especially auth tokens
5. **Observability**: Add logging and metrics to middleware

## Migration Between Transports

### Code That Doesn't Change

The beauty of mcp-go's architecture is that your tool/prompt/resource registration code **never changes**:

```go
// This code works with ANY transport
server.RegisterTool("analyze", "Analyzes data", func(args struct{
    Data string `json:"data" jsonschema:"required"`
}) (*mcpgo.ToolResponse, error) {
    // Your logic here
    return mcpgo.NewToolResponse(
        mcpgo.NewTextContent("Analysis result"),
    ), nil
})
```

### Switching Transports

**From Stdio to SSE**:

```go
// Before (Stdio)
transport := stdio.NewServerTransport()

// After (SSE)
transport := sse.NewServerTransport(":8001", "/mcp/sse")

// Everything else stays the same!
server := mcpgo.NewServer(transport, ...)
```

**From SSE to Streamable HTTP**:

```go
// Before (SSE)
transport := sse.NewServerTransport(":8001", "/mcp/sse")

// After (Streamable)
transport := streamable.NewServerTransport(":8000", "/mcp")

// Everything else stays the same!
server := mcpgo.NewServer(transport, ...)
```

### Multi-Transport Support

Run multiple transports simultaneously:

```go
func main() {
    // Create server logic once
    createServer := func(transport types.Transport) *mcpgo.Server {
        server := mcpgo.NewServer(transport, mcpgo.WithName("multi-transport"))
        
        // Register tools (shared across all transports)
        server.RegisterTool("echo", "Echoes", echoHandler)
        
        return server
    }
    
    // Start on multiple transports
    go createServer(stdio.NewServerTransport()).Start()
    go createServer(sse.NewServerTransport(":8001", "/sse")).Start()
    createServer(streamable.NewServerTransport(":8000", "/http")).Start()
}
```

## Troubleshooting

### Common Issues

#### Stdio: "EOF" or "broken pipe"

**Cause**: Server process terminated or stdin/stdout closed

**Solution**: 
- Check server logs
- Ensure server doesn't write to stdout except JSON-RPC
- Verify server process hasn't crashed

#### SSE: "Connection timeout"

**Cause**: Client didn't receive endpoint event

**Solution**:
- Check server sent `endpoint` event immediately after connection
- Verify no firewalls blocking SSE
- Check server logs for connection establishment

#### HTTP: "404 Not Found"

**Cause**: Wrong endpoint path

**Solution**:
- Verify server path matches client path
- Check server is actually running
- Confirm port is correct

### Debug Mode

Enable verbose logging:

```go
// Set environment variable
os.Setenv("MCP_DEBUG", "true")

// Or use custom logger
server := mcpgo.NewServer(
    transport,
    mcpgo.WithLogger(yourLogger),
)
```

### Testing Transports

Use the mock transport for unit tests:

```go
import "github.com/DR1N0/mcp-go/transport"

mock := transport.NewMock()
server := mcpgo.NewServer(mock)

// Test without network
mock.SimulateReceive(ctx, testMessage)
```

## Best Practices

### General

1. **Initialize Once**: Create server/client once, reuse for all requests
2. **Handle Errors**: Always check and handle errors from MCP calls
3. **Use Contexts**: Pass appropriate contexts with timeouts
4. **Clean Shutdown**: Always call `Close()` when done

### Transport-Specific

**Stdio**:
- Build server binary separately
- Use absolute paths in Claude Desktop config
- Don't write debug logs to stdout (use stderr or files)

**SSE**:
- Use HTTPS in production
- Implement connection retry logic
- Monitor connection count
- Set appropriate timeouts

**Streamable HTTP**:
- Use connection pooling
- Implement retries with exponential backoff
- Consider rate limiting
- Use health check endpoints

## Performance Tuning

### Stdio

```go
// Pre-compile and cache server binary
// Use smallest possible Docker image if containerized
// Minimize startup time
```

### SSE

```go
transport := sse.NewServerTransport(
    ":8001", "/mcp/sse",
    sse.WithReadTimeout(30*time.Second),
    sse.WithWriteTimeout(30*time.Second),
    sse.WithMaxConnections(1000),
)
```

### Streamable HTTP

```go
// Increase connection pool
httpClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        1000,
        MaxIdleConnsPerHost: 100,
        IdleConnTimeout:     90 * time.Second,
    },
}

transport := streamable.NewClientTransport(
    url,
    streamable.WithHTTPClient(httpClient),
)
