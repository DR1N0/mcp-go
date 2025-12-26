# mcp-go

> A clean, well-architected Go implementation of the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/)

[![Go Version](https://img.shields.io/github/go-mod/go-version/DR1N0/mcp-go)](https://github.com/DR1N0/mcp-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/DR1N0/mcp-go)](https://goreportcard.com/report/github.com/DR1N0/mcp-go)
[![License](https://img.shields.io/github/license/DR1N0/mcp-go)](https://github.com/DR1N0/mcp-go/blob/master/LICENSE)
[![GitHub last commit](https://img.shields.io/github/last-commit/DR1N0/mcp-go)](https://github.com/DR1N0/mcp-go/commits/master)

## Table of Contents

- [Overview](#overview)
- [Why mcp-go?](#why-mcp-go)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Server Example](#server-example)
  - [Client Example](#client-example)
  - [Python Integration](#python-integration)
- [Features](#features)
- [Architecture](#architecture)
- [Transports](#transports)
- [Examples](#examples)
- [Documentation](#documentation)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Overview

`mcp-go` provides a production-ready implementation of MCP in Go, featuring a clean layered architecture and innovative transport options. Built for seamless integration with AI frameworks like [pydantic-ai](https://ai.pydantic.dev/) and tools like [Claude Desktop](https://claude.ai/desktop).

## Why mcp-go?

🏗️ **Clean Architecture** - Properly separated protocol, transport, and application layers make the codebase maintainable and extensible

🚀 **Streamable HTTP Transport** - Industry-first stateless HTTP transport designed for cloud-native deployments and microservices architectures

🐍 **Python-First Design** - Built from the ground up for seamless pydantic-ai integration with first-class HTTP client support

🧪 **Test-Driven** - Comprehensive test coverage across all layers ensures reliability in production

📦 **Modular** - Use just the components you need: transport layer, protocol handler, or full server/client

## Quick Start

### Installation

```bash
go get github.com/DR1N0/mcp-go
```

### Server Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    mcp "github.com/DR1N0/mcp-go"
    "github.com/DR1N0/mcp-go/transport/stdio"
)

// Define tool arguments as Go structs
type GetTimeArgs struct {
    Timezone string `json:"timezone" jsonschema:"required,description=Timezone name (e.g., America/New_York, UTC, Asia/Tokyo)"`
    Format   string `json:"format" jsonschema:"description=Time format (default: RFC3339)"`
}

func main() {
    // Create server with stdio transport
    server := mcp.NewServer(
        stdio.NewStdioServerTransport(),
        mcp.WithName("time-server"),
        mcp.WithVersion("1.0.0"),
    )
    
    // Register tool with automatic schema generation
    err := server.RegisterTool(
        "get_time",
        "Get current time in a specific timezone",
        func(ctx context.Context, args GetTimeArgs) (*mcp.ToolResponse, error) {
            // Load the timezone
            loc, err := time.LoadLocation(args.Timezone)
            if err != nil {
                return nil, fmt.Errorf("invalid timezone: %w", err)
            }
            
            // Get current time in that timezone
            now := time.Now().In(loc)
            
            // Format the time
            format := args.Format
            if format == "" {
                format = time.RFC3339
            }
            
            result := fmt.Sprintf("Current time in %s: %s", args.Timezone, now.Format(format))
            return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
        },
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Start serving
    if err := server.Serve(); err != nil {
        log.Fatal(err)
    }
}
```

### Client Example

```go
package main

import (
    "context"
    "log"
    
    mcp "github.com/DR1N0/mcp-go"
    "github.com/DR1N0/mcp-go/transport/stdio"
)

func main() {
    // Create client
    client := mcp.NewClient(
        stdio.NewStdioClientTransport("go", "run", "./server/main.go"),
    )
    
    // Initialize connection
    if _, err := client.Initialize(context.Background()); err != nil {
        log.Fatal(err)
    }
    
    // Call a tool
    response, err := client.CallTool(context.Background(), "get_time", map[string]string{
        "timezone": "America/New_York",
        "format":   "2006-01-02 15:04:05 MST",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Result: %s", response.Content[0].Text)
}
```

### Python Integration

Connect your Go MCP server to pydantic-ai using the streamable HTTP transport:

**Go Server:**
```go
server := mcp.NewServer(
    streamable.NewServerTransport("/mcp", ":8080"),
    mcp.WithName("my-tools"),
)
server.RegisterTool("get_time", "Get current time in timezone", getTimeHandler)
server.Serve()
```

**Python Client (pydantic-ai):**
```python
from pydantic_ai import Agent
from pydantic_ai.mcp import MCPServerStreamableHTTP

# Connect to Go MCP server
server = MCPServerStreamableHTTP(
    url="http://localhost:8080/mcp",
    tool_prefix="mytools_"
)

agent = Agent(
    model="anthropic:claude-sonnet-4",
    toolsets=[server]
)

# Agent can now use tools from the Go server
result = agent.run_sync("What time is it in Tokyo?")
```

## Features

- ✅ **Multiple Transport Options**
  - stdio - Standard I/O for subprocess communication
  - SSE - Server-Sent Events for real-time updates
  - Streamable HTTP - Stateless HTTP for cloud deployments
  
- ✅ **Type-Safe Tool Registration**
  - Automatic JSON schema generation from Go structs
  - Reflection-based argument validation
  - Context support for cancellation and timeouts

- ✅ **Full MCP Support**
  - Tools - Expose functions to AI agents
  - Prompts - Reusable prompt templates
  - Resources - Access to external data sources
  - Sampling - LLM completion requests

- ✅ **Production Ready**
  - Comprehensive test coverage
  - Clean error handling
  - Graceful shutdown
  - Request/response correlation

## Architecture

```
┌─────────────────────────────────────────────────┐
│              Application Layer                  │
│  (Your Tools, Prompts, Resources, Handlers)     │
└────────────────┬────────────────────────────────┘
                 │
┌────────────────┴────────────────────────────────┐
│            Server / Client Layer                │
│  (Registration, Schema Gen, Lifecycle)          │
└────────────────┬────────────────────────────────┘
                 │
┌────────────────┴────────────────────────────────┐
│            Protocol Layer                       │
│  (JSON-RPC 2.0, Request/Response Correlation)   │
└────────────────┬────────────────────────────────┘
                 │
┌────────────────┴────────────────────────────────┐
│            Transport Layer                      │
│  (stdio / SSE / Streamable HTTP)                │
└─────────────────────────────────────────────────┘
```

Each layer has clear interfaces and can be tested independently. See [docs/architecture.md](docs/architecture.md) for details.

## Transports

### stdio
Perfect for Claude Desktop integration and subprocess-based tools.
```go
transport := stdio.NewStdioServerTransport()
```

### SSE (Server-Sent Events)
Ideal for long-lived HTTP connections with server push capabilities.
```go
transport := sse.NewSSEServerTransport("/events", ":8080")
```

### Streamable HTTP
Industry-first stateless HTTP transport for cloud-native deployments.
```go
transport := streamable.NewServerTransport("/mcp", ":8080")
```

See [docs/transport-guide.md](docs/transport-guide.md) for detailed comparison and usage.

## Examples

| Example | Description | Transport |
|---------|-------------|-----------|
| [stdio/](examples/stdio/) | Basic tool server with subprocess communication | stdio |
| [sse/](examples/sse/) | Real-time updates with Server-Sent Events | SSE |
| [streamable_http/](examples/streamable_http/) | Stateless HTTP server for cloud deployment | Streamable HTTP |

Each example includes both Go server and Python client implementations.

## Documentation

- **[Architecture Guide](docs/architecture.md)** - System design and layer responsibilities
- **[Transport Guide](docs/transport-guide.md)** - Choosing and configuring transports
- **[API Reference](docs/api-reference.md)** - Complete API documentation

## Development

```bash
# Clone the repository
git clone https://github.com/DR1N0/mcp-go.git
cd mcp-go

# Run tests
make test

# Run a specific example
cd examples/stdio/server
go run main.go

# Run linter
make lint
```

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

Areas where we'd love help:
- Additional transport implementations
- More example servers
- Documentation improvements
- Performance optimizations

## License

MIT License - See [LICENSE](LICENSE) for details.

---

**Built with ❤️ for the MCP community**
