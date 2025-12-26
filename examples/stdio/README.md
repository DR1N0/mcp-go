# Stdio Transport Example

Standard I/O transport where client spawns server as subprocess and communicates via stdin/stdout pipes.

See [Examples Overview](../README.md) for transport comparison and common setup.

## 🚀 Quick Start

### Run Go Client

The stdio client automatically spawns the server:

```bash
# From project root
make client-stdio

# Or manually
go run ./examples/stdio/clients/go/main.go
```

### Run Python Client

```bash
make client-stdio-python

# Or manually
uv run ./examples/stdio/clients/python/main.py
```

### Run Server Standalone (Manual Testing)

```bash
# Build server first
make build-stdio-server

# Run and type JSON-RPC messages manually
./bin/stdio_server
```

Example input:
```json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}
```

## 📝 How It Works

**Communication**: Line-delimited JSON-RPC over stdin/stdout

1. Client spawns server process with `go run`
2. Client writes JSON-RPC to server's stdin (with `\n`)
3. Server reads from stdin, processes request
4. Server writes JSON-RPC response to stdout (with `\n`)
5. Client reads from server's stdout

**Lifecycle**: Server runs until client closes stdin (on exit)

## 📝 Server Details

**Tools**:
- `echo` - Echoes back a message
- `add` - Adds two numbers

**Prompts**:
- `greeting` - Personalized greeting

**Resources**:
- `config://server` - Server configuration (JSON)
- `lyrics://never-gonna-give-you-up` - Song lyrics (text)

## 🔧 Using a Different Server

```go
// Spawn any MCP server executable
transport := stdio.NewClientTransport(
    "/path/to/server",
    []string{"--arg1", "value1"},
)
```

## 💡 Troubleshooting

**Server exits immediately:**
```bash
# Check if binary exists
ls -la ./bin/stdio_server

# Rebuild
make build-stdio-server
```

**Client hangs:**
```bash
# Server must flush stdout after each message
# Check server implementation
```

**Permission denied:**
```bash
chmod +x ./bin/stdio_server
```

## 📚 Python Client Usage

```python
from pydantic_ai.mcp import MCPServerStdio

server = MCPServerStdio(
    command="./bin/stdio_server",
    args=[],
)

# Use with pydantic_ai Agent
agent = Agent(model, toolsets=[server])
```

## 🔗 Claude Desktop Integration

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "mcp-go-example": {
      "command": "/path/to/mcp-go/bin/stdio_server",
      "args": []
    }
  }
}
