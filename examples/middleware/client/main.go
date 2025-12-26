package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	mcpgo "github.com/DR1N0/mcp-go"
	"github.com/DR1N0/mcp-go/transport/streamable"
)

func main() {
	fmt.Println("=================================================================================")
	fmt.Println("MCP Client - Testing Middleware Authentication")
	fmt.Println("=================================================================================")
	fmt.Println()

	serverURL := "http://localhost:8080/mcp"

	// Test 1: Try without authentication (should fail)
	fmt.Println("Test 1: Calling server WITHOUT authentication...")
	testWithoutAuth(serverURL)
	fmt.Println()

	// Test 2: Try with valid Bearer token (should succeed)
	fmt.Println("Test 2: Calling server WITH authentication...")
	testWithAuth(serverURL)
	fmt.Println()

	fmt.Println("=================================================================================")
	fmt.Println("✅ All tests completed!")
	fmt.Println("=================================================================================")
}

func testWithoutAuth(serverURL string) {
	// Create transport without authentication
	transport := streamable.NewClientTransport(serverURL)
	client := mcpgo.NewClient(transport)

	ctx := context.Background()
	_, err := client.Initialize(ctx)

	if err != nil {
		fmt.Printf("❌ Expected failure: %v\n", err)
		fmt.Println("   (This is correct - server requires authentication)")
	} else {
		fmt.Println("⚠️  Unexpected: Server accepted request without auth!")
	}
}

func testWithAuth(serverURL string) {
	// Create custom HTTP client with Authorization header
	httpClient := &http.Client{
		Transport: &authTransport{
			token: "demo-token",
			base:  http.DefaultTransport,
		},
	}

	// Create transport with custom HTTP client
	transport := streamable.NewClientTransport(serverURL, streamable.WithHTTPClient(httpClient))
	client := mcpgo.NewClient(transport)

	ctx := context.Background()

	// Initialize connection
	result, err := client.Initialize(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		return
	}

	fmt.Printf("✅ Connected to server: %s v%s\n", result.ServerInfo.Name, result.ServerInfo.Version)

	// List tools
	toolsResp, err := client.ListTools(ctx, nil)
	if err != nil {
		fmt.Printf("❌ Failed to list tools: %v\n", err)
		return
	}

	fmt.Printf("✅ Found %d tool(s):\n", len(toolsResp.Tools))
	for _, tool := range toolsResp.Tools {
		desc := "No description"
		if tool.Description != nil {
			desc = *tool.Description
		}
		fmt.Printf("   - %s: %s\n", tool.Name, desc)
	}

	// Call echo tool
	echoArgs := map[string]interface{}{
		"message": "Hello from authenticated client!",
	}

	response, err := client.CallTool(ctx, "echo", echoArgs)
	if err != nil {
		fmt.Printf("❌ Failed to call tool: %v\n", err)
		return
	}

	if len(response.Content) > 0 && response.Content[0].Text != nil {
		fmt.Printf("✅ Tool response: %s\n", *response.Content[0].Text)
	}

	// Close client
	if err := client.Close(); err != nil {
		log.Printf("Error closing client: %v", err)
	}
}

// authTransport wraps http.RoundTripper to add Authorization header
type authTransport struct {
	token string
	base  http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(req)
}
