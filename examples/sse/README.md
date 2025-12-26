# SSE Transport Example

Server-Sent Events transport with persistent connections for real-time communication. Implements the official MCP SSE protocol compatible with pydantic_ai and MCP SDK.

See [Examples Overview](../README.md) for transport comparison and common setup.

## 🚀 Quick Start

### 1. Start the Server

```bash
# From project root
make server-sse

# Or manually
go run ./examples/sse/server/main.go
```

Server starts on `http://localhost:8001/mcp/sse`

### 2. Run Go Client

```bash
make client-sse

# Or manually
go run ./examples/sse/clients/go/main.go
```

### 3. Run Python Client

```bash
make client-sse-python

# Or manually
uv run ./examples/sse/clients/python/main.py
```

## 📝 How It Works

**MCP SSE Protocol** (compatible with pydantic_ai/MCP SDK):

1. Client connects via `GET /mcp/sse` (establishes SSE stream)
2. Server sends `endpoint` event with message URL containing session_id
   ```
   event: endpoint
   data: /mcp/sse/message?session_id=abc123def456
   ```
3. Client sends requests via `POST` to that URL
4. Server responds with `202 Accepted`
5. Server sends responses via SSE `message` events
   ```
   event: message
   data: {"jsonrpc":"2.0","id":1,"result":{...}}
   ```

**Key Features**:
- Persistent SSE connection for server-to-client messages
- Query parameter session management
- Bi-directional communication (SSE + POST)
- Compatible with all MCP clients

## 📝 Server Details

**SSE Endpoint**: `http://localhost:8001/mcp/sse` (GET)
**Message Endpoint**: `http://localhost:8001/mcp/sse/message?session_id=...` (POST)
**Health Check**: `http://localhost:8001/mcp/sse/health`

**Tools**:
- `echo` - Echoes back a message
- `add` - Adds two numbers

**Prompts**:
- `greeting` - Personalized greeting

**Resources**:
- `config://server` - Server configuration (JSON)
- `lyrics://never-gonna-give-you-up` - Song lyrics (text)

## 🔧 Customization

Same as other transports - see [Examples Overview](../README.md#common-patterns).

## 💡 Troubleshooting

**Port already in use:**
```bash
lsof -i :8001
kill -9 <PID>
```

**Client connection fails:**
```bash
# Check server is running
curl -N http://localhost:8001/mcp/sse

# Should see endpoint event
```

**Python client timeout:**
```bash
# Verify server is accessible
curl http://localhost:8001/mcp/sse/health
```

**Go client timeout:**
```bash
# Check if endpoint event is received
# Server must send event: endpoint on connection
```

## 📚 Python Client Usage

```python
from pydantic_ai.mcp import MCPServerSSE

server = MCPServerSSE("http://localhost:8001/mcp/sse")

# Use with pydantic_ai Agent
agent = Agent(model, toolsets=[server])
```

## 🔍 Manual Testing

Watch SSE stream:
```bash
# See endpoint and message events
curl -N http://localhost:8001/mcp/sse
```

Send request (after getting session_id from endpoint event):
```bash
curl -X POST "http://localhost:8001/mcp/sse/message?session_id=YOUR_SESSION_ID" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"ping"}'
