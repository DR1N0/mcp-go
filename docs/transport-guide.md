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
| Microservice / REST API | **Streamable HTTP** | Stateless, scales horizontally |
| Real-time collaboration | **SSE** | Push notifications built-in |
| Serverless / Lambda | **Streamable HTTP** | No persistent state required |
| Desktop app with subprocess | **Stdio** | Process lifecycle management |
| Mobile app backend | **Streamable HTTP** | Works with standard HTTP clients |

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
     │─── GET /mcp/sse (establish) ─────────→│
     │                                        │
     │←─── event: endpoint ──────────────────│
     │     data: /message?session_id=abc     │
     │                                        │
     │─── POST /message?session_id=abc ─────→│
     │     {method: tools/call}              │
     │                                        │
     │←─── event: message ────────────────────│
     │     data: {result: ...}               │
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
     │─── POST /mcp ────────────────────────→│
     │     {method: initialize}              │
     │                                        │
     │←─── 200 OK ────────────────────────────│
     │     {result: {capabilities: ...}}     │
     │                                        │
     │─── POST /mcp ────────────────────────→│
     │     {method: tools/call}              │
     │                                        │
     │←─── 200 OK ────────────────────────────│
     │     {result: {content: [...]}}        │
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
