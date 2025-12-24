package mcpgo

// Re-export types from types package for convenience
import (
	"github.com/DR1N0/mcp-go/types"
)

// Re-export message types
type BaseJSONRPCMessage = types.BaseJSONRPCMessage
type RPCError = types.RPCError

// Re-export transport types
type Transport = types.Transport
type MessageHandler = types.MessageHandler
type ErrorHandler = types.ErrorHandler
type CloseHandler = types.CloseHandler

// Re-export MCP types
type InitializeResponse = types.InitializeResponse
type ServerInfo = types.ServerInfo
type ServerCapabilities = types.ServerCapabilities
type ToolsCapability = types.ToolsCapability
type PromptsCapability = types.PromptsCapability
type ResourcesCapability = types.ResourcesCapability
type LoggingCapability = types.LoggingCapability
type ToolsResponse = types.ToolsResponse
type Tool = types.Tool
type ToolResponse = types.ToolResponse
type Content = types.Content
type ListPromptsResponse = types.ListPromptsResponse
type Prompt = types.Prompt
type PromptArgument = types.PromptArgument
type PromptResponse = types.PromptResponse
type PromptMessage = types.PromptMessage
type MessageRole = types.MessageRole
type ListResourcesResponse = types.ListResourcesResponse
type Resource = types.Resource
type ResourceResponse = types.ResourceResponse
type ResourceContent = types.ResourceContent

// Re-export constants
const (
	RoleUser      = types.RoleUser
	RoleAssistant = types.RoleAssistant
)

// Re-export helper functions
var (
	NewTextContent      = types.NewTextContent
	NewToolResponse     = types.NewToolResponse
	NewPromptMessage    = types.NewPromptMessage
	NewPromptResponse   = types.NewPromptResponse
	NewTextResource     = types.NewTextResource
	NewResourceResponse = types.NewResourceResponse
)
