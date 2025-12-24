package types

// InitializeResponse is returned when initializing the server
type InitializeResponse struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
	Instructions    *string            `json:"instructions,omitempty"`
}

// ServerInfo provides information about the server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerCapabilities describes what the server can do
type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Logging   *LoggingCapability   `json:"logging,omitempty"`
}

// ToolsCapability describes tool-related capabilities
type ToolsCapability struct {
	ListChanged *bool `json:"listChanged,omitempty"`
}

// PromptsCapability describes prompt-related capabilities
type PromptsCapability struct {
	ListChanged *bool `json:"listChanged,omitempty"`
}

// ResourcesCapability describes resource-related capabilities
type ResourcesCapability struct {
	Subscribe   *bool `json:"subscribe,omitempty"`
	ListChanged *bool `json:"listChanged,omitempty"`
}

// LoggingCapability describes logging capabilities
type LoggingCapability struct{}

// ToolsResponse is the response to a tools/list request
type ToolsResponse struct {
	Tools      []Tool  `json:"tools"`
	NextCursor *string `json:"nextCursor,omitempty"`
}

// Tool represents a tool that can be called
type Tool struct {
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolResponse is the result of a tool call
type ToolResponse struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content represents a piece of content in a response
type Content struct {
	Type     string  `json:"type"`
	Text     *string `json:"text,omitempty"`
	Data     *string `json:"data,omitempty"`
	MimeType *string `json:"mimeType,omitempty"`
}

// NewTextContent creates text content
func NewTextContent(text string) *Content {
	return &Content{
		Type: "text",
		Text: &text,
	}
}

// NewToolResponse creates a new tool response
func NewToolResponse(content ...*Content) *ToolResponse {
	return &ToolResponse{
		Content: contentPtrSliceToSlice(content),
		IsError: false,
	}
}

// ListPromptsResponse is the response to a prompts/list request
type ListPromptsResponse struct {
	Prompts    []Prompt `json:"prompts"`
	NextCursor *string  `json:"nextCursor,omitempty"`
}

// Prompt represents a prompt template
type Prompt struct {
	Name        string           `json:"name"`
	Description *string          `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// PromptArgument describes an argument for a prompt
type PromptArgument struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Required    *bool   `json:"required,omitempty"`
}

// PromptResponse is the result of getting a prompt
type PromptResponse struct {
	Description *string         `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages"`
}

// PromptMessage represents a message in a prompt
type PromptMessage struct {
	Role    MessageRole `json:"role"`
	Content Content     `json:"content"`
}

// MessageRole represents the role of a message sender
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
)

// NewPromptMessage creates a new prompt message
func NewPromptMessage(content *Content, role MessageRole) PromptMessage {
	return PromptMessage{
		Role:    role,
		Content: *content,
	}
}

// NewPromptResponse creates a new prompt response
func NewPromptResponse(description string, messages ...PromptMessage) *PromptResponse {
	return &PromptResponse{
		Description: &description,
		Messages:    messages,
	}
}

// ListResourcesResponse is the response to a resources/list request
type ListResourcesResponse struct {
	Resources  []Resource `json:"resources"`
	NextCursor *string    `json:"nextCursor,omitempty"`
}

// Resource represents a resource that can be read
type Resource struct {
	URI         string  `json:"uri"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	MimeType    *string `json:"mimeType,omitempty"`
}

// ResourceResponse is the result of reading a resource
type ResourceResponse struct {
	Contents []ResourceContent `json:"contents"`
}

// ResourceContent represents the content of a resource
type ResourceContent struct {
	URI      string  `json:"uri"`
	MimeType *string `json:"mimeType,omitempty"`
	Text     *string `json:"text,omitempty"`
	Blob     *string `json:"blob,omitempty"`
}

// NewTextResource creates a text resource content
func NewTextResource(uri, text, mimeType string) ResourceContent {
	return ResourceContent{
		URI:      uri,
		Text:     &text,
		MimeType: &mimeType,
	}
}

// NewResourceResponse creates a new resource response
func NewResourceResponse(contents ...ResourceContent) *ResourceResponse {
	return &ResourceResponse{
		Contents: contents,
	}
}

// Helper function to convert []*Content to []Content
func contentPtrSliceToSlice(ptrs []*Content) []Content {
	result := make([]Content, len(ptrs))
	for i, ptr := range ptrs {
		if ptr != nil {
			result[i] = *ptr
		}
	}
	return result
}
