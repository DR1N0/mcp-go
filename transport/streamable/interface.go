package streamable

import (
	"net/http"

	"github.com/DR1N0/mcp-go/transport"
	"github.com/DR1N0/mcp-go/types"
)

// HTTPHandler provides HTTP-specific functionality
type HTTPHandler interface {
	// ServeHTTP handles HTTP requests for the MCP endpoint
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// ServerTransport combines Transport with HTTP handling for streamable HTTP servers
type ServerTransport interface {
	transport.ServerTransport
	HTTPHandler
	// WithMiddleware adds HTTP middleware to be chained before the MCP handler
	WithMiddleware(middleware ...types.HTTPMiddleware) ServerTransport
}

// ClientTransport is a streamable HTTP client transport
type ClientTransport interface {
	transport.ClientTransport
}
