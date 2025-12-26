package mcpgo

import "context"

// Server represents an MCP server that can register and serve tools, prompts, and resources
type Server interface {
	// RegisterTool registers a new tool with the server
	// The handler should be a function with signature:
	// func(ctx context.Context, args YourArgsStruct) (*ToolResponse, error)
	// or
	// func(args YourArgsStruct) (*ToolResponse, error)
	RegisterTool(name, description string, handler interface{}) error

	// RegisterPrompt registers a new prompt with the server
	RegisterPrompt(name, description string, handler interface{}) error

	// RegisterResource registers a new resource with the server
	RegisterResource(uri, name, description, mimeType string, handler interface{}) error

	// DeregisterTool removes a tool from the server
	DeregisterTool(name string) error

	// DeregisterPrompt removes a prompt from the server
	DeregisterPrompt(name string) error

	// DeregisterResource removes a resource from the server
	DeregisterResource(uri string) error

	// HasTool checks if a tool is registered
	HasTool(name string) bool

	// HasPrompt checks if a prompt is registered
	HasPrompt(name string) bool

	// HasResource checks if a resource is registered
	HasResource(uri string) bool

	// Serve starts the server and begins handling requests
	Serve() error

	// Close shuts down the server gracefully
	Close() error
}

// Client represents an MCP client that can connect to and interact with MCP servers
type Client interface {
	// Initialize connects to the server and retrieves its capabilities
	Initialize(ctx context.Context) (*InitializeResponse, error)

	// ListTools retrieves the list of available tools from the server
	ListTools(ctx context.Context, cursor *string) (*ToolsResponse, error)

	// CallTool calls a specific tool on the server with the provided arguments
	CallTool(ctx context.Context, name string, args interface{}) (*ToolResponse, error)

	// ListPrompts retrieves the list of available prompts from the server
	ListPrompts(ctx context.Context, cursor *string) (*ListPromptsResponse, error)

	// GetPrompt retrieves a specific prompt from the server
	GetPrompt(ctx context.Context, name string, args interface{}) (*PromptResponse, error)

	// ListResources retrieves the list of available resources from the server
	ListResources(ctx context.Context, cursor *string) (*ListResourcesResponse, error)

	// ReadResource reads a specific resource from the server
	ReadResource(ctx context.Context, uri string) (*ResourceResponse, error)

	// Ping sends a ping request to check server connectivity
	Ping(ctx context.Context) error

	// GetCapabilities returns the server capabilities obtained during initialization
	GetCapabilities() *ServerCapabilities

	// Close closes the client connection
	Close() error
}
