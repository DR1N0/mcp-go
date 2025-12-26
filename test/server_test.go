package mcpgo_test

import (
	"testing"

	mcpgo "github.com/DR1N0/mcp-go"
)

type TestToolArgs struct {
	Message string `json:"message" jsonschema:"required,description=Test message"`
}

type TestPromptArgs struct {
	Name string `json:"name" jsonschema:"required,description=User name"`
}

func testToolHandler(args TestToolArgs) (*mcpgo.ToolResponse, error) {
	return mcpgo.NewToolResponse(
		mcpgo.NewTextContent("Test: " + args.Message),
	), nil
}

func testPromptHandler(args TestPromptArgs) (*mcpgo.PromptResponse, error) {
	return mcpgo.NewPromptResponse(
		"Test prompt",
		mcpgo.NewPromptMessage(
			mcpgo.NewTextContent("Hello "+args.Name),
			mcpgo.RoleAssistant,
		),
	), nil
}

func testResourceHandler() (*mcpgo.ResourceResponse, error) {
	return mcpgo.NewResourceResponse(
		mcpgo.NewTextResource("test://resource", "test content", "text/plain"),
	), nil
}

func TestServer_RegisterTool(t *testing.T) {
	server := mcpgo.NewServer(nil)

	err := server.RegisterTool("test-tool", "Test tool", testToolHandler)
	if err != nil {
		t.Fatalf("RegisterTool failed: %v", err)
	}

	// Register multiple tools
	err = server.RegisterTool("test-tool-2", "Test tool 2", testToolHandler)
	if err != nil {
		t.Fatalf("Second RegisterTool failed: %v", err)
	}
}

func TestServer_RegisterPrompt(t *testing.T) {
	server := mcpgo.NewServer(nil)

	err := server.RegisterPrompt("test-prompt", "Test prompt", testPromptHandler)
	if err != nil {
		t.Fatalf("RegisterPrompt failed: %v", err)
	}

	// Register multiple prompts
	err = server.RegisterPrompt("test-prompt-2", "Test prompt 2", testPromptHandler)
	if err != nil {
		t.Fatalf("Second RegisterPrompt failed: %v", err)
	}
}

func TestServer_RegisterResource(t *testing.T) {
	server := mcpgo.NewServer(nil)

	err := server.RegisterResource(
		"test://resource",
		"Test Resource",
		"A test resource",
		"text/plain",
		testResourceHandler,
	)
	if err != nil {
		t.Fatalf("RegisterResource failed: %v", err)
	}

	// Register multiple resources
	err = server.RegisterResource(
		"test://resource2",
		"Test Resource 2",
		"Another resource",
		"text/plain",
		testResourceHandler,
	)
	if err != nil {
		t.Fatalf("Second RegisterResource failed: %v", err)
	}
}

func TestServer_WithServerInfo(t *testing.T) {
	server := mcpgo.NewServer(nil, mcpgo.WithName("test-server"), mcpgo.WithVersion("1.0.0"))

	// Can't directly test the server info, but we can verify the server was created
	if server == nil {
		t.Fatal("Server should not be nil")
	}
}
