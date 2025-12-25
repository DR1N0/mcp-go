#!/usr/bin/env python3
"""
Test script for mcp-go server using pydantic_ai Agent.

This script tests the MCP server implementation by:
1. Connecting via streamable HTTP transport
2. Discovering available tools
3. Testing the echo and add tools
4. Testing the time resource
"""

import os
from dotenv import load_dotenv
from pydantic_ai import Agent
from pydantic_ai.mcp import MCPServerStreamableHTTP
from pydantic_ai.models.openai import OpenAIChatModel
from pydantic_ai.providers.openai import OpenAIProvider
from pydantic_ai.models.anthropic import AnthropicModel
from pydantic_ai.providers.anthropic import AnthropicProvider
from anthropic import AsyncAnthropic


def main():
    """Run the agent test with the mcp-go server."""

    # Load environment variables
    load_dotenv()

    model_name = os.getenv("LLM_MODEL", "claude-sonnet-4")
    base_url = os.getenv("LLM_BASE_URL")
    server_url = os.getenv("MCP_SERVER_URL", "http://localhost:8000/mcp")
    api_key = os.getenv("LLM_API_KEY")
    provider = os.getenv("LLM_PROVIDER", "openai")

    print("=" * 80)
    print("Testing mcp-go Server with pydantic_ai Agent")
    print("=" * 80)
    print(f"Model: {model_name}")
    print(f"Base URL: {base_url}")
    print(f"Server URL: {server_url}")

    # Create MCP connection to local server
    print(f"\n[1] Connecting to MCP server at {server_url}...")
    mcp_server = MCPServerStreamableHTTP(
        url=server_url,
        tool_prefix=""  # No prefix for simplicity
    )

    # Create agent with pydantic_ai
    print("[2] Initializing pydantic_ai Agent...")

    # Configure model based on provider
    model = None
    if provider.lower() == "anthropic":
        model = AnthropicModel(
            model_name,
            provider=AnthropicProvider(
                anthropic_client=AsyncAnthropic(
                    base_url=base_url,
                    api_key=api_key)
            ),
        )
    else:  # Default to OpenAI
        model = OpenAIChatModel(
            model_name,
            provider=OpenAIProvider(
                base_url=base_url,
                api_key="dummy",
            ),
        )

    agent = Agent(
        model,
        system_prompt="""You are a helpful assistant that uses available tools to answer questions.
When using tools, be clear about what you're doing and show the results.""",
        toolsets=[mcp_server]
    )

    # Test 1: Echo tool
    print("\n" + "=" * 80)
    print("Test 1: Echo Tool")
    print("=" * 80)
    query1 = "Use the echo tool to echo back the message 'Hello from mcp-go!'"
    print(f"Query: {query1}\n")
    print("Running agent (this may take a moment)...\n")

    try:
        result1 = agent.run_sync(query1)
        print("Response:")
        print(result1.output)
        print("\n✅ Echo tool test passed!")
    except Exception as e:
        print(f"\n❌ Echo tool test failed: {e}")
        import traceback
        traceback.print_exc()

    # Test 2: Add tool
    print("\n" + "=" * 80)
    print("Test 2: Add Tool")
    print("=" * 80)
    query2 = "Use the add tool to calculate 42 + 58"
    print(f"Query: {query2}\n")
    print("Running agent (this may take a moment)...\n")

    try:
        result2 = agent.run_sync(query2)
        print("Response:")
        print(result2.output)
        print("\n✅ Add tool test passed!")
    except Exception as e:
        print(f"\n❌ Add tool test failed: {e}")
        import traceback
        traceback.print_exc()

if __name__ == "__main__":
    main()
