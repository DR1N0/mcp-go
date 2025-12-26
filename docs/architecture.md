# MCP-Go Architecture

This document describes the architecture and design principles of mcp-go, a Go implementation of the Model Context Protocol (MCP).

## Table of Contents

- [Overview](#overview)
- [Core Design Principles](#core-design-principles)
- [System Architecture](#system-architecture)
- [Component Details](#component-details)
- [Message Flow](#message-flow)
- [Extension Points](#extension-points)

## Overview

mcp-go is designed as a **transport-agnostic, layered architecture** that separates protocol handling from transport implementation. This enables the same MCP server/client logic to work seamlessly across different communication mechanisms.

```
┌─────────────────────────────────────────────────┐
│              Application Layer                  │
│         (Tools, Prompts, Resources)             │
└─────────────────────────────────────────────────┘
                      ↕
┌─────────────────────────────────────────────────┐
│           MCP Server / Client Layer             │
│       (Business Logic, Registration)            │
└─────────────────────────────────────────────────┘
                      ↕
┌─────────────────────────────────────────────────┐
│            Protocol Layer                       │
│         (JSON-RPC 2.0 Processing)               │
└─────────────────────────────────────────────────┘
                      ↕
┌─────────────────────────────────────────────────┐
│            Transport Layer                      │
│    (HTTP, SSE, Stdio - Pluggable)               │
└─────────────────────────────────────────────────┘
```

## Core Design Principles

### 1. Interface-First Design

Every layer is defined by clear interfaces before implementation:

```go
// Transport interface - any transport must implement this
type Transport interface {
    Start(ctx context.Context) error
    Send(ctx context.Context, msg *BaseJSONRPCMessage) error
    Close() error
    SetMessageHandler(handler func(ctx context.Context, msg *BaseJSONRPCMessage))
    SetErrorHandler(handler func(error))
    SetCloseHandler(handler func())
}
```

### 2. Transport Agnostic

Server and client code doesn't know or care about the transport mechanism:

```go
// Create server with ANY transport
server := NewServer(transport)

// Works with stdio
server := NewServer(stdio.NewServerTransport())

// Works with SSE
server := NewServer(sse.NewServerTransport(":8001"))

// Works with HTTP
server := NewServer(streamable.NewServerTransport(":8000"))
```

### 3. Type Safety

Reflection-based JSON schema generation ensures compile-time type safety:

```go
type MyToolArgs struct {
    Name string `json:"name" jsonschema:"required,description=User name"`
    Age  int    `json:"age" jsonschema:"description=User age"`
}

// Schema automatically generated from struct
server.RegisterTool("myTool", "Description", func(args MyToolArgs) (*ToolResponse, error) {
    // args is already properly typed
    return NewToolResponse(NewTextContent("Hello " + args.Name)), nil
})
```

### 4. Polyglot Support

First-class Python compatibility through standard MCP protocol:

- Works with pydantic_ai out of the box
- Compatible with MCP SDK
- Standard JSON-RPC 2.0 messaging

## System Architecture

### High-Level Component Diagram

```
┌──────────────────────────────────────────────────────────┐
│                      Application                         │
│  ┌────────────┐  ┌────────────┐  ┌──────────────┐        │
│  │   Tools    │  │  Prompts   │  │  Resources   │        │
│  └────────────┘  └────────────┘  └──────────────┘        │
└──────────────────────────────────────────────────────────┘
                         ↕
┌──────────────────────────────────────────────────────────┐
│                    MCP Layer                             │
│  ┌────────────────────────────────────────────────────┐  │
│  │  Server                    Client                  │  │
│  │  - RegisterTool()          - ListTools()           │  │
│  │  - RegisterPrompt()        - CallTool()            │  │
│  │  - RegisterResource()      - GetPrompt()           │  │
│  │  - CallTool()              - ReadResource()        │  │
│  └────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
                         ↕
┌──────────────────────────────────────────────────────────┐
│                  Protocol Layer                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │  JSON-RPC 2.0 Protocol                             │  │
│  │  - Request/Response Correlation                    │  │
│  │  - Notification Handling                           │  │
│  │  - Error Handling                                  │  │
│  │  - ID Normalization                                │  │
│  └────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
                         ↕
┌──────────────────────────────────────────────────────────┐
│                  Transport Layer                         │
│  ┌───────────┐  ┌─────────┐  ┌──────────────────────┐    │
│  │  Stdio    │  │   SSE   │  │  Streamable HTTP     │    │
│  │  Transport│  │Transport│  │     Transport        │    │
│  └───────────┘  └─────────┘  └──────────────────────┘    │
└──────────────────────────────────────────────────────────┘
```

## Component Details

### Server Component

**Location**: `server.go`

**Responsibilities**:
- Tool/Prompt/Resource registration
- Request routing
- Response handling
- Schema generation

**Key Features**:
```go
// Automatic schema generation
server.RegisterTool("add", "Adds numbers", func(args struct{
    A int `json:"a" jsonschema:"required"`
    B int `json:"b" jsonschema:"required"`
}) (*ToolResponse, error) {
    return NewToolResponse(NewTextContent(fmt.Sprintf("%d", args.A + args.B))), nil
})

// Dynamic content generation
server.RegisterPrompt("greeting", "Greeting", func(args struct{
    Name string `json:"name" jsonschema:"required"`
}) (*PromptResponse, error) {
    return NewPromptResponse(
        "Personalized greeting",
        NewPromptMessage(NewTextContent("Hello "+args.Name), RoleAssistant),
    ), nil
})
```

### Client Component

**Location**: `client.go`

**Responsibilities**:
- Server initialization
- Tool invocation
- Prompt retrieval
- Resource reading

**Key Features**:
```go
client := NewClient(transport)

// Initialize connection
serverInfo, err := client.Initialize(ctx)

// Call tools
response, err := client.CallTool(ctx, "toolName", args)

// Get prompts
prompt, err := client.GetPrompt(ctx, "promptName", args)

// Read resources
resource, err := client.ReadResource(ctx, "resource://uri")
```

### Protocol Layer

**Location**: `protocol/protocol.go`

**Responsibilities**:
- JSON-RPC 2.0 message formatting
- Request/response correlation
- ID normalization (handles int/float/string IDs)
- Error marshaling

**Message Types**:

```go
// Request
{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {...}
}

// Response
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {...}
}

// Notification (no ID)
{
    "jsonrpc": "2.0",
    "method": "notification",
    "params": {...}
}

// Error
{
    "jsonrpc": "2.0",
    "id": 1,
    "error": {
        "code": -32600,
        "message": "Invalid Request"
    }
}
```

### Transport Layer

**Location**: `transport/`

Three implementations with identical interface:

#### 1. Stdio Transport

**Use Case**: CLI tools, subprocess communication, Claude Desktop

```go
// Server spawned by client
transport := stdio.NewServerTransport()

// Client spawns server process
transport := stdio.NewClientTransport(
    "/path/to/server",
    []string{"--arg1", "value1"},
)
```

**Communication**:
- Messages sent via stdin/stdout
- Line-delimited JSON-RPC
- Synchronous, process-bound

#### 2. SSE Transport

**Use Case**: Web applications, real-time updates

```go
// Server
transport := sse.NewServerTransport(":8001", "/mcp/sse")

// Client
transport := sse.NewClientTransport("http://localhost:8001/mcp/sse")
```

**Communication**:
- Server-Sent Events for server → client
- HTTP POST for client → server
- Session-based with query parameters
- Bi-directional, persistent connection

#### 3. Streamable HTTP Transport

**Use Case**: Stateless services, microservices, REST APIs

```go
// Server
transport := streamable.NewServerTransport(":8000", "/mcp")

// Client
transport := streamable.NewClientTransport("http://localhost:8000/mcp")
```

**Communication**:
- HTTP POST for all messages
- Stateless request/response
- No persistent connection

## Message Flow

### Tool Invocation Flow (SSE Example)

```
Client                   SSE Transport              Protocol              Server
  │                            │                        │                    │
  │─── CallTool("echo", args)──→                        │                    │
  │                            │                        │                    │
  │                            │──── POST /message ────→│                    │
  │                            │   {method: tools/call} │                    │
  │                            │                        │                    │
  │                            │                        │── Route Request ──→│
  │                            │                        │                    │
  │                            │                        │                    │─── Execute Tool
  │                            │                        │                    │
  │                            │                        │←── Tool Response ──│
  │                            │                        │                    │
  │                            │←─── SSE Event ─────────│                    │
  │                            │   event: message       │                    │
  │                            │   data: {result: ...}  │                    │
  │                            │                        │                    │
  │←── ToolResponse ───────────│                        │                    │
  │                            │                        │                    │
```

### Server Initialization Flow

```
1. Client creates transport
2. Client creates MCP client with transport
3. Client calls Initialize()
4. Protocol layer sends initialize request
5. Transport sends message to server
6. Server transport receives message
7. Server protocol layer routes to initialize handler
8. Server responds with capabilities + server info
9. Response flows back through layers
10. Client receives ServerInfo
```

## Extension Points

### Adding a New Transport

1. **Implement Transport Interface**:

```go
type MyTransport struct {
    // Your fields
}

func (t *MyTransport) Start(ctx context.Context) error { ... }
func (t *MyTransport) Send(ctx context.Context, msg *BaseJSONRPCMessage) error { ... }
func (t *MyTransport) Close() error { ... }
func (t *MyTransport) SetMessageHandler(handler func(ctx context.Context, msg *BaseJSONRPCMessage)) { ... }
func (t *MyTransport) SetErrorHandler(handler func(error)) { ... }
func (t *MyTransport) SetCloseHandler(handler func()) { ... }
```

2. **Handle Message Flow**:
   - Call `messageHandler` when receiving messages
   - Call `errorHandler` on errors
   - Call `closeHandler` on close

3. **Use with Server/Client**:

```go
server := NewServer(MyTransport{...})
client := NewClient(MyTransport{...})
```

### Adding Custom Tool Types

Create strongly-typed tool handlers:

```go
type FileReadArgs struct {
    Path string `json:"path" jsonschema:"required,description=File path to read"`
}

func ReadFileToolHandler(args FileReadArgs) (*ToolResponse, error) {
    content, err := os.ReadFile(args.Path)
    if err != nil {
        return nil, err
    }
    return NewToolResponse(NewTextContent(string(content))), nil
}

server.RegisterTool("readFile", "Reads a file", ReadFileToolHandler)
```

### Testing with Mock Transport

Use the provided mock transport for testing:

```go
mock := transport.NewMock()
client := NewClient(mock)

// Simulate responses
go func() {
    msgs := mock.GetSentMessages()
    mock.SimulateReceive(ctx, response)
}()

// Test client behavior
result, err := client.CallTool(ctx, "test", args)
```

## Performance Considerations

### Transport Selection

- **Stdio**: Lowest latency, tightly coupled
- **SSE**: Moderate latency, persistent connection overhead
- **HTTP**: Higher latency, stateless scales better

### Schema Generation

- Schemas cached per tool/prompt
- Reflection cost paid once at registration
- Zero runtime overhead after registration

### Message Correlation

- Lock-free for reads (RLock)
- Minimal locking for writes
- O(1) request/response matching

## Security Considerations

### Transport Security

- **Stdio**: Inherits process permissions
- **SSE**: Use HTTPS, validate origins
- **HTTP**: Use HTTPS, implement authentication

### Input Validation

- JSON schema validation at protocol layer
- Type checking via Go structs
- User-provided validators possible

### Resource Access

- Implement authorization in resource handlers
- Validate URIs before access
- Sandbox file system access

## Future Architecture

Planned additions:

1. **Sampling Support**: Allow servers to request LLM completions
2. **Roots Support**: Multi-directory workspaces
3. **Logging**: Structured logging throughout
4. **Metrics**: Prometheus-compatible metrics
5. **Middleware**: Request/response interceptors
