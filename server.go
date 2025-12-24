package mcpgo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/DR1N0/mcp-go/protocol"
)

// registeredTool holds a tool's metadata and handler
type registeredTool struct {
	name        string
	description *string
	handler     interface{}
	inputSchema map[string]interface{}
}

// MCPServer implements the Server interface with automatic tool management
type MCPServer struct {
	transport Transport
	protocol  protocol.Protocol
	info      ServerInfo
	tools     map[string]*registeredTool
}

// ServerOption configures the server
type ServerOption func(*MCPServer)

// WithName sets the server name
func WithName(name string) ServerOption {
	return func(s *MCPServer) {
		s.info.Name = name
	}
}

// WithVersion sets the server version
func WithVersion(version string) ServerOption {
	return func(s *MCPServer) {
		s.info.Version = version
	}
}

// NewServer creates a new MCP server
func NewServer(transport Transport, opts ...ServerOption) Server {
	server := &MCPServer{
		transport: transport,
		protocol:  protocol.NewProtocol(),
		info: ServerInfo{
			Name:    "mcp-server",
			Version: "0.1.0",
		},
		tools: make(map[string]*registeredTool),
	}

	// Apply options
	for _, opt := range opts {
		opt(server)
	}

	// Register MCP protocol handlers
	server.protocol.SetRequestHandler("initialize", server.handleInitialize)
	server.protocol.SetRequestHandler("tools/list", server.handleToolsList)
	server.protocol.SetRequestHandler("tools/call", server.handleToolCall)
	server.protocol.SetRequestHandler("ping", server.handlePing)

	return server
}

// RegisterTool registers a tool with automatic schema generation
func (s *MCPServer) RegisterTool(name, description string, handler interface{}) error {
	// Validate handler is a function
	handlerType := reflect.TypeOf(handler)
	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("handler must be a function")
	}

	// Generate input schema from function signature
	schema, err := GenerateSchema(handler)
	if err != nil {
		return fmt.Errorf("failed to generate schema: %w", err)
	}

	desc := &description
	s.tools[name] = &registeredTool{
		name:        name,
		description: desc,
		handler:     handler,
		inputSchema: schema,
	}

	log.Printf("Registered tool: %s", name)
	return nil
}

// RegisterPrompt registers a prompt (placeholder for now)
func (s *MCPServer) RegisterPrompt(name, description string, handler interface{}) error {
	return fmt.Errorf("prompts not yet implemented")
}

// RegisterResource registers a resource (placeholder for now)
func (s *MCPServer) RegisterResource(uri, name, description, mimeType string, handler interface{}) error {
	return fmt.Errorf("resources not yet implemented")
}

// Serve starts the server
func (s *MCPServer) Serve() error {
	// Connect protocol to transport
	if err := s.protocol.Connect(s.transport); err != nil {
		return fmt.Errorf("failed to connect protocol: %w", err)
	}

	log.Printf("MCP server '%s' v%s started", s.info.Name, s.info.Version)
	return nil
}

// Close shuts down the server
func (s *MCPServer) Close() error {
	return s.protocol.Close()
}

// handleInitialize handles the initialize request
func (s *MCPServer) handleInitialize(ctx context.Context, params interface{}) (interface{}, error) {
	log.Println("Handling initialize request")

	return InitializeResponse{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: boolPtr(false),
			},
		},
		ServerInfo: s.info,
	}, nil
}

// handleToolsList handles the tools/list request
func (s *MCPServer) handleToolsList(ctx context.Context, params interface{}) (interface{}, error) {
	log.Println("Handling tools/list request")

	tools := make([]Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, Tool{
			Name:        tool.name,
			Description: tool.description,
			InputSchema: tool.inputSchema,
		})
	}

	return ToolsResponse{
		Tools:      tools,
		NextCursor: nil,
	}, nil
}

// handleToolCall handles the tools/call request
func (s *MCPServer) handleToolCall(ctx context.Context, params interface{}) (interface{}, error) {
	log.Println("Handling tools/call request")

	// Parse params
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid params type")
	}

	toolName, ok := paramsMap["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing tool name")
	}

	arguments, ok := paramsMap["arguments"]
	if !ok {
		arguments = make(map[string]interface{})
	}

	// Look up the tool
	tool, ok := s.tools[toolName]
	if !ok {
		return ToolResponse{
			Content: []Content{
				{Type: "text", Text: strPtr(fmt.Sprintf("Unknown tool: %s", toolName))},
			},
			IsError: true,
		}, nil
	}

	log.Printf("Calling tool: %s with args: %v", toolName, arguments)

	// Call the handler
	result, err := s.callToolHandler(tool.handler, arguments)
	if err != nil {
		return ToolResponse{
			Content: []Content{
				{Type: "text", Text: strPtr(fmt.Sprintf("Error: %v", err))},
			},
			IsError: true,
		}, nil
	}

	return result, nil
}

// callToolHandler calls the tool handler with proper argument unmarshaling
func (s *MCPServer) callToolHandler(handler interface{}, arguments interface{}) (*ToolResponse, error) {
	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()

	// Determine if handler has context parameter
	numIn := handlerType.NumIn()
	hasContext := numIn > 0 && handlerType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem())

	var argIndex int
	if hasContext {
		argIndex = 1
	} else {
		argIndex = 0
	}

	// If no args expected, call with no args
	if numIn == argIndex {
		return s.invokeHandler(handlerValue, hasContext, reflect.Value{})
	}

	// Marshal arguments to JSON then unmarshal to the expected type
	argType := handlerType.In(argIndex)
	argValue := reflect.New(argType)

	if arguments != nil {
		argBytes, err := json.Marshal(arguments)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal arguments: %w", err)
		}

		if err := json.Unmarshal(argBytes, argValue.Interface()); err != nil {
			return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
		}
	}

	return s.invokeHandler(handlerValue, hasContext, argValue.Elem())
}

// invokeHandler invokes the handler function
func (s *MCPServer) invokeHandler(handlerValue reflect.Value, hasContext bool, argValue reflect.Value) (*ToolResponse, error) {
	var args []reflect.Value
	if hasContext {
		args = append(args, reflect.ValueOf(context.Background()))
	}
	if argValue.IsValid() {
		args = append(args, argValue)
	}

	results := handlerValue.Call(args)

	// Handle return values
	if len(results) == 0 {
		return NewToolResponse(), nil
	}

	// Check for error
	if len(results) == 2 {
		if !results[1].IsNil() {
			err := results[1].Interface().(error)
			return nil, err
		}
	}

	// First result should be *ToolResponse
	response, ok := results[0].Interface().(*ToolResponse)
	if !ok {
		return nil, fmt.Errorf("handler must return *ToolResponse")
	}

	return response, nil
}

// handlePing handles the ping request
func (s *MCPServer) handlePing(ctx context.Context, params interface{}) (interface{}, error) {
	log.Println("Handling ping request")
	return map[string]interface{}{}, nil
}

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
