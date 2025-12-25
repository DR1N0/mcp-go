# Streamable HTTP Transport Example

This example demonstrates the **Streamable HTTP** transport for MCP (Model Context Protocol), featuring a Go server and clients in both Go and Python.

## 📁 Structure

```
streamable_http/
├── server/
│   └── main.go          # MCP server with streamable HTTP transport
├── clients/
│   ├── go/
│   │   └── main.go      # Go client using mcp-go
│   └── python/
│       ├── main.py      # Python pydantic_ai client tests
│       └── agent_e2e.py # pydantic_ai Agent integration test
└── README.md            # This file
```

## 🎯 What This Example Demonstrates

### Server (`server/main.go`)
A complete MCP server exposing:
- **Tools**: `echo`, `add`
- **Prompts**: `greeting`
- **Resources**: `config://server`, `lyrics://never-gonna-give-you-up`

### Go Client (`clients/go/main.go`)
Demonstrates all MCP operations using the Go client:
1. Initialize connection
2. List and call tools
3. List and get prompts
4. List and read resources
5. Ping server

### Python Clients (`clients/python/`)
Two Python clients using `pydantic_ai`:

1. **`main.py`** - Direct API testing
   - Health check
   - Tool listing and execution
   - Resource listing and reading
   - Error handling

2. **`agent_e2e.py`** - AI Agent integration
   - Uses LLM to intelligently call tools
   - Demonstrates agent-driven MCP usage
   - Shows real-world AI assistant scenario

## 🚀 Quick Start

### Prerequisites

**For Go:**
```bash
# Go 1.21+ required
go version
```

**For Python:**
```bash
# Python 3.12+ and uv required
python --version
uv --version

# Install uv if needed
pip install uv
```

### 1. Start the Server

```bash
# From project root
make server-streamable

# Or directly
go run ./examples/streamable_http/server/main.go
```

You should see:
```
✅ MCP Server started successfully!
   - Tools: echo, add
   - Prompts: greeting, system
   - Resources: config://server, time://current

Press Ctrl+C to stop...
```

The server listens on `http://localhost:8000/mcp`

### 2. Run the Go Client

**Terminal 2:**
```bash
# From project root
make client-streamable

# Or directly
go run ./examples/streamable_http/clients/go/main.go
```

Expected output:
```
=================================================================================
MCP Go Client Example
=================================================================================
[1] Initializing client...
✅ Connected to server: example-server v1.0.0
   Protocol version: 2024-11-05

[2] Listing available tools...
✅ Found 2 tools:
   - echo: Echoes back the provided message
   - add: Adds two numbers together

[3] Calling 'echo' tool...
✅ Echo response: Echo: Hello from Go client!

...
✅ All operations completed successfully!
```

### 3. Run Python Client Tests

**Terminal 3:**
```bash
# From project root
make client-streamable-python

# Or directly
uv run ./examples/streamable_http/clients/python/main.py
```

Expected output:
```
=================================================================================
MCP-Go Streamable HTTP Server Tests
=================================================================================
Server URL: http://localhost:8000/mcp

Running Synchronous Tests
✅ Health check passed

Running Asynchronous Tests
--- Tool Tests ---
✅ List tools passed - found 2 tools
✅ Echo tool test passed: Echo: Hello from test!
✅ Add tool test passed: 42 + 58 = 100
...
All tests completed!
```

### 4. Run Python Agent Integration (Optional)

Requires LLM API access (OpenAI, Anthropic, or compatible):

```bash
# Set up environment variables
export LLM_PROVIDER=openai
export LLM_BASE_URL=https://api.openai.com/v1
export LLM_API_KEY=your-api-key
export LLM_MODEL=gpt-4

# Run agent test
uv run ./examples/streamable_http/clients/python/agent_e2e.py
```

The agent will:
1. Connect to the MCP server
2. Use the LLM to intelligently call tools
3. Demonstrate AI-driven MCP usage

## 🏗️ Architecture

```
┌─────────────────────┐
│   Python Client     │ ← pydantic_ai (MCPServerStreamableHTTP)
│   (pydantic_ai)     │
└──────────┬──────────┘
           │
           │ HTTP POST
           │ (JSON-RPC 2.0)
           ↓
┌─────────────────────┐
│    Go MCP Server    │ ← Streamable HTTP Transport
│   (mcp-go)          │    Port: 8000
└──────────┬──────────┘    Endpoint: /mcp
           ↑
           │ HTTP POST
           │ (JSON-RPC 2.0)
           │
┌──────────┴──────────┐
│    Go Client        │ ← mcp-go Client
│    (mcp-go)         │
└─────────────────────┘
```

### Transport Flow

1. **Client** → HTTP POST to `/mcp` with JSON-RPC request
2. **Server** → Processes request, calls handler
3. **Server** → Returns HTTP response with JSON-RPC result
4. **Client** → Receives and processes result

## 📚 What You'll Learn

### Server Implementation
- Registering tools with typed arguments
- Registering prompts with dynamic content
- Serving resources (static and dynamic)
- Clean shutdown handling

### Client Implementation (Go)
- Connecting to MCP servers
- Making synchronous RPC calls
- Handling responses and errors
- Resource lifecycle management

### Client Implementation (Python)
- Using `pydantic_ai` MCP client
- Direct API testing patterns
- AI agent integration
- Cross-language interoperability

## 🔧 Customization

### Adding New Tools

In `server/main.go`:

```go
type YourArgs struct {
    Field string `json:"field" jsonschema:"required,description=Field description"`
}

func yourTool(args YourArgs) (*mcpgo.ToolResponse, error) {
    return mcpgo.NewToolResponse(
        mcpgo.NewTextContent("Your response"),
    ), nil
}

// Register it
server.RegisterTool("your-tool", "Description", yourTool)
```

### Adding New Resources

```go
func yourResource() (*mcpgo.ResourceResponse, error) {
    content := "Your resource content"
    return mcpgo.NewResourceResponse(
        mcpgo.NewTextResource("custom://uri", content, "text/plain"),
    ), nil
}

// Register it
server.RegisterResource(
    "custom://uri",
    "Resource Name",
    "Description",
    "text/plain",
    yourResource,
)
```

## 🧪 Testing

### Unit Testing
```bash
# Test the server
go test ./examples/streamable_http/server/...

# Test the Go client
go test ./examples/streamable_http/clients/go/...
```

### Integration Testing
```bash
# Start server in background
make server-streamable &

# Run all clients
make client-streamable
make client-streamable-python

# Stop server
pkill -f "streamable_http_server"
```

## 📖 Next Steps

- **Add More Transports**: Try `stdio` or `sse` transports
- **Build Your Own**: Create custom tools and resources
- **Integrate with AI**: Connect to Claude, GPT-4, or other LLMs
- **Deploy**: Run the server in production environments

## 🔗 Related

- [MCP Protocol Specification](https://spec.modelcontextprotocol.io/)
- [pydantic_ai Documentation](https://ai.pydantic.dev/)
- [Project Root README](../../README.md)

## 💡 Troubleshooting

### Server won't start
```bash
# Check if port 8000 is in use
lsof -i :8000

# Kill the process using the port
kill -9 <PID>
```

### Python client fails
```bash
# Ensure uv is installed
pip install uv

# Check server is running
curl http://localhost:8000/health
```

### Go client connection refused
```bash
# Verify server is running on correct port
netstat -an | grep 8000

# Check server logs for errors
```

## 📝 Notes

- The server automatically generates JSON schemas from Go struct tags
- Python clients use `pydantic_ai` for type-safe MCP communication
- All communication uses JSON-RPC 2.0 over HTTP
- The transport is stateless - each request is independent
