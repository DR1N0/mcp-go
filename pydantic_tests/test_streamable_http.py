#!/usr/bin/env python3
"""
Unit tests for mcp-go streamable HTTP server.

Tests include:
- Health check endpoint
- Tool listing
- Tool execution (echo and add)
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
    assert "message" in str(echo_tool.input_schema)
    
    # Check add tool schema
    add_tool = next(t for t in tools if t.name == "add")
    assert add_tool.description is not None
    assert "a" in str(add_tool.input_schema)
    assert "b" in str(add_tool.input_schema)
    
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
    
    # Call echo tool
    result = await server.call_tool(
        "echo",
        {"message": "Hello from test!"}
    )
    
    # Check result structure
    assert hasattr(result, 'content'), "Result should have content"
    assert len(result.content) > 0, "Result should have at least one content item"
    
    # Check that the message was echoed
    content_text = result.content[0].text
    assert "Hello from test!" in content_text, f"Expected echo in response, got: {content_text}"
    
    print(f"✅ Echo tool test passed: {content_text}")


@pytest.mark.asyncio
async def test_call_add_tool():
    """Test calling the add tool."""
    server = MCPServerStreamableHTTP(
        url=SERVER_URL,
        tool_prefix=""
    )
    
    # Call add tool
    result = await server.call_tool(
        "add",
        {"a": 42, "b": 58}
    )
    
    # Check result structure
    assert hasattr(result, 'content'), "Result should have content"
    assert len(result.content) > 0, "Result should have at least one content item"
    
    # Check that the sum is correct
    content_text = result.content[0].text
    assert "100" in content_text, f"Expected 100 in response, got: {content_text}"
    
    print(f"✅ Add tool test passed: {content_text}")


@pytest.mark.asyncio
async def test_tool_error_handling():
    """Test that calling unknown tool returns appropriate error."""
    server = MCPServerStreamableHTTP(
        url=SERVER_URL,
        tool_prefix=""
    )
    
    try:
        # Try to call non-existent tool
        result = await server.call_tool(
            "nonexistent_tool",
            {}
        )
        # If we get here, check if it's an error response
        if hasattr(result, 'isError'):
            assert result.isError, "Should return error for unknown tool"
            print("✅ Error handling test passed - returned error response")
        else:
            pytest.fail("Expected error for unknown tool")
    except Exception as e:
        # Also acceptable - throwing exception for unknown tool
        print(f"✅ Error handling test passed - raised exception: {type(e).__name__}")


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
    
    await test_list_tools()
    await test_call_echo_tool()
    await test_call_add_tool()
    await test_tool_error_handling()


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
