#!/usr/bin/env python3
"""
Stdio transport test for mcp-go server using pydantic_ai.

This script tests the MCP server implementation by:
1. Spawning the Go server as a subprocess via stdio
2. Discovering available tools
3. Testing the echo and add tools
4. Testing prompts and resources
"""

import asyncio
import os
from dotenv import load_dotenv
from pydantic_ai.mcp import MCPServerStdio

# Load environment variables
load_dotenv()


async def test_list_tools(server):
    """Test getting tool list from server."""
    print("\n--- Tool Tests ---")
    
    tools = await server.list_tools()
    
    # Should have 2 tools: echo and add
    assert len(tools) == 2, f"Expected 2 tools, got {len(tools)}"
    
    tool_names = [tool.name for tool in tools]
    assert "echo" in tool_names, "Echo tool not found"
    assert "add" in tool_names, "Add tool not found"
    
    print(f"✅ List tools passed - found {len(tools)} tools:")
    for tool in tools:
        print(f"   - {tool.name}: {tool.description}")


async def test_call_echo_tool(server):
    """Test calling the echo tool."""
    result = await server.direct_call_tool(
        "echo",
        {"message": "Hello from Python stdio client!"}
    )
    
    assert isinstance(result, str), f"Expected string result, got {type(result)}"
    assert "Hello from Python stdio client!" in result
    
    print(f"✅ Echo tool test passed: {result}")


async def test_call_add_tool(server):
    """Test calling the add tool."""
    result = await server.direct_call_tool(
        "add",
        {"a": 42, "b": 58}
    )
    
    assert isinstance(result, str), f"Expected string result, got {type(result)}"
    assert "100" in result
    
    print(f"✅ Add tool test passed: {result}")


async def test_list_resources(server):
    """Test getting resource list from server."""
    print("\n--- Resource Tests ---")
    
    resources = await server.list_resources()
    
    # Should have 2 resources
    assert len(resources) == 2, f"Expected 2 resources, got {len(resources)}"
    
    print(f"✅ List resources passed - found {len(resources)} resources:")
    for resource in resources:
        print(f"   - {resource.uri} ({resource.mime_type}): {resource.name}")


async def test_read_config_resource(server):
    """Test reading the config resource."""
    result = await server.read_resource("config://server")
    
    # Result is a string (JSON text)
    assert isinstance(result, str), f"Expected string result, got {type(result)}"
    
    # Verify it's valid JSON with expected fields
    import json
    config = json.loads(result)
    assert "server" in config, "Config should have 'server' field"
    assert "version" in config, "Config should have 'version' field"
    assert "transport" in config, "Config should have 'transport' field"
    assert config["transport"] == "stdio", "Transport should be 'stdio'"
    
    print(f"✅ Read config resource test passed")
    print(f"   Config: {json.dumps(config, indent=2)}")


async def run_tests():
    """Run all async tests."""
    print("=" * 80)
    print("MCP-Go Stdio Transport Tests (Python Client)")
    print("=" * 80)
    
    # Create MCP connection via stdio
    # The server will be spawned as a subprocess
    server = MCPServerStdio(
        command="go",
        args=["run", "./examples/stdio/server/main.go"],
        tool_prefix=""
    )
    
    # Use a single context manager for all operations to reuse the connection
    # This avoids reconnecting for each operation, which is much faster
    async with server:
        print("✅ Connected to stdio server")
        
        # Run all tests using the same connection
        await test_list_tools(server)
        await test_call_echo_tool(server)
        await test_call_add_tool(server)
        await test_list_resources(server)
        await test_read_config_resource(server)
        
        print("\n" + "=" * 80)
        print("All tests completed!")
        print("=" * 80)
    
    # Server process will be terminated automatically when exiting the context


if __name__ == "__main__":
    asyncio.run(run_tests())
