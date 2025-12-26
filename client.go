package mcpgo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/DR1N0/mcp-go/protocol"
)

// mcpClient implements the Client interface
type mcpClient struct {
	transport    Transport
	protocol     protocol.Protocol
	capabilities *ServerCapabilities
	initialized  bool
}

// NewClient creates a new MCP client that returns the Client interface
func NewClient(transport Transport) Client {
	return &mcpClient{
		transport:   transport,
		protocol:    protocol.NewProtocol(),
		initialized: false,
	}
}

// Initialize connects to the server and retrieves its capabilities
func (c *mcpClient) Initialize(ctx context.Context) (*InitializeResponse, error) {
	if c.initialized {
		return nil, fmt.Errorf("client already initialized")
	}

	// Connect protocol to transport
	if err := c.protocol.Connect(c.transport); err != nil {
		return nil, fmt.Errorf("failed to connect protocol: %w", err)
	}

	// Send initialize request
	params := map[string]interface{}{
		"protocolVersion": "2025-12-25",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "mcp-go-client",
			"version": "0.1.0",
		},
	}

	response, err := c.protocol.Request(ctx, "initialize", params)
	if err != nil {
		return nil, fmt.Errorf("initialize request failed: %w", err)
	}

	// Parse initialize response
	var initResp InitializeResponse
	if err := unmarshalResponse(response, &initResp); err != nil {
		return nil, fmt.Errorf("failed to parse initialize response: %w", err)
	}

	c.capabilities = &initResp.Capabilities
	c.initialized = true

	log.Printf("Initialized MCP client: server=%s v%s", initResp.ServerInfo.Name, initResp.ServerInfo.Version)
	return &initResp, nil
}

// ListTools retrieves the list of available tools from the server
func (c *mcpClient) ListTools(ctx context.Context, cursor *string) (*ToolsResponse, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	params := map[string]interface{}{}
	if cursor != nil {
		params["cursor"] = *cursor
	}

	response, err := c.protocol.Request(ctx, "tools/list", params)
	if err != nil {
		return nil, fmt.Errorf("tools/list request failed: %w", err)
	}

	var toolsResp ToolsResponse
	if err := unmarshalResponse(response, &toolsResp); err != nil {
		return nil, fmt.Errorf("failed to parse tools/list response: %w", err)
	}

	return &toolsResp, nil
}

// CallTool calls a specific tool on the server with the provided arguments
func (c *mcpClient) CallTool(ctx context.Context, name string, args interface{}) (*ToolResponse, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	params := map[string]interface{}{
		"name":      name,
		"arguments": args,
	}

	response, err := c.protocol.Request(ctx, "tools/call", params)
	if err != nil {
		return nil, fmt.Errorf("tools/call request failed: %w", err)
	}

	var toolResp ToolResponse
	if err := unmarshalResponse(response, &toolResp); err != nil {
		return nil, fmt.Errorf("failed to parse tools/call response: %w", err)
	}

	return &toolResp, nil
}

// ListPrompts retrieves the list of available prompts from the server
func (c *mcpClient) ListPrompts(ctx context.Context, cursor *string) (*ListPromptsResponse, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	params := map[string]interface{}{}
	if cursor != nil {
		params["cursor"] = *cursor
	}

	response, err := c.protocol.Request(ctx, "prompts/list", params)
	if err != nil {
		return nil, fmt.Errorf("prompts/list request failed: %w", err)
	}

	var promptsResp ListPromptsResponse
	if err := unmarshalResponse(response, &promptsResp); err != nil {
		return nil, fmt.Errorf("failed to parse prompts/list response: %w", err)
	}

	return &promptsResp, nil
}

// GetPrompt retrieves a specific prompt from the server
func (c *mcpClient) GetPrompt(ctx context.Context, name string, args interface{}) (*PromptResponse, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	params := map[string]interface{}{
		"name":      name,
		"arguments": args,
	}

	response, err := c.protocol.Request(ctx, "prompts/get", params)
	if err != nil {
		return nil, fmt.Errorf("prompts/get request failed: %w", err)
	}

	var promptResp PromptResponse
	if err := unmarshalResponse(response, &promptResp); err != nil {
		return nil, fmt.Errorf("failed to parse prompts/get response: %w", err)
	}

	return &promptResp, nil
}

// ListResources retrieves the list of available resources from the server
func (c *mcpClient) ListResources(ctx context.Context, cursor *string) (*ListResourcesResponse, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	params := map[string]interface{}{}
	if cursor != nil {
		params["cursor"] = *cursor
	}

	response, err := c.protocol.Request(ctx, "resources/list", params)
	if err != nil {
		return nil, fmt.Errorf("resources/list request failed: %w", err)
	}

	var resourcesResp ListResourcesResponse
	if err := unmarshalResponse(response, &resourcesResp); err != nil {
		return nil, fmt.Errorf("failed to parse resources/list response: %w", err)
	}

	return &resourcesResp, nil
}

// ReadResource reads a specific resource from the server
func (c *mcpClient) ReadResource(ctx context.Context, uri string) (*ResourceResponse, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	params := map[string]interface{}{
		"uri": uri,
	}

	response, err := c.protocol.Request(ctx, "resources/read", params)
	if err != nil {
		return nil, fmt.Errorf("resources/read request failed: %w", err)
	}

	var resourceResp ResourceResponse
	if err := unmarshalResponse(response, &resourceResp); err != nil {
		return nil, fmt.Errorf("failed to parse resources/read response: %w", err)
	}

	return &resourceResp, nil
}

// Ping sends a ping request to check server connectivity
func (c *mcpClient) Ping(ctx context.Context) error {
	if !c.initialized {
		return fmt.Errorf("client not initialized")
	}

	_, err := c.protocol.Request(ctx, "ping", nil)
	if err != nil {
		return fmt.Errorf("ping request failed: %w", err)
	}

	return nil
}

// GetCapabilities returns the server capabilities obtained during initialization
func (c *mcpClient) GetCapabilities() *ServerCapabilities {
	return c.capabilities
}

// Close closes the client connection
func (c *mcpClient) Close() error {
	return c.protocol.Close()
}

// unmarshalResponse is a helper to unmarshal JSON-RPC response results
func unmarshalResponse(response interface{}, target interface{}) error {
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
