# API Reference

Complete API reference for mcp-go.

## Table of Contents

- [Server API](#server-api)
- [Client API](#client-api)
- [Types](#types)
- [Transport Interface](#transport-interface)
- [Schema Generation](#schema-generation)

## Server API

### Creating a Server

```go
func NewServer(transport types.Transport, opts ...ServerOption) *Server
```

**Parameters**:
- `transport`: Any implementation of `types.Transport` interface
- `opts`: Optional configuration functions

**Server Options**:

```go
// Set server name
func WithName(name string) ServerOption

// Set server version
func WithVersion(version string) ServerOption

// Set custom logger
func WithLogger(logger Logger) ServerOption
```

**Example**:

```go
server := mcpgo.NewServer(
    transport,
    mcpgo.WithName("my-server"),
    mcpgo.WithVersion("1.0.0"),
)
```

### Registering Tools

```go
func (s *Server) RegisterTool(name, description string, handler interface{}) error
```

**Parameters**:
- `name`: Unique tool identifier
- `description`: Human-readable description
- `handler`: Function with signature `func(Args) (*ToolResponse, error)`

**Handler Signature**:

```go
// Args is any struct with json/jsonschema tags
type MyArgs struct {
    Input string `json:"input" jsonschema:"required,description=Input text"`
}

func myHandler(args MyArgs) (*ToolResponse, error) {
    return mcpgo.NewToolResponse(
        mcpgo.NewTextContent("Result"),
    ), nil
}
```

**Returns**: Error if tool name already registered

### Registering Prompts

```go
func (s *Server) RegisterPrompt(name, description string, handler interface{}) error
```

**Handler Signature**:

```go
func myPromptHandler(args MyArgs) (*PromptResponse, error) {
    return mcpgo.NewPromptResponse(
        "description",
        mcpgo.NewPromptMessage(
            mcpgo.NewTextContent("content"),
            mcpgo.RoleAssistant,
        ),
    ), nil
}
```

### Registering Resources

```go
func (s *Server) RegisterResource(
    uri, name, description, mimeType string,
    handler interface{},
) error
```

**Handler Signature**:

```go
func myResourceHandler() (*ResourceResponse, error) {
    return mcpgo.NewResourceResponse(
        mcpgo.NewTextResource(uri, content, mimeType),
    ), nil
}
```

### Starting Server

```go
func (s *Server) Start() error
```

Starts the server (blocks). Returns error if server fails to start.

## Client API

### Creating a Client

```go
func NewClient(transport types.Transport) *Client
```

### Initialize Connection

```go
func (c *Client) Initialize(ctx context.Context) (*ServerInfo, error)
```

**Returns**: `ServerInfo` containing server name, version, and capabilities

### Tool Operations

#### List Tools

```go
func (c *Client) ListTools(ctx context.Context, cursor *string) (*ToolsResponse, error)
```

#### Call Tool

```go
func (c *Client) CallTool(
    ctx context.Context,
    name string,
    arguments interface{},
) (*ToolResponse, error)
```

**Example**:

```go
args := map[string]interface{}{
    "input": "test",
}
result, err := client.CallTool(ctx, "myTool", args)
```

### Prompt Operations

#### List Prompts

```go
func (c *Client) ListPrompts(ctx context.Context, cursor *string) (*ListPromptsResponse, error)
```

#### Get Prompt

```go
func (c *Client) GetPrompt(
    ctx context.Context,
    name string,
    arguments interface{},
) (*PromptResponse, error)
```

### Resource Operations

#### List Resources

```go
func (c *Client) ListResources(ctx context.Context, cursor *string) (*ListResourcesResponse, error)
```

#### Read Resource

```go
func (c *Client) ReadResource(
    ctx context.Context,
    uri string,
) (*ResourceResponse, error)
```

### Utility

#### Ping

```go
func (c *Client) Ping(ctx context.Context) error
```

#### Close

```go
func (c *Client) Close() error
```

## Types

### Tool Types

```go
type Tool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
}

type ToolResponse struct {
    Content []Content `json:"content"`
    IsError bool      `json:"isError,omitempty"`
}
```

### Prompt Types

```go
type Prompt struct {
    Name        string                  `json:"name"`
    Description string                  `json:"description,omitempty"`
    Arguments   []PromptArgument        `json:"arguments,omitempty"`
}

type PromptResponse struct {
    Description string          `json:"description,omitempty"`
    Messages    []PromptMessage `json:"messages"`
}

type PromptMessage struct {
    Role    string  `json:"role"`
    Content Content `json:"content"`
}
```

### Resource Types

```go
type Resource struct {
    URI         string `json:"uri"`
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    MimeType    string `json:"mimeType,omitempty"`
}

type ResourceContents struct {
    URI      string `json:"uri"`
    MimeType string `json:"mimeType,omitempty"`
    Text     string `json:"text,omitempty"`
    Blob     string `json:"blob,omitempty"`
}
```

### Content Types

```go
type Content struct {
    Type        string                 `json:"type"`
    Text        string                 `json:"text,omitempty"`
    Data        string                 `json:"data,omitempty"`
    MimeType    string                 `json:"mimeType,omitempty"`
    Annotations map[string]interface{} `json:"annotations,omitempty"`
}
```

### Helper Functions

```go
// Create text content
func NewTextContent(text string) Content

// Create image content
func NewImageContent(data, mimeType string) Content

// Create embedded resource content
func NewResourceContent(uri, mimeType string) Content

// Create tool response
func NewToolResponse(contents ...Content) *ToolResponse

// Create prompt response
func NewPromptResponse(description string, messages ...PromptMessage) *PromptResponse

// Create prompt message
func NewPromptMessage(content Content, role string) PromptMessage

// Create resource response
func NewResourceResponse(contents ...*ResourceContents) *ResourceResponse

// Create text resource
func NewTextResource(uri, text, mimeType string) *ResourceContents

// Create blob resource
func NewBlobResource(uri, blob, mimeType string) *ResourceContents
```

### Role Constants

```go
const (
    RoleUser      = "user"
    RoleAssistant = "assistant"
)
```

## Transport Interface

```go
type Transport interface {
    // Start initializes the transport
    Start(ctx context.Context) error

    // Send sends a message
    Send(ctx context.Context, msg *BaseJSONRPCMessage) error

    // Close shuts down the transport
    Close() error

    // Set handlers
    SetMessageHandler(handler func(ctx context.Context, msg *BaseJSONRPCMessage))
    SetErrorHandler(handler func(error))
    SetCloseHandler(handler func())
}
```

### Available Transports

**Stdio**:

```go
// Server
transport := stdio.NewServerTransport()

// Client
transport := stdio.NewClientTransport(command string, args []string)
```

**SSE**:

```go
// Server
transport := sse.NewServerTransport(addr, path string, opts ...Option)

// Client
transport := sse.NewClientTransport(url string, opts ...Option)
```

**Streamable HTTP**:

```go
// Server
transport := streamable.NewServerTransport(addr, path string, opts ...Option)

// Client
transport := streamable.NewClientTransport(url string, opts ...Option)
```

## Schema Generation

### Automatic Schema Generation

Schemas are automatically generated from struct tags:

```go
type Args struct {
    // Required string field
    Name string `json:"name" jsonschema:"required,description=User name"`
    
    // Optional integer field
    Age int `json:"age" jsonschema:"description=User age"`
    
    // Array field
    Tags []string `json:"tags" jsonschema:"description=User tags"`
    
    // Nested object
    Address struct {
        City string `json:"city" jsonschema:"required"`
    } `json:"address" jsonschema:"required"`
}
```

### Manual Schema Generation

```go
func GenerateSchema(handler interface{}) (map[string]interface{}, error)
```

Generates a JSON schema from a handler function's argument type.

### Supported Types

- `string` → `"string"`
- `int`, `int64`, etc. → `"integer"`
- `float32`, `float64` → `"number"`
- `bool` → `"boolean"`
- `[]T` → `"array"` with items
- `map[string]T` → `"object"`
- Nested structs → `"object"` with properties

### Tags

**`json` tag**: Specifies JSON field name

**`jsonschema` tag**: Comma-separated options:
- `required`: Field is required
- `description=text`: Field description

**Example**:

```go
Field string `json:"field_name" jsonschema:"required,description=Field description"`
```

## Error Handling

All API functions return `error`. Common error types:

```go
// Tool not found
fmt.Errorf("tool not found: %s", name)

// Invalid arguments
fmt.Errorf("failed to parse arguments: %w", err)

// Transport error
fmt.Errorf("failed to send request: %w", err)

// Context timeout
context.DeadlineExceeded

// Context canceled
context.Canceled
```

## Best Practices

1. **Always use contexts with timeouts**:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
result, err := client.CallTool(ctx, "tool", args)
```

2. **Check errors**:

```go
if err != nil {
    log.Printf("Failed to call tool: %v", err)
    return
}
```

3. **Close clients when done**:

```go
defer client.Close()
```

4. **Use typed arguments for tools**:

```go
type MyArgs struct {
    Field string `json:"field" jsonschema:"required"`
}

func handler(args MyArgs) (*ToolResponse, error) {
    // args.Field is guaranteed to be a string
}
```

5. **Provide clear descriptions**:

```go
server.RegisterTool(
    "analyze",
    "Analyzes text and returns insights", // Clear description
    handler,
)
```

## Complete Example

```go
package main

import (
    "context"
    "log"
    
    mcpgo "github.com/DR1N0/mcp-go"
    "github.com/DR1N0/mcp-go/transport/stdio"
)

type EchoArgs struct {
    Message string `json:"message" jsonschema:"required,description=Message to echo"`
}

func main() {
    // Create transport
    transport := stdio.NewServerTransport()
    
    // Create server
    server := mcpgo.NewServer(
        transport,
        mcpgo.WithName("echo-server"),
        mcpgo.WithVersion("1.0.0"),
    )
    
    // Register tool
    server.RegisterTool("echo", "Echoes a message", func(args EchoArgs) (*mcpgo.ToolResponse, error) {
        return mcpgo.NewToolResponse(
            mcpgo.NewTextContent("Echo: " + args.Message),
        ), nil
    })
    
    // Register prompt
    server.RegisterPrompt("greeting", "Generates greeting", func(args struct{
        Name string `json:"name" jsonschema:"required"`
    }) (*mcpgo.PromptResponse, error) {
        return mcpgo.NewPromptResponse(
            "A greeting",
            mcpgo.NewPromptMessage(
                mcpgo.NewTextContent("Hello "+args.Name),
                mcpgo.RoleAssistant,
            ),
        ), nil
    })
    
    // Register resource
    server.RegisterResource(
        "config://app",
        "App Config",
        "Application configuration",
        "application/json",
        func() (*mcpgo.ResourceResponse, error) {
            return mcpgo.NewResourceResponse(
                mcpgo.NewTextResource(
                    "config://app",
                    `{"version": "1.0.0"}`,
                    "application/json",
                ),
            ), nil
        },
    )
    
    // Start server
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}
