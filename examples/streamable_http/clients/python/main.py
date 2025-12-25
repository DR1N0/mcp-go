#!/usr/bin/env python3
"""
Unit tests for mcp-go streamable HTTP server.

Tests include:
- Health check endpoint
- Tool listing and execution (echo and add)
- Resource listing and reading (config and time)

Note: Prompt tests are not included as pydantic_ai's MCPServer client
does not expose prompt methods (prompts are meant to be used by agents
internally, not called directly).
"""

import asyncio
import os
import pytest
import requests
from dotenv import load_dotenv
from pydantic_ai.mcp import MCPServerStreamableHTTP


# Load environment variables
load_dotenv()
SERVER_URL = os.getenv("MCP_SERVER_URL", "http://localhost:8000/mcp")
BASE_URL = SERVER_URL.rsplit("/", 1)[0]  # Get base URL without /mcp


def test_health_check():
    """Test server health endpoint returns 200."""
    response = requests.get(f"{BASE_URL}/health")
    assert response.status_code == 200
    assert response.text == "OK"
    print("✅ Health check passed")


@pytest.mark.asyncio
async def test_list_tools():
    """Test getting tool list from server."""
    server = MCPServerStreamableHTTP(
        url=SERVER_URL,
        tool_prefix=""
    )
    
    tools = await server.list_tools()
    
    # Should have 2 tools: echo and add
    assert len(tools) == 2, f"Expected 2 tools, got {len(tools)}"
    
    tool_names = [tool.name for tool in tools]
    assert "echo" in tool_names, "Echo tool not found"
    assert "add" in tool_names, "Add tool not found"
    
    # Check echo tool schema
    echo_tool = next(t for t in tools if t.name == "echo")
    assert echo_tool.description is not None
    assert "message" in str(echo_tool.inputSchema)
    
    # Check add tool schema
    add_tool = next(t for t in tools if t.name == "add")
    assert add_tool.description is not None
    assert "a" in str(add_tool.inputSchema)
    assert "b" in str(add_tool.inputSchema)
    
    print(f"✅ List tools passed - found {len(tools)} tools:")
    for tool in tools:
        print(f"   - {tool.name}: {tool.description}")


@pytest.mark.asyncio
async def test_call_echo_tool():
    """Test calling the echo tool."""
    server = MCPServerStreamableHTTP(
        url=SERVER_URL,
        tool_prefix=""
    )
    
    # Call echo tool using direct_call_tool
    result = await server.direct_call_tool(
        "echo",
        {"message": "Hello from test!"}
    )
    
    # Result is a string (the tool returns text content)
    assert isinstance(result, str), f"Expected string result, got {type(result)}"
    assert "Hello from test!" in result, f"Expected echo in response, got: {result}"
    
    print(f"✅ Echo tool test passed: {result}")


@pytest.mark.asyncio
async def test_call_add_tool():
    """Test calling the add tool."""
    server = MCPServerStreamableHTTP(
        url=SERVER_URL,
        tool_prefix=""
    )
    
    # Call add tool using direct_call_tool
    result = await server.direct_call_tool(
        "add",
        {"a": 42, "b": 58}
    )
    
    # Result is a string (the tool returns text content)
    assert isinstance(result, str), f"Expected string result, got {type(result)}"
    assert "100" in result, f"Expected 100 in response, got: {result}"
    
    print(f"✅ Add tool test passed: {result}")


@pytest.mark.asyncio
async def test_tool_error_handling():
    """Test that calling unknown tool returns appropriate error."""
    server = MCPServerStreamableHTTP(
        url=SERVER_URL,
        tool_prefix=""
    )
    
    try:
        # Try to call non-existent tool
        await server.direct_call_tool("nonexistent_tool", {})
        pytest.fail("Expected error for unknown tool")
    except Exception as e:
        # Should throw ModelRetry exception for unknown tool
        print(f"✅ Error handling test passed - raised exception: {type(e).__name__}")


@pytest.mark.asyncio
async def test_list_resources():
    """Test getting resource list from server."""
    server = MCPServerStreamableHTTP(
        url=SERVER_URL,
        tool_prefix=""
    )
    
    resources = await server.list_resources()
    
    # Should have 2 resources: lyrics://never-gonna-give-you-up and config://server
    assert len(resources) == 2, f"Expected 2 resources, got {len(resources)}"
    
    # Check lyrics resource
    lyrics_resource = next(r for r in resources if r.uri == "lyrics://never-gonna-give-you-up")
    assert lyrics_resource.name is not None
    assert lyrics_resource.mime_type == "text/plain"
    
    print(f"✅ List resources passed - found {len(resources)} resources:")
    for resource in resources:
        print(f"   - {resource.uri} ({resource.mime_type}): {resource.name}")


@pytest.mark.asyncio
async def test_read_lyrics_resource():
    """Test reading the lyrics resource."""
    server = MCPServerStreamableHTTP(
        url=SERVER_URL,
        tool_prefix=""
    )
    
    # Read lyrics resource - returns string directly for single content
    result = await server.read_resource("lyrics://never-gonna-give-you-up")
    
    # Result is a string (lyrics text)
    assert isinstance(result, str), f"Expected string result, got {type(result)}"
    assert "Never gonna give you up" in result, "Lyrics content mismatch"
    
    print(f"✅ Read lyrics resource test passed")
    print(f"   Lyrics: {result[:20]}...")


@pytest.mark.asyncio
async def test_resource_error_handling():
    """Test that reading unknown resource returns appropriate error."""
    server = MCPServerStreamableHTTP(
        url=SERVER_URL,
        tool_prefix=""
    )
    
    try:
        # Try to read non-existent resource
        result = await server.read_resource("unknown://resource")
        pytest.fail("Expected error for unknown resource")
    except Exception as e:
        # Should throw exception for unknown resource
        print(f"✅ Resource error handling test passed - raised exception: {type(e).__name__}")


def run_sync_tests():
    """Run synchronous tests."""
    print("\n" + "=" * 80)
    print("Running Synchronous Tests")
    print("=" * 80)
    
    test_health_check()


async def run_async_tests():
    """Run asynchronous tests."""
    print("\n" + "=" * 80)
    print("Running Asynchronous Tests")
    print("=" * 80)
    
    print("\n--- Tool Tests ---")
    await test_list_tools()
    await test_call_echo_tool()
    await test_call_add_tool()
    await test_tool_error_handling()
    
    print("\n--- Resource Tests ---")
    await test_list_resources()
    await test_read_lyrics_resource()
    await test_resource_error_handling()


if __name__ == "__main__":
    print("=" * 80)
    print("MCP-Go Streamable HTTP Server Tests")
    print("=" * 80)
    print(f"Server URL: {SERVER_URL}")
    print(f"Base URL: {BASE_URL}")
    
    # Run sync tests
    run_sync_tests()
    
    # Run async tests
    asyncio.run(run_async_tests())
    
    print("\n" + "=" * 80)
    print("All tests completed!")
    print("=" * 80)
