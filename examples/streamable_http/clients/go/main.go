package main

import (
	"context"
	"fmt"
	"log"
	"time"

	mcpgo "github.com/DR1N0/mcp-go"
	"github.com/DR1N0/mcp-go/transport/streamable"
)

func main() {
	fmt.Println("=================================================================================")
	fmt.Println("MCP Go Client Example")
	fmt.Println("=================================================================================")
	fmt.Println("This example demonstrates connecting a Go client to an MCP server")
	fmt.Println()

	// Create client transport
	serverURL := "http://localhost:8000/mcp"
	fmt.Printf("Connecting to MCP server at %s...\n", serverURL)

	transport := streamable.NewClientTransport(
		serverURL,
		streamable.WithTimeout(10*time.Second),
	)

	// Create MCP client
	client := mcpgo.NewClient(transport)

	// Initialize connection
	ctx := context.Background()
	fmt.Println("\n[1] Initializing client...")
	initResp, err := client.Initialize(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	fmt.Printf("✅ Connected to server: %s v%s\n", initResp.ServerInfo.Name, initResp.ServerInfo.Version)
	fmt.Printf("   Protocol version: %s\n", initResp.ProtocolVersion)

	// List tools
	fmt.Println("\n[2] Listing available tools...")
	toolsResp, err := client.ListTools(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}

	fmt.Printf("✅ Found %d tools:\n", len(toolsResp.Tools))
	for _, tool := range toolsResp.Tools {
		desc := "No description"
		if tool.Description != nil {
			desc = *tool.Description
		}
		fmt.Printf("   - %s: %s\n", tool.Name, desc)
	}

	// Call echo tool
	fmt.Println("\n[3] Calling 'echo' tool...")
	echoArgs := map[string]interface{}{
		"message": "Hello from Go client!",
	}
	echoResp, err := client.CallTool(ctx, "echo", echoArgs)
	if err != nil {
		log.Fatalf("Failed to call echo tool: %v", err)
	}

	if len(echoResp.Content) > 0 && echoResp.Content[0].Text != nil {
		fmt.Printf("✅ Echo response: %s\n", *echoResp.Content[0].Text)
	}

	// Call add tool
	fmt.Println("\n[4] Calling 'add' tool...")
	addArgs := map[string]interface{}{
		"a": 42,
		"b": 58,
	}
	addResp, err := client.CallTool(ctx, "add", addArgs)
	if err != nil {
		log.Fatalf("Failed to call add tool: %v", err)
	}

	if len(addResp.Content) > 0 && addResp.Content[0].Text != nil {
		fmt.Printf("✅ Add response: %s\n", *addResp.Content[0].Text)
	}

	// List prompts
	fmt.Println("\n[5] Listing available prompts...")
	promptsResp, err := client.ListPrompts(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list prompts: %v", err)
	}

	fmt.Printf("✅ Found %d prompts:\n", len(promptsResp.Prompts))
	for _, prompt := range promptsResp.Prompts {
		desc := "No description"
		if prompt.Description != nil {
			desc = *prompt.Description
		}
		fmt.Printf("   - %s: %s\n", prompt.Name, desc)
	}

	// Get a prompt
	if len(promptsResp.Prompts) > 0 {
		fmt.Println("\n[6] Getting 'greeting' prompt...")
		promptArgs := map[string]interface{}{
			"name": "Go Developer",
		}
		promptResp, err := client.GetPrompt(ctx, "greeting", promptArgs)
		if err != nil {
			log.Fatalf("Failed to get prompt: %v", err)
		}

		if promptResp.Description != nil {
			fmt.Printf("✅ Prompt: %s\n", *promptResp.Description)
		}
		if len(promptResp.Messages) > 0 {
			fmt.Printf("   Messages: %d\n", len(promptResp.Messages))
			for i, msg := range promptResp.Messages {
				fmt.Printf("   [%d] Role: %s\n", i+1, msg.Role)
				if msg.Content.Type == "text" && msg.Content.Text != nil {
					fmt.Printf("       Text: %s\n", *msg.Content.Text)
				}
			}
		}
	}

	// List resources
	fmt.Println("\n[7] Listing available resources...")
	resourcesResp, err := client.ListResources(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list resources: %v", err)
	}

	fmt.Printf("✅ Found %d resources:\n", len(resourcesResp.Resources))
	for _, resource := range resourcesResp.Resources {
		desc := "No description"
		if resource.Description != nil {
			desc = *resource.Description
		}
		fmt.Printf("   - %s (%s): %s\n", resource.Name, resource.URI, desc)
	}

	// Read a resource
	if len(resourcesResp.Resources) > 0 {
		fmt.Println("\n[8] Reading 'config://server' resource...")
		resourceResp, err := client.ReadResource(ctx, "config://server")
		if err != nil {
			log.Fatalf("Failed to read resource: %v", err)
		}

		fmt.Printf("✅ Resource contents: %d items\n", len(resourceResp.Contents))
		for i, content := range resourceResp.Contents {
			fmt.Printf("   [%d] URI: %s\n", i+1, content.URI)
			if content.MimeType != nil {
				fmt.Printf("       MimeType: %s\n", *content.MimeType)
			}
			if content.Text != nil {
				textPreview := *content.Text
				if len(textPreview) > 100 {
					textPreview = textPreview[:100] + "..."
				}
				fmt.Printf("       Text: %s\n", textPreview)
			}
		}
	}

	// Ping test
	fmt.Println("\n[9] Testing server connectivity...")
	if err := client.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping server: %v", err)
	}
	fmt.Println("✅ Ping successful!")

	// Close client
	fmt.Println("\n[10] Closing client...")
	if err := client.Close(); err != nil {
		log.Printf("Error closing client: %v", err)
	}

	fmt.Println("\n=================================================================================")
	fmt.Println("✅ All operations completed successfully!")
	fmt.Println("=================================================================================")
}
