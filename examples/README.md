# MCP-Go Examples

This directory contains complete examples demonstrating all three MCP transport implementations.

## 📁 Directory Structure

```
examples/
├── README.md                    # This file - overview and comparison
├── streamable_http/             # HTTP-based transport
│   ├── server/main.go
│   ├── clients/go/main.go
│   ├── clients/python/main.py
│   └── README.md
├── stdio/                       # Standard I/O transport  
│   ├── server/main.go
│   ├── clients/go/main.go
│   ├── clients/python/main.py
│   └── README.md
└── sse/                         # Server-Sent Events transport
    ├── server/main.go
    ├── clients/go/main.go
    ├── clients/python/main.py
    └── README.md
```

## 🚀 Quick Start

### Prerequisites

**Go (all examples):**
```bash
go version  # Requires Go 1.21+
```

**Python (for Python clients):**
```bash
python --version  # Requires Python 3.12+
uv --version      # Install: pip install uv
```

### Try All Transports

```bash
# Streamable HTTP
make server-streamable
make client-streamable
make client-streamable-python

# Stdio (auto-spawns server)
make client-stdio
make client-stdio-python

# SSE
make server-sse
make client-sse
make client-sse-python
```

## 📊 Transport Comparison

| Feature | Streamable HTTP | Stdio | SSE |
|---------|----------------|-------|-----|
| **Process Model** | Independent server | Client spawns server | Independent server |
| **Communication** | HTTP POST requests | stdin/stdout pipes | SSE stream + POST |
| **Port Required** | Yes (8000) | No | Yes (8001) |
| **Network Support** | ✅ Remote | ❌ Local only | ✅ Remote |
| **Server Lifecycle** | Independent | Tied to client | Independent |
| **Session Management** | Stateless/Per-request | Per-subprocess | Query parameter |
| **Bi-directional** | Request/Response | Full duplex | Async responses |
| **Best For** | Web services, APIs | CLI tools, plugins | Real-time updates |

## 🎯 When to Use Each Transport

### Streamable HTTP
**Use when:**
- ✅ Building web services or APIs
- ✅ Need to support multiple remote clients
- ✅ Want stateless, scalable architecture
- ✅ Deploying to cloud/containers

**Example use cases:**
- REST API backend for MCP
- Microservices architecture
- Cloud-deployed AI assistants
- Load-balanced services

### Stdio
**Use when:**
- ✅ Building CLI tools
- ✅ IDE/editor plugins (like Claude Desktop)
- ✅ Local-only development tools
- ✅ Process isolation is important

**Example use cases:**
- Code analysis tools
- Local AI assistants
- Development utilities
- Subprocess-based services

### SSE
**Use when:**
- ✅ Need server-initiated updates
- ✅ Building real-time applications  
- ✅ Want persistent connections
- ✅ Compatible with pydantic_ai/MCP SDK

**Example use cases:**
- Real-time dashboards
- Live data streaming
- Event-driven applications
- Browser-based MCP clients

## 🏗️ Architecture Overview

### Streamable HTTP
```
┌─────────┐                    ┌─────────┐
│ Client  │──HTTP POST────────►│ Server  │
│         │   (JSON-RPC)       │  :8000  │
│         │◄──HTTP Response────│         │
└─────────┘   (JSON-RPC)       └─────────┘
```

### Stdio
```
┌─────────┐ spawns subprocess  ┌─────────┐
│ Client  │───────────────────►│ Server  │
│         │                    │ Process │
│         │◄─stdin/stdout pipe─┤         │
└─────────┘                    └─────────┘
```

### SSE
```
┌─────────┐      GET /sse      ┌─────────┐
│ Client  │───────────────────►│ Server  │
│         │◄─SSE stream────────│  :8001  │
│         │  (endpoint event)  │         │
│         │                    │         │
│         │  POST /message?id  │         │
│         │───────────────────►│         │
│         │◄─202 Accepted──────│         │
│         │◄─SSE: response─────│         │
└─────────┘                    └─────────┘
```

## 📦 What's Included

Each transport example includes:

### Server
- **Tools**: `echo`, `add`
- **Prompts**: `greeting`  
- **Resources**: `config://server`, `lyrics://never-gonna-give-you-up`
- Clean shutdown handling
- Error handling

### Go Client
- Connection initialization
- Tool listing and calling
- Prompt listing and retrieval
- Resource listing and reading
- Ping/health checks
- Comprehensive error handling

### Python Client
- Direct API testing with `pydantic_ai`
- Tool execution
- Resource access
- Automated test suite
- Cross-language interoperability demos

## 🔧 Common Patterns

### Registering Tools

All transports use the same server-side API:

```go
type EchoArgs struct {
    Message string `json:"message" jsonschema:"required,description=Message to echo"`
}

func echoTool(args EchoArgs) (*mcpgo.ToolResponse, error) {
    return mcpgo.NewToolResponse(
        mcpgo.NewTextContent(fmt.Sprintf("Echo: %s", args.Message)),
    ), nil
}

server.RegisterTool("echo", "Echoes back the provided message", echoTool)
```

### Registering Resources

```go
func configResource() (*mcpgo.ResourceResponse, error) {
    config := `{"server": "example", "version": "1.0.0"}`
    return mcpgo.NewResourceResponse(
        mcpgo.NewTextResource("config://server", config, "application/json"),
    ), nil
}

server.RegisterResource(
    "config://server",
    "Server Configuration",
    "Configuration details",
    "application/json",
    configResource,
)
```

### Python Client Usage

```python
from pydantic_ai.mcp import MCPServerStdio, MCPServerHTTP, MCPServerSSE

# Streamable HTTP
server = MCPServerHTTP("http://localhost:8000/mcp")

# Stdio
server = MCPServerStdio("./bin/server", args=[])

# SSE
server = MCPServerSSE("http://localhost:8001/mcp/sse")

# Use with agent
agent = Agent(model, toolsets=[server])
```

## 📚 Related

- **[MCP Protocol Specification](https://spec.modelcontextprotocol.io/)** - Official protocol docs
- **[pydantic_ai](https://ai.pydantic.dev/)** - Python MCP client library
- **[Project Root README](../README.md)** - mcp-go library docs

## 💡 Next Steps

1. **Start Simple**: Try Streamable HTTP first (easiest to debug)
2. **Explore Transports**: Compare behavior across all three
3. **Build Custom Tools**: Add your own tools and resources
4. **Integrate AI**: Connect to Claude, GPT-4, or other LLMs
5. **Deploy**: Choose transport based on your deployment needs
