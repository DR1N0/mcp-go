# MCP-Go Test Client

This directory contains a Python test client to verify the mcp-go server works correctly with pydantic-ai.

## Prerequisites

- Python 3.11 or higher
- [uv](https://github.com/astral-sh/uv) package manager (recommended) or pip
- The mcp-go server compiled and ready to run

## Setup

### Install Dependencies

Using uv (recommended):
```bash
cd local
uv sync
```

Using pip:
```bash
cd local
pip install -e .
```

### AI Model Configuration

The test uses **CATAgent** with eBay's internal LLM proxy - **no API key needed!**

```python
agent = CATAgent(
    model="hubgpt-chat-completions-claude-sonnet-4.5",
    base_url="https://proxy-llmproxy-staging.qa.ebay.com",
    provider_type="anthropic"
)
```

This configuration:
- ✅ Uses eBay's internal infrastructure
- ✅ No external API keys required
- ✅ Same setup as other internal tools

## Running the Tests

### Step 1: Start the Go Server

In the root directory:
```bash
./bin/simple_server
```

You should see:
```
2024/12/23 16:00:00 Streamable HTTP server starting on http://localhost:8000/mcp
2024/12/23 16:00:00 MCP server started on http://localhost:8000/mcp
2024/12/23 16:00:00 Health check: http://localhost:8000/health
2024/12/23 16:00:00 Press Ctrl+C to stop
```

### Step 2: Run the Python Test

In another terminal:
```bash
cd local
uv run test_agent.py
# or: python test_agent.py
```

## Expected Output

```
================================================================================
Testing mcp-go Server with pydantic-ai
================================================================================

[1] Connecting to MCP server at http://localhost:8000/mcp...
[2] Initializing AI agent...

================================================================================
Test 1: Echo Tool
================================================================================
Query: Use the echo tool to echo back the message 'Hello from mcp-go!'

Response:
Echo: Hello from mcp-go!

✅ Echo tool test passed!

================================================================================
Test 2: Add Tool
================================================================================
Query: Use the add tool to calculate 42 + 58

Response:
Result: 100.00

✅ Add tool test passed!

================================================================================
Test 3: List Available Tools
================================================================================
Found 2 tools:
  - echo: Echoes back the provided message
  - add: Adds two numbers together

✅ Tool listing test passed!

================================================================================
All tests completed!
================================================================================
```

## Manual Testing with curl

You can also test the server directly without AI:

### Initialize
```bash
curl -X POST http://localhost:8000/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'
```

### List Tools
```bash
curl -X POST http://localhost:8000/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
```

### Call Echo Tool
```bash
curl -X POST http://localhost:8000/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc":"2.0",
    "id":3,
    "method":"tools/call",
    "params":{
      "name":"echo",
      "arguments":{"message":"Hello MCP!"}
    }
  }'
```

### Call Add Tool
```bash
curl -X POST http://localhost:8000/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc":"2.0",
    "id":4,
    "method":"tools/call",
    "params":{
      "name":"add",
      "arguments":{"a":42,"b":58}
    }
  }'
```

## Troubleshooting

### Server Not Running
```
Error: Connection refused
```
**Solution:** Make sure the Go server is running on port 8000

### API Key Not Set
```
Error: OpenAI API key not set
```
**Solution:** Export your API key: `export OPENAI_API_KEY="sk-..."`

### Python Packages Not Found
```
ModuleNotFoundError: No module named 'pydantic_ai'
```
**Solution:** Run `uv sync` or `pip install -e .` in the local directory

### Tools Not Being Called
If the AI doesn't use the tools, try:
1. Be more explicit in your query: "Use the echo tool to..."
2. Check server logs for errors
3. Verify tools are listed correctly

## What This Tests

✅ **Streamable HTTP Transport** - Connection via HTTP POST  
✅ **JSON-RPC Protocol** - Request/response handling  
✅ **MCP Methods** - initialize, tools/list, tools/call  
✅ **Tool Execution** - Both echo and add tools  
✅ **Error Handling** - Proper error responses  
✅ **pydantic-ai Integration** - Full compatibility

## Next Steps

Once these tests pass, you can:
1. Add more complex tools to the Go server
2. Test with different AI models
3. Implement the full Server API with reflection-based schemas
4. Add stdio and SSE transports
5. Create production-ready applications
