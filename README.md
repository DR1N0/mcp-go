# mcp-go

A Go implementation of the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) with support for multiple transports including streamable HTTP, SSE, and stdio.

> **Status:** 🚧 Early Development - Foundation complete, core implementation in progress

## Overview

`mcp-go` provides a clean, type-safe implementation of MCP in Go, designed to work seamlessly with AI frameworks like [pydantic-ai](https://ai.pydantic.dev/) and tools like [Claude Desktop](https://claude.ai/desktop).

### Key Features

- ✅ **Multiple Transports**: stdio, SSE, streamable HTTP
- ✅ **Type-Safe**: Reflection-based schema generation from Go structs
- ✅ **Clean API**: Simple, intuitive interface inspired by popular Go frameworks
- ✅ **Interface-Driven**: Clear separation of concerns, easy to test and extend
- 🚧 **Python Compatible**: First-class support for pydantic-ai (in progress)

## Architecture

```
mcp-go/
├── interface.go              # Core interfaces (Server, Client, Transport)
├── types.go                  # MCP protocol types
├── server.go                 # Server implementation
├── client.go                 # Client implementation
├── schema.go                 # JSON schema generation
│
├── transport/
│   ├── interface.go          # Transport interface
│   ├── stdio/                # Standard I/O transport
│   ├── sse/                  # Server-Sent Events transport
│   └── streamable/           # Streamable HTTP transport
│
├── protocol/
│   ├── protocol.go           # JSON-RPC 2.0 handler
│   └── messages.go           # Message utilities
│
├── examples/                 # Example servers and clients
└── tests/                    # Integration and E2E tests
```

## Quick Start (Planned API)

### Server Example

```go
package main

import (
    "context"
    
    mcpgo "github.com/DR1N0/mcp-go"
    "github.com/DR1N0/mcp-go/transport/streamable"
)

type GetReposArgs struct {
    Namespace string `json:"namespace" jsonschema:"required,description=Kubernetes namespace"`
    Name      string `json:"name" jsonschema:"required,description=Service account name"`
}

func main() {
    // Create server with streamable HTTP transport
    server := mcpgo.NewServer(
        streamable.NewServerTransport("/mcp", ":8000"),
        mcpgo.WithName("my-mcp-server"),
        mcpgo.WithVersion("1.0.0"),
    )
    
    // Register a tool with automatic schema generation
    server.RegisterTool(
        "get_repos",
        "Get GitHub repositories for a service account",
        func(ctx context.Context, args GetReposArgs) (* mcpgo.ToolResponse, error) {
            // Your tool logic here
            return mcpgo.NewToolResponse(
                mcpgo.NewTextContent("Result here"),
            ), nil
        },
    )
    
    // Start serving
    server.Serve()
}
```

### Client Example

```go
package main

import (
    "context"
    
    mcpgo "github.com/DR1N0/mcp-go"
    "github.com/DR1N0/mcp-go/transport/streamable"
)

func main() {
    client := mcpgo.NewClient(
        streamable.NewClientTransport("http://localhost:8000/mcp"),
    )
    
    // Initialize connection
    _, err := client.Initialize(context.Background())
    if err != nil {
        panic(err)
    }
    
    // List available tools
    tools, err := client.ListTools(context.Background(), nil)
    if err != nil {
        panic(err)
    }
    
    // Call a tool
    result, err := client.CallTool(context.Background(), "get_repos", map[string]string{
        "namespace": "default",
        "name": "my-sa",
    })
    if err != nil {
        panic(err)
    }
}
```

### Python Integration (pydantic-ai)

```python
from pydantic_ai import Agent
from pydantic_ai.mcp import MCPServerStreamableHTTP

# Connect to Go MCP server
server = MCPServerStreamableHTTP(
    url="http://localhost:8000/mcp",
    tool_prefix="mytools_"
)

agent = Agent(
    model="anthropic:claude-sonnet-4",
    toolsets=[server]
)

# Agent can now use tools from the Go server
result = agent.run_sync("Get repos for service account 'default' in namespace 'prod'")
```

## Current Status

### ✅ Completed
- Core interfaces and types
- MCP protocol type definitions
- Transport layer architecture
- Basic streamable HTTP server structure

### 🚧 In Progress
- Streamable HTTP request/response correlation
- Protocol layer (JSON-RPC 2.0 handling)
- Server implementation with reflection-based schemas
- Client implementation

### 📋 Planned
- stdio transport (standard MCP approach)
- SSE transport (for real-time updates)
- Comprehensive tests
- Working examples
- Full documentation

See [PROJECT_STATUS.md](PROJECT_STATUS.md) for detailed progress tracking.

## Design Principles

### 1. Interface-First
Every package has an `interface.go` defining clear contracts:
```go
type Server interface {
    RegisterTool(name, description string, handler interface{}) error
    RegisterPrompt(name, description string, handler interface{}) error
    RegisterResource(uri, name, description, mimeType string, handler interface{}) error
    Serve() error
    Close() error
}
```

### 2. Type-Safe Tool Registration
Tools are defined using Go structs with automatic JSON schema generation:
```go
type MyToolArgs struct {
    Name     string  `json:"name" jsonschema:"required,description=User name"`
    Age      *int    `json:"age" jsonschema:"description=Optional age"`
}

server.RegisterTool("my_tool", "Description", 
    func(ctx context.Context, args MyToolArgs) (* mcpgo.ToolResponse, error) {
        // Type-safe access to args.Name and args.Age
    },
)
```

### 3. Transport Agnostic
Server and client work with any transport implementation:
- **stdio**: For Claude Desktop and subprocess-based tools
- **SSE**: For long-lived HTTP connections with server push
- **streamable HTTP**: For stateless HTTP environments

### 4. Testing Strategy
- **Unit tests**: Co-located with implementation (`server_test.go`)
- **Integration tests**: In `tests/integration/` for transport testing
- **E2E tests**: In `tests/e2e/` for real-world scenarios

## Development

```bash
# Clone the repository
git clone https://github.com/DR1N0/mcp-go.git
cd mcp-go

# Run tests (once implemented)
go test ./...

# Run integration tests
go test ./tests/integration/...

# Run examples
go run examples/streamable_server/main.go
```

## Inspiration and References

This project builds upon:
- **[metoro-io/mcp-golang](https://github.com/metoro-io/mcp-golang)**: Excellent server architecture and reflection-based schema generation
- **[MCP Specification](https://spec.modelcontextprotocol.io/)**: Official MCP protocol specification
- **[pydantic-ai](https://ai.pydantic.dev/)**: Python AI framework with MCP client implementations

## Contributing

This project is in active development. Contributions welcome!

Areas where help is needed:
- Completing the streamable HTTP transport
- Implementing stdio and SSE transports
- Writing tests and examples
- Documentation and guides

## License

MIT License - See [LICENSE](LICENSE) for details

## Related Projects

- [MCP Specification](https://modelcontextprotocol.io/)
- [pydantic-ai](https://ai.pydantic.dev/) - Python AI framework
- [Claude Desktop](https://claude.ai/desktop) - AI assistant with MCP support
- [metoro-io/mcp-golang](https://github.com/metoro-io/mcp-golang) - Another Go MCP implementation

---

**Note**: This project is under active development. The API may change as we iterate toward 1.0.
