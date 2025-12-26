package mcpgo

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"
	"sync"

	"github.com/DR1N0/mcp-go/protocol"
)

// registeredTool holds a tool's metadata and handler
type registeredTool struct {
	name        string
	description *string
	handler     interface{}
	inputSchema map[string]interface{}
}

// registeredPrompt holds a prompt's metadata and handler
type registeredPrompt struct {
	name        string
	description *string
	arguments   []PromptArgument
	handler     interface{}
}

// registeredResource holds a resource's metadata and handler
type registeredResource struct {
	uri         string
	name        string
	description *string
	mimeType    *string
	handler     interface{}
}

// MCPServer implements the Server interface with automatic tool management
type MCPServer struct {
	transport       Transport
	protocol        protocol.Protocol
	info            ServerInfo
	paginationLimit int
	started         bool         // Tracks if Serve() has been called
	mu              sync.RWMutex // Protects tools, prompts, resources, started
	tools           map[string]*registeredTool
	prompts         map[string]*registeredPrompt
	resources       map[string]*registeredResource
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

// WithPaginationLimit sets the pagination limit for list responses
func WithPaginationLimit(limit int) ServerOption {
	return func(s *MCPServer) {
		s.paginationLimit = limit
	}
}

// NewServer creates a new MCP server
func NewServer(transport Transport, opts ...ServerOption) Server {
	server := &MCPServer{
		transport:       transport,
		protocol:        protocol.NewProtocol(),
		paginationLimit: 10, // Default pagination limit
		info: ServerInfo{
			Name:    "mcp-server",
			Version: "0.1.0",
		},
		tools:     make(map[string]*registeredTool),
		prompts:   make(map[string]*registeredPrompt),
		resources: make(map[string]*registeredResource),
	}

	// Apply options
	for _, opt := range opts {
		opt(server)
	}

	// Register MCP protocol handlers
	server.protocol.SetRequestHandler("initialize", server.handleInitialize)
	server.protocol.SetRequestHandler("tools/list", server.handleToolsList)
	server.protocol.SetRequestHandler("tools/call", server.handleToolCall)
	server.protocol.SetRequestHandler("prompts/list", server.handlePromptsList)
	server.protocol.SetRequestHandler("prompts/get", server.handlePromptsGet)
	server.protocol.SetRequestHandler("resources/list", server.handleResourcesList)
	server.protocol.SetRequestHandler("resources/read", server.handleResourceRead)
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

	s.mu.Lock()
	desc := &description
	s.tools[name] = &registeredTool{
		name:        name,
		description: desc,
		handler:     handler,
		inputSchema: schema,
	}
	s.mu.Unlock()

	log.Printf("Registered tool: %s", name)

	// Send change notification (after releasing lock to avoid deadlock)
	s.sendToolsListChangedNotification()

	return nil
}

// RegisterPrompt registers a prompt with automatic argument extraction
func (s *MCPServer) RegisterPrompt(name, description string, handler interface{}) error {
	// Validate handler is a function
	handlerType := reflect.TypeOf(handler)
	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("handler must be a function")
	}

	// TODO: Extract arguments from handler signature
	// For now, use empty arguments list
	arguments := []PromptArgument{}

	s.mu.Lock()
	desc := &description
	s.prompts[name] = &registeredPrompt{
		name:        name,
		description: desc,
		arguments:   arguments,
		handler:     handler,
	}
	s.mu.Unlock()

	log.Printf("Registered prompt: %s", name)

	// Send change notification (after releasing lock to avoid deadlock)
	s.sendPromptsListChangedNotification()

	return nil
}

// RegisterResource registers a resource
func (s *MCPServer) RegisterResource(uri, name, description, mimeType string, handler interface{}) error {
	// Validate handler is a function
	handlerType := reflect.TypeOf(handler)
	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("handler must be a function")
	}

	s.mu.Lock()
	desc := &description
	mime := &mimeType
	s.resources[uri] = &registeredResource{
		uri:         uri,
		name:        name,
		description: desc,
		mimeType:    mime,
		handler:     handler,
	}
	s.mu.Unlock()

	log.Printf("Registered resource: %s (%s)", name, uri)

	// Send change notification (after releasing lock to avoid deadlock)
	s.sendResourcesListChangedNotification()

	return nil
}

// Serve starts the server
func (s *MCPServer) Serve() error {
	// Connect protocol to transport
	if err := s.protocol.Connect(s.transport); err != nil {
		return fmt.Errorf("failed to connect protocol: %w", err)
	}

	// Mark server as started
	s.mu.Lock()
	s.started = true
	s.mu.Unlock()

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

	capabilities := ServerCapabilities{}

	// Advertise tools capability (always, even if empty, to support dynamic registration)
	capabilities.Tools = &ToolsCapability{
		ListChanged: boolPtr(true), // Support dynamic registration
	}

	// Advertise prompts capability
	capabilities.Prompts = &PromptsCapability{
		ListChanged: boolPtr(true), // Support dynamic registration
	}

	// Advertise resources capability
	capabilities.Resources = &ResourcesCapability{
		Subscribe:   boolPtr(false),
		ListChanged: boolPtr(true), // Support dynamic registration
	}

	return InitializeResponse{
		ProtocolVersion: "2024-11-05",
		Capabilities:    capabilities,
		ServerInfo:      s.info,
	}, nil
}

// handleToolsList handles the tools/list request
func (s *MCPServer) handleToolsList(ctx context.Context, params interface{}) (interface{}, error) {
	log.Println("Handling tools/list request")

	// Parse cursor from params
	var cursor *string
	if params != nil {
		if paramsMap, ok := params.(map[string]interface{}); ok {
			if cursorVal, ok := paramsMap["cursor"].(string); ok {
				cursor = &cursorVal
			}
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Collect all tools
	allTools := make([]Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		allTools = append(allTools, Tool{
			Name:        tool.name,
			Description: tool.description,
			InputSchema: tool.inputSchema,
		})
	}

	// Sort by name for consistent pagination
	sort.Slice(allTools, func(i, j int) bool {
		return allTools[i].Name < allTools[j].Name
	})

	// Apply pagination
	startIndex := 0
	if cursor != nil && *cursor != "" {
		// Decode cursor (base64-encoded last item name)
		decoded, err := base64.StdEncoding.DecodeString(*cursor)
		if err == nil {
			lastItem := string(decoded)
			// Find first item after cursor
			for i, tool := range allTools {
				if tool.Name > lastItem {
					startIndex = i
					break
				}
			}
		}
	}

	// Determine end index based on pagination limit
	endIndex := len(allTools)
	var nextCursor *string
	if s.paginationLimit > 0 && startIndex+s.paginationLimit < len(allTools) {
		endIndex = startIndex + s.paginationLimit
		// Generate next cursor (base64-encode the last item name)
		lastItemName := allTools[endIndex-1].Name
		encoded := base64.StdEncoding.EncodeToString([]byte(lastItemName))
		nextCursor = &encoded
	}

	// Return paginated results
	tools := allTools[startIndex:endIndex]

	return ToolsResponse{
		Tools:      tools,
		NextCursor: nextCursor,
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

// handlePromptsList handles the prompts/list request
func (s *MCPServer) handlePromptsList(ctx context.Context, params interface{}) (interface{}, error) {
	log.Println("Handling prompts/list request")

	// Parse cursor from params
	var cursor *string
	if params != nil {
		if paramsMap, ok := params.(map[string]interface{}); ok {
			if cursorVal, ok := paramsMap["cursor"].(string); ok {
				cursor = &cursorVal
			}
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Collect all prompts
	allPrompts := make([]Prompt, 0, len(s.prompts))
	for _, prompt := range s.prompts {
		allPrompts = append(allPrompts, Prompt{
			Name:        prompt.name,
			Description: prompt.description,
			Arguments:   prompt.arguments,
		})
	}

	// Sort by name for consistent pagination
	sort.Slice(allPrompts, func(i, j int) bool {
		return allPrompts[i].Name < allPrompts[j].Name
	})

	// Apply pagination
	startIndex := 0
	if cursor != nil && *cursor != "" {
		decoded, err := base64.StdEncoding.DecodeString(*cursor)
		if err == nil {
			lastItem := string(decoded)
			for i, prompt := range allPrompts {
				if prompt.Name > lastItem {
					startIndex = i
					break
				}
			}
		}
	}

	// Determine end index based on pagination limit
	endIndex := len(allPrompts)
	var nextCursor *string
	if s.paginationLimit > 0 && startIndex+s.paginationLimit < len(allPrompts) {
		endIndex = startIndex + s.paginationLimit
		lastItemName := allPrompts[endIndex-1].Name
		encoded := base64.StdEncoding.EncodeToString([]byte(lastItemName))
		nextCursor = &encoded
	}

	// Return paginated results
	prompts := allPrompts[startIndex:endIndex]

	return ListPromptsResponse{
		Prompts:    prompts,
		NextCursor: nextCursor,
	}, nil
}

// handlePromptsGet handles the prompts/get request
func (s *MCPServer) handlePromptsGet(ctx context.Context, params interface{}) (interface{}, error) {
	log.Println("Handling prompts/get request")

	// Parse params
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid params type")
	}

	promptName, ok := paramsMap["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing prompt name")
	}

	arguments := paramsMap["arguments"]

	// Look up the prompt
	prompt, ok := s.prompts[promptName]
	if !ok {
		return nil, fmt.Errorf("unknown prompt: %s", promptName)
	}

	log.Printf("Getting prompt: %s with args: %v", promptName, arguments)

	// Call the handler (similar to tools but returns PromptResponse)
	result, err := s.callPromptHandler(prompt.handler, arguments)
	if err != nil {
		return nil, fmt.Errorf("error calling prompt handler: %w", err)
	}

	return result, nil
}

// handleResourcesList handles the resources/list request
func (s *MCPServer) handleResourcesList(ctx context.Context, params interface{}) (interface{}, error) {
	log.Println("Handling resources/list request")

	// Parse cursor from params
	var cursor *string
	if params != nil {
		if paramsMap, ok := params.(map[string]interface{}); ok {
			if cursorVal, ok := paramsMap["cursor"].(string); ok {
				cursor = &cursorVal
			}
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Collect all resources
	allResources := make([]Resource, 0, len(s.resources))
	for _, resource := range s.resources {
		allResources = append(allResources, Resource{
			URI:         resource.uri,
			Name:        resource.name,
			Description: resource.description,
			MimeType:    resource.mimeType,
		})
	}

	// Sort by URI for consistent pagination
	sort.Slice(allResources, func(i, j int) bool {
		return allResources[i].URI < allResources[j].URI
	})

	// Apply pagination
	startIndex := 0
	if cursor != nil && *cursor != "" {
		decoded, err := base64.StdEncoding.DecodeString(*cursor)
		if err == nil {
			lastURI := string(decoded)
			for i, resource := range allResources {
				if resource.URI > lastURI {
					startIndex = i
					break
				}
			}
		}
	}

	// Determine end index based on pagination limit
	endIndex := len(allResources)
	var nextCursor *string
	if s.paginationLimit > 0 && startIndex+s.paginationLimit < len(allResources) {
		endIndex = startIndex + s.paginationLimit
		lastURI := allResources[endIndex-1].URI
		encoded := base64.StdEncoding.EncodeToString([]byte(lastURI))
		nextCursor = &encoded
	}

	// Return paginated results
	resources := allResources[startIndex:endIndex]

	return ListResourcesResponse{
		Resources:  resources,
		NextCursor: nextCursor,
	}, nil
}

// handleResourceRead handles the resources/read request
func (s *MCPServer) handleResourceRead(ctx context.Context, params interface{}) (interface{}, error) {
	log.Println("Handling resources/read request")

	// Parse params
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid params type")
	}

	uri, ok := paramsMap["uri"].(string)
	if !ok {
		return nil, fmt.Errorf("missing resource URI")
	}

	// Look up the resource
	resource, ok := s.resources[uri]
	if !ok {
		return nil, fmt.Errorf("unknown resource: %s", uri)
	}

	log.Printf("Reading resource: %s (%s)", resource.name, uri)

	// Call the handler
	result, err := s.callResourceHandler(resource.handler)
	if err != nil {
		return nil, fmt.Errorf("error calling resource handler: %w", err)
	}

	return result, nil
}

// callPromptHandler calls a prompt handler
func (s *MCPServer) callPromptHandler(handler interface{}, arguments interface{}) (*PromptResponse, error) {
	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()

	var args []reflect.Value

	// Check if handler takes arguments
	if handlerType.NumIn() > 0 && arguments != nil {
		argType := handlerType.In(0)
		argValue := reflect.New(argType)

		argBytes, err := json.Marshal(arguments)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal arguments: %w", err)
		}

		if err := json.Unmarshal(argBytes, argValue.Interface()); err != nil {
			return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
		}

		args = append(args, argValue.Elem())
	}

	results := handlerValue.Call(args)

	// Handle return values
	if len(results) == 0 {
		return nil, fmt.Errorf("prompt handler must return a value")
	}

	// Check for error
	if len(results) == 2 {
		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}
	}

	// First result should be *PromptResponse
	response, ok := results[0].Interface().(*PromptResponse)
	if !ok {
		return nil, fmt.Errorf("handler must return *PromptResponse")
	}

	return response, nil
}

// callResourceHandler calls a resource handler
func (s *MCPServer) callResourceHandler(handler interface{}) (*ResourceResponse, error) {
	handlerValue := reflect.ValueOf(handler)

	results := handlerValue.Call(nil)

	// Handle return values
	if len(results) == 0 {
		return nil, fmt.Errorf("resource handler must return a value")
	}

	// Check for error
	if len(results) == 2 {
		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}
	}

	// First result should be *ResourceResponse
	response, ok := results[0].Interface().(*ResourceResponse)
	if !ok {
		return nil, fmt.Errorf("handler must return *ResourceResponse")
	}

	return response, nil
}

// handlePing handles the ping request
func (s *MCPServer) handlePing(ctx context.Context, params interface{}) (interface{}, error) {
	log.Println("Handling ping request")
	return map[string]interface{}{}, nil
}

// DeregisterTool removes a tool from the server
func (s *MCPServer) DeregisterTool(name string) error {
	s.mu.Lock()
	if _, exists := s.tools[name]; !exists {
		s.mu.Unlock()
		return fmt.Errorf("tool not found: %s", name)
	}

	delete(s.tools, name)
	s.mu.Unlock()

	log.Printf("Deregistered tool: %s", name)

	// Send change notification (after releasing lock to avoid deadlock)
	s.sendToolsListChangedNotification()

	return nil
}

// DeregisterPrompt removes a prompt from the server
func (s *MCPServer) DeregisterPrompt(name string) error {
	s.mu.Lock()
	if _, exists := s.prompts[name]; !exists {
		s.mu.Unlock()
		return fmt.Errorf("prompt not found: %s", name)
	}

	delete(s.prompts, name)
	s.mu.Unlock()

	log.Printf("Deregistered prompt: %s", name)

	// Send change notification (after releasing lock to avoid deadlock)
	s.sendPromptsListChangedNotification()

	return nil
}

// DeregisterResource removes a resource from the server
func (s *MCPServer) DeregisterResource(uri string) error {
	s.mu.Lock()
	if _, exists := s.resources[uri]; !exists {
		s.mu.Unlock()
		return fmt.Errorf("resource not found: %s", uri)
	}

	delete(s.resources, uri)
	s.mu.Unlock()

	log.Printf("Deregistered resource: %s", uri)

	// Send change notification (after releasing lock to avoid deadlock)
	s.sendResourcesListChangedNotification()

	return nil
}

// HasTool checks if a tool is registered
func (s *MCPServer) HasTool(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.tools[name]
	return exists
}

// HasPrompt checks if a prompt is registered
func (s *MCPServer) HasPrompt(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.prompts[name]
	return exists
}

// HasResource checks if a resource is registered
func (s *MCPServer) HasResource(uri string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.resources[uri]
	return exists
}

// sendToolsListChangedNotification sends a notification that the tools list has changed
func (s *MCPServer) sendToolsListChangedNotification() {
	// Only send notifications if server has been started
	s.mu.RLock()
	started := s.started
	s.mu.RUnlock()

	if !started {
		// Server not started yet, skip notification
		return
	}

	// Send notification to client
	if err := s.protocol.Notification("notifications/tools/list_changed", nil); err != nil {
		log.Printf("Failed to send tools list changed notification: %v", err)
	} else {
		log.Println("Sent tools list changed notification")
	}
}

// sendPromptsListChangedNotification sends a notification that the prompts list has changed
func (s *MCPServer) sendPromptsListChangedNotification() {
	// Only send notifications if server has been started
	s.mu.RLock()
	started := s.started
	s.mu.RUnlock()

	if !started {
		// Server not started yet, skip notification
		return
	}

	// Send notification to client
	if err := s.protocol.Notification("notifications/prompts/list_changed", nil); err != nil {
		log.Printf("Failed to send prompts list changed notification: %v", err)
	} else {
		log.Println("Sent prompts list changed notification")
	}
}

// sendResourcesListChangedNotification sends a notification that the resources list has changed
func (s *MCPServer) sendResourcesListChangedNotification() {
	// Only send notifications if server has been started
	s.mu.RLock()
	started := s.started
	s.mu.RUnlock()

	if !started {
		// Server not started yet, skip notification
		return
	}

	// Send notification to client
	if err := s.protocol.Notification("notifications/resources/list_changed", nil); err != nil {
		log.Printf("Failed to send resources list changed notification: %v", err)
	} else {
		log.Println("Sent resources list changed notification")
	}
}

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
