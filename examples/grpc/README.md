# gRPC Transport Example

This example demonstrates how to use the gRPC transport for MCP (Model Context Protocol) communication. gRPC enables remote server hosting and provides better scalability compared to local-only transports.

## Features

- **Bidirectional Streaming**: Full-duplex communication between client and server
- **Remote Hosting**: Deploy MCP servers anywhere, not just locally
- **Production Ready**: Built on gRPC's battle-tested infrastructure
- **Interceptor Support**: Add authentication, logging, and other middleware
- **TLS Support**: Secure communication with transport layer security

## Architecture

```
┌─────────────┐         gRPC Stream           ┌─────────────┐
│             │◄─────────────────────────────►│             │
│   Client    │   Bidirectional Streaming     │   Server    │
│             │         Port: 50051           │             │
└─────────────┘                               └─────────────┘
```

## Running the Example

### 1. Start the Server

```bash
# From the project root
go run ./examples/grpc/server

# Or from the example directory
cd examples/grpc/server
go run main.go
```

Expected output:
```
✅ MCP gRPC Server started successfully!
   - gRPC endpoint: localhost:50051
   - Protocol: bidirectional streaming
   - Tools: echo, add, get_weather
   - Prompts: greeting
   - Resources: config://server, lyrics://sample-song

Press Ctrl+C to stop...
```

### 2. Run the Client

In a separate terminal:

```bash
# From the project root
go run ./examples/grpc/clients/go

# Or from the example directory
cd examples/grpc/clients/go
go run main.go
```

## Server Implementation

```go
import (
    mcpgo "github.com/DR1N0/mcp-go"
    "github.com/DR1N0/mcp-go/transport/grpc"
)

// Create server with gRPC transport (default port 50051)
server := mcpgo.NewServer(
    grpc.NewServerTransport(),
    mcpgo.WithName("my-server"),
    mcpgo.WithVersion("1.0.0"),
)

// Register tools, prompts, and resources...

// Start the server
if err := server.Serve(); err != nil {
    log.Fatal(err)
}
```

### Custom Configuration

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

// With interceptors (for auth, logging, etc.)
transport := grpc.NewServerTransport().
    WithInterceptor(myAuthInterceptor).
    WithStreamInterceptor(myLoggingInterceptor)
```

## Client Implementation

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

### Secure Client Connection

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

## Available Tools

- **echo**: Echoes back your message
- **add**: Adds two numbers
- **get_weather**: Fetches weather data for given coordinates

## Available Prompts

- **greeting**: Generates a personalized greeting

## Available Resources

- **config://server**: Server configuration
- **lyrics://sample-song**: Sample song lyrics

## Production Deployment

### Using TLS (Recommended)

Generate certificates:
```bash
# Self-signed for testing
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
```

Server with TLS:
```go
creds, err := credentials.NewServerTLSFromFile("cert.pem", "key.pem")
if err != nil {
    log.Fatal(err)
}

transport := grpc.NewServerTransport(
    grpc.WithServerGRPCOptions(grpc.Creds(creds)),
)
```

Client with TLS:
```go
creds, err := credentials.NewClientTLSFromFile("cert.pem", "")
if err != nil {
    log.Fatal(err)
}

transport := grpc.NewClientTransport(
    "myserver.com:50051",
    grpc.WithClientGRPCDialOptions(grpc.WithTransportCredentials(creds)),
)
```

### Docker Deployment

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./examples/grpc/server

FROM alpine:latest
COPY --from=builder /app/server /server
EXPOSE 50051
CMD ["/server"]
```

### Kubernetes Deployment

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

## Advantages of gRPC Transport

1. **Remote Deployment**: Host servers anywhere, not limited to local processes
2. **Scalability**: Load balance across multiple server instances
3. **Performance**: Efficient binary protocol with HTTP/2
4. **Infrastructure**: Reuse existing gRPC tooling (auth, monitoring, tracing)
5. **Cross-Language**: Easy to create clients in other languages
6. **Streaming**: Native bidirectional streaming support

## Comparison with Other Transports

| Feature | stdio | SSE | Streamable HTTP | gRPC |
|---------|-------|-----|-----------------|------|
| Remote Hosting | ❌ | ✅ | ✅ | ✅ |
| Bidirectional | ✅ | ❌ | ✅ | ✅ |
| Load Balancing | ❌ | ⚠️ | ⚠️ | ✅ |
| Binary Protocol | ❌ | ❌ | ❌ | ✅ |
| Interceptors | ❌ | ✅ | ✅ | ✅ |
| Setup Complexity | Low | Medium | Medium | Medium |

## Troubleshooting

### Connection Refused
- Ensure server is running: `netstat -an | grep 50051`
- Check firewall rules
- Verify correct address and port

### TLS Errors
- Ensure certificates are valid
- Check certificate paths
- Verify hostname matches certificate

### Performance Issues
- Enable connection pooling
- Use compression: `grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip"))`
- Monitor with gRPC metrics

## Learn More

- [gRPC Documentation](https://grpc.io/docs/)
- [MCP Specification](https://modelcontextprotocol.io)
- [Transport Guide](../../docs/transport-guide.md)
