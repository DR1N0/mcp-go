# Streamable HTTP Transport Example

HTTP-based MCP transport with independent server process and stateless request/response model.

See [Examples Overview](../README.md) for transport comparison and common setup.

## 🚀 Quick Start

### 1. Start the Server

```bash
# From project root
make server-streamable

# Or manually
go run ./examples/streamable_http/server/main.go
```

Server starts on `http://localhost:8000/mcp`

### 2. Run Go Client

```bash
make client-streamable

# Or manually  
go run ./examples/streamable_http/clients/go/main.go
```

### 3. Run Python Client

```bash
make client-streamable-python

# Or manually
uv run ./examples/streamable_http/clients/python/main.py
```

### 4. Run Python Agent (Optional)

Requires LLM API credentials:

```bash
export LLM_PROVIDER=openai
export LLM_BASE_URL=https://api.openai.com/v1
export LLM_API_KEY=your-key
export LLM_MODEL=gpt-4

uv run ./examples/streamable_http/clients/python/agent_e2e.py
```

## 📝 Server Details

**Endpoint**: `http://localhost:8000/mcp`

**Tools**:
- `echo` - Echoes back a message
- `add` - Adds two numbers

**Prompts**:
- `greeting` - Personalized greeting

**Resources**:
- `config://server` - Server configuration (JSON)
- `lyrics://never-gonna-give-you-up` - Song lyrics (text)

## 🔧 Customization

### Add a Tool

```go
type MyArgs struct {
    Input string `json:"input" jsonschema:"required,description=Input value"`
}

func myTool(args MyArgs) (*mcpgo.ToolResponse, error) {
    return mcpgo.NewToolResponse(
        mcpgo.NewTextContent("Result: " + args.Input),
    ), nil
}

server.RegisterTool("my-tool", "Tool description", myTool)
```

### Add a Resource

```go
func myResource() (*mcpgo.ResourceResponse, error) {
    return mcpgo.NewResourceResponse(
        mcpgo.NewTextResource("custom://uri", "content", "text/plain"),
    ), nil
}

server.RegisterResource("custom://uri", "Name", "Description", "text/plain", myResource)
```

## 💡 Troubleshooting

**Port already in use:**
```bash
lsof -i :8000
kill -9 <PID>
```

**Python client fails:**
```bash
# Check server is running
curl http://localhost:8000/health

# Verify uv is installed
pip install uv
```

**Connection refused:**
```bash
# Verify server is running
netstat -an | grep 8000
```

## 📚 Python Client Usage

```python
from pydantic_ai.mcp import MCPServerStreamableHTTP

server = MCPServerStreamableHTTP("http://localhost:8000/mcp")

# Use with pydantic_ai Agent
agent = Agent(model, toolsets=[server])
